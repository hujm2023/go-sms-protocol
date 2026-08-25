# go-sms-protocol

English | [简体中文](README.zh-CN.md)

[![CI](https://github.com/hujm2023/go-sms-protocol/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/hujm2023/go-sms-protocol/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/hujm2023/go-sms-protocol.svg)](https://pkg.go.dev/github.com/hujm2023/go-sms-protocol)
[![Go version](https://img.shields.io/github/go-mod/go-version/hujm2023/go-sms-protocol)](go.mod)
[![License](https://img.shields.io/github/license/hujm2023/go-sms-protocol)](LICENSE)

`go-sms-protocol` is a Go library for encoding and decoding selected PDUs from SMPP 3.4, CMPP 2.0/3.0, SGIP 1.2, and SMGP 3.0. It also provides CMPP/SMPP stream framing, SMS text codecs, concatenated-message helpers, and a `netpoll`-based server building block.

This repository is a library, not a ready-to-run SMS gateway or command-line application.

## Installation

The minimum supported Go version is 1.20.

```bash
go get github.com/hujm2023/go-sms-protocol@latest
```

## Quick start

The following program creates an SMPP 3.4 `enquire_link` PDU, encodes it to wire bytes, and decodes it through the protocol dispatcher.

```go
package main

import (
	"fmt"
	"log"

	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
)

func main() {
	request := &smpp34.EnquireLink{
		Header: smpp.Header{
			ID:       smpp.ENQUIRE_LINK,
			Sequence: 42,
		},
	}

	wire, err := request.IEncode()
	if err != nil {
		log.Fatal(err)
	}

	pdu, err := smpp34.DecodeSMPP34(wire)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s sequence=%d\n", pdu.GetCommand().String(), pdu.GetSequenceID())
}
```

Expected output:

```text
SMPP_ENQUIRE_LINK sequence=42
```

Use the dispatcher that matches the wire protocol and version:

| Protocol | Dispatcher |
| --- | --- |
| CMPP 2.0 | `cmpp20.DecodeCMPP20` |
| CMPP 3.0 | `cmpp30.DecodeCMPP30` |
| SGIP 1.2 | `sgip12.DecodeSGIP12` |
| SMGP 3.0 | `smgp30.DecodeSMGP30` |
| SMPP 3.4 | `smpp34.DecodeSMPP34` |

Every successfully decoded value implements the root `protocol.PDU` interface, including wire encoding/decoding, command and sequence access, response construction, and string formatting.

## Protocol coverage

Coverage in this table means that the package contains concrete PDU types and that its dispatcher recognizes the listed command families. Command constants outside this table may exist without a corresponding PDU implementation.

| Version | Implemented PDU families |
| --- | --- |
| CMPP 2.0 | `CONNECT`, `SUBMIT`, `QUERY`, `DELIVER`, `ACTIVE_TEST`, `TERMINATE`, and their responses |
| CMPP 3.0 | CMPP 2.0 families plus `CANCEL` and its response |
| SGIP 1.2 | `BIND`, `UNBIND`, `SUBMIT`, `REPORT`, `DELIVER`, and their responses |
| SMGP 3.0 | `LOGIN`, `SUBMIT`, `DELIVER`, `ACTIVE_TEST`, `EXIT`, and their responses |
| SMPP 3.4 | bind receiver/transmitter/transceiver, `SUBMIT_SM`, `DELIVER_SM`, `ENQUIRE_LINK`, `UNBIND`, `GENERIC_NACK`, and responses where defined |

SMPP 5.0 is not implemented. A protocol PDF or version constant in the repository does not imply decoder or PDU support.

## Packages

| Package | Purpose |
| --- | --- |
| Root package | Shared `PDU` contract, content helpers, CMPP/SMPP long-message splitting, UDH parsing, and batch data-coding selection |
| [`cmpp/cmpp20`](cmpp/cmpp20), [`cmpp/cmpp30`](cmpp/cmpp30) | CMPP PDU types and version-specific dispatchers |
| [`sgip/sgip12`](sgip/sgip12) | SGIP 1.2 PDU types and dispatcher |
| [`smgp/smgp30`](smgp/smgp30) | SMGP 3.0 PDU types and dispatcher |
| [`smpp/smpp34`](smpp/smpp34) | SMPP 3.4 PDU types, validation, TLVs, delivery receipts, and dispatcher |
| [`codec`](codec) | Blocking and non-blocking TCP frame extraction for CMPP and SMPP |
| [`datacoding`](datacoding) | ASCII, GB18030, GSM 7-bit packed/unpacked, Latin-1, and UCS-2 codecs |
| [`packet`](packet) | Binary packet reader, writer, and PDU string formatting helpers |
| [`nioserver`](nioserver) | Generic `netpoll` TCP server primitives with bounded handler and write admission |

The complete exported API is available on [pkg.go.dev](https://pkg.go.dev/github.com/hujm2023/go-sms-protocol).

## Message content and concatenated SMS

The root package provides:

- `EncodeCMPPContentAndSplit` and `EncodeSMPPContentAndSplit` for encoding text and adding a six-byte UDH when segmentation is required.
- `ParseLongSmsContent` and `ParseLongSmsContentBytes` for parsing one segment with a supported six- or seven-byte UDH.
- `DecodeCMPPCContent` and `DecodeSMPPCContent` variants for decoding message content.
- `BatchDataCodingEncoder` for trying multiple CMPP or SMPP data codings and choosing the result with the fewest segments.

The library does not maintain cross-segment state or assemble a complete message from multiple received segments. Applications must group segments by their UDH reference and order them by segment index.

## TCP framing and server integration

`codec.NewCMPPCodec` and `codec.NewSMPPCodec` provide non-blocking `Decode` and blocking `DecodeBlocked` paths. The non-blocking path returns `codec.ErrPacketNotComplete` without discarding a partial frame. SGIP and SMGP do not currently have equivalent stream codecs in this repository.

`nioserver.BaseServer` is a generic server primitive. Callers must provide both an `UnpackFunc` and a `HandleFunc`; no protocol dispatcher is selected automatically.

`codec.ConnReader` and `netpoll.Reader` are different interfaces. Using a `codec` implementation inside `nioserver` requires an adapter or a protocol-specific `UnpackFunc`.

Important lifecycle rules:

- `Run` binds and serves synchronously; the application owns process signals.
- `Shutdown(ctx)` performs a context-bounded graceful shutdown.
- A server instance is single-use.
- The default per-connection handler limit is one, preserving request/response order. Raising it opts into concurrent handling and possible response reordering.
- Use `ISMSConn.Write` or `WriteAndClose` for error-returning writes. `AsyncWrite` is deprecated.

## Scope and limitations

- There is no built-in SMS gateway process, SMPP/CMPP client session manager, persistence layer, routing policy, or delivery retry engine.
- Protocol coverage is limited to the PDU families listed above.
- Long-message reassembly state belongs to the caller.
- Authentication credentials, connection state, sequence allocation, and business-level delivery semantics remain application responsibilities.

## Development

Run the following commands from the repository root:

```bash
go test ./...
go test -race ./...
go vet ./...
```

GitHub Actions additionally tests Go 1.20 and the current stable Go release, checks formatting and module tidiness, runs Staticcheck and Actionlint, enforces the configured coverage threshold, performs security scans, and runs the fuzz target matrix on a schedule.

When changing a wire decoder, include valid round-trip cases, malformed-length cases, and a fuzz seed when the new path accepts untrusted bytes.

## Contributing

Issues and pull requests are welcome. For protocol changes, identify the protocol version and command ID, describe the expected wire layout, and include tests that demonstrate the behavior.

## License

This project is licensed under the [MIT License](LICENSE).

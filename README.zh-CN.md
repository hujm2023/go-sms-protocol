# go-sms-protocol

[English](README.md) | 简体中文

[![CI](https://github.com/hujm2023/go-sms-protocol/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/hujm2023/go-sms-protocol/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/hujm2023/go-sms-protocol.svg)](https://pkg.go.dev/github.com/hujm2023/go-sms-protocol)
[![Go version](https://img.shields.io/github/go-mod/go-version/hujm2023/go-sms-protocol)](go.mod)
[![License](https://img.shields.io/github/license/hujm2023/go-sms-protocol)](LICENSE)

`go-sms-protocol` 是一个 Go 协议库，用于编解码 SMPP 3.4、CMPP 2.0/3.0、SGIP 1.2 和 SMGP 3.0 中已经实现的 PDU。它还提供 CMPP/SMPP 流式拆包、短信内容编码、长短信分片辅助函数，以及基于 `netpoll` 的服务端基础组件。

本仓库提供的是库，不是可以直接运行的短信网关或命令行程序。

## 安装

最低支持 Go 1.20。

```bash
go get github.com/hujm2023/go-sms-protocol@latest
```

## 快速开始

下面的程序创建一个 SMPP 3.4 `enquire_link` PDU，将其编码为线上的二进制数据，再通过协议分发器解码。

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

预期输出：

```text
SMPP_ENQUIRE_LINK sequence=42
```

处理线上数据时，应选择与协议及版本对应的分发器：

| 协议 | 分发器 |
| --- | --- |
| CMPP 2.0 | `cmpp20.DecodeCMPP20` |
| CMPP 3.0 | `cmpp30.DecodeCMPP30` |
| SGIP 1.2 | `sgip12.DecodeSGIP12` |
| SMGP 3.0 | `smgp30.DecodeSMGP30` |
| SMPP 3.4 | `smpp34.DecodeSMPP34` |

成功解码后的值都实现根包中的 `protocol.PDU` 接口，包括二进制编解码、命令与序列号访问、响应构造和字符串格式化。

## 协议覆盖范围

下表中的“支持”表示对应包中存在具体 PDU 类型，并且分发器能够识别列出的命令族。仓库中可能还定义了其他命令常量，但没有相应的 PDU 实现。

| 版本 | 已实现的 PDU 命令族 |
| --- | --- |
| CMPP 2.0 | `CONNECT`、`SUBMIT`、`QUERY`、`DELIVER`、`ACTIVE_TEST`、`TERMINATE` 及其响应 |
| CMPP 3.0 | CMPP 2.0 的命令族，以及 `CANCEL` 和对应响应 |
| SGIP 1.2 | `BIND`、`UNBIND`、`SUBMIT`、`REPORT`、`DELIVER` 及其响应 |
| SMGP 3.0 | `LOGIN`、`SUBMIT`、`DELIVER`、`ACTIVE_TEST`、`EXIT` 及其响应 |
| SMPP 3.4 | receiver/transmitter/transceiver bind、`SUBMIT_SM`、`DELIVER_SM`、`ENQUIRE_LINK`、`UNBIND`、`GENERIC_NACK`，以及协议中定义的相应响应 |

当前没有实现 SMPP 5.0。仓库中存在协议 PDF 或版本常量，不代表已有对应的分发器和 PDU 实现。

## 主要包

| 包 | 用途 |
| --- | --- |
| 根包 | 公共 `PDU` 契约、短信内容辅助函数、CMPP/SMPP 长短信拆分、UDH 解析和批量编码选择 |
| [`cmpp/cmpp20`](cmpp/cmpp20)、[`cmpp/cmpp30`](cmpp/cmpp30) | CMPP PDU 类型和版本分发器 |
| [`sgip/sgip12`](sgip/sgip12) | SGIP 1.2 PDU 类型和分发器 |
| [`smgp/smgp30`](smgp/smgp30) | SMGP 3.0 PDU 类型和分发器 |
| [`smpp/smpp34`](smpp/smpp34) | SMPP 3.4 PDU、校验、TLV、状态报告和分发器 |
| [`codec`](codec) | CMPP 和 SMPP 的阻塞/非阻塞 TCP 帧提取 |
| [`datacoding`](datacoding) | ASCII、GB18030、GSM 7-bit packed/unpacked、Latin-1 和 UCS-2 编解码 |
| [`packet`](packet) | 二进制 Reader、Writer 和 PDU 字符串格式化辅助函数 |
| [`nioserver`](nioserver) | 带 handler 和写入准入上限的通用 `netpoll` TCP 服务端组件 |

完整的导出 API 可在 [pkg.go.dev](https://pkg.go.dev/github.com/hujm2023/go-sms-protocol) 查看。

## 短信内容与长短信

根包提供：

- `EncodeCMPPContentAndSplit` 和 `EncodeSMPPContentAndSplit`：编码文本，并在需要分片时添加六字节 UDH。
- `ParseLongSmsContent` 和 `ParseLongSmsContentBytes`：解析带有受支持的六字节或七字节 UDH 的单个分片。
- `DecodeCMPPCContent` 和 `DecodeSMPPCContent` 的相关版本：解码短信内容。
- `BatchDataCodingEncoder`：尝试多种 CMPP 或 SMPP 编码，并选择分片数最少的结果。

本库不维护跨分片状态，也不会把多个接收分片自动组装成完整短信。调用方需要根据 UDH 引用号对分片分组，并按分片序号排序。

## TCP 拆包与服务端集成

`codec.NewCMPPCodec` 和 `codec.NewSMPPCodec` 同时提供非阻塞的 `Decode` 和阻塞的 `DecodeBlocked`。非阻塞路径遇到不完整输入时返回 `codec.ErrPacketNotComplete`，且不会丢弃半个数据帧。当前仓库没有对应的 SGIP/SMGP 流式 codec。

`nioserver.BaseServer` 是通用服务端组件。调用方必须同时提供 `UnpackFunc` 和 `HandleFunc`，服务端不会自动选择协议分发器。

`codec.ConnReader` 与 `netpoll.Reader` 是不同的接口。在 `nioserver` 中使用 `codec` 实现时，需要额外的适配器或协议专属 `UnpackFunc`。

主要生命周期约束如下：

- `Run` 负责绑定并同步运行；进程信号由宿主应用管理。
- `Shutdown(ctx)` 在给定 context 截止时间内执行优雅关闭。
- 每个 server 实例只能运行一次。
- 默认每条连接只允许一个 handler，以保持请求和响应顺序。提高该上限意味着显式接受并发处理和响应乱序的可能。
- 需要获知写入错误时使用 `ISMSConn.Write` 或 `WriteAndClose`；`AsyncWrite` 已弃用。

## 范围与限制

- 本库不提供内置短信网关进程、SMPP/CMPP 客户端会话管理、持久化、路由策略或投递重试引擎。
- 协议覆盖范围仅限上表中的 PDU 命令族。
- 长短信重组状态由调用方维护。
- 认证凭据、连接状态、序列号分配和业务层投递语义仍由应用负责。

## 开发

在仓库根目录运行：

```bash
go test ./...
go test -race ./...
go vet ./...
```

GitHub Actions 还会使用 Go 1.20 和当前稳定版 Go 运行测试，检查格式和 `go.mod`/`go.sum` 是否整洁，执行 Staticcheck、Actionlint 和覆盖率门禁，并运行安全扫描及定时 fuzz 目标矩阵。

修改二进制解码器时，应添加有效数据的 round-trip 用例、异常长度用例；如果新路径接收不可信字节，还应增加相应的 fuzz seed。

## 贡献

欢迎提交 issue 和 pull request。修改协议实现时，请注明协议版本和命令 ID，描述预期的二进制布局，并添加能够说明行为的测试。

## 许可证

本项目使用 [MIT License](LICENSE)。

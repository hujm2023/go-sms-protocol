package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
)

type config struct {
	exampleutil.Credentials
	address, source, destination, message string
	timeout                               time.Duration
}

func main() { os.Exit(realMain(os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr)) }

func realMain(args []string, lookup func(string) (string, bool), stdout, stderr io.Writer) int {
	cfg, err := parseConfig(args, lookup, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "configuration error: %v\n", err)
		return 2
	}
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, cfg.timeout)
	defer cancel()
	if err := runClient(ctx, cfg, stdout); err != nil {
		fmt.Fprintf(stderr, "smpp client: %v\n", err)
		return 1
	}
	return 0
}

func parseConfig(args []string, lookup func(string) (string, bool), output io.Writer) (config, error) {
	var cfg config
	f := flag.NewFlagSet("smpp34client", flag.ContinueOnError)
	f.SetOutput(output)
	f.StringVar(&cfg.address, "server", "127.0.0.1:2775", "SMSC address")
	f.StringVar(&cfg.source, "source", "106900001", "source address")
	f.StringVar(&cfg.destination, "destination", "13800138000", "simulated recipient")
	f.StringVar(&cfg.message, "message", "hello SMPP", "ASCII message, at most 140 bytes")
	f.DurationVar(&cfg.timeout, "timeout", 30*time.Second, "deadline for the complete exchange")
	if err := f.Parse(args); err != nil {
		return config{}, err
	}
	if f.NArg() != 0 || cfg.timeout <= 0 {
		return config{}, errors.New("unexpected arguments or nonpositive timeout")
	}
	credentials, err := exampleutil.LoadCredentials(lookup, "SMPP34", 15, 8)
	if err != nil {
		return config{}, err
	}
	cfg.Credentials = credentials
	for field, value := range map[string]string{smpp.SYSTEM_ID: cfg.Account, smpp.PASSWORD: cfg.Password} {
		maximum := 16
		if field == smpp.PASSWORD {
			maximum = 9
		}
		if err := smpp.ValidateCString(field, value, maximum); err != nil {
			return config{}, err
		}
	}
	for _, value := range []string{cfg.source, cfg.destination} {
		if err := exampleutil.TextField(value, 1, 20); err != nil {
			return config{}, err
		}
		if err := smpp.ValidateCString("address", value, 21); err != nil {
			return config{}, err
		}
	}
	return cfg, exampleutil.ASCIIMessage(cfg.message, 140)
}

func runClient(ctx context.Context, cfg config, output io.Writer) error {
	conn, cleanup, err := exampleutil.Dial(ctx, cfg.address)
	if err != nil {
		return err
	}
	defer cleanup()
	return runSession(conn, cfg, output)
}

func runSession(conn io.ReadWriter, cfg config, output io.Writer) error {
	bind := &smpp34.Bind{
		Header:   smpp.Header{ID: smpp.BIND_TRANSCEIVER, Sequence: 1},
		SystemID: cfg.Account, Password: cfg.Password, InterfaceVersion: 0x34,
	}
	if err := exampleutil.Write(conn, bind); err != nil {
		return err
	}
	var bound, heartbeat, exiting bool
	var submitID, reportID string
	for {
		pdu, err := exampleutil.Read(conn, 16, smpp.MAX_PDU_SIZE, smpp34.DecodeSMPP34)
		if err != nil {
			return err
		}
		switch p := pdu.(type) {
		case *smpp34.BindResp:
			if bound || p.ID != smpp.BIND_TRANSCEIVER_RESP || p.Sequence != 1 || p.Status != smpp.ESME_ROK {
				return fmt.Errorf("bind rejected or unexpected: status=%d", p.Status)
			}
			bound = true
			fmt.Fprintln(output, "bind accepted mode=transceiver")
			submit := &smpp34.SubmitSm{
				Header:     smpp.Header{ID: smpp.SUBMIT_SM, Sequence: 2},
				SourceAddr: cfg.source, DestinationAddr: cfg.destination,
				RegisteredDelivery: 1, DataCoding: smpp.ENCODING_IA5,
				SmLength: uint8(len(cfg.message)), ShortMessage: []byte(cfg.message),
			}
			if err := exampleutil.Write(conn, submit); err != nil {
				return err
			}
			if err := exampleutil.Write(conn, &smpp34.EnquireLink{Header: smpp.Header{ID: smpp.ENQUIRE_LINK, Sequence: 3}}); err != nil {
				return err
			}
		case *smpp34.SubmitSmResp:
			if !bound || p.Sequence != 2 || p.Status != smpp.ESME_ROK || p.MessageID == "" {
				return fmt.Errorf("submit rejected or unexpected: status=%d", p.Status)
			}
			submitID = p.MessageID
			fmt.Fprintf(output, "submit accepted msg_id=%s\n", submitID)
		case *smpp34.DeliverSm:
			if !bound || p.ESMClass&0x3c != 0x04 {
				return errors.New("expected an SMSC delivery receipt")
			}
			var receipt smpp34.DeliveryReceipt
			if err := receipt.IDecode(p.ShortMessage); err != nil {
				return err
			}
			if !receipt.Valid() || receipt.Stat != smpp.DELIVERED || receipt.Err != "000" {
				return errors.New("invalid delivery receipt")
			}
			reportID = receipt.ID
			if err := exampleutil.Write(conn, p.GenEmptyResponse()); err != nil {
				return err
			}
			fmt.Fprintf(output, "receipt received msg_id=%s status=DELIVRD acknowledged\n", reportID)
		case *smpp34.EnquireLink:
			if err := exampleutil.Write(conn, p.GenEmptyResponse()); err != nil {
				return err
			}
		case *smpp34.EnquireLinkResp:
			if !bound || p.Sequence != 3 || p.Status != smpp.ESME_ROK {
				return errors.New("unexpected heartbeat response")
			}
			heartbeat = true
			fmt.Fprintln(output, "heartbeat acknowledged")
		case *smpp34.UnBindResp:
			if !exiting || p.Sequence != 4 || p.Status != smpp.ESME_ROK {
				return errors.New("unexpected unbind response")
			}
			fmt.Fprintln(output, "unbind acknowledged")
			return nil
		case *smpp34.Unbind:
			if err := exampleutil.Write(conn, p.GenEmptyResponse()); err != nil {
				return err
			}
			return errors.New("server unbound before the exchange completed")
		default:
			return fmt.Errorf("unexpected command: %s", pdu.GetCommand())
		}
		if !exiting && submitID != "" && reportID != "" && heartbeat {
			if submitID != reportID {
				return errors.New("receipt does not match the submission")
			}
			if err := exampleutil.Write(conn, &smpp34.Unbind{Header: smpp.Header{ID: smpp.UNBIND, Sequence: 4}}); err != nil {
				return err
			}
			exiting = true
		}
	}
}

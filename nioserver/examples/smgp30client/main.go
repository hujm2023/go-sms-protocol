package main

import (
	"context"
	"crypto/md5"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/smgp"
	"github.com/hujm2023/go-sms-protocol/smgp/smgp30"
)

type config struct {
	address     string
	account     string
	password    string
	destination string
	source      string
	message     string
	timeout     time.Duration
}

func main() {
	os.Exit(realMain(os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr))
}

func realMain(args []string, lookupEnv func(string) (string, bool), stdout, stderr io.Writer) int {
	cfg, err := parseConfig(args, lookupEnv, stderr)
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
		fmt.Fprintf(stderr, "smgp client: %v\n", err)
		return 1
	}
	return 0
}

func parseConfig(args []string, lookupEnv func(string) (string, bool), output io.Writer) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("smgp30client", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&cfg.address, "server", "127.0.0.1:8890", "SMGP server address")
	flags.StringVar(&cfg.destination, "destination", "13800138000", "recipient for the simulated submit")
	flags.StringVar(&cfg.source, "source", "118900001", "SP source number")
	flags.StringVar(&cfg.message, "message", "hello SMGP", "ASCII message (at most 140 bytes)")
	flags.DurationVar(&cfg.timeout, "timeout", 30*time.Second, "deadline for the complete example exchange")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	if flags.NArg() != 0 {
		return config{}, errors.New("unexpected positional arguments")
	}
	cfg.account, _ = lookupEnv("SMGP30_ACCOUNT")
	cfg.password, _ = lookupEnv("SMGP30_PASSWORD")
	if cfg.account == "" || len(cfg.account) > 8 || strings.ContainsAny(cfg.account, "\x00\r\n\t") {
		return config{}, errors.New("SMGP30_ACCOUNT must be 1 to 8 bytes without control characters")
	}
	if cfg.password == "" || len(cfg.password) > 15 {
		return config{}, errors.New("SMGP30_PASSWORD must be 1 to 15 bytes")
	}
	if cfg.destination == "" || len(cfg.destination) > 21 || strings.ContainsRune(cfg.destination, 0) {
		return config{}, errors.New("destination must be 1 to 21 bytes without NUL")
	}
	if cfg.source == "" || len(cfg.source) > 21 || strings.ContainsRune(cfg.source, 0) {
		return config{}, errors.New("source must be 1 to 21 bytes without NUL")
	}
	if len(cfg.message) > 140 {
		return config{}, errors.New("message must be at most 140 ASCII bytes")
	}
	for _, b := range []byte(cfg.message) {
		if b > 127 {
			return config{}, errors.New("message must contain only ASCII bytes")
		}
	}
	if cfg.timeout <= 0 {
		return config{}, errors.New("timeout must be positive")
	}
	if _, err := net.ResolveTCPAddr("tcp", cfg.address); err != nil {
		return config{}, fmt.Errorf("resolve server address: %w", err)
	}
	return cfg, nil
}

func runClient(ctx context.Context, cfg config, output io.Writer) error {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", cfg.address)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return err
		}
	}
	// Go 1.20: close the socket on cancellation to interrupt blocked reads.
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()
	defer func() { close(done); <-stopped }()

	login := smgp30.NewLogin(cfg.account, cfg.password, 1)
	if err := writePDU(conn, login); err != nil {
		return err
	}
	var authenticated, heartbeatOK, exiting bool
	var submitID, receiptID string
	for {
		pdu, err := readPDU(conn)
		if err != nil {
			return fmt.Errorf("receive smgp pdu: %w", err)
		}
		switch p := pdu.(type) {
		case *smgp30.LoginResp:
			if authenticated || p.GetSequenceID() != login.GetSequenceID() {
				return errors.New("unexpected login response")
			}
			if err := verifyLoginResponse(login, p, cfg.password); err != nil {
				return err
			}
			authenticated = true
			fmt.Fprintln(output, "login accepted login_mode=2")
			submit := &smgp30.Submit{
				Header:  smgp.NewHeader(0, smgp.CommandSubmit, 2),
				MsgType: smgp.MT, NeedReport: smgp.NEED_REPORT, FeeType: "00",
				MsgFormat: smgp.ASCII, SrcTermID: cfg.source,
				DestTermIDCount: 1, DestTermID: []string{cfg.destination},
				MsgLength: uint8(len(cfg.message)), MsgContent: []byte(cfg.message),
			}
			if err := writePDU(conn, submit); err != nil {
				return err
			}
			heartbeat := &smgp30.ActiveTest{Header: smgp.NewHeader(0, smgp.CommandActiveTest, 3)}
			if err := writePDU(conn, heartbeat); err != nil {
				return err
			}
		case *smgp30.SubmitResp:
			if !authenticated || p.GetSequenceID() != 2 || submitID != "" {
				return errors.New("unexpected submit response")
			}
			if p.Status != 0 {
				return fmt.Errorf("submit rejected: status=%d", p.Status)
			}
			submitID = p.MsgID
			fmt.Fprintf(output, "submit accepted msg_id=%s\n", submitID)
		case *smgp30.Deliver:
			if !authenticated || p.IsReport != smgp.IS_REPORT || p.MsgFormat != smgp.ASCII {
				return errors.New("unexpected delivery; expected an ASCII status report")
			}
			var receipt smgp30.DeliveryReceipt
			if err := receipt.IDecode(p.MsgContent); err != nil {
				return fmt.Errorf("decode receipt: %w", err)
			}
			if !receipt.Valid() || receipt.Stat != "DELIVRD" || receipt.Err != "000" {
				return fmt.Errorf("unexpected receipt status=%q error=%q", receipt.Stat, receipt.Err)
			}
			receiptID = receipt.ID
			ack := p.GenEmptyResponse().(*smgp30.DeliverResp)
			// GenEmptyResponse sets only the sequence, so echo MsgID explicitly.
			ack.MsgID = p.MsgID
			if err := writePDU(conn, ack); err != nil {
				return err
			}
			fmt.Fprintf(output, "receipt received msg_id=%s status=%s acknowledged\n", receiptID, receipt.Stat)
		case *smgp30.ActiveTest:
			if err := writePDU(conn, p.GenEmptyResponse()); err != nil {
				return err
			}
		case *smgp30.ActiveTestResp:
			if !authenticated || p.GetSequenceID() != 3 {
				return errors.New("unexpected heartbeat response")
			}
			heartbeatOK = true
			fmt.Fprintln(output, "heartbeat acknowledged")
		case *smgp30.ExitResp:
			if !exiting || p.GetSequenceID() != 4 {
				return errors.New("unexpected exit response")
			}
			fmt.Fprintln(output, "exit acknowledged")
			return nil
		case *smgp30.Exit:
			if err := writePDU(conn, p.GenEmptyResponse()); err != nil {
				return err
			}
			return errors.New("server closed the session before the example completed")
		default:
			return fmt.Errorf("unexpected command: %s", pdu.GetCommand())
		}
		if !exiting && submitID != "" && receiptID != "" && heartbeatOK {
			if submitID != receiptID {
				return errors.New("receipt does not match the submitted message ID")
			}
			if err := writePDU(conn, &smgp30.Exit{Header: smgp.NewHeader(0, smgp.CommandExit, 4)}); err != nil {
				return err
			}
			exiting = true
		}
	}
}

func verifyLoginResponse(login *smgp30.Login, response *smgp30.LoginResp, password string) error {
	if response.Status != 0 {
		return fmt.Errorf("login rejected: status=%d", response.Status)
	}
	if response.ServerVersion != 0x30 {
		return fmt.Errorf("unsupported server version: %#x", response.ServerVersion)
	}
	data := make([]byte, 4, 4+md5.Size+len(password))
	binary.BigEndian.PutUint32(data, uint32(response.Status))
	data = append(data, login.AuthenticatorClient...)
	data = append(data, password...)
	// SMGP protocol exception: 3.0.3 section 6.2.7 requires MD5 to verify AuthenticatorServer.
	want := md5.Sum(data)
	if subtle.ConstantTimeCompare([]byte(response.AuthenticatorServer), want[:]) != 1 {
		return errors.New("server authenticator mismatch")
	}
	return nil
}

func readPDU(reader io.Reader) (protocol.PDU, error) {
	var length [4]byte
	if _, err := io.ReadFull(reader, length[:]); err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(length[:])
	if size < smgp.HeaderLength || size > 64*1024 {
		return nil, fmt.Errorf("invalid frame length: %d", size)
	}
	frame := make([]byte, int(size))
	copy(frame, length[:])
	if _, err := io.ReadFull(reader, frame[4:]); err != nil {
		return nil, err
	}
	return smgp30.DecodeSMGP30(frame)
}

func writePDU(writer io.Writer, pdu protocol.PDU) error {
	data, err := pdu.IEncode()
	if err != nil {
		return fmt.Errorf("encode %s: %w", pdu.GetCommand(), err)
	}
	for len(data) > 0 {
		n, err := writer.Write(data)
		if err != nil {
			return fmt.Errorf("write %s: %w", pdu.GetCommand(), err)
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}

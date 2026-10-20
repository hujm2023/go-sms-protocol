package exampleutil

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"
)

type ServerConfig struct {
	Listen          string
	Workers         int
	Queue           int
	MinDelay        time.Duration
	MaxDelay        time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func (c *ServerConfig) Flags(f *flag.FlagSet, address string) {
	f.StringVar(&c.Listen, "listen", address, "TCP listen address")
	f.IntVar(&c.Workers, "receipt-workers", 4, "number of simulated receipt workers")
	f.IntVar(&c.Queue, "receipt-queue", 1024, "maximum queued submissions awaiting receipts")
	f.DurationVar(&c.MinDelay, "receipt-min-delay", 100*time.Millisecond, "minimum simulated receipt delay")
	f.DurationVar(&c.MaxDelay, "receipt-max-delay", 500*time.Millisecond, "maximum simulated receipt delay")
	f.DurationVar(&c.ReadTimeout, "read-timeout", 2*time.Minute, "per-read timeout")
	f.DurationVar(&c.WriteTimeout, "write-timeout", 5*time.Second, "per-write timeout")
	f.DurationVar(&c.IdleTimeout, "idle-timeout", 2*time.Minute, "idle timeout")
	f.DurationVar(&c.ShutdownTimeout, "shutdown-timeout", 10*time.Second, "graceful shutdown timeout")
}

func (c ServerConfig) Validate() error {
	if c.Listen == "" {
		return errors.New("listen address is required")
	}
	if _, err := net.ResolveTCPAddr("tcp", c.Listen); err != nil {
		return fmt.Errorf("resolve listen address: %w", err)
	}
	if c.Workers < 1 || c.Queue < 1 {
		return errors.New("receipt workers and queue must be positive")
	}
	if c.MinDelay <= 0 || c.MaxDelay < c.MinDelay {
		return errors.New("receipt delays must be positive, with max >= min")
	}
	if c.ReadTimeout <= 0 || c.WriteTimeout <= 0 || c.IdleTimeout <= 0 || c.ShutdownTimeout <= 0 {
		return errors.New("transport and shutdown timeouts must be positive")
	}
	return nil
}

type Credentials struct{ Account, Password string }

func LoadCredentials(lookup func(string) (string, bool), prefix string, accountMax, passwordMax int) (Credentials, error) {
	account, _ := lookup(prefix + "_ACCOUNT")
	password, _ := lookup(prefix + "_PASSWORD")
	if err := TextField(account, 1, accountMax); err != nil {
		return Credentials{}, fmt.Errorf("%s_ACCOUNT: %w", prefix, err)
	}
	if err := TextField(password, 1, passwordMax); err != nil {
		return Credentials{}, fmt.Errorf("%s_PASSWORD: %w", prefix, err)
	}
	return Credentials{Account: account, Password: password}, nil
}

func TextField(value string, minimum, maximum int) error {
	if len(value) < minimum || len(value) > maximum || strings.IndexFunc(value, func(r rune) bool { return r < 32 || r == 127 }) >= 0 {
		return fmt.Errorf("must be %d..%d bytes without control characters", minimum, maximum)
	}
	return nil
}

func ASCIIMessage(value string, maximum int) error {
	if len(value) > maximum {
		return fmt.Errorf("message exceeds %d bytes", maximum)
	}
	for _, b := range []byte(value) {
		if b > 127 {
			return errors.New("this example accepts ASCII messages only")
		}
	}
	return nil
}

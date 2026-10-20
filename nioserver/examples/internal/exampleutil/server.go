package exampleutil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/hujm2023/hlog"

	"github.com/hujm2023/go-sms-protocol/nioserver"
)

type Lifecycle struct {
	Name            string
	ShutdownTimeout time.Duration
	Ready           func(net.Addr)
	Stop            func()
	Wait            func(context.Context) error
}

// Run waits for binding, reports readiness, and owns the Run goroutine through shutdown.
func Run[T any](ctx context.Context, server *nioserver.BaseServer[T], logger hlog.FullLogger, life Lifecycle) error {
	result := make(chan error, 1)
	go func() { result <- server.Run() }()
	finish := func(runErr error, waitRun bool) error {
		if life.Stop != nil {
			life.Stop()
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), life.ShutdownTimeout)
		defer cancel()
		serverErr := server.Shutdown(shutdownCtx)
		var servicesErr error
		if life.Wait != nil {
			servicesErr = life.Wait(shutdownCtx)
		}
		if waitRun {
			select {
			case runErr = <-result:
			case <-shutdownCtx.Done():
				runErr = shutdownCtx.Err()
			}
		}
		return errors.Join(serverErr, servicesErr, runErr)
	}
	startup := time.NewTimer(5 * time.Second)
	defer startup.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for server.Addr() == nil {
		select {
		case err := <-result:
			if err == nil {
				err = errors.New("server stopped before binding")
			}
			return finish(err, false)
		case <-ctx.Done():
			return finish(nil, true)
		case <-startup.C:
			return errors.Join(errors.New("server binding timed out"), finish(nil, true))
		case <-ticker.C:
		}
	}
	logger.Infof("%s listening address=%s", life.Name, server.Addr())
	if life.Ready != nil {
		life.Ready(server.Addr())
	}
	select {
	case err := <-result:
		if err == nil {
			err = fmt.Errorf("%s stopped unexpectedly", life.Name)
		}
		return finish(err, false)
	case <-ctx.Done():
		return finish(nil, true)
	}
}

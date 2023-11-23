package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

// Fatal calls the default logger's Fatal method and then os.Exit(1).
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Error calls the default logger's Error method.
func Error(v ...interface{}) {
	logger.Error(v...)
}

// Warn calls the default logger's Warn method.
func Warn(v ...interface{}) {
	logger.Warn(v...)
}

// Notice calls the default logger's Notice method.
func Notice(v ...interface{}) {
	logger.Notice(v...)
}

// Info calls the default logger's Info method.
func Info(v ...interface{}) {
	logger.Info(v...)
}

// Debug calls the default logger's Debug method.
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Trace calls the default logger's Trace method.
func Trace(v ...interface{}) {
	logger.Trace(v...)
}

// Fatalf calls the default logger's Fatalf method and then os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// Errorf calls the default logger's Errorf method.
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Warnf calls the default logger's Warnf method.
func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

// Noticef calls the default logger's Noticef method.
func Noticef(format string, v ...interface{}) {
	logger.Noticef(format, v...)
}

// Infof calls the default logger's Infof method.
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Debugf calls the default logger's Debugf method.
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Tracef calls the default logger's Tracef method.
func Tracef(format string, v ...interface{}) {
	logger.Tracef(format, v...)
}

// CtxFatal calls the default logger's CtxFatal method and then os.Exit(1).
func CtxFatal(ctx context.Context, format string, v ...interface{}) {
	logger.CtxFatalf(ctx, format, v...)
}

// CtxError calls the default logger's CtxError method.
func CtxError(ctx context.Context, format string, v ...interface{}) {
	logger.CtxErrorf(ctx, format, v...)
}

// CtxWarn calls the default logger's CtxWarn method.
func CtxWarn(ctx context.Context, format string, v ...interface{}) {
	logger.CtxWarnf(ctx, format, v...)
}

// CtxNotice calls the default logger's CtxNotice method.
func CtxNotice(ctx context.Context, format string, v ...interface{}) {
	logger.CtxNoticef(ctx, format, v...)
}

// CtxInfo calls the default logger's CtxInfo method.
func CtxInfo(ctx context.Context, format string, v ...interface{}) {
	logger.CtxInfof(ctx, format, v...)
}

// CtxDebug calls the default logger's CtxDebug method.
func CtxDebug(ctx context.Context, format string, v ...interface{}) {
	logger.CtxDebugf(ctx, format, v...)
}

// CtxTrace calls the default logger's CtxTrace method.
func CtxTrace(ctx context.Context, format string, v ...interface{}) {
	logger.CtxTracef(ctx, format, v...)
}

type defaultLogger struct {
	stdlog *log.Logger
	level  Level
	depth  int
}

func (ll *defaultLogger) SetOutput(w io.Writer) {
	ll.stdlog.SetOutput(w)
}

func (ll *defaultLogger) SetLevel(lv Level) {
	ll.level = lv
}

func (ll *defaultLogger) logf(lv Level, format *string, v ...interface{}) {
	if ll.level > lv {
		return
	}
	msg := lv.toString()
	if format != nil {
		if len(v) > 0 {
			msg += fmt.Sprintf(*format, v...)
		} else {
			msg += *format
		}
	} else {
		msg += fmt.Sprint(v...)
	}

	ll.stdlog.Output(ll.depth, msg)
	if lv == LevelFatal {
		os.Exit(1)
	}
}

func (ll *defaultLogger) Fatal(v ...interface{}) {
	ll.logf(LevelFatal, nil, v...)
}

func (ll *defaultLogger) Error(v ...interface{}) {
	ll.logf(LevelError, nil, v...)
}

func (ll *defaultLogger) Warn(v ...interface{}) {
	ll.logf(LevelWarn, nil, v...)
}

func (ll *defaultLogger) Notice(v ...interface{}) {
	ll.logf(LevelNotice, nil, v...)
}

func (ll *defaultLogger) Info(v ...interface{}) {
	ll.logf(LevelInfo, nil, v...)
}

func (ll *defaultLogger) Debug(v ...interface{}) {
	ll.logf(LevelDebug, nil, v...)
}

func (ll *defaultLogger) Trace(v ...interface{}) {
	ll.logf(LevelTrace, nil, v...)
}

func (ll *defaultLogger) Fatalf(format string, v ...interface{}) {
	ll.logf(LevelFatal, &format, v...)
}

func (ll *defaultLogger) Errorf(format string, v ...interface{}) {
	ll.logf(LevelError, &format, v...)
}

func (ll *defaultLogger) Warnf(format string, v ...interface{}) {
	ll.logf(LevelWarn, &format, v...)
}

func (ll *defaultLogger) Noticef(format string, v ...interface{}) {
	ll.logf(LevelNotice, &format, v...)
}

func (ll *defaultLogger) Infof(format string, v ...interface{}) {
	ll.logf(LevelInfo, &format, v...)
}

func (ll *defaultLogger) Debugf(format string, v ...interface{}) {
	ll.logf(LevelDebug, &format, v...)
}

func (ll *defaultLogger) Tracef(format string, v ...interface{}) {
	ll.logf(LevelTrace, &format, v...)
}

func (ll *defaultLogger) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelFatal, &format, v...)
}

func (ll *defaultLogger) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelError, &format, v...)
}

func (ll *defaultLogger) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelWarn, &format, v...)
}

func (ll *defaultLogger) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelNotice, &format, v...)
}

func (ll *defaultLogger) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelInfo, &format, v...)
}

func (ll *defaultLogger) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelDebug, &format, v...)
}

func (ll *defaultLogger) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	ll.logf(LevelTrace, &format, v...)
}

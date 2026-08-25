package logger

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
)

func TestDefaultLoggerFiltersAndFormatsLevels(t *testing.T) {
	var buf bytes.Buffer
	ll := &defaultLogger{
		stdlog: log.New(&buf, "", 0),
		level:  LevelInfo,
		depth:  1,
	}

	ll.Trace("hidden")
	ll.Infof("value=%d", 7)
	ll.CtxWarnf(context.Background(), "warning=%s", "visible")

	got := buf.String()
	if strings.Contains(got, "hidden") {
		t.Fatalf("trace message was not filtered: %q", got)
	}
	for _, want := range []string{"[Info] value=7", "[Warn] warning=visible"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output %q does not contain %q", got, want)
		}
	}
}

func TestGlobalLoggerAndSystemPrefix(t *testing.T) {
	oldLogger, oldSysLogger, oldSilent := logger, sysLogger, silentMode
	t.Cleanup(func() {
		logger = oldLogger
		sysLogger = oldSysLogger
		silentMode = oldSilent
	})

	var buf bytes.Buffer
	base := &defaultLogger{stdlog: log.New(&buf, "", 0), depth: 1}
	SetLogger(base)
	SetLevel(LevelInfo)
	SetOutput(&buf)
	SetSilentMode(false)

	Infof("default=%s", "ok")
	SystemLogger().Warnf("system=%s", "ok")
	CtxInfo(context.Background(), "context")

	got := buf.String()
	for _, want := range []string{
		"[Info] default=ok",
		"[Warn] [go-sms-protocol]: system=ok",
		"[Info] context",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output %q does not contain %q", got, want)
		}
	}
}

func TestSystemLoggerSilentModeOnlySuppressesEngineError(t *testing.T) {
	oldLogger, oldSysLogger, oldSilent := logger, sysLogger, silentMode
	t.Cleanup(func() {
		logger = oldLogger
		sysLogger = oldSysLogger
		silentMode = oldSilent
	})

	var buf bytes.Buffer
	base := &defaultLogger{stdlog: log.New(&buf, "", 0), depth: 1}
	SetLogger(base)
	SetLevel(LevelTrace)
	SetSilentMode(true)

	SystemLogger().Errorf(EngineErrorFormat, "read failed", "127.0.0.1:1")
	SystemLogger().Errorf("other error: %s", "visible")

	got := buf.String()
	if strings.Contains(got, "read failed") {
		t.Fatalf("engine error was not silenced: %q", got)
	}
	if !strings.Contains(got, "[Error] [go-sms-protocol]: other error: visible") {
		t.Fatalf("non-engine error missing from output: %q", got)
	}
}

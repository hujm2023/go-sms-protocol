package logger

import (
	"context"
	"testing"
)

func TestLog(t *testing.T) {
	CtxInfo(context.Background(), "hello %s", "world")
}

package datacoding

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGSM7Packed(t *testing.T) {
	s := `[ByteDance] Hello world. 123`
	t.Log([]byte(s))

	data, err := GSM7Packed(s).Encode()
	assert.Nil(t, err)
	t.Log(data)
	t.Log(len([]byte(s)), len(data))

	ss, err := GSM7Packed(data).Decode()
	assert.Nil(t, err)
	t.Log(ss)

	assert.Equal(t, s, string(ss))

	want := []byte("\xC8\x32\x9B\xFD\x06\xDD\xDF\x72\x36\x19")
	text := []byte("Hello world")
	sss := GSM7Packed(text)
	have, _ := sss.Encode()
	if !bytes.Equal(want, have) {
		t.Fatalf("Unexpected text; want %q, have %q", want, have)
	}
}

package datacoding

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAscii_Encode(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		s := "Hello, playground"
		encoded, err := Ascii(s).Encode()
		assert.Nil(t, err)
		assert.True(t, bytes.Equal(encoded, []byte(s)))

		data, err := Ascii(encoded).Decode()
		assert.Nil(t, err)
		assert.Equal(t, s, string(data))
	})
	t.Run("非ascii-encode", func(t *testing.T) {
		s := "【你好】哈哈"
		encoded, err := Ascii(s).Encode()
		assert.Equal(t, err, ErrInvalidCharacter)
		assert.Nil(t, encoded)
	})
	t.Run("非ascii-decode", func(t *testing.T) {
		s := "【你好】哈哈"
		encoded, err := Ascii(s).Decode()
		assert.Equal(t, err, ErrInvalidCharacter)
		assert.Nil(t, encoded)
	})
}

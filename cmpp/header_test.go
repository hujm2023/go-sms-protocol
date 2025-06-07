package cmpp

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryReadHeader(t *testing.T) {
	t.Run("正常长度的 byte", func(t *testing.T) {
		h := Header{
			TotalLength: 39,
			CommandID:   CommandConnect,
			SequenceID:  123,
		}
		h1, err := PeekHeader(h.Bytes())
		assert.Nil(t, err)
		reflect.DeepEqual(h1, h)
	})
	t.Run("原来的数据很长，验证读取 cmpp.Header 后是否影响原有的 []byte", func(t *testing.T) {
		h := Header{
			TotalLength: 39,
			CommandID:   CommandConnect,
			SequenceID:  123,
		}
		data := h.Bytes()
		data = append(data, []byte{0, 1, 4, 6, 3}...)
		l1 := len(data)
		tmp := make([]byte, l1)
		copy(tmp, data)

		h, err := PeekHeader(data)
		if err != nil {
			t.Fatal(err)
		}
		l2 := len(data)

		assert.Equal(t, true, reflect.DeepEqual(h, NewHeader(39, CommandConnect, 123)))
		assert.Equal(t, l1, l2)
		assert.Equal(t, true, bytes.Equal(tmp, data))
	})
}

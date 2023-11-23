package cmpp20

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConnect(t *testing.T) {
	c := NewConnect("900001", "888888", 0x17)
	data, err := c.IEncode()
	assert.Nil(t, err)
	t.Log(data)
}

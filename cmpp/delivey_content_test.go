package cmpp

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/packet"
)

func TestSubPduDeliveryContent_IEncodeFieldLengthError(t *testing.T) {
	content := &SubPduDeliveryContent{
		Stat: strings.Repeat("X", 8),
	}

	data, err := content.IEncode()
	assert.Nil(t, data)
	if !assert.Error(t, err) {
		return
	}

	var lengthErr *packet.FieldLengthError
	if !errors.As(err, &lengthErr) {
		t.Fatalf("errors.As(%T) did not expose FieldLengthError: %v", err, err)
	}
	assert.Equal(t, "Stat", lengthErr.Field)
	assert.Equal(t, 8, lengthErr.Actual)
	assert.Equal(t, 7, lengthErr.Limit)
	assert.ErrorIs(t, err, packet.ErrFieldLengthExceeded)
	assert.Contains(t, err.Error(), `field "Stat"`)
	assert.Contains(t, err.Error(), "actual length 8")
	assert.Contains(t, err.Error(), "limit 7")
}

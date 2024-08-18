package smpp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	s := ESME_RINVDSTADR
	assert.Equal(t, "SMPP_STATUS_11:Invalid Dest Addr", s.String())

	h := Header{
		Length:   16,
		ID:       SUBMIT_SM,
		Status:   s,
		Sequence: 1,
	}

	assert.Equal(
		t,
		"{Length=16, ID=SMPP_SUBMIT_SM, Status=(valueInt=11,valueString=Invalid Dest Addr), Sequence=1}",
		h.String(),
	)
}

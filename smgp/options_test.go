package smgp

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/packet"
)

func TestParseOptions(t *testing.T) {
	rawData := []byte{
		0x00, 0x01, // TAG_TP_pid
		0x00, 0x02, // Length: 2
		0x12, 0x34, // Value

		0x00, 0x02, // TAG_TP_udhi
		0x00, 0x01, // Length: 1
		0x56, // Value

		0x00, 0x03, // TAG_LinkID
		0x00, 0x02, // Length: 2
		0xAB, 0xCD, // Value
	}

	expectedOptions := Options{}
	expectedOptions.Add(NewOption(
		TAG_TP_pid, []byte{0x12, 0x34},
	))
	expectedOptions.Add(NewOption(
		TAG_TP_udhi, []byte{0x56},
	))
	expectedOptions.Add(NewOption(
		TAG_LinkID, []byte{0xAB, 0xCD},
	))

	options, err := ParseOptions(rawData)
	assert.NoError(t, err)
	assert.Equal(t, expectedOptions, options)
}

func TestOptionsLen(t *testing.T) {
	options := Options{}

	options.Add(NewOption(
		TAG_TP_pid, []byte{0x12, 0x34},
	))
	options.Add(NewOption(
		TAG_TP_udhi, []byte{0x56},
	))
	options.Add(NewOption(
		TAG_LinkID, []byte{0xAB, 0xCD},
	))

	expectedLen := 4 + len(options[TAG_TP_pid].ValueBytes) + 4 + len(options[TAG_TP_udhi].ValueBytes) + 4 + len(options[TAG_LinkID].ValueBytes)
	assert.Equal(t, expectedLen, options.Len())
}

func TestOptionsSerialize_1(t *testing.T) {
	options := Options{}
	options.Add(NewOption(
		TAG_TP_pid, []byte{0x12, 0x34},
	))

	expectedBytes := []byte{
		0x00, 0x01, // TAG_TP_pid
		0x00, 0x02, // Length: 2
		0x12, 0x34, // Value

		// 0x00, 0x02, // TAG_TP_udhi
		// 0x00, 0x01, // Length: 1
		// 0x56, // Value

		// 0x00, 0x03, // TAG_LinkID
		// 0x00, 0x02, // Length: 2
		// 0xAB, 0xCD, // Value
	}

	// note: Serialize map 遍历是乱序的
	serializedBytes := options.Serialize()
	assert.Equal(t, expectedBytes, serializedBytes)
}

func TestOptionsTP_udhi(t *testing.T) {
	options := Options{}
	options.Add(NewOption(
		TAG_TP_udhi, []byte{0x56},
	))

	expectedValue := uint8(0x56)
	assert.Equal(t, expectedValue, options.TP_udhi())

	options = Options{}
	expectedDefaultValue := uint8(0)
	assert.Equal(t, expectedDefaultValue, options.TP_udhi())
}

func TestOptionsAddNilDoesNotPanicOrMutate(t *testing.T) {
	var options Options

	assert.NotPanics(t, func() {
		options.Add(NewOption(TAG_TP_pid, []byte{0x01}))
	})
	assert.Nil(t, options)
}

func TestOptionsTPUdhiEmptyValueIsSafe(t *testing.T) {
	options := Options{TAG_TP_udhi: NewOption(TAG_TP_udhi, nil)}

	assert.NotPanics(t, func() {
		assert.Zero(t, options.TP_udhi())
	})
}

func TestOptionsSerializeParseSemanticRoundTrip(t *testing.T) {
	want := Options{}
	want.Add(NewOption(TAG_TP_pid, []byte{0x12, 0x34}))
	want.Add(NewOption(TAG_TP_udhi, []byte{0x01}))
	want.Add(NewOption(TAG_LinkID, []byte("link")))

	raw := want.Serialize()
	got, err := ParseOptions(raw)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, want.Len(), len(raw))
}

func TestParseOptionsRejectsTruncatedInput(t *testing.T) {
	raw := NewOption(TAG_TP_pid, []byte{0x12, 0x34}).Bytes()
	for i := 1; i < len(raw); i++ {
		t.Run(fmt.Sprintf("truncated_at_%d", i), func(t *testing.T) {
			_, err := ParseOptions(raw[:i])
			if !errors.Is(err, ErrLength) {
				t.Fatalf("ParseOptions(%x) error=%v, want ErrLength", raw[:i], err)
			}
		})
	}
}

func TestReadOptionsPreservesTruncationError(t *testing.T) {
	raw := NewOption(TAG_TP_pid, []byte{0x12, 0x34}).Bytes()
	for i := 1; i < len(raw); i++ {
		t.Run(fmt.Sprintf("truncated_at_%d", i), func(t *testing.T) {
			r := packet.NewPacketReader(raw[:i])
			defer r.Release()

			if got := ReadOptions(r); got != nil {
				t.Fatalf("ReadOptions(%x)=%v, want nil on truncation", raw[:i], got)
			}
			if r.Error() == nil {
				t.Fatalf("ReadOptions(%x) cleared the reader error", raw[:i])
			}
		})
	}
}

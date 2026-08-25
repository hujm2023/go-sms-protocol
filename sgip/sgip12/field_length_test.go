package sgip12

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hujm2023/go-sms-protocol/packet"
)

func TestIEncodeFieldLengthErrors(t *testing.T) {
	tests := []struct {
		name   string
		encode func() ([]byte, error)
		field  string
		actual int
		limit  int
	}{
		{
			name: "bind name",
			encode: func() ([]byte, error) {
				return (&Bind{Name: strings.Repeat("n", 17)}).IEncode()
			},
			field:  "Name",
			actual: 17,
			limit:  16,
		},
		{
			name: "submit indexed user number",
			encode: func() ([]byte, error) {
				return (&Submit{
					UserCount: 2,
					UserNumber: []string{
						"valid",
						strings.Repeat("u", 22),
					},
				}).IEncode()
			},
			field:  "UserNumber[1]",
			actual: 22,
			limit:  21,
		},
		{
			name: "deliver user number",
			encode: func() ([]byte, error) {
				return (&Deliver{UserNumber: strings.Repeat("u", 22)}).IEncode()
			},
			field:  "UserNumber",
			actual: 22,
			limit:  21,
		},
		{
			name: "report reserved",
			encode: func() ([]byte, error) {
				return (&Report{Reserved: strings.Repeat("r", 9)}).IEncode()
			},
			field:  "Reserved",
			actual: 9,
			limit:  8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.encode()
			if err == nil {
				t.Fatal("IEncode accepted an over-limit field")
			}

			var lengthErr *packet.FieldLengthError
			if !errors.As(err, &lengthErr) {
				t.Fatalf("errors.As(%T) did not expose FieldLengthError: %v", err, err)
			}
			if lengthErr.Field != tt.field {
				t.Errorf("Field = %q, want %q", lengthErr.Field, tt.field)
			}
			if lengthErr.Actual != tt.actual {
				t.Errorf("Actual = %d, want %d", lengthErr.Actual, tt.actual)
			}
			if lengthErr.Limit != tt.limit {
				t.Errorf("Limit = %d, want %d", lengthErr.Limit, tt.limit)
			}
			if !errors.Is(err, packet.ErrFieldLengthExceeded) {
				t.Errorf("errors.Is(%v, ErrFieldLengthExceeded) = false", err)
			}

			for _, metadata := range []string{
				"WriteFixedLenStringField write error",
				fmt.Sprintf("field %q", tt.field),
				fmt.Sprintf("actual length %d", tt.actual),
				fmt.Sprintf("limit %d", tt.limit),
				fmt.Sprintf("excess %d", tt.actual-tt.limit),
			} {
				if !strings.Contains(err.Error(), metadata) {
					t.Errorf("error %q does not contain metadata %q", err, metadata)
				}
			}
		})
	}
}

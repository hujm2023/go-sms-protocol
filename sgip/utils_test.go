package sgip

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeStamp(t *testing.T) {
	n := time.Now()
	tests := []struct {
		name string
		want uint32
	}{
		{
			name: "testTimeStamp",
			want: uint32((n.Month()) * 100000000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Timestamp(time.Now())
			if got < tt.want {
				t.Error("not actual: expect ", got, tt.want)
			}
		})
	}
}

func Test_genSgipMobile(t *testing.T) {
	tests := []struct {
		name   string
		mobile string
		want   string
	}{
		{
			"有86", "8618800110011", "8618800110011",
		},
		{
			"无86", "18800110011", "8618800110011",
		},
		{
			"+86", "+8618800110011", "8618800110011",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, FixSGIPMobile(tt.mobile), "genSgipMobile(%v)", tt.mobile)
		})
	}
}

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
	// 注意：函数名应该与被测试函数名一致
	// FixSGIPMobile 是实际的函数名
	// 但为了保持与现有代码的一致性，这里暂时保留原名称
	// 建议后续修改为 TestFixSGIPMobile

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

// TestSequenceIDString 测试SequenceIDString函数
func TestSequenceIDString(t *testing.T) {
	tests := []struct {
		name       string
		sequenceID [3]uint32
		want       string
	}{{
		name:       "正常情况",
		sequenceID: [3]uint32{1, 2, 3},
		want:       "1:2:3",
	}, {
		name:       "包含0的情况",
		sequenceID: [3]uint32{0, 4, 5},
		want:       "0:4:5",
	}, {
		name:       "较大数值的情况",
		sequenceID: [3]uint32{9999, 8888, 7777},
		want:       "9999:8888:7777",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SequenceIDString(tt.sequenceID)
			assert.Equalf(t, tt.want, got, "SequenceIDString(%v)", tt.sequenceID)
		})
	}
}

// TestSequenceIDFromString 测试SequenceIDFromString函数
func TestSequenceIDFromString(t *testing.T) {
	tests := []struct {
		name       string
		sequenceID string
		want       [3]uint32
	}{{
		name:       "正常格式",
		sequenceID: "1:2:3",
		want:       [3]uint32{1, 2, 3},
	}, {
		name:       "包含0的格式",
		sequenceID: "0:4:5",
		want:       [3]uint32{0, 4, 5},
	}, {
		name:       "较大数值的格式",
		sequenceID: "9999:8888:7777",
		want:       [3]uint32{9999, 8888, 7777},
	}, {
		name:       "格式错误 - 返回默认值",
		sequenceID: "invalid",
		want:       [3]uint32{0, 0, 0},
	}, {
		name:       "格式不完整 - 返回默认值",
		sequenceID: "1:2",
		want:       [3]uint32{0, 0, 0},
	}, {
		name:       "多余的冒号 - 返回默认值",
		sequenceID: "1:2:3:4",
		want:       [3]uint32{0, 0, 0},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SequenceIDFromString(tt.sequenceID)
			assert.Equalf(t, tt.want, got, "SequenceIDFromString(%v)", tt.sequenceID)
		})
	}
}

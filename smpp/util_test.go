package smpp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToValidatePeriod1(t *testing.T) {
	parseTime := func(t *testing.T, s string) time.Time {
		tt, err := time.Parse("2006-01-02 15:04:05", s)
		if err != nil {
			t.Fatal(err)
		}
		return tt
	}

	t.Run("test_timeToSMPPTimeFormatRelativce", func(t *testing.T) {
		t.Log(timeToSMPPTimeFormatRelativce(time.Second * 5))
		tests := []struct {
			name string
			args time.Duration
			want string
		}{
			{name: "秒级", args: time.Second * 10, want: "000000000010000R"},
			{name: "分钟", args: time.Minute * 10, want: "000000001000000R"},
			{name: "小时", args: time.Hour * 10, want: "000000100000000R"},
			{name: "天", args: time.Hour * 24 * 10, want: "000010000000000R"},
			{name: "月，不起作用", args: time.Hour * 24 * 31 * 10, want: ""},
			{name: "年，不起作用", args: time.Hour * 24 * 31 * 12 * 10, want: ""},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := timeToSMPPTimeFormatRelativce(tt.args); got != tt.want {
					t.Errorf("timeToSMPPTimeFormatRelativce() = %v, want %v", got, tt.want)
				}
			})
		}
	})
	t.Run("absolute", func(t *testing.T) {
		now := parseTime(t, "2023-04-19 18:38:25")
		tests := []struct {
			name string
			args time.Duration
			want string
		}{
			{name: "秒级", args: time.Second * 10, want: "230419183835000+"},
			{name: "分钟", args: time.Minute * 10, want: "230419184825000+"},
			{name: "分钟，跨小时", args: time.Minute * 40, want: "230419191825000+"},
			{name: "小时", args: time.Hour * 10, want: "230420043825000+"},
			{name: "天", args: now.AddDate(0, 0, 1).Sub(now), want: "230420183825000+"},
			{name: "月", args: now.AddDate(0, 1, 0).Sub(now), want: "230519183825000+"},
			{name: "年", args: now.AddDate(1, 0, 0).Sub(now), want: "240419183825000+"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := timeToSMPPTimeFormatAbsolute(now, now.Add(tt.args)); got != tt.want {
					t.Errorf("timeToSMPPTimeFormatAbsolute() = %v, want %v", got, tt.want)
				}
			})
		}
	})
	t.Run("case1", func(t *testing.T) {
		for _, item := range []struct {
			now         string
			period      string
			isRelative  bool
			expectError bool
			expectValue string
		}{
			{now: "2023-04-18 15:46:30", period: "-15s", isRelative: true, expectError: true, expectValue: "invalid target date"},
			{now: "2023-04-18 15:46:30", period: "15ahaha", isRelative: true, expectError: true, expectValue: "parse duration error"},
			{now: "2023-04-18 15:46:30", period: "15s", isRelative: true, expectError: false, expectValue: "000000000015000R"},
		} {
			now := parseTime(t, item.now)
			res, err := ToValidatePeriod(now, item.period, item.isRelative)
			if item.expectError {
				assert.ErrorContains(t, err, item.expectValue)
				continue
			}
			assert.Nil(t, err)
			assert.Equal(t, item.expectValue, res)
		}
	})

	t.Run("demo", func(t *testing.T) {
	})
}

func Test_isDigest(t *testing.T) {
	for _, item := range []struct {
		s      string
		expect bool
	}{
		{s: "12273784", expect: true},
		{s: "245d", expect: false},
		{s: "18337381198", expect: true},
		{s: "82082", expect: true},
		{s: "Lark", expect: false},
		{s: "+1234567890", expect: false},
	} {
		if v := isDigit(item.s); v != item.expect {
			t.Fatalf("%+v failed", item)
		}
	}
}

func TestGenerateSourceAddress(t *testing.T) {
	// - 纯数字(包括手机号 或 10DLC)，长码(长度>=10) 用 1/1，短码(长度<10) 用 3/0；
	// - 带有字母，用 5/0
	testCases := []struct {
		addr         string
		expectedTON  int
		expectedNPI  int
		expectedAddr string
	}{
		{"1234567890", 1, 1, "1234567890"},   // 纯数字，长码，1/1
		{"12345", 3, 0, "12345"},             // 纯数字，短码，3/0
		{"abc123", 5, 0, "abc123"},           // 带字母，5/0
		{"+1234567890", 5, 0, "+1234567890"}, // 非纯数字, 5/0
		{"+12345", 5, 0, "+12345"},           // 纯数字，5/0
		{"18552949988", 1, 1, "18552949988"}, // 10DLC，纯数字，长码，5/0
		{"15706612302", 1, 1, "15706612302"}, // 10DLC，纯数字，长码，5/0
		{"1234567890", 1, 1, "1234567890"},   // 10DLC，纯数字，长码，5/0
		{"82082", 3, 0, "82082"},             // 纯数字，短码，3/0
		{"10690055", 3, 0, "10690055"},       // 纯数字，短码，3/0
	}

	for _, testCase := range testCases {
		ton, npi, addr := GenerateSourceAddress(testCase.addr)

		if ton != testCase.expectedTON {
			t.Errorf("Expected TON %d, but got %d for address %s", testCase.expectedTON, ton, testCase.addr)
		}
		if npi != testCase.expectedNPI {
			t.Errorf("Expected NPI %d, but got %d for address %s", testCase.expectedNPI, npi, testCase.addr)
		}
		if addr != testCase.expectedAddr {
			t.Errorf("Expected address %s, but got %s", testCase.expectedAddr, addr)
		}
	}
}

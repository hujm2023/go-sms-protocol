package smpp34

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractDeliveryReceipt(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		s := "id:1235 sub:001 dlvrd:1 submit date:2107081716 done date:2210131801 stat:DELIVRD err:0 text:aha"
		d, err := ExtractDeliveryReceipt(s)
		assert.Nil(t, err)
		expectDeliveryReceipt := DeliveryReceipt{
			ID:       "1235",
			Sub:      "001",
			Dlvrd:    "1",
			SubDate:  "2107081716",
			DoneDate: "2210131801",
			Stat:     "DELIVRD",
			Err:      "0",
			Text:     "aha",
		}
		assert.Equal(t, expectDeliveryReceipt, d)
		t.Logf("%+v", d)
	})
	t.Run("参数错位", func(t *testing.T) {
		s := "id:1235 sub:001 dlvrd:1 done date:2210131801 submit date:2107081716 stat:DELIVRD err:0 text:aha"
		d, err := ExtractDeliveryReceipt(s)
		assert.Nil(t, err)
		expectDeliveryReceipt := DeliveryReceipt{
			ID:       "1235",
			Sub:      "001",
			Dlvrd:    "1",
			SubDate:  "2107081716",
			DoneDate: "2210131801",
			Stat:     "DELIVRD",
			Err:      "0",
			Text:     "aha",
		}
		assert.Equal(t, expectDeliveryReceipt, d)
		t.Logf("%+v", d)
	})
	t.Run("empty source", func(t *testing.T) {
		d, err := ExtractDeliveryReceipt("")
		assert.Nil(t, err)
		assert.Equal(t, DeliveryReceipt{}, d)
	})
	t.Run("demo", func(t *testing.T) {
		d := ""
		dd, err := ExtractDeliveryReceipt(d)
		t.Log(err)
		t.Log(dd)
	})
}

func TestFindSubValue(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		s := "id:1235 sub:001 dlvrd:1 submit date:2107081716 done date:2210131801 stat:DELIVRD err:0 text:"
		for _, item := range []struct {
			key         string
			expectValue string
			maxSize     int
		}{
			{key: "id", expectValue: "1235", maxSize: 10},
			{key: "sub", expectValue: "001", maxSize: 3},
			{key: "dlvrd", expectValue: "1", maxSize: 3},
			{key: "submit date", expectValue: "2107081716", maxSize: 10},
			{key: "done date", expectValue: "2210131801", maxSize: 10},
			{key: "stat", expectValue: "DELIVRD", maxSize: 7},
			{key: "err", expectValue: "0", maxSize: 3},
			{key: "text", expectValue: "", maxSize: 20},
		} {
			v := findSubValue(s, item.key, item.maxSize)
			assert.Equal(t, item.expectValue, v)
		}
	})
	t.Run("顺序不对", func(t *testing.T) {
		s := "id:1235 submit date:2107081716 sub:001 dlvrd:1 done date:2210131801 stat:DELIVRD err:0 text:"
		for _, item := range []struct {
			key         string
			maxSize     int
			expectValue string
		}{
			{key: "id", expectValue: "1235", maxSize: 10},
			{key: "sub", expectValue: "001", maxSize: 3},
			{key: "dlvrd", expectValue: "1", maxSize: 3},
			{key: "submit date", expectValue: "2107081716", maxSize: 10},
			{key: "done date", expectValue: "2210131801", maxSize: 10},
			{key: "stat", expectValue: "DELIVRD", maxSize: 7},
			{key: "err", expectValue: "0", maxSize: 3},
			{key: "text", expectValue: "", maxSize: 20},
		} {
			v := findSubValue(s, item.key, item.maxSize)
			assert.Equal(t, item.expectValue, v)
		}
	})
	t.Run("少一个key", func(t *testing.T) {
		s := "id:1235 submit date:2107081716 dlvrd:1 done date:2210131801 stat:DELIVRD err:0 text:"
		for _, item := range []struct {
			key         string
			maxSize     int
			expectValue string
		}{
			{key: "id", expectValue: "1235", maxSize: 10},
			{key: "sub", expectValue: "", maxSize: 3},
			{key: "dlvrd", expectValue: "1", maxSize: 3},
			{key: "submit date", expectValue: "2107081716", maxSize: 10},
			{key: "done date", expectValue: "2210131801", maxSize: 10},
			{key: "stat", expectValue: "DELIVRD", maxSize: 7},
			{key: "err", expectValue: "0", maxSize: 3},
			{key: "text", expectValue: "", maxSize: 20},
		} {
			v := findSubValue(s, item.key, item.maxSize)
			assert.Equal(t, item.expectValue, v)
		}
	})
	t.Run("中间的 key 为空", func(t *testing.T) {
		s := "id:1235 submit date:2107081716 sub: dlvrd:1 done date:2210131801 stat:DELIVRD err:0"
		for _, item := range []struct {
			key         string
			maxSize     int
			expectValue string
		}{
			{key: "id", expectValue: "1235", maxSize: 10},
			{key: "sub", expectValue: "", maxSize: 3},
			{key: "dlvrd", expectValue: "1", maxSize: 3},
			{key: "submit date", expectValue: "2107081716", maxSize: 10},
			{key: "done date", expectValue: "2210131801", maxSize: 10},
			{key: "stat", expectValue: "DELIVRD", maxSize: 7},
			{key: "err", expectValue: "0", maxSize: 3},
			{key: "text", expectValue: "", maxSize: 20},
		} {
			v := findSubValue(s, item.key, item.maxSize)
			assert.Equal(t, item.expectValue, v)
		}
	})
}

var _s = "id:1235 sub:001 dlvrd:1 submit date:2107081716 done date:2210131801 stat:DELIVRD err:0 text:aha"

func BenchmarkExtractDeliveryReceipt(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractDeliveryReceipt(_s)
	}
	b.StopTimer()
}

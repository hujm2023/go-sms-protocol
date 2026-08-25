package protocol

import "testing"

func TestParseLongSmsContentSixAndSevenByteHeaders(t *testing.T) {
	for _, tt := range []struct {
		name      string
		content   []byte
		wantKey   int
		wantTotal int
		wantIndex int
		wantBody  []byte
	}{
		{
			name:      "six-byte header",
			content:   []byte{0x05, 0x00, 0x03, 0x7f, 0x02, 0x01, 'a', 0x00},
			wantKey:   0x7f,
			wantTotal: 2,
			wantIndex: 1,
			wantBody:  []byte{'a', 0x00},
		},
		{
			name:      "seven-byte header uses network-order reference",
			content:   []byte{0x06, 0x08, 0x04, 0x12, 0x34, 0x02, 0x01, 'b', 0x00},
			wantKey:   0x1234,
			wantTotal: 2,
			wantIndex: 1,
			wantBody:  []byte{'b', 0x00},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			key, total, index, body, valid := ParseLongSmsContentBytes(tt.content)
			if !valid {
				t.Fatal("ParseLongSmsContentBytes() marked a valid header invalid")
			}
			if key != tt.wantKey || total != tt.wantTotal || index != tt.wantIndex {
				t.Fatalf("parsed header = (%d, %d, %d), want (%d, %d, %d)", key, total, index, tt.wantKey, tt.wantTotal, tt.wantIndex)
			}
			if string(body) != string(tt.wantBody) {
				t.Fatalf("parsed body = %x, want %x", body, tt.wantBody)
			}

			stringKey, stringTotal, stringIndex, stringBody, stringValid := ParseLongSmsContent(string(tt.content))
			if stringKey != key || stringTotal != total || stringIndex != index || stringValid != valid || string(stringBody) != string(body) {
				t.Fatalf("string/bytes parsing differs: string=(%d, %d, %d, %q, %t), bytes=(%d, %d, %d, %q, %t)",
					stringKey, stringTotal, stringIndex, stringBody, stringValid, key, total, index, body, valid)
			}
		})
	}
}

func TestParseLongSmsContentPreservesInvalidInput(t *testing.T) {
	for _, content := range [][]byte{
		{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		{0x05, 0x00, 0x03, 0x01, 0x01},
	} {
		_, _, _, body, valid := ParseLongSmsContentBytes(content)
		if valid {
			t.Fatalf("ParseLongSmsContentBytes(%x) unexpectedly valid", content)
		}
		if string(body) != string(content) {
			t.Fatalf("invalid input body = %x, want original %x", body, content)
		}
	}
}

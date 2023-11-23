package protocol

import (
	"context"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/datacoding"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
)

func TestSplitSMPPLongContent(t *testing.T) {
	ctx := context.Background()

	checkSplitContents := func(t *testing.T, contents [][]byte, maxLengthPerSeg int) {
		assert.True(t, len(contents) > 0)

		if len(contents) == 1 {
			assert.True(t, len(contents[0]) <= maxLengthPerSeg)
			return
		}
		for _, content := range contents {
			assert.True(t, len(content) > 6)
			assert.Equal(t, LongMsgHeader6ByteFrameKey, content[0])
			assert.Equal(t, LongMsgHeader6ByteFrameTotal, content[1])
			assert.Equal(t, LongMsgHeader6ByteFrameNum, content[2])
		}
	}

	t.Run("gsm", func(t *testing.T) {
		t.Run("no escape", func(t *testing.T) {
			t.Run("<=160", func(t *testing.T) {
				s := strings.Repeat("a", 159) + "b"
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, Default6FrameKey)
				assert.Nil(t, err)
				assert.Equal(t, msgFmt, datacoding.SMPP_CODING_GSM7_PACKED)
				assert.Equal(t, 1, len(contents))
				checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			})
			t.Run(">160", func(t *testing.T) {
				s := strings.Repeat("a", 161)
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, Default6FrameKey)
				assert.Nil(t, err)
				assert.Equal(t, msgFmt, datacoding.SMPP_CODING_GSM7_PACKED)
				assert.Equal(t, 2, len(contents))
				checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			})
		})
		t.Run("has escape", func(t *testing.T) {
			// `| ^ € { } [ ] ~ \`，扩展表中的字符，需要采用 2 个字节去编码，所以出现这几个字符，意味着不能 pack
			t.Run("总数不超过160", func(t *testing.T) {
				s := strings.Repeat("a", 158) + "]"
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, Default6FrameKey)
				assert.Nil(t, err)
				assert.Equal(t, datacoding.SMPP_CODING_GSM7_PACKED, msgFmt)
				assert.Equal(t, 1, len(contents))
				checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			})
			t.Run("含有escape字符，不能被gsm7编码", func(t *testing.T) {
				s := strings.Repeat("a", 140) + string(byte(0x1B)) + "a" // 142个字符，一个ucs2占两个字节，需切成ceil(142/70)=3条
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, Default6FrameKey)
				assert.Nil(t, err)
				assert.Equal(t, msgFmt, datacoding.SMPP_CODING_UCS2)
				assert.Equal(t, 3, len(contents))
				checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			})
			t.Run(">160", func(t *testing.T) {
				s := strings.Repeat("a", 161) + "{"
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, Default6FrameKey)
				assert.Nil(t, err)
				assert.Equal(t, msgFmt, datacoding.SMPP_CODING_GSM7_PACKED)
				assert.Equal(t, 2, len(contents))
				checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			})
			t.Run("只有7个字符", func(t *testing.T) {
				s := strings.Repeat("a", 7)
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, Default6FrameKey)
				assert.Nil(t, err)
				assert.Equal(t, msgFmt, datacoding.SMPP_CODING_GSM7_PACKED)
				assert.Equal(t, 1, len(contents))
				t.Log(contents)
				checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			})
		})
		t.Run("has non gsm", func(t *testing.T) {
			// 含有非gsm7编码，应该回退成 ucs2
			s := strings.Repeat("a", 120) + "你好"
			contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, Default6FrameKey)
			assert.Nil(t, err)
			assert.Equal(t, msgFmt, datacoding.SMPP_CODING_UCS2)
			assert.Equal(t, 2, len(contents)) // 全被当成了中文，122长度应该被切成两条
			checkSplitContents(t, contents, datacoding.MaxLongSmsLength)
		})
		t.Run("gsm7 unpacked,以160切片", func(t *testing.T) {
			s := strings.Repeat("a", 161) + "b"
			contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_UNPACKED, Default6FrameKey)
			assert.Nil(t, err)
			assert.Equal(t, msgFmt, datacoding.SMPP_CODING_GSM7_UNPACKED)
			assert.Equal(t, 2, len(contents))
			checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			assert.Equal(t, datacoding.MaxGSM7Length-1, len(contents[0])) // 切成两条，第一条长度为 159
		})
		t.Run("gsm7(packed)长短信末尾第一个字符是escape", func(t *testing.T) {
			content := strings.Repeat("a", 152) + "[" + "bbbbbbbbbbbbb"
			contents, actualCoding, err := EncodeSMPPContentAndSplit(ctx, content, datacoding.SMPP_CODING_GSM7_PACKED, Default6FrameKey)
			if err != nil {
				t.Fatal(err)
			}
			// 第一个 part 去掉长短信头的长度应该是 133 ((153-1) * 7 / 8) - 6)
			assert.Equal(t, 133, len(contents[0])-6)
			assert.Equal(t, datacoding.SMPP_CODING_GSM7_PACKED, actualCoding)
			res := ""
			for i := 0; i < len(contents); i++ {
				t.Log(contents[i])
				content, err := DecodeSMPPCContent(context.Background(), string(contents[i][6:]), actualCoding.ToInt())
				if err != nil {
					t.Fatal(err)
				}
				res += content
			}
			assert.Equal(t, content, res)
		})
	})
}

// 测试本系统的序列化反序列化和其他系统的对比
func TestSMPPSplitCompareWithOtherSystem(t *testing.T) {
	ctx := context.Background()
	frameKey := byte(0x12)

	contentBytes2HexSlice := func(contentBytes [][]byte) []string {
		res := make([]string, 0, len(contentBytes))
		for _, item := range contentBytes {
			res = append(res, strings.ToUpper(hex.EncodeToString(item)))
		}
		return res
	}

	hex2Contents := func(t *testing.T, src []string) [][]byte {
		res := make([][]byte, 0, len(src))
		for i := 0; i < len(src); i++ {
			data, err := hex.DecodeString(strings.ToLower(src[i]))
			if err != nil {
				t.Fatal(err)
			}
			res = append(res, data)
		}
		return res
	}

	compareHexSlice := func(t *testing.T, name string, a, b []string) {
		if len(a) != len(b) {
			t.Logf("a: %v\nb: %v\n", a, b)
			t.Fatalf("%s length not equal. len(a)=%d, len(b)=%d", name, len(a), len(b))
		}
		var failed bool
		for i := 0; i < len(a); i++ {
			if a[i] != b[i] {
				t.Logf("%s NOT EQUAL: \nshould be: %s\nbut got: %s\n", name, a[i], b[i])
				failed = true
			}
		}
		if failed {
			t.Fatal()
		}
	}

	compareSrcContent := func(t *testing.T, name string, a, b string) {
		if a != b {
			t.Fatalf("%s NOT EQUAL.  %s vs %s", name, strconv.Quote(a), strconv.Quote(b))
		}
	}

	decodeAndCombine := func(t *testing.T, dataCoding int, src []string) (string, error) {
		srcContents := hex2Contents(t, src)
		if len(srcContents) == 0 {
			t.Fatal("empty src")
		}
		if len(srcContents) == 1 {
			return DecodeSMPPCContent(context.Background(), string(srcContents[0]), dataCoding)
		}
		res := ""
		for i := 0; i < len(srcContents); i++ {
			content, err := DecodeSMPPCContent(context.Background(), string(srcContents[i][6:]), dataCoding)
			if err != nil {
				return "", err
			}
			res += content
		}
		return res, nil
	}

	for _, item := range []struct {
		dataCoding  datacoding.SMPPDataCoding
		content     string
		targetBytes []string
		isLong      bool
	}{
		// 短短信
		{
			dataCoding: datacoding.SMPP_CODING_ASCII,
			content:    "1234567ahifbewibaiownfe",
			targetBytes: []string{
				"3132333435363761686966626577696261696F776E6665",
			},
		},
		{
			dataCoding: datacoding.SMPP_CODING_GSM7_PACKED,
			content:    "1234567ahifbewibaiownfe", // 会出现\r
			targetBytes: []string{
				"31D98C56B3DDC2E8B4595CBEA7C5E1F4FBEE36971B",
			},
		},
		{
			dataCoding: datacoding.SMPP_CODING_GSM7_PACKED,
			content:    "1234567ahifbewibaiownfehahahahahahahahah",
			targetBytes: []string{
				"31D98C56B3DDC2E8B4595CBEA7C5E1F4FBEE3697D16174181D4687D16174181D4687D1",
			},
		},
		{
			dataCoding: datacoding.SMPP_CODING_GSM7_PACKED,
			content:    "1234567ahifbewibaiownf[e]h{a}hahahahahahahah",
			targetBytes: []string{
				"31D98C56B3DDC2E8B4595CBEA7C5E1F4FBEE366F78E58D0FBD4185372974181D4687D16174181D4687D1",
			},
		},
		{
			dataCoding: datacoding.SMPP_CODING_Latin1,
			content:    "1234567ahifbewibaiownfe",
			targetBytes: []string{
				"3132333435363761686966626577696261696F776E6665",
			},
		},
		{
			dataCoding: datacoding.SMPP_CODING_UCS2,
			content:    "【测试】这是一条测试短信",
			targetBytes: []string{
				"30106D4B8BD530118FD9662F4E0067616D4B8BD577ED4FE1",
			},
		},
		// 长短信
		{
			dataCoding: datacoding.SMPP_CODING_ASCII,
			content:    strings.Repeat("a", 164),
			targetBytes: []string{
				"0500031202016161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161616161",
				"050003120202616161616161616161616161616161616161616161616161616161616161",
			},
			isLong: true,
		},
		{
			dataCoding: datacoding.SMPP_CODING_GSM7_PACKED,
			content:    "[1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a][1234a]",
			targetBytes: []string{
				"0500031203011B5E4C36A38537BE0D2F269BD1C21BDF861793CD68E18D6FC38BC966B4F0C6B7E1C564335A78E3DBF062B2192DBCF16D7831D98C16DEF836BC986C460B6F7C1B5E4C36A38537BE0D2F269BD1C21BDF861793CD68E18D6FC38BC966B4F0C6B7E1C564335A78E3DBF062B2192DBCF16D7831D98C16DEF836BC986C460B6F7C1B5E4C36A385373E",
				"0500031203021B5E4C36A38537BE0D2F269BD1C21BDF861793CD68E18D6FC38BC966B4F0C6B7E1C564335A78E3DBF062B2192DBCF16D7831D98C16DEF836BC986C460B6F7C1B5E4C36A38537BE0D2F269BD1C21BDF861793CD68E18D6FC38BC966B4F0C6B7E1C564335A78E3DBF062B2192DBCF16D7831D98C16DEF836BC986C460B6F7C1B5E4C36A385373E",
				"0500031203031B5E4C36A38537BE0D2F269BD1C21BDF861793CD68E18D6FC38BC966B4F0C6B7E1C564335A78E3DBF062B2192DBCF16D7831D98C16DEF81A",
			},
			isLong: true,
		},
		{
			dataCoding: datacoding.SMPP_CODING_GSM7_UNPACKED,
			content:    "1234567ahifbewibaiownf[e]h{a}hahahahahahahah1234567ahifbewibaiownf[e]h{a}hahahahahahahah1234567ahifbewibaiownf[e]h{a}hahahahahahahah1234567ahifbewibaiownf[e]h{a}hahahahahahahah1234567ahifbewibaiownf[e]h{a}hahahahahahahah1234567ahifbewibaiownf[e]h{a}hahahahahahahah1234567ahifbewibaiownf[e]h{a}hahahahahahahah1234567ahifbewibaiownf[e]h{a}hahahahahahahah1234567ahifbewibaiownf[e]h{a}hahahahahahahah",
			targetBytes: []string{
				"0500031203013132333435363761686966626577696261696F776E661B3C651B3E681B28611B296861686168616861686168616861683132333435363761686966626577696261696F776E661B3C651B3E681B28611B296861686168616861686168616861683132333435363761686966626577696261696F776E661B3C651B3E681B28611B29686168616861686168616861686168313233343536376168",
				"0500031203026966626577696261696F776E661B3C651B3E681B28611B296861686168616861686168616861683132333435363761686966626577696261696F776E661B3C651B3E681B28611B296861686168616861686168616861683132333435363761686966626577696261696F776E661B3C651B3E681B28611B29686168616861686168616861686168313233343536376168696662657769626169",
				"0500031203036F776E661B3C651B3E681B28611B296861686168616861686168616861683132333435363761686966626577696261696F776E661B3C651B3E681B28611B296861686168616861686168616861683132333435363761686966626577696261696F776E661B3C651B3E681B28611B29686168616861686168616861686168",
			},
			isLong: true,
		},
	} {
		contents, actualDataCoding, err := EncodeSMPPContentAndSplit(ctx, item.content, item.dataCoding, frameKey)
		assert.Nil(t, err)
		assert.Equal(t, item.dataCoding, actualDataCoding)
		compareHexSlice(t, item.dataCoding.String()+" isLong: "+strconv.FormatBool(item.isLong), item.targetBytes, contentBytes2HexSlice(contents))

		content, err := decodeAndCombine(t, actualDataCoding.ToInt(), item.targetBytes)
		assert.Nil(t, err)
		compareSrcContent(t, item.dataCoding.String()+" isLong: "+strconv.FormatBool(item.isLong), item.content, content)

		contents, actualDataCoding1, err := NewBatchDataCodingEncoder().
			Protocol(SMPP).
			Content(item.content, frameKey).
			DataCodings([]datacoding.ProtocolDataCoding{item.dataCoding}).
			Build(ctx)
		assert.Nil(t, err)
		assert.Equal(t, item.dataCoding, actualDataCoding1)
		compareHexSlice(t, item.dataCoding.String()+" isLong: "+strconv.FormatBool(item.isLong), item.targetBytes, contentBytes2HexSlice(contents))

		content, err = decodeAndCombine(t, actualDataCoding1.ToInt(), item.targetBytes)
		assert.Nil(t, err)
		compareSrcContent(t, item.dataCoding.String()+" isLong: "+strconv.FormatBool(item.isLong), item.content, content)
	}
}

func TestGSM7Packed(t *testing.T) {
	// 1234567ahifbewibaimwnfe
	content := "1234567890abcdefghijklm"
	contents, _, err := EncodeSMPPContentAndSplit(context.Background(), content, datacoding.SMPP_CODING_GSM7_PACKED, byte(0x12))
	assert.Nil(t, err)
	t.Logf("编码前: %s, 长度: %d", strconv.Quote(content), len(content)) // 编码前: "1234567ahifbewibaimwnfe", 长度: 23
	t.Logf("编码后: %v, 长度: %d", strings.ToUpper(hex.EncodeToString(contents[0])), len(contents[0]))
	cc, err := DecodeSMPPCContent(context.Background(), string(contents[0]), datacoding.SMPP_CODING_GSM7_PACKED.ToInt())
	assert.Nil(t, err)
	t.Logf("解码后: %s，长度为：%d", strconv.Quote(cc), len(cc)) // 解码后: "1234567ahifbewibaimwnfe\r"，长度为：24
}

func TestA(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := (i * 8) + 1
		if v%7 == 0 {
			// t.Log(((i * 8) + 1) / 7)
			if v%153 == 0 {
				t.Logf("=====> %d", v)
			}
		}
	}
}

func TestEncodeSMPPContentAndSplit(t *testing.T) {
	content := "I'll compress the primary AGP bandwidth, that should bandwidth the SAS driver! Use the bluetooth GB port, then you can calculate the open-source application! I'll quantify the 1080p XSS hard drive, that should hard drive the XML transmitter!"
	contents, coding, err := EncodeSMPPContentAndSplit(context.Background(), content, datacoding.SMPP_CODING_GSM7_PACKED, 0x12)
	assert.Nil(t, err)
	assert.Equal(t, datacoding.SMPP_CODING_GSM7_PACKED, coding)
	assert.Equal(t, 2, len(contents))
	assert.Equal(t, "050003120201C9139B0D1ABFDB7079793E07D1D165105C9E6E87E57950F0080589C36EF23D4DA6A359203A3A4C07CDD1EF3A9B0C1287DDE47B9A4C4783E8E832681A9C82C8F2B4BD2C0F81AAF332888E2E83C4EC7A99FE7ED3D1A0A310047FCBE92C101D5D7683F2EF3A681C7683C661F6B8CE0ED3CB203ABA0C7AC3CBEED6FC5D978FCBA0301CCE4E8FC374", strings.ToUpper(hex.EncodeToString(contents[0])))
	assert.Equal(t, "050003120202E9B73B044A9ED86C50BC1E76D3D3E63C888E2E8362301C0C0EC24EA72074584E0691E5697B9905A2A3C374D01CFDAEB3C92074584E0691E5697B1944479741D82613449787DDF3769A4E2FCB43", strings.ToUpper(hex.EncodeToString(contents[1])))
}

func TestBBA(t *testing.T) {
	b := []byte{
		0, 0, 0, 212, 0, 0, 0, 4, 0, 0, 0, 0, 0, 2, 184, 78, 0, 5, 0, 66, 121, 116, 101, 80, 108, 117, 115, 0, 1, 1, 57, 53, 57, 54, 57, 57, 48, 50, 56, 52, 53, 52, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 159, 27, 60, 97, 99, 99, 99, 27, 62, 121, 111, 117, 114, 32, 118, 101, 114, 105, 102, 105, 99, 97, 116, 105, 111, 110, 32, 99, 111, 100, 101, 32, 105, 115, 32, 55, 54, 50, 49, 49, 32, 44, 32, 27, 40, 112, 108, 101, 97, 115, 101, 27, 41, 32, 105, 110, 112, 117, 116, 32, 105, 116, 32, 119, 105, 116, 104, 105, 110, 32, 53, 27, 47, 32, 109, 105, 110, 115, 27, 61, 32, 109, 105, 110, 115, 27, 20, 32, 97, 116, 116, 101, 110, 116, 105, 111, 110, 32, 105, 115, 32, 111, 107, 2, 44, 32, 112, 108, 115, 32, 27, 101, 32, 100, 111, 32, 110, 111, 116, 32, 102, 111, 114, 119, 97, 114, 100, 32, 116, 111, 32, 27, 64, 32, 116, 111, 32, 111, 116, 104, 101, 114, 115, 44, 100, 111, 0, 32, 110, 111, 116, 32, 102, 111, 114, 119, 97, 114, 100, 46,
	}
	t.Log(hex.EncodeToString(b))
}

func TestSubmitUnpack(t *testing.T) {
	b := []byte{
		0, 0, 0, 212, 0, 0, 0, 4, 0, 0, 0, 0, 0, 2, 184, 78, 0, 5, 0, 66, 121, 116, 101, 80, 108, 117, 115, 0, 1, 1, 57, 53, 57, 54, 57, 57, 48, 50, 56, 52, 53, 52, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 159, 27, 60, 97, 99, 99, 99, 27, 62, 121, 111, 117, 114, 32, 118, 101, 114, 105, 102, 105, 99, 97, 116, 105, 111, 110, 32, 99, 111, 100, 101, 32, 105, 115, 32, 55, 54, 50, 49, 49, 32, 44, 32, 27, 40, 112, 108, 101, 97, 115, 101, 27, 41, 32, 105, 110, 112, 117, 116, 32, 105, 116, 32, 119, 105, 116, 104, 105, 110, 32, 53, 27, 47, 32, 109, 105, 110, 115, 27, 61, 32, 109, 105, 110, 115, 27, 20, 32, 97, 116, 116, 101, 110, 116, 105, 111, 110, 32, 105, 115, 32, 111, 107, 2, 44, 32, 112, 108, 115, 32, 27, 101, 32, 100, 111, 32, 110, 111, 116, 32, 102, 111, 114, 119, 97, 114, 100, 32, 116, 111, 32, 27, 64, 32, 116, 111, 32, 111, 116, 104, 101, 114, 115, 44, 100, 111, 0, 32, 110, 111, 116, 32, 102, 111, 114, 119, 97, 114, 100, 46,
	}
	submit := new(smpp34.SubmitSm)
	assert.Nil(t, submit.IDecode(b))
	dc := submit.DataCoding
	shortMessage := submit.ShortMessage
	t.Log(dc, shortMessage)
	t.Log(DecodeSMPPCContent(context.Background(), string(shortMessage), int(dc)))
}

func TestDelivery(t *testing.T) {
	b := []byte{
		0, 0, 0, 191, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0, 0, 37, 0, 1, 1, 50, 49, 51, 55, 57, 56, 51, 50, 50, 54, 52, 51, 0, 5, 0, 84, 105, 107, 84, 111, 107, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 122, 105, 100, 58, 48, 54, 48, 52, 56, 48, 54, 57, 48, 55, 32, 115, 117, 98, 58, 48, 48, 49, 32, 100, 108, 118, 114, 100, 58, 48, 48, 49, 32, 115, 117, 98, 109, 105, 116, 32, 100, 97, 116, 101, 58, 50, 51, 48, 50, 50, 51, 48, 54, 51, 53, 32, 100, 111, 110, 101, 32, 100, 97, 116, 101, 58, 50, 51, 48, 50, 50, 51, 48, 54, 51, 53, 32, 115, 116, 97, 116, 58, 68, 69, 76, 73, 86, 82, 68, 32, 101, 114, 114, 58, 48, 48, 48, 32, 116, 101, 120, 116, 58, 91, 35, 93, 91, 84, 105, 107, 84, 111, 107, 93, 32, 49, 55, 49, 51, 32, 101, 115, 116, 4, 39, 0, 1, 2, 0, 30, 0, 9, 50, 52, 48, 67, 57, 69, 70, 66, 0,
	}
	pdu := new(smpp34.DeliverSm)
	assert.Nil(t, pdu.IDecode(b))
	shortMessage := pdu.ShortMessage
	delivery, err := smpp34.ExtractDeliveryReceipt(string(shortMessage))
	assert.Nil(t, err)
	t.Log(delivery)
	t.Log(delivery.ID)
}

func TestBind(t *testing.T) {
	b := []byte{0, 0, 0, 38, 0, 0, 0, 9, 0, 0, 0, 0, 0, 30, 166, 148, 84, 105, 107, 84, 111, 107, 67, 79, 98, 0, 80, 51, 56, 113, 97, 53, 0, 0, 52, 0, 0, 0}
	bb := new(smpp34.Bind)
	assert.Nil(t, bb.IDecode(b))
	t.Log(bb)
	t.Log(hex.EncodeToString(b))
}

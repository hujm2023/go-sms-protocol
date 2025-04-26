package protocol

import (
	"bytes"
	"context"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/datacoding"
)

func TestSplitLongContent(t *testing.T) {
	content := "【德邦快递】尊敬客户您好，您反馈的问题将由我（工号：7142）负责专职跟进，问题目前正在处理中 在此期间，您不需要进行任何操作，有最新处理进展我们将及时联系您，请保持电话畅通。期间如有问题咨询可回拨95353热线1-4号键转工号7142，会由我来回复您，祝您生活愉快！"
	data, coding, err := EncodeCMPPContentAndSplit(context.Background(), content, datacoding.CMPP_CODING_UCS2, 123)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, datacoding.CMPP_CODING_UCS2, coding)
	assert.Equal(t, 2, len(data))
	for idx, cc := range data {
		assert.True(t, bytes.Equal([]byte{0x05, 0x00, 0x03, 123, 2, byte(idx + 1)}, []byte(cc[:6])))
	}
}

func TestCeil(t *testing.T) {
	for _, i := range []struct {
		total  int
		split  int
		result int
	}{
		{total: 142, split: 134, result: 2},
		{total: 268, split: 134, result: 2},
		{total: 161, split: 153, result: 2},
		{total: 120, split: 134, result: 1},
		{total: 1, split: 134, result: 1},
	} {
		if v := ceil(i.total, i.split); v != i.result {
			t.Fatalf("[ceil] failed. total: %d, splitWithUDHI: %d, want: %d, actial: %d", i.total, i.split, i.result, v)
		}
	}
}

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
			assert.Equal(t, longMsgHeader6ByteFrameKey, content[0])
			assert.Equal(t, longMsgHeader6ByteFrameTotal, content[1])
			assert.Equal(t, longMsgHeader6ByteFrameNum, content[2])
		}
	}

	t.Run("gsm", func(t *testing.T) {
		t.Run("no escape", func(t *testing.T) {
			t.Run("<=160", func(t *testing.T) {
				s := strings.Repeat("a", 159) + "b"
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, default6FrameKey)
				assert.Nil(t, err)
				assert.Equal(t, msgFmt, datacoding.SMPP_CODING_GSM7_PACKED)
				assert.Equal(t, 1, len(contents))
				checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			})
			t.Run(">160", func(t *testing.T) {
				s := strings.Repeat("a", 161)
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, default6FrameKey)
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
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, default6FrameKey)
				assert.Nil(t, err)
				assert.Equal(t, datacoding.SMPP_CODING_GSM7_PACKED, msgFmt)
				assert.Equal(t, 1, len(contents))
				checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			})
			t.Run("含有escape字符，不能被gsm7编码", func(t *testing.T) {
				s := strings.Repeat("a", 140) + string(byte(0x1B)) + "a" // 142个字符，一个ucs2占两个字节，需切成ceil(142/70)=3条
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, default6FrameKey)
				assert.Nil(t, err)
				assert.Equal(t, msgFmt, datacoding.SMPP_CODING_UCS2)
				assert.Equal(t, 3, len(contents))
				checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			})
			t.Run(">160", func(t *testing.T) {
				s := strings.Repeat("a", 161) + "{"
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, default6FrameKey)
				assert.Nil(t, err)
				assert.Equal(t, msgFmt, datacoding.SMPP_CODING_GSM7_PACKED)
				assert.Equal(t, 2, len(contents))
				checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			})
			t.Run("只有7个字符", func(t *testing.T) {
				s := strings.Repeat("a", 7)
				contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, default6FrameKey)
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
			contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_PACKED, default6FrameKey)
			assert.Nil(t, err)
			assert.Equal(t, msgFmt, datacoding.SMPP_CODING_UCS2)
			assert.Equal(t, 2, len(contents)) // 全被当成了中文，122长度应该被切成两条
			checkSplitContents(t, contents, datacoding.MaxLongSmsLength)
		})
		t.Run("gsm7 unpacked,以160切片", func(t *testing.T) {
			s := strings.Repeat("a", 161) + "b"
			contents, msgFmt, err := EncodeSMPPContentAndSplit(ctx, s, datacoding.SMPP_CODING_GSM7_UNPACKED, default6FrameKey)
			assert.Nil(t, err)
			assert.Equal(t, msgFmt, datacoding.SMPP_CODING_GSM7_UNPACKED)
			assert.Equal(t, 2, len(contents))
			checkSplitContents(t, contents, datacoding.MaxGSM7Length)
			assert.Equal(t, datacoding.MaxGSM7Length-1, len(contents[0])) // 切成两条，第一条长度为 159
		})
		t.Run("gsm7(packed)长短信末尾第一个字符是escape", func(t *testing.T) {
			content := strings.Repeat("a", 152) + "[" + "bbbbbbbbbbbbb"
			contents, actualCoding, err := EncodeSMPPContentAndSplit(ctx, content, datacoding.SMPP_CODING_GSM7_PACKED, default6FrameKey)
			if err != nil {
				t.Fatal(err)
			}
			// 第一个 part 去掉长短信头的长度应该是 133 ((153-1) * 7 / 8) - 6)
			assert.Equal(t, 133, len(contents[0])-6)
			assert.Equal(t, datacoding.SMPP_CODING_GSM7_PACKED, actualCoding)
			res := ""
			for i := 0; i < len(contents); i++ {
				t.Log(contents[i])
				c, err := DecodeSMPPCContent(context.Background(), string(contents[i][6:]), actualCoding.ToInt())
				if err != nil {
					t.Fatal(err)
				}
				res += c
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

func TestEncodeSMPPContentAndSplit(t *testing.T) {
	content := "I'll compress the primary AGP bandwidth, that should bandwidth the SAS driver! Use the bluetooth GB port, then you can calculate the open-source application! I'll quantify the 1080p XSS hard drive, that should hard drive the XML transmitter!"
	contents, coding, err := EncodeSMPPContentAndSplit(context.Background(), content, datacoding.SMPP_CODING_GSM7_PACKED, 0x12)
	assert.Nil(t, err)
	assert.Equal(t, datacoding.SMPP_CODING_GSM7_PACKED, coding)
	assert.Equal(t, 2, len(contents))
	assert.Equal(t, "050003120201C9139B0D1ABFDB7079793E07D1D165105C9E6E87E57950F0080589C36EF23D4DA6A359203A3A4C07CDD1EF3A9B0C1287DDE47B9A4C4783E8E832681A9C82C8F2B4BD2C0F81AAF332888E2E83C4EC7A99FE7ED3D1A0A310047FCBE92C101D5D7683F2EF3A681C7683C661F6B8CE0ED3CB203ABA0C7AC3CBEED6FC5D978FCBA0301CCE4E8FC374", strings.ToUpper(hex.EncodeToString(contents[0])))
	assert.Equal(t, "050003120202E9B73B044A9ED86C50BC1E76D3D3E63C888E2E8362301C0C0EC24EA72074584E0691E5697B9905A2A3C374D01CFDAEB3C92074584E0691E5697B1944479741D82613449787DDF3769A4E2FCB43", strings.ToUpper(hex.EncodeToString(contents[1])))
}

package protocol

import (
	"context"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/datacoding"
)

func TestEncodeSMPPContentAndSplitBatch(t *testing.T) {
	printData := func(t *testing.T, d [][]byte) {
		for i, dd := range d {
			t.Logf("==> part %d, length: %d, data: %+v\n", i+1, len(dd), dd)
		}
	}
	ctx := context.Background()
	t.Run("case1", func(t *testing.T) {
		dataCoding := []datacoding.ProtocolDataCoding{
			datacoding.SMPP_CODING_GSM7_UNPACKED,
			datacoding.SMPP_CODING_Latin1,
			datacoding.SMPP_CODING_UCS2,
			datacoding.SMPP_CODING_ASCII,
		}
		for _, item := range []struct {
			protocol           Protocol
			dataCodings        []datacoding.ProtocolDataCoding
			content            string
			expectError        error
			expectedDataCoding datacoding.ProtocolDataCoding
			expectParts        int
			printData          bool
		}{
			{
				protocol:           SMPP,
				content:            strings.Repeat("a", 158), // 全都是 ascii，应该能用 gsm7_packed 一条发完
				dataCodings:        dataCoding,
				expectError:        nil,
				expectedDataCoding: datacoding.SMPP_CODING_GSM7_UNPACKED,
				expectParts:        1,
			},
			{
				protocol:           SMPP,
				content:            strings.Repeat("a", 138) + "Á", // 含有 Latin1 字符，不能用 gsm7，但能被Latin1一条送走
				dataCodings:        dataCoding,
				expectError:        nil,
				expectedDataCoding: datacoding.SMPP_CODING_Latin1,
				expectParts:        1,
			},
			{
				protocol:           SMPP,
				content:            strings.Repeat("a", 138) + "啊哈", // 含有汉字，只能用 ucs2，按照 70(实际是67)去切分(一个字符占 2 字节)，且 3 条
				dataCodings:        dataCoding,
				expectError:        nil,
				expectedDataCoding: datacoding.SMPP_CODING_UCS2,
				expectParts:        3,
			},
			{
				protocol: CMPP,
				// 含有汉字，能用gbk和ucs2，但是ucs2固定2字节，只能切按照 70 去切;
				// gbk 是变长的，1 2 4 字节不等，前面的"a"只占1字节，后面的两个汉字各占2字节，按照 140 去切。
				// 所以此处会选择 GBK，编码后共 142 字节，按照 140 去切，两条，最后一条内容长度为 8 个字节；按ucs2去切，会切 3 条
				content: strings.Repeat("a", 138) + "啊哈",
				dataCodings: []datacoding.ProtocolDataCoding{
					datacoding.CMPP_CODING_GBK,
					datacoding.CMPP_CODING_UCS2,
				},
				expectError:        nil,
				expectedDataCoding: datacoding.CMPP_CODING_GBK,
				expectParts:        2,
				printData:          true,
			},
		} {
			data, f, err := NewBatchDataCodingEncoder().
				Protocol(item.protocol).
				Content(item.content, byte(123)).
				DataCodings(item.dataCodings).
				Build(ctx)
			assert.Equal(t, err, item.expectError)
			assert.Equal(t, f, item.expectedDataCoding)
			if item.expectParts != len(data) {
				t.Log(data)
				t.Fatalf("expected parts: %d, actual parts: %d", item.expectParts, len(data))
			}
			if item.printData {
				printData(t, data)
			}
		}
	})

	t.Run("Latin1", func(t *testing.T) {
		content := strings.Repeat("a", 135) + "Á"
		dataCoding := []datacoding.ProtocolDataCoding{
			datacoding.SMPP_CODING_GSM7_UNPACKED,
			datacoding.SMPP_CODING_Latin1,
			datacoding.SMPP_CODING_UCS2,
		}
		data, f, err := NewBatchDataCodingEncoder().
			Protocol(SMPP).
			Content(content, byte(123)).
			DataCodings(dataCoding).
			Build(ctx)
		assert.Nil(t, err)
		// gsm7不能编码 Á，所以排除。latin1按照 140 去切分，切 1 条；ucs2 按照 70(实际67) 去切，切 3 条
		// 最终选择切 1 条的 Latin1
		assert.Equal(t, datacoding.SMPP_CODING_Latin1, f)
		assert.Equal(t, 1, len(data))
	})

	t.Run("online", func(t *testing.T) {
		t.Run("smpp", func(t *testing.T) {
			content := "[#][TikTok] 123456 adalah kode verifikasi Anda\nDaewVlZQ+ns"
			dataCodings := []datacoding.ProtocolDataCoding{
				datacoding.SMPP_CODING_GSM7_UNPACKED,
				datacoding.SMPP_CODING_UCS2,
			}
			data, f, err := NewBatchDataCodingEncoder().
				Protocol(SMPP).
				Content(content, byte(123)).
				DataCodings(dataCodings).
				Build(ctx)
			assert.Nil(t, err)
			// 都能切，且长度都是 1，选择优先级最高的 ucs2
			assert.Equal(t, datacoding.SMPP_CODING_UCS2, f)
			assert.Equal(t, 1, len(data))
			t.Log(len(content))
			t.Log(f)
			t.Log(data)
		})

		t.Run("长度在70~160之间", func(t *testing.T) {
			content := "your verify code is 123456.your verify code is 123456.your verify code is 123456."
			dataCodings := []datacoding.ProtocolDataCoding{
				datacoding.SMPP_CODING_GSM7_UNPACKED,
				datacoding.SMPP_CODING_UCS2,
			}
			data, f, err := NewBatchDataCodingEncoder().
				Protocol(SMPP).
				Content(content, byte(123)).
				DataCodings(dataCodings).
				Build(ctx)
			t.Log(len(content))
			assert.Nil(t, err)
			// 长度为81，gsm7_unpakced编码为 1 条，ucs2编码为 2 条，选择gsm7_unpacked
			assert.Equal(t, datacoding.SMPP_CODING_GSM7_UNPACKED, f)
			assert.Equal(t, 1, len(data))
		})
	})
	t.Run("原始编码设置", func(t *testing.T) {
		content := "your verify code is 123456.your verify code is 123456.your verify code is 123456."
		dataCodings := []datacoding.ProtocolDataCoding{
			datacoding.SMPP_CODING_GSM7_UNPACKED,
			datacoding.SMPP_CODING_UCS2,
		}
		data, f, err := NewBatchDataCodingEncoder().
			Protocol(SMPP).
			Content(content, byte(123)).
			DataCodings(dataCodings).
			OriginDataCoding(datacoding.SMPP_CODING_UCS2).
			Build(ctx)
		t.Log(len(content))
		assert.Nil(t, err)
		// 长度为81，gsm7_unpakced编码为 1 条，ucs2编码为 2 条，选择gsm7_unpacked
		assert.Equal(t, datacoding.SMPP_CODING_GSM7_UNPACKED, f)
		assert.Equal(t, 1, len(data))
	})
	t.Run("ucs2兜底", func(t *testing.T) {
		content := "your verify code is 123456.your verify code is 123456.your verify code is 123456." + "啊哈"
		dataCodings := []datacoding.ProtocolDataCoding{
			datacoding.SMPP_CODING_GSM7_UNPACKED,
			datacoding.SMPP_CODING_Latin1,
		}
		data, f, err := NewBatchDataCodingEncoder().
			Protocol(SMPP).
			Content(content, byte(123)).
			DataCodings(dataCodings).
			Build(ctx)
		t.Log(len(content))
		assert.Nil(t, err)
		// 长度为83个字符，gsm7 和 latin1 无法编码，使用 ucs2 兜底。且两条
		assert.Equal(t, datacoding.SMPP_CODING_UCS2, f)
		assert.Equal(t, 2, len(data))
	})
}

func BenchmarkEncodeSMPPContentAndSplitBatch(b *testing.B) {
	content := strings.Repeat("a", 135) + "Á" + "啊"
	dataCoding := []datacoding.ProtocolDataCoding{
		datacoding.SMPP_CODING_GSM7_UNPACKED,
		datacoding.SMPP_CODING_Latin1,
		datacoding.SMPP_CODING_UCS2,
	}
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _, _ = NewBatchDataCodingEncoder().
			Protocol(SMPP).
			Content(content, byte(123)).
			DataCodings(dataCoding).
			Build(ctx)
	}
}

func TestBatchEncoderSorter(t *testing.T) {
	randomData := func(n int) []byte {
		buf := make([]byte, n)
		_, _ = rand.Read(buf)
		return buf
	}
	makeData := func(n int, per int) [][]byte {
		if n <= 0 {
			return [][]byte{randomData(per)}
		}
		res := make([][]byte, 0, n)
		for i := 0; i < n; i++ {
			res = append(res, randomData(per))
		}
		return res
	}

	t.Run("cmpp", func(t *testing.T) {
		t.Run("长度一致", func(t *testing.T) {
			encoders := []*encoder{
				{
					msgFmt: datacoding.CMPP_CODING_GBK,
					data:   makeData(1, 120),
				},
				{
					msgFmt: datacoding.CMPP_CODING_UCS2,
					data:   makeData(1, 120),
				},
			}
			encoderOrderBy(byLength, byDataCoding).Sort(encoders)
			// ucs2优先级大于gbk
			assert.Equal(t, datacoding.CMPP_CODING_UCS2, encoders[0].msgFmt)
		})
		t.Run("长度不一致", func(t *testing.T) {
			encoders := []*encoder{
				{
					msgFmt: datacoding.CMPP_CODING_GBK,
					data:   makeData(1, 69),
				},
				{
					msgFmt: datacoding.CMPP_CODING_UCS2,
					data:   makeData(2, 133),
				},
			}
			encoderOrderBy(byLength, byDataCoding).Sort(encoders)
			// 尽管ucs2优先级大于gbk，但 GBK 长度短，优先判断长度
			assert.Equal(t, datacoding.CMPP_CODING_GBK, encoders[0].msgFmt)
		})
	})

	t.Run("smpp", func(t *testing.T) {
		t.Run("长度一致", func(t *testing.T) {
			encoders := []*encoder{
				{
					msgFmt: datacoding.SMPP_CODING_GSM7_UNPACKED,
					data:   makeData(1, 120),
				},
				{
					msgFmt: datacoding.SMPP_CODING_UCS2,
					data:   makeData(1, 120),
				},
			}
			encoderOrderBy(byLength, byDataCoding).Sort(encoders)
			// ucs2优先级大于gsm7
			assert.Equal(t, datacoding.SMPP_CODING_UCS2, encoders[0].msgFmt)
		})
		t.Run("长度不一致", func(t *testing.T) {
			encoders := []*encoder{
				{
					msgFmt: datacoding.SMPP_CODING_GSM7_UNPACKED,
					data:   makeData(2, 200),
				},
				{
					msgFmt: datacoding.SMPP_CODING_UCS2,
					data:   makeData(1, 67),
				},
			}
			encoderOrderBy(byLength, byDataCoding).Sort(encoders)
			// ucs2优先级大于gsm7，但优先判断长度，所以选择 ucs2
			assert.Equal(t, datacoding.SMPP_CODING_UCS2, encoders[0].msgFmt)
		})
	})
}

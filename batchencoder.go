package protocol

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"golang.org/x/sync/errgroup"

	"github.com/samber/lo"

	"github.com/hujm2023/go-sms-protocol/datacoding"
	gsm7 "github.com/hujm2023/go-sms-protocol/datacoding/gsm7encoding"
	"github.com/hujm2023/go-sms-protocol/logger"
)

// Protocol represents the enumeration values for supported protocol types.
type Protocol string

const (
	CMPP Protocol = "CMPP"
	SMPP Protocol = "SMPP"
	SMGP Protocol = "SMGP"
	SGIP Protocol = "SGIP"
)

func (p Protocol) String() string {
	return string(p)
}

// BatchDataCodingEncoder ...
type BatchDataCodingEncoder struct {
	protocol         Protocol
	content          string
	dataCodings      []datacoding.ProtocolDataCoding
	originDataCoding datacoding.ProtocolDataCoding
	frameKey         byte
}

// NewBatchDataCodingEncoder ...
func NewBatchDataCodingEncoder() *BatchDataCodingEncoder {
	return new(BatchDataCodingEncoder)
}

// Protocol sets protocol for a BatchDataCodingEncoder.
func (b *BatchDataCodingEncoder) Protocol(p Protocol) *BatchDataCodingEncoder {
	b.protocol = p
	return b
}

// Content sets content and frameKey for a BatchDataCodingEncoder.
func (b *BatchDataCodingEncoder) Content(content string, frameKey byte) *BatchDataCodingEncoder {
	b.content = content
	b.frameKey = frameKey
	return b
}

// DataCodings sets dataCodings for a BatchDataCodingEncoder.
func (b *BatchDataCodingEncoder) DataCodings(d []datacoding.ProtocolDataCoding) *BatchDataCodingEncoder {
	b.dataCodings = d
	return b
}

// OriginDataCoding sets dataCodings for a BatchDataCodingEncoder.
func (b *BatchDataCodingEncoder) OriginDataCoding(d datacoding.ProtocolDataCoding) *BatchDataCodingEncoder {
	b.originDataCoding = d
	return b
}

func (b *BatchDataCodingEncoder) allDataCodings() map[datacoding.ProtocolDataCoding]struct{} {
	res := make(map[datacoding.ProtocolDataCoding]struct{}, len(b.dataCodings)+1)
	for _, msgFmt := range b.dataCodings {
		res[msgFmt] = struct{}{}
	}
	if datacoding.IsValidProtoDataCoding(b.originDataCoding) {
		res[b.originDataCoding] = struct{}{}
	}
	return res
}

func (b *BatchDataCodingEncoder) findOriginEncoder(encoders []*encoder) *encoder {
	if !datacoding.IsValidProtoDataCoding(b.originDataCoding) {
		return nil
	}

	return lo.FindOrElse(encoders, nil, func(item *encoder) bool {
		return item.msgFmt == b.originDataCoding
	})
}

// Build encodes the content according to dataCodings respectively. It returns the one which has the shortest segmented length.
//
// Please note that if there is no ucs2 in the incoming dataCodings, and none of the incoming dataCodings can encode,
// ucs2 will be used as a bottom-up encoding.
func (b *BatchDataCodingEncoder) Build(ctx context.Context) (contents [][]byte, actualMsgFmt datacoding.ProtocolDataCoding, err error) {
	if b == nil || b.content == "" || len(b.dataCodings) == 0 {
		return nil, nil, fmt.Errorf("invalid batch encoder builder")
	}

	var hasUcs2 bool
	encoders := make([]*encoder, 0, len(b.dataCodings)+1)
	eg := new(errgroup.Group)
	for msgFmt := range b.allDataCodings() {
		if msgFmt == datacoding.SMPP_CODING_UCS2 {
			hasUcs2 = true
		}
		encoder := newBatchEncoder(b.protocol, msgFmt, b.content, b.frameKey, datacoding.IsValidProtoDataCoding(b.originDataCoding) && msgFmt == b.originDataCoding)
		encoders = append(encoders, encoder)
		eg.Go(func() error {
			encoder.Run(ctx)
			return nil
		})
	}
	_ = eg.Wait()

	// 过滤掉不能编码的
	encoders = lo.Filter(encoders, func(encoder *encoder, _ int) bool {
		if !encoder.canEncode {
			return false
		}
		return true
	})

	// 都不能编码
	if len(encoders) == 0 {
		// 传入的配置中没有ucs2，则使用ucs2兜底
		if !hasUcs2 {
			logger.CtxInfo(ctx, "[EncodeSMPPContentAndSplitBatch] all dataCoding failed. use ucs2 as default")
			var defaultEncoder *encoder
			switch b.protocol {
			case CMPP:
				defaultEncoder = newBatchEncoder(b.protocol, datacoding.CMPP_CODING_UCS2, b.content, b.frameKey, datacoding.CMPP_CODING_UCS2 == b.originDataCoding)
			case SMPP:
				defaultEncoder = newBatchEncoder(b.protocol, datacoding.SMPP_CODING_UCS2, b.content, b.frameKey, datacoding.SMPP_CODING_UCS2 == b.originDataCoding)
			}
			if defaultEncoder != nil {
				defaultEncoder.Run(ctx)
				return defaultEncoder.Result()
			}
		}

		logger.CtxError(ctx, "[EncodeSMPPContentAndSplitBatch] all dataCoding failed")
		return nil, datacoding.UnknownProtocolDataCoding, fmt.Errorf("all dataCoding failed")
	}

	// 取切割条数最短的那个(按照长度->编码优先级排序)
	encoderOrderBy(byLength, byDataCoding).Sort(encoders)

	return encoders[0].Result()
}

type encoder struct {
	protocol       Protocol                      // protocol
	msgFmt         datacoding.ProtocolDataCoding // encoding
	content        string                        // source content
	frameKey       byte                          // frameKey
	isOriginMsgFmt bool                          // whether s it the msgFmt from the original configuration

	canEncode bool   // Can the current msgFmt be used for encoding?
	reason    string // Reasons for the inability to encode

	data [][]byte // Encoded content. Valid when canEncode is true.
}

func newBatchEncoder(protocol Protocol, msgFmt datacoding.ProtocolDataCoding, content string, frameKey byte, isOriginMsgFmt bool) *encoder {
	return &encoder{
		protocol:       protocol,
		msgFmt:         msgFmt,
		content:        content,
		frameKey:       frameKey,
		isOriginMsgFmt: isOriginMsgFmt,
	}
}

func (s *encoder) Name() string {
	return fmt.Sprintf("[Encoder:%s]", s.msgFmt.String())
}

func (s *encoder) Run(ctx context.Context) {
	var encoder datacoding.Codec
	switch s.protocol {
	case SMPP:
		encoder = datacoding.NewSMPPCodec(s.msgFmt.(datacoding.SMPPDataCoding), s.content)
	case CMPP:
		encoder = datacoding.NewCMPPCodec(s.msgFmt.(datacoding.CMPPDataCoding), s.content)
	}
	if encoder == nil {
		s.canEncode = false
		s.reason = fmt.Sprintf("unknown dataCoding: %d", s.msgFmt)
		return
	}

	// Special handling for GSM7 (packed)
	if s.msgFmt == datacoding.SMPP_CODING_GSM7_PACKED {
		if datacoding.CanEncodeByGSM7(s.content) {
			contents, _, err := encodeAndSplitGSM7Packed(s.content, s.frameKey)
			if err != nil {
				s.canEncode = false
				s.reason = fmt.Sprintf("%s encode error: %v", s.Name(), err)
				return
			}
			s.canEncode = true
			s.data = contents
		} else {
			s.canEncode = false
			s.reason = gsm7.ErrInvalidCharacter.Error()
		}
		return
	}

	encodedData, err := encoder.Encode()
	if err != nil {
		s.canEncode = false
		s.reason = fmt.Sprintf("%s encode error: %v", s.Name(), err)
		return
	}

	s.canEncode = true

	maxLongLength, perMsgLength := encoder.SplitBy()
	// short message
	if len(encodedData) <= maxLongLength {
		s.data = [][]byte{encodedData}
		return
	}

	s.data = splitWithUDHI(encodedData, perMsgLength, s.frameKey)
}

func (s *encoder) Result() (contents [][]byte, actualMsgFmt datacoding.ProtocolDataCoding, err error) {
	if s == nil {
		return nil, datacoding.UnknownProtocolDataCoding, fmt.Errorf("invalid encoder")
	}

	if !s.canEncode {
		return nil, datacoding.UnknownProtocolDataCoding, errors.New(lo.Ternary(len(s.reason) > 0, s.reason, "encode failed"))
	}
	return s.data, s.msgFmt, nil
}

// --------

type encoderCompareFunc func(p, q *encoder) bool

// byLength: Ordered by encoded length
func byLength(p, q *encoder) bool {
	// tips: the length of 'p' is smaller, so 'p' is used.
	return len(p.data) < len(q.data)
}

// byDataCoding: Ordered by dataCoding's priority
func byDataCoding(p, q *encoder) bool {
	// UCS2>GSM>latin1>other
	// Priority returns a numerical value, where a smaller value indicates a higher priority.
	// tips: When the Priority() of 'p' is smaller, it means it has a higher priority and will be used.
	return p.msgFmt.Priority() < q.msgFmt.Priority()
}

type batchEncoderSorter struct {
	// encoders
	encoders []*encoder
	// Specific function used for comparison. Compares elements one by one in array order.
	compareFuncs []encoderCompareFunc
}

func encoderOrderBy(by ...encoderCompareFunc) *batchEncoderSorter {
	return &batchEncoderSorter{compareFuncs: by}
}

// Sort sorts the encoders with `compareFuncs`.
func (b *batchEncoderSorter) Sort(encoders []*encoder) {
	b.encoders = encoders
	sort.Sort(b)
}

func (b *batchEncoderSorter) Len() int {
	return len(b.encoders)
}

func (b *batchEncoderSorter) Less(i, j int) bool {
	if len(b.encoders) == 0 || len(b.compareFuncs) == 0 {
		return false
	}

	// Performs comparisons one by one according to the array order
	// for the first `len(b.compareFuncs)-1` `compareFunc`
	var idx int
	for idx = 0; idx < len(b.compareFuncs)-1; idx++ {
		less := b.compareFuncs[idx]
		switch {
		case less(b.encoders[i], b.encoders[j]):
			return true
		case less(b.encoders[j], b.encoders[i]):
			return false
		}
		// For the current 'less', when the comparison conditions are the same, proceed to the next 'compareFunc'.
		continue
	}

	// When all preceding compareFuncs are equal, use the result of the last one as a fallback.
	return b.compareFuncs[idx](b.encoders[i], b.encoders[j])
}

func (b *batchEncoderSorter) Swap(i, j int) {
	b.encoders[i], b.encoders[j] = b.encoders[j], b.encoders[i]
}

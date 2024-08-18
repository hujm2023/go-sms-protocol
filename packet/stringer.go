package packet

import (
	"fmt"
	"strings"
	"sync"
)

type PDUStringer struct {
	e  error
	op string

	buf *strings.Builder
}

func NewPDUStringer() *PDUStringer {
	p := &PDUStringer{
		buf: borrorStringBuilder(),
	}
	_, _ = p.buf.WriteString("\n== Start ==\n")
	return p
}

func (p *PDUStringer) Release() {
	restoreStringBuilder(p.buf)
	p.e = nil
	p.op = ""
}

func (p *PDUStringer) composeKV(field string, v any, appendBytes bool) string {
	switch vv := v.(type) {
	case string:
		if appendBytes {
			return fmt.Sprintf(`key=%s, valueString="%s", valueBytes=%v`, field, vv, []byte(vv))
		} else {
			return fmt.Sprintf(`key=%s, valueString="%s"`, field, vv)
		}
	case byte, int, int16, int32, int64, uint, uint16, uint32, uint64:
		return fmt.Sprintf(`key=%s, valueNumber=%d`, field, vv)
	case float32, float64:
		return fmt.Sprintf(`key=%s, valueReal=%0.02f`, field, vv)
	case bool:
		return fmt.Sprintf(`key=%s, valueBoolean=%t`, field, vv)
	case []byte:
		return fmt.Sprintf(`key=%s, valueString="%s", valueBytes=%v`, field, string(vv), vv)
	case fmt.Stringer:
		s := vv.String()
		if appendBytes {
			return fmt.Sprintf(`key=%s, valueString="%s", valueBytes=%+v`, field, s, []byte(s))
		}
		return fmt.Sprintf(`key=%s, valueStringer=%s`, field, vv.String())
	default:
		return fmt.Sprintf(`key=%s, valueAny=%v`, field, vv)
	}
}

func (p *PDUStringer) writeString(s string) {
	if p.e != nil {
		return
	}
	_, err := p.buf.WriteString(s + "\n")
	if err != nil {
		p.e = err
		p.op = "write error for '%s'" + s
	}
}

func (p *PDUStringer) Write(field string, v any) {
	p.writeString(p.composeKV(field, v, false))
}

func (p *PDUStringer) WriteWithBytes(field string, v any) {
	p.writeString(p.composeKV(field, v, true))
}

func (p *PDUStringer) OmitWrite(k, v string) {
	if v == "" {
		return
	}
	p.writeString(p.composeKV(k, v, false))
}

func (p *PDUStringer) String() string {
	if p.e != nil {
		return fmt.Sprintf("<%s: %s>", p.op, p.e.Error())
	}
	_, _ = p.buf.WriteString("== End ==\n")
	return p.buf.String()
}

// -------------------------------

var stringBuilderPool = sync.Pool{New: func() any {
	sb := new(strings.Builder)
	sb.Grow(1800)
	return sb
}}

func borrorStringBuilder() *strings.Builder {
	return stringBuilderPool.Get().(*strings.Builder)
}

func restoreStringBuilder(sb *strings.Builder) {
	sb.Reset()
	stringBuilderPool.Put(sb)
}

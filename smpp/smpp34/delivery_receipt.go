package smpp34

import (
	"strings"

	"github.com/hujm2023/go-sms-protocol/packet"
)

var deliveryReceiptFields = []string{
	"id",
	"sub",
	"dlvrd",
	"submit date",
	"done date",
	"stat",
	"err",
	"text",
}

// DeliveryReceipt is the model representation of short_message for SMPP PDU delivery_sm.
// SMPP provides for return of an SMSC delivery receipt via the deliver_sm or data_sm PDU,which indicates the delivery status of the message.
type DeliveryReceipt struct {
	ID       string // id，10，C-Octet String (Decimal)，The message ID allocated to the message by the SMSC when originally submitted.
	Sub      string // sub， 3， C-Octet String (Decimal)，Number of short messages originally submitted. This is only relevant when the original message was submitted to a distribution list.The value is padded with leading zeros if necessary.
	Dlvrd    string // dlvrd， 3， C-Octet Fixed Length String (Decimal)， Number of short messages delivered. This is only relevant where the original message was submitted to a distribution list.The value is padded with leading zeros if necessary.
	SubDate  string // submit date 10，C-Octet Fixed Length String The time and date at which the short message was submitted. In the case of a message which has been replaced, this is the date that the original message was replaced.The format is as follows: YYMMDDhhmm where: YY = last two digits of the year (00-99) MM = month (01-12) ，DD= day (01-31) hh = hour (00-23) mm = minute (00-59)
	DoneDate string // done date，10，C-Octet Fixed Length String，The time and date at which the short message reached it’s final state. The format is the same as for the submit date.
	Stat     string // stat，7，C-Octet Fixed Length String，The final status of the message. For settings for this field see Table B-2.
	Err      string // err，3，C-Octet Fixed Length String，Where appropriate this may hold a Network specific error code or an SMSC error code for the attempted delivery of the message. These errors are Network or SMSC specific and are not included here.
	Text     string // text，20，Octet String，The first 20 characters of the short message.
}

func (d DeliveryReceipt) Valid() bool {
	// id 和 stat 都有值时才有效
	if d.ID != "" && d.Stat != "" {
		return true
	}
	return false
}

// ExtractDeliveryReceipt extracts the short_message string into a DeliveryReceipt struct.
func (d *DeliveryReceipt) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	buf.WriteString("id:")
	buf.WriteString(d.ID)
	buf.WriteString(" sub:")
	buf.WriteString(d.Sub)
	buf.WriteString(" dlvrd:")
	buf.WriteString(d.Dlvrd)
	buf.WriteString(" submit date:")
	buf.WriteString(d.SubDate)
	buf.WriteString(" done date:")
	buf.WriteString(d.DoneDate)
	buf.WriteString(" stat:")
	buf.WriteString(d.Stat)
	buf.WriteString(" err:")
	buf.WriteString(d.Err)
	buf.WriteString(" text:")
	buf.WriteString(d.Text)
	return buf.Bytes()
}

func (d *DeliveryReceipt) IDecode(data []byte) error {
	s := strings.TrimRight(string(data), "\x00")
	d.ID = findSubValue(s, "id", 10)
	d.Sub = findSubValue(s, "sub", 3)
	d.Dlvrd = findSubValue(s, "dlvrd", 3)
	d.SubDate = findSubValue(s, "submit date", 10)
	d.DoneDate = findSubValue(s, "done date", 10)
	d.Stat = findSubValue(s, "stat", 7)
	d.Err = findSubValue(s, "err", 3)
	d.Text = findSubValue(s, "text", 20)
	return nil
}

func ExtractDeliveryReceipt(s string) (d DeliveryReceipt, err error) {
	// TODO: Consider the case where the key is in uppercase
	d.ID = findSubValue(s, "id", 10)
	d.Sub = findSubValue(s, "sub", 3)
	d.Dlvrd = findSubValue(s, "dlvrd", 3)
	d.SubDate = findSubValue(s, "submit date", 10)
	d.DoneDate = findSubValue(s, "done date", 10)
	d.Stat = findSubValue(s, "stat", 7)
	d.Err = findSubValue(s, "err", 3)
	d.Text = findSubValue(s, "text", 20)
	return
}

func findSubValue(s string, sub string, maxSize int) (value string) {
	sub = sub + ":"
	n := findDeliveryReceiptFieldStart(s, sub)
	if n == -1 {
		return ""
	}

	start := n + len(sub)
	if strings.TrimSuffix(sub, ":") == "text" {
		// The text field is an Octet String and may contain spaces. It normally
		// terminates the receipt, so preserve the complete remainder. If a
		// receipt contains a known field after text, use that field as the
		// delimiter while still ignoring vendor-specific fields.
		if next := nextDeliveryReceiptField(s, start); next >= 0 {
			return s[start:next]
		}
		return s[start:]
	}

	// 当前 key 后面的下一个空格
	spaceIdx := strings.Index(s[start:], " ")

	// 后面再无空格，说明当前 key 是最后一个，直接返回即可
	if spaceIdx == -1 {
		value = s[start:]
	} else {
		// 空格之前的就是我们要找的 value
		value = s[start : start+spaceIdx]
	}

	// maxSize is retained for source compatibility with the original helper,
	// but receipt IDs and vendor-specific values must not be truncated.
	return
}

func findDeliveryReceiptFieldStart(s, field string) int {
	if strings.HasPrefix(s, field) {
		return 0
	}
	if n := strings.Index(s, " "+field); n >= 0 {
		return n + 1
	}
	return -1
}

func nextDeliveryReceiptField(s string, start int) int {
	next := -1
	for _, field := range deliveryReceiptFields {
		marker := " " + field + ":"
		if idx := strings.Index(s[start:], marker); idx >= 0 {
			idx += start
			if next == -1 || idx < next {
				next = idx
			}
		}
	}
	return next
}

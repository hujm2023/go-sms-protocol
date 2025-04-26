package cmpp20

//goland:noinspection WeakCrypto
import (
	"bytes"
	"crypto/md5"
	"strconv"
	"time"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
)

// NewConnect creates a new PduConnect PDU.
func NewConnect(account, passwd string, seqID uint32) *PduConnect {
	t, ts := now()
	md5Bytes := md5.Sum(
		bytes.Join([][]byte{
			[]byte(account),
			make([]byte, 9),
			[]byte(passwd),
			[]byte(t),
		},
			nil))
	connectPdu := &PduConnect{
		Header:              cmpp.NewHeader(0, cmpp.CommandConnect, seqID),
		SourceAddr:          account,
		AuthenticatorSource: string(md5Bytes[:]),
		Version:             cmpp.Version20,
		Timestamp:           ts,
	}
	return connectPdu
}

// NewTerminatePacket creates a new PduTerminate PDU and encodes it into a byte slice.
func NewTerminatePacket(seqID uint32) []byte {
	pdu := &PduTerminate{Header: cmpp.NewHeader(0, cmpp.CommandTerminate, seqID)}
	data, _ := pdu.IEncode()
	return data
}

// NewActiveTestPacket creates a new PduActiveTest PDU and encodes it into a byte slice.
func NewActiveTestPacket(seqID uint32) []byte {
	pdu := &PduActiveTest{Header: cmpp.NewHeader(MaxActiveTestLength, cmpp.CommandActiveTest, seqID)}
	data, _ := pdu.IEncode()
	return data
}

// now generates the current timestamp in MMDDHHMMSS format (string and uint32).
func now() (string, uint32) {
	s := time.Now().Format("0102150405")
	i, _ := strconv.Atoi(s)
	return s, uint32(i)
}

// DecodeCMPP20 decodes the given byte slice into a corresponding CMPP 2.0 PDU.
func DecodeCMPP20(data []byte) (sms.PDU, error) {
	header, err := cmpp.PeekHeader(data)
	if err != nil {
		return nil, err
	}

	var pdu sms.PDU
	switch header.CommandID {
	case cmpp.CommandConnect:
		pdu = new(PduConnect)
	case cmpp.CommandConnectResp:
		pdu = new(PduConnectResp)
	case cmpp.CommandSubmit:
		pdu = new(PduSubmit)
	case cmpp.CommandSubmitResp:
		pdu = new(PduSubmitResp)
	case cmpp.CommandDeliver:
		pdu = new(PduDeliver)
	case cmpp.CommandDeliverResp:
		pdu = new(PduDeliverResp)
	case cmpp.CommandActiveTest:
		pdu = new(PduActiveTest)
	case cmpp.CommandActiveTestResp:
		pdu = new(PduActiveTestResp)
	case cmpp.CommandTerminate:
		pdu = new(PduTerminate)
	case cmpp.CommandTerminateResp:
		pdu = new(PduTerminateResp)
	}

	if pdu == nil {
		return nil, sms.ErrUnsupportedPacket
	}

	if err = pdu.IDecode(data); err != nil {
		return nil, err
	}
	return pdu, nil
}

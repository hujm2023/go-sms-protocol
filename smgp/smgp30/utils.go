package smgp30

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"time"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/smgp"
)

// NewLogin ...
func NewLogin(account, passwd string, seqID uint32) *Login {
	timeStr := genTimestamp()
	auth, _ := genAuthenticatorClient(account, passwd, timeStr)
	connectPdu := &Login{
		Header:              smgp.NewHeader(0, smgp.CommandLogin, seqID),
		ClientID:            account,
		AuthenticatorClient: string(auth),
		Version:             0x30,
		Timestamp:           timeStr,
		LoginMode:           smgp.TRANSMIT_MODE,
	}
	return connectPdu
}

func NewActiveTestPacket(seqID uint32) []byte {
	pdu := &ActiveTest{Header: smgp.NewHeader(smgp.MaxActiveTestRespLength, smgp.CommandActiveTest, seqID)}
	data, _ := pdu.IEncode()
	return data
}

func DecodeSMGP30(data []byte) (sms.PDU, error) {
	header, err := smgp.PeekHeader(data)
	if err != nil {
		return nil, err
	}

	var pdu sms.PDU
	switch header.CommandID {
	case smgp.CommandLogin:
		pdu = new(Login)
	case smgp.CommandLoginResp:
		pdu = new(LoginResp)
	case smgp.CommandSubmit:
		pdu = new(Submit)
	case smgp.CommandSubmitResp:
		pdu = new(SubmitResp)
	case smgp.CommandDeliver:
		pdu = new(Deliver)
	case smgp.CommandDeliverResp:
		pdu = new(DeliverResp)
	case smgp.CommandActiveTest:
		pdu = new(ActiveTest)
	case smgp.CommandActiveTestResp:
		pdu = new(ActiveTestResp)
	case smgp.CommandExit:
		pdu = new(Exit)
	case smgp.CommandExitResp:
		pdu = new(ExitResp)
	}

	if pdu == nil {
		return nil, sms.ErrUnsupportedPacket
	}

	if err = pdu.IDecode(data); err != nil {
		return nil, err
	}
	return pdu, nil
}

func genTimestamp() uint32 {
	t := time.Now()
	return uint32(int(t.Month())*100000000 + t.Day()*1000000 +
		t.Hour()*10000 + t.Minute()*100 + t.Second())
}

// 生成客户端认证码
// 其值通过单向MD5 hash计算得出，表示如下：
// AuthenticatorClient =MD5（ClientID+7 字节的二进制0（0x00） + Shared secret+Timestamp）
// Shared secret 由服务器端与客户端事先商定，最长15字节。
// 此处Timestamp格式为：MMDDHHMMSS（月日时分秒），经TimeStamp字段值转换成字符串，转换后右对齐，左补0x30得到。
// 例如3月1日0时0分0秒，TimeStamp字段值为0x11F0E540，此处为0301000000。
func genAuthenticatorClient(clientId, secret string, timestamp uint32) ([]byte, error) {
	buf := new(bytes.Buffer)

	buf.WriteString(clientId)
	buf.Write([]byte{0, 0, 0, 0, 0, 0, 0})
	buf.WriteString(secret)
	buf.WriteString(fmt.Sprintf("%010d", timestamp))

	h := md5.New()
	_, err := h.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

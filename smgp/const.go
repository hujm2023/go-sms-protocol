package smgp

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	MinSMGPPduLength  = HeaderLength // smgp PduSMGP 最小的长度
	MaxExitRespLength = HeaderLength

	MaxActiveTestRespLength = 13
	MaxDeliverRespLength    = HeaderLength + 10 + 4
)

type CommandID uint32

func (c CommandID) String() string {
	switch c {
	case CommandLogin:
		return "SMGP_LOGIN"
	case CommandLoginResp:
		return "SMGP_LOGIN_RESP"
	case CommandActiveTest:
		return "SMGP_ACTIVE_TEST"
	case CommandActiveTestResp:
		return "SMGP_ACTIVE_TEST_RESP"
	case CommandSubmit:
		return "SMGP_SUBMIT"
	case CommandSubmitResp:
		return "SMGP_SUBMIT_RESP"
	case CommandDeliver:
		return "SMGP_DELIVERY"
	case CommandDeliverResp:
		return "SMGP_DELIVERY_RESP"
	case CommandExit:
		return "SMGP_EXIT"
	case CommandExitResp:
		return "SMGP_EXIT_RESP"
	}
	return fmt.Sprintf("unknown(%d)", uint32(c))
}

func (c CommandID) ToUint32() uint32 {
	return uint32(c)
}

// client发起
const (
	CommandRequestNone    CommandID = 0x00000000 + iota
	CommandLogin                    // 请求连接
	CommandSubmit                   // 提交短信
	CommandDeliver                  // 短信下发
	CommandActiveTest               // 激活测试
	CommandFwd                      // 消息前转
	CommandExit                     // 退出请求
	CommandQuery                    // SP统计查询
	CommandQueryTERoute             // 查询TE路由
	CommandQuerySPRoute             // 查询SP路由
	CommandPaymentRequest           // 扣款请求

)

// server 应答
const (
	CommandResponseNone   CommandID = 0x80000000 + iota
	CommandLoginResp                // 登陆应答
	CommandSubmitResp               // 提交短信应答
	CommandDeliverResp              // 短信下发应答
	CommandActiveTestResp           // 激活测试应答
	CommandFwdResp                  // 消息前转应答
	CommandExitResp                 // 终止连接应答
)

var ErrInvalidPudLength = errors.New("invalid pdu length")

// 客户端用来登录服务器端的登录模式
const (
	SEND_MODE = iota
	RECEIVE_MODE
	TRANSMIT_MODE
)

// MsgType
const (
	MO uint8 = 0 // MO消息（终端发给SP）
	MT uint8 = 6 // MT消息（SP发给终端，包括WEB上发送的点对点短消息）
)

// MsgFormat
// 短消息内容体的编码格式
// 对于文字短消息，要求MsgFormat＝15, 对于回执消息，要求MsgFormat＝0
const (
	ASCII   uint8 = 0  // ASCII编码
	BINARY  uint8 = 4  // 二进制短消息
	UCS2    uint8 = 8  // UCS2编码
	GB18030 uint8 = 15 // GB18030编码
)

const (
	NOT_REPORT uint8 = 0 // 不是状态报告
	IS_REPORT  uint8 = 1 // 是状态报告
)

// 是否要求返回状态报告
const (
	NO_NEED_REPORT uint8 = 0
	NEED_REPORT    uint8 = 1
)

// 短消息发送优先级
const (
	LOW_PRIORITY uint8 = iota
	NORMAL_PRIORITY
	HIGHER_PRIORITY
	HIGHEST_PRIORITY
)

type Status uint32

func (s Status) Data() uint32 {
	return uint32(s)
}

func (s Status) Error() error {
	return errors.New(strconv.Itoa(int(s)) + " : " + s.String())
}

func (s Status) String() string {
	var msg string
	switch s {
	case 0:
		msg = "成功"
	case 1:
		msg = "系统忙"
	case 2:
		msg = "超过最大连接数"
	case 10:
		msg = "消息结构错"
	case 11:
		msg = "命令字错"
	case 12:
		msg = "序列号重复"
	case 20:
		msg = "IP地址错"
	case 21:
		msg = "认证错"
	case 22:
		msg = "版本太高"
	case 30:
		msg = "非法消息类型（MsgType）"
	case 31:
		msg = "非法优先级（Priority）"
	case 32:
		msg = "非法资费类型（FeeType）"
	case 33:
		msg = "非法资费代码（FeeCode）"
	case 34:
		msg = "非法短消息格式（MsgFormat）"
	case 35:
		msg = "非法时间格式"
	case 36:
		msg = "非法短消息长度（MsgLength）"
	case 37:
		msg = "有效期已过"
	case 38:
		msg = "非法查询类别（QueryType）"
	case 39:
		msg = "路由错误"
	case 40:
		msg = "非法包月费/封顶费（FixedFee）"
	case 41:
		msg = "非法更新类型（UpdateType）"
	case 42:
		msg = "非法路由编号（RouteId）"
	case 43:
		msg = "非法服务代码（ServiceId）"
	case 44:
		msg = "非法有效期（ValidTime）"
	case 45:
		msg = "非法定时发送时间（AtTime）"
	case 46:
		msg = "非法发送用户号码（SrcTermId）"
	case 47:
		msg = "非法接收用户号码（DestTermId）"
	case 48:
		msg = "非法计费用户号码（ChargeTermId）"
	case 49:
		msg = "非法SP服务代码（SPCode）"
	case 56:
		msg = "非法源网关代码（SrcGatewayID）"
	case 57:
		msg = "非法查询号码（QueryTermID）"
	case 58:
		msg = "没有匹配路由"
	case 59:
		msg = "非法SP类型（SPType）"
	case 60:
		msg = "非法上一条路由编号（LastRouteID）"
	case 61:
		msg = "非法路由类型（RouteType）"
	case 62:
		msg = "非法目标网关代码（DestGatewayID）"
	case 63:
		msg = "非法目标网关IP（DestGatewayIP）"
	case 64:
		msg = "非法目标网关端口（DestGatewayPort）"
	case 65:
		msg = "非法路由号码段（TermRangeID）"
	case 66:
		msg = "非法终端所属省代码（ProvinceCode）"
	case 67:
		msg = "非法用户类型（UserType）"
	case 68:
		msg = "本节点不支持路由更新"
	case 69:
		msg = "非法SP企业代码（SPID）"
	case 70:
		msg = "非法SP接入类型（SPAccessType）"
	case 71:
		msg = "路由信息更新失败"
	case 72:
		msg = "非法时间戳（Time）"
	case 73:
		msg = "非法业务代码（MServiceID）"
	case 74:
		msg = "SP禁止下发时段"
	case 75:
		msg = "SP发送超过日流量"
	case 76:
		msg = "SP帐号过有效期"

	default:
		msg = "Status Unknown: " + strconv.Itoa(int(s))
	}

	return msg
}

const (
	StatOk Status = iota
)

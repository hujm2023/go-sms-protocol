package cmpp

import (
	"fmt"
)

const (
	MinCMPPPduLength = HeaderLength // cmpp PduCMPP 最小的长度
)

type Version uint8

func (v Version) String() string {
	switch v {
	case Version20:
		return "CMPP V2.0"
	case Version30:
		return "CMPP V3.0"
	default:
		return "Unknown"
	}
}

const (
	VersionUnknown = 0
	Version20      = 0x20
	Version30      = 0x30
)

type CommandID uint32

func (c CommandID) String() string {
	switch c {
	case CommandConnect:
		return "CMPP_CONNECT"
	case CommandConnectResp:
		return "CMPP_CONNECT_RESP"
	case CommandActiveTest:
		return "CMPP_ACTIVE_TEST"
	case CommandActiveTestResp:
		return "CMPP_ACTIVE_TEST_RESP"
	case CommandSubmit:
		return "CMPP_SUBMIT"
	case CommandSubmitResp:
		return "CMPP_SUBMIT_RESP"
	case CommandDeliver:
		return "CMPP_DELIVERY"
	case CommandDeliverResp:
		return "CMPP_DELIVERY_RESP"
	case CommandTerminate:
		return "CMPP_TERMINATE"
	case CommandTerminateResp:
		return "CMPP_TERMINATE_RESP"
	}
	return fmt.Sprintf("unknown(%d)", uint32(c))
}

func (c CommandID) ToUint32() uint32 {
	return uint32(c)
}

func (c CommandID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, c.String())), nil
}

func (c *CommandID) UnmarshalJSON(b []byte) error {
	fmt.Println(string(b))
	switch string(b) {
	case `"CMPP_CONNECT"`:
		*c = CommandConnect
	case `"CMPP_CONNECT_RESP"`:
		*c = CommandConnectResp
	case `"CMPP_ACTIVE_TEST"`:
		*c = CommandActiveTest
	case `"CMPP_ACTIVE_TEST_RESP"`:
		*c = CommandActiveTestResp
	case `"CMPP_SUBMIT"`:
		*c = CommandSubmit
	case `"CMPP_SUBMIT_RESP"`:
		*c = CommandSubmitResp
	case `"CMPP_DELIVERY"`:
		*c = CommandDeliver
	case `"CMPP_DELIVERY_RESP"`:
		*c = CommandDeliverResp
	case `"CMPP_TERMINATE"`:
		*c = CommandTerminate
	case `"CMPP_TERMINATE_RESP"`:
		*c = CommandTerminateResp
	default:
		return fmt.Errorf("invalid command id: %s", string(b))
	}
	return nil
}

const (
	CommandRequestNone       CommandID = 0x00000000 + iota
	CommandConnect                     // 请求连接
	CommandTerminate                   // 终止连接
	_                                  // 保留
	CommandSubmit                      // 提交短信
	CommandDeliver                     // 短信下发
	CommandQuery                       // 发送短信状态查询
	CommandCancel                      // 删除短信
	CommandActiveTest                  // 激活测试
	CommandFwd                         // 消息前转
	CommandMtRoute                     // MT 路由请求
	CommandMoRoute                     // MO 路由请求
	CommandGetRoute                    // 获取路由请求
	CommandMtRouteUpdate               // MT 路由更新
	CommandMoRouteUpdate               // MO 路由更新
	CommandPushMtRouteUpdate           // MT 路由更新
	CommandPushMoRouteUpdate           // MO 路由更新
)

const (
	CommandResponseNone          CommandID = 0x80000000 + iota
	CommandConnectResp                     // 请求连接应答
	CommandTerminateResp                   // 终止连接应答
	_                                      // 保留
	CommandSubmitResp                      // 提交短信应答
	CommandDeliverResp                     // 短信下发应答
	CommandQueryResp                       // 发送短信状态查询应答
	CommandCancelResp                      // 删除短信应答
	CommandActiveTestResp                  // 激活测试应答
	CommandFwdResp                         // 消息前转应答
	CommandMtRouteResp                     // MT 路由请求应答
	CommandMoRouteResp                     // MO 路由请求应答
	CommandGetRouteResp                    // 获取路由请求应答
	CommandMtRouteUpdateResp               // MT 路由更新应答
	CommandMoRouteUpdateResp               // MO 路由更新应答
	CommandPushMtRouteUpdateResp           // MT 路由更新应答
	CommandPushMoRouteUpdateResp           // MO 路由更新应答
)

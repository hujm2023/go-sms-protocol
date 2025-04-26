package cmpp

import (
	"fmt"
)

const (
	// MinCMPPPduLength defines the minimum length of a CMPP PDU, which is the header length.
	MinCMPPPduLength = HeaderLength // cmpp PduCMPP 最小的长度
)

// Version represents the CMPP protocol version.
type Version uint8

// String returns the string representation of the CMPP version.
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
	// VersionUnknown represents an unknown CMPP version.
	VersionUnknown = 0
	// Version20 represents CMPP version 2.0.
	Version20 = 0x20
	// Version30 represents CMPP version 3.0.
	Version30 = 0x30
)

// CommandID represents the command identifier for CMPP PDUs.
type CommandID uint32

// String returns the string representation of the CommandID.
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

// ToUint32 converts the CommandID to its uint32 representation.
func (c CommandID) ToUint32() uint32 {
	return uint32(c)
}

// MarshalJSON implements the json.Marshaler interface for CommandID.
func (c CommandID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, c.String())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for CommandID.
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

// CMPP Command IDs for requests.
const (
	// CommandRequestNone is a placeholder for no request command.
	CommandRequestNone CommandID = 0x00000000 + iota
	// CommandConnect represents the CMPP_CONNECT request.
	CommandConnect // 请求连接
	// CommandTerminate represents the CMPP_TERMINATE request.
	CommandTerminate // 终止连接
	_                // 保留
	// CommandSubmit represents the CMPP_SUBMIT request.
	CommandSubmit // 提交短信
	// CommandDeliver represents the CMPP_DELIVER request (used for MO messages).
	CommandDeliver // 短信下发
	// CommandQuery represents the CMPP_QUERY request.
	CommandQuery // 发送短信状态查询
	// CommandCancel represents the CMPP_CANCEL request.
	CommandCancel // 删除短信
	// CommandActiveTest represents the CMPP_ACTIVE_TEST request.
	CommandActiveTest // 激活测试
	// CommandFwd represents the CMPP_FWD request (message forwarding).
	CommandFwd // 消息前转
	// CommandMtRoute represents the CMPP_MT_ROUTE request.
	CommandMtRoute // MT 路由请求
	// CommandMoRoute represents the CMPP_MO_ROUTE request.
	CommandMoRoute // MO 路由请求
	// CommandGetRoute represents the CMPP_GET_ROUTE request.
	CommandGetRoute // 获取路由请求
	// CommandMtRouteUpdate represents the CMPP_MT_ROUTE_UPDATE request.
	CommandMtRouteUpdate // MT 路由更新
	// CommandMoRouteUpdate represents the CMPP_MO_ROUTE_UPDATE request.
	CommandMoRouteUpdate // MO 路由更新
	// CommandPushMtRouteUpdate represents the CMPP_PUSH_MT_ROUTE_UPDATE request.
	CommandPushMtRouteUpdate // MT 路由更新
	// CommandPushMoRouteUpdate represents the CMPP_PUSH_MO_ROUTE_UPDATE request.
	CommandPushMoRouteUpdate // MO 路由更新
)

// CMPP Command IDs for responses.
const (
	// CommandResponseNone is a placeholder for no response command.
	CommandResponseNone CommandID = 0x80000000 + iota
	// CommandConnectResp represents the CMPP_CONNECT_RESP response.
	CommandConnectResp // 请求连接应答
	// CommandTerminateResp represents the CMPP_TERMINATE_RESP response.
	CommandTerminateResp // 终止连接应答
	_                    // 保留
	// CommandSubmitResp represents the CMPP_SUBMIT_RESP response.
	CommandSubmitResp // 提交短信应答
	// CommandDeliverResp represents the CMPP_DELIVER_RESP response.
	CommandDeliverResp // 短信下发应答
	// CommandQueryResp represents the CMPP_QUERY_RESP response.
	CommandQueryResp // 发送短信状态查询应答
	// CommandCancelResp represents the CMPP_CANCEL_RESP response.
	CommandCancelResp // 删除短信应答
	// CommandActiveTestResp represents the CMPP_ACTIVE_TEST_RESP response.
	CommandActiveTestResp // 激活测试应答
	// CommandFwdResp represents the CMPP_FWD_RESP response.
	CommandFwdResp // 消息前转应答
	// CommandMtRouteResp represents the CMPP_MT_ROUTE_RESP response.
	CommandMtRouteResp // MT 路由请求应答
	// CommandMoRouteResp represents the CMPP_MO_ROUTE_RESP response.
	CommandMoRouteResp // MO 路由请求应答
	// CommandGetRouteResp represents the CMPP_GET_ROUTE_RESP response.
	CommandGetRouteResp // 获取路由请求应答
	// CommandMtRouteUpdateResp represents the CMPP_MT_ROUTE_UPDATE_RESP response.
	CommandMtRouteUpdateResp // MT 路由更新应答
	// CommandMoRouteUpdateResp represents the CMPP_MO_ROUTE_UPDATE_RESP response.
	CommandMoRouteUpdateResp // MO 路由更新应答
	// CommandPushMtRouteUpdateResp represents the CMPP_PUSH_MT_ROUTE_UPDATE_RESP response.
	CommandPushMtRouteUpdateResp // MT 路由更新应答
	// CommandPushMoRouteUpdateResp represents the CMPP_PUSH_MO_ROUTE_UPDATE_RESP response.
	CommandPushMoRouteUpdateResp // MO 路由更新应答
)

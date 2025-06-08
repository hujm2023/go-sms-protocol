package sgip

import (
	"errors"
	"fmt"
	"strconv"
)

type RespStatus uint8

const (
	// 系统中每个消息包最大不超过2K字节
	MAX_OP_SIZE = 2048

	// 群发接收手机号码最大为100
	MAX_USER_COUNT

	MinSGIPPduLength = HeaderLength // smgp PduSMGP 最小的长度

	MaxBindLength = HeaderLength + 1 + 16 + 16 + 8

	MaxRespLength = HeaderLength + 1 + 8
)

// LoginType 登录类型
/*
	1:SP 向 SMG 建立的连接，用于发送命令  我们连运营商是1
	2:SMG 向 SP 建立的连接，用于发送命令  运营商连我们是2
	3:SMG 之间建立的连接，用于转发命令
	4:SMG 向 GNS 建立的连接，用于路由表的检索和 维护
	5:GNS 向 SMG 建立的连接，用于路由表的更新 6:主备 GNS 之间建立的连接，用于主备路由表的 一致性
	11:SP 与 SMG 以及 SMG 之间建立的测试连接， 用于跟踪测试
*/
type LoginType uint8

const (
	SP_SMG LoginType = iota + 1
	SMG_SP
	SMG_SMG
)

var ErrInvalidPudLength = errors.New("invalid pdu length")

type CommandID uint32

func (c CommandID) String() string {
	switch c {
	case SGIP_BIND:
		return "SGIP_BIND"
	case SGIP_UNBIND:
		return "SGIP_UNBIND"
	case SGIP_SUBMIT:
		return "SGIP_SUBMIT"
	case SGIP_DELIVER:
		return "SGIP_DELIVER"
	case SGIP_REPORT:
		return "SGIP_REPORT"
	case SGIP_BIND_REP:
		return "SGIP_BIND_REP"
	case SGIP_UNBIND_REP:
		return "SGIP_UNBIND_REP"
	case SGIP_SUBMIT_REP:
		return "SGIP_SUBMIT_REP"
	case SGIP_DELIVER_REP:
		return "SGIP_DELIVER_REP"
	case SGIP_REPORT_REP:
		return "SGIP_REPORT_REP"
	}
	return fmt.Sprintf("unknown(%d)", uint32(c))
}

func (c CommandID) ToUint32() uint32 {
	return uint32(c)
}

// Command ID
const (
	SGIP_BIND    CommandID = 0x00000001 + iota // 建连请求
	SGIP_UNBIND                                // 断连连请求
	SGIP_SUBMIT                                // 提交请求
	SGIP_DELIVER                               // 上行请求
	SGIP_REPORT                                // 回执请求
	SGIP_ADDSP
	SGIP_MODIFYSP
	SGIP_DELETESP
	SGIP_QUERYROUTE
	SGIP_ADDTELESEG
	SGIP_MODIFYTELESEG
	SGIP_DELETETELESEG
	SGIP_ADDSMG
	SGIP_MODIFYSMG
	SGIP_DELETESMG
	SGIP_CHECKUSER
	SGIP_USERRPT
)

const (
	SGIP_BIND_REP    CommandID = 0x80000001 + iota // 建连响应
	SGIP_UNBIND_REP                                // 断连响应
	SGIP_SUBMIT_REP                                // 提交响应
	SGIP_DELIVER_REP                               // 上行响应
	SGIP_REPORT_REP                                // 回执响应
	SGIP_ADDSP_REP
	SGIP_MODIFYSP_REP
	SGIP_DELETESP_REP
	SGIP_QUERYROUTE_REP
	SGIP_ADDTELESEG_REP
	SGIP_MODIFYTELESEG_REP
	SGIP_DELETETELESEG_REP
	SGIP_ADDSMG_REP
	SGIP_MODIFYSMG_REP
	SGIP_DELETESMG_REP
	SGIP_CHECKUSER_REP
	SGIP_USERRPT_REP
)

func (rs RespStatus) String() string {
	switch rs {
	case STAT_OK:
		return "STAT_OK"
	case STAT_ILLLOGIN:
		return "非法登录，如登录名、口令出错、登录名与口令不符等"
	case STAT_RPTLOGIN:
		return "重复登录，如在同一TCP/IP连接中连续两次以上请求登录"
	case STAT_MUCHCONN:
		return "连接过多，指单个节点要求同时建立的连接数过多"
	case STAT_ERLGNTYPE:
		return "登录类型错，指bind命令中的logintype字段出错"
	case STAT_ERARGFMT:
		return "参数格式错，指命令中参数值与参数类型不符或与协议规定的范围不符"
	case STAT_ILLUSRNUM:
		return "非法手机号码，协议中所有手机号码字段出现非86130号码或手机号码前未加“86”时都应报错"
	case STAT_ERSEQ:
		return "消息ID错"
	case STAT_ERLEN:
		return "非法序列号，包括序列号重复、序列号格式错误等"
	case STAT_ILLSEQ:
		return "非法操作GNS"
	case STAT_ILLOPGNS:
		return "非法操作GNS"
	case STAT_NODEBUSY:
		return "节点忙，指本节点存储队列满或其他原因，暂时不能提供服务的情况"
	case STAT_DSTCNTRCH:
		return "目的地址不可达，指路由表存在路由且消息路由正确但被路由的节点暂时不能提供服务的情况"
	case STAT_ROUTER:
		return "路由错，指路由表存在路由但消息路由出错的情况，如转错SMG等"
	case STAT_ROUTENEST:
		return "路由不存在，指消息路由的节点在路由表中不存在"
	case STAT_INVCHGNUM:
		return "计费号码无效，鉴权不成功时反馈的错误信息"
	case STAT_USRCNTRCH:
		return "用户不能通信（如不在服务区、未开机等情况)"
	case STAT_MEMFULL:
		return "手机内存不足"
	case STAT_NTSPTSMS:
		return "手机不支持短消息"
	case STAT_RCVERR:
		return "手机接收短消息出现错误"
	case STAT_UNKNUSR:
		return "不知道的用户"
	case STAT_NTSPTFUN:
		return "不提供此功能"
	case STAT_ILLDEV:
		return "非法设备"
	case STAT_SYSFAIL:
		return "系统失败"
	case STAT_SMSCFULL:
		return "短信中心队列满"
	default:
		return "Status Unknown: " + strconv.Itoa(int(rs))
	}

}

const (
	STAT_OK RespStatus = iota // 无错误，命令正确接收

	// 1-20所指错误一般在各类命令的应答中用到

	STAT_ILLLOGIN  // 非法登录，如登录名、口令出错、登录名与口令不符等
	STAT_RPTLOGIN  // 重复登录，如在同一TCP/IP连接中连续两次以上请求登录
	STAT_MUCHCONN  // 连接过多，指单个节点要求同时建立的连接数过多
	STAT_ERLGNTYPE // 登录类型错，指bind命令中的logintype字段出错
	STAT_ERARGFMT  // 参数格式错，指命令中参数值与参数类型不符或与协议规定的范围不符
	STAT_ILLUSRNUM // 非法手机号码，协议中所有手机号码字段出现非86130号码或手机号码前未加“86”时都应报错
	STAT_ERSEQ     // 消息ID错
	STAT_ERLEN     // 信息长度错
	STAT_ILLSEQ    // 非法序列号，包括序列号重复、序列号格式错误等
	STAT_ILLOPGNS  // 非法操作GNS
	STAT_NODEBUSY  // 节点忙，指本节点存储队列满或其他原因，暂时不能提供服务的情况

	// 21-32所指错误一般在report命令中用到

	STAT_DSTCNTRCH = 21 + iota // 目的地址不可达，指路由表存在路由且消息路由正确但被路由的节点暂时不能提供服务的情况
	STAT_ROUTER                // 路由错，指路由表存在路由但消息路由出错的情况，如转错SMG等
	STAT_ROUTENEST             // 路由不存在，指消息路由的节点在路由表中不存在
	STAT_INVCHGNUM             // 计费号码无效，鉴权不成功时反馈的错误信息
	STAT_USRCNTRCH             // 用户不能通信（如不在服务区、未开机等情况）
	STAT_MEMFULL               // 手机内存不足
	STAT_NTSPTSMS              // 手机不支持短消息
	STAT_RCVERR                // 手机接收短消息出现错误
	STAT_UNKNUSR               // 不知道的用户
	STAT_NTSPTFUN              // 不提供此功能
	STAT_ILLDEV                // 非法设备
	STAT_SYSFAIL               // 系统失败
	STAT_SMSCFULL              // 短信中心队列满
)

// sgip 回执的错误状态枚举
const (
	REPORT_OK   RespStatus = 0
	REPORT_WAIT RespStatus = 1
	REPORT_FAIL RespStatus = 2
)

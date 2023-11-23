package smpp34

import (
	"fmt"

	"github.com/hujm2023/go-sms-protocol/packet"
)

type CMDStatus uint32

type CMDId uint32

type Header struct {
	Length   uint32
	ID       CMDId
	Status   CMDStatus
	Sequence uint32
}

func (h Header) String() string {
	return fmt.Sprintf("Length:%d, Id:%s, Status:%d, Sequence:%d", h.Length, h.ID, h.Status, h.Sequence)
}

func NewPduHeader(l uint32, id CMDId, status CMDStatus, seq uint32) *Header {
	return &Header{l, id, status, seq}
}

func ReadHeader(r *packet.PacketReader) Header {
	var h Header
	r.ReadNumeric(&h.Length)
	r.ReadNumeric(&h.ID)
	r.ReadNumeric(&h.Status)
	r.ReadNumeric(&h.Sequence)
	return h
}

func (s CMDId) Error() string {
	return s.String()
}

func (s CMDId) String() string {
	switch s {
	case GENERIC_NACK:
		return "SMPP_GENERIC_NACK"
	case BIND_RECEIVER:
		return "SMPP_BIND_RECEIVER"
	case BIND_RECEIVER_RESP:
		return "SMPP_BIND_RECEIVER_RESP"
	case BIND_TRANSMITTER:
		return "SMPP_BIND_TRANSMITTER"
	case BIND_TRANSMITTER_RESP:
		return "SMPP_BIND_TRANSMITTER_RESP"
	case QUERY_SM:
		return "SMPP_QUERY_SM"
	case QUERY_SM_RESP:
		return "SMPP_QUERY_SM_RESP"
	case SUBMIT_SM:
		return "SMPP_SUBMIT_SM"
	case SUBMIT_SM_RESP:
		return "SMPP_SUBMIT_SM_RESP"
	case DELIVER_SM:
		return "SMPP_DELIVER_SM"
	case DELIVER_SM_RESP:
		return "SMPP_DELIVER_SM_RESP"
	case UNBIND:
		return "SMPP_UNBIND"
	case UNBIND_RESP:
		return "SMPP_UNBIND_RESP"
	case REPLACE_SM:
		return "SMPP_REPLACE_SM"
	case REPLACE_SM_RESP:
		return "SMPP_REPLACE_SM_RESP"
	case CANCEL_SM:
		return "SMPP_CANCEL_SM"
	case CANCEL_SM_RESP:
		return "SMPP_CANCEL_SM_RESP"
	case BIND_TRANSCEIVER:
		return "SMPP_BIND_TRANSCEIVER"
	case BIND_TRANSCEIVER_RESP:
		return "SMPP_BIND_TRANSCEIVER_RESP"
	case OUTBIND:
		return "SMPP_OUTBIND"
	case ENQUIRE_LINK:
		return "SMPP_ENQUIRE_LINK"
	case ENQUIRE_LINK_RESP:
		return "SMPP_ENQUIRE_LINK_RESP"
	case SUBMIT_MULTI:
		return "SMPP_SUBMIT_MULTI"
	case SUBMIT_MULTI_RESP:
		return "SMPP_SUBMIT_MULTI_RESP"
	case ALERT_NOTIFICATION:
		return "SMPP_ALERT_NOTIFICATION"
	case DATA_SM:
		return "SMPP_DATA_SM"
	case DATA_SM_RESP:
		return "SMPP_DATA_SM_RESP"
	default:
		return fmt.Sprintf("UNKNOWN PDU %d", s)
	}
}

func (s CMDStatus) Error() string {
	switch s {
	default:
		return fmt.Sprint("Unknown Status:", uint32(s))
	case ESME_ROK:
		return "No Error"
	case ESME_RINVMSGLEN:
		return "Message Length is invalid"
	case ESME_RINVCMDLEN:
		return "Command Length is invalid"
	case ESME_RINVCMDID:
		return "Invalid Command ID"
	case ESME_RINVBNDSTS:
		return "Incorrect BIND Status for given command"
	case ESME_RALYBND:
		return "ESME Already in Bound State"
	case ESME_RINVPRTFLG:
		return "Invalid Priority Flag"
	case ESME_RINVREGDLVFLG:
		return "Invalid Registered CMPPDelivery Flag"
	case ESME_RSYSERR:
		return "System Error"
	case ESME_RINVSRCADR:
		return "Invalid Source Address"
	case ESME_RINVDSTADR:
		return "Invalid Dest Addr"
	case ESME_RINVMSGID:
		return "Message ID is invalid"
	case ESME_RBINDFAIL:
		return "Bind Failed"
	case ESME_RINVPASWD:
		return "Invalid Password"
	case ESME_RINVSYSID:
		return "Invalid System ID"
	case ESME_RCANCELFAIL:
		return "Cancel SM Failed"
	case ESME_RREPLACEFAIL:
		return "Replace SM Failed"
	case ESME_RMSGQFUL:
		return "Message Queue Full"
	case ESME_RINVSERTYP:
		return "Invalid Service Type"
	case ESME_RINVNUMDESTS:
		return "Invalid number of destinations"
	case ESME_RINVDLNAME:
		return "Invalid Distribution List name"
	case ESME_RINVDESTFLAG:
		return "Destination flag is invalid"
	case ESME_RINVSUBREP:
		return "Invalid 'submit with replace' request"
	case ESME_RINVESMCLASS:
		return "Invalid esm_class field data"
	case ESME_RCNTSUBDL:
		return "Cannot Submit to Distribution List"
	case ESME_RSUBMITFAIL:
		return "submit_sm or submit_multi failed"
	case ESME_RINVSRCTON:
		return "Invalid Source address TON"
	case ESME_RINVSRCNPI:
		return "Invalid Source address NPI"
	case ESME_RINVDSTTON:
		return "Invalid Destination address TON"
	case ESME_RINVDSTNPI:
		return "Invalid Destination address NPI"
	case ESME_RINVSYSTYP:
		return "Invalid system_type field"
	case ESME_RINVREPFLAG:
		return "Invalid replace_if_present flag"
	case ESME_RINVNUMMSGS:
		return "Invalid number of messages"
	case ESME_RTHROTTLED:
		return "Throttling error (ESME has exceeded allowed message limit"
	case ESME_RINVSCHED:
		return "Invalid IsScheduled CMPPDelivery Time"
	case ESME_RINVEXPIRY:
		return "Invalid message validity period (Expiry time)"
	case ESME_RINVDFTMSGID:
		return "Predefined Message Invalid or Not Found"
	case ESME_RX_T_APPN:
		return "ESME Receiver Temporary App Error Code"
	case ESME_RX_P_APPN:
		return "ESME Receiver Permanent App Error Code"
	case ESME_RX_R_APPN:
		return "ESME Receiver Reject Message Error Code"
	case ESME_RQUERYFAIL:
		return "Query_sm request failed"
	case ESME_RINVOPTPARSTREAM:
		return "Error in the optional part of the PDU Body."
	case ESME_ROPTPARNOTALLWD:
		return "Optional Parameter not allowed"
	case ESME_RINVPARLEN:
		return "Invalid Parameter Length."
	case ESME_RMISSINGOPTPARAM:
		return "Expected Optional Parameter missing"
	case ESME_RINVOPTPARAMVAL:
		return "Invalid Optional Parameter Value"
	case ESME_RDELIVERYFAILURE:
		return "CMPPDelivery Failure (used for data_sm_resp)"
	case ESME_RUNKNOWNERR:
		return "Unknown Error"
	}
}

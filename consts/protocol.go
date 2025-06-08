package consts

// Protocol 支持的协议类型
type Protocol string

const (
	ProtocolUnknown Protocol = "UNKNOWN"
	ProtocolCMPP    Protocol = "CMPP"
	ProtocolSMPP    Protocol = "SMPP"
	ProtocolSMGP    Protocol = "SMGP"
	ProtocolSGIP    Protocol = "SGIP"
)

func (p Protocol) String() string {
	return string(p)
}

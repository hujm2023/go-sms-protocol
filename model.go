package protocol

// Protocol 支持的协议类型
type Protocol string

const (
	CMPP Protocol = "CMPP"
	SMPP Protocol = "SMPP"
)

func (p Protocol) String() string {
	return string(p)
}

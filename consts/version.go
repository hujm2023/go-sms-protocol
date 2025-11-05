package consts

import (
	"fmt"
	"strings"
)

type Version interface {
	String() string
	Protocol() Protocol
	ToInt8() int8
	ToInt() int
}

// ---- cmpp version ----

const (
	CMPPVersion2_0 CMPPVersion = 0x20
	CMPPVersion2_1 CMPPVersion = 0x21
	CMPPVersion3_0 CMPPVersion = 0x30
)

// CMPPVersion cmpp 协议版本
type CMPPVersion int8

func (cv CMPPVersion) String() string {
	if cv == 0 {
		return "unknown"
	}
	// 高4位表示主版本，低4位表示次版本
	return VersionString(cv)
}

func (cv CMPPVersion) Protocol() Protocol {
	return ProtocolCMPP
}

func (cv CMPPVersion) ToInt8() int8 {
	return int8(cv)
}

func (cv CMPPVersion) ToInt() int {
	return int(cv)
}

// ---- smpp version ----

type SMPPVersion int8

const (
	SMPPVersion3_4 SMPPVersion = 0x34
	SMPPVersion5_0 SMPPVersion = 0x50
)

func (sv SMPPVersion) String() string {
	if sv == 0 {
		return "unknown"
	}
	return VersionString(sv)
}

func (sv SMPPVersion) Protocol() Protocol {
	return ProtocolSMPP
}

func (sv SMPPVersion) ToInt8() int8 {
	return int8(sv)
}

func (sv SMPPVersion) ToInt() int {
	return int(sv)
}

// ---- smgp version ----

const SMGPVersion3_0 SMGPVersion = 0x30

// SMGPVersion smgp 协议版本
type SMGPVersion int8

func (cv SMGPVersion) String() string {
	if cv == 0 {
		return "unknown"
	}
	// 高4位表示主版本，低4位表示次版本
	return VersionString(cv)
}

func (cv SMGPVersion) Protocol() Protocol {
	return ProtocolSMGP
}

func (cv SMGPVersion) ToInt8() int8 {
	return int8(cv)
}

func (cv SMGPVersion) ToInt() int {
	return int(cv)
}

// ---- sgip version ----

// SGIPVersion 协议版本
type SGIPVersion int8

const SGIPVersion1_2 SGIPVersion = 0x12

func (sv SGIPVersion) String() string {
	if sv == 0 {
		return "unknown"
	}
	// 高4位表示主版本，低4位表示次版本
	return VersionString(sv)
}

func (sv SGIPVersion) Protocol() Protocol {
	return ProtocolSGIP
}

func (sv SGIPVersion) ToInt8() int8 {
	return int8(sv)
}

func (sv SGIPVersion) ToInt() int {
	return int(sv)
}

// ---- unknown version ----

type UnknownVersion int8

func (u UnknownVersion) String() string {
	if u == 0 {
		return "unknown"
	}
	return VersionString(u)
}

func (u UnknownVersion) Protocol() Protocol {
	return ProtocolUnknown
}

func (u UnknownVersion) ToInt8() int8 {
	return int8(u)
}

func (u UnknownVersion) ToInt() int {
	return int(u)
}

// ---- utils ----

var (
	smpp34String = pvString(ProtocolSMPP, SMPPVersion3_4)
	cmpp20String = pvString(ProtocolCMPP, CMPPVersion2_0)
	cmpp30String = pvString(ProtocolCMPP, CMPPVersion3_0)
	sgmp30String = pvString(ProtocolSMGP, SMGPVersion3_0)
	sgip12String = pvString(ProtocolSGIP, SGIPVersion1_2)
)

func ProtocolVersionString(p Protocol, v Version) string {
	// fast path, to avoid too many allocs
	switch {
	case p == ProtocolCMPP && v == CMPPVersion2_0:
		return cmpp20String
	case p == ProtocolCMPP && v == CMPPVersion3_0:
		return cmpp30String
	case p == ProtocolSMPP && v == SMPPVersion3_4:
		return smpp34String
	case p == ProtocolSMGP && v == SMGPVersion3_0:
		return sgmp30String
	case p == ProtocolSGIP && v == SGIPVersion1_2:
		return sgip12String
	}

	return pvString(p, v)
}

func pvString(p Protocol, v Version) string {
	var version string
	if v == nil {
		version = "0"
	} else {
		version = v.String()
	}
	return p.String() + "_" + version
}

func ProtocolVersionFromString(s string) (p Protocol, v Version) {
	temp := strings.Split(s, "_")
	if len(temp) != 2 {
		return ProtocolUnknown, UnknownVersion(0)
	}
	p = Protocol(temp[0])
	vv := VersionFromString(temp[1])
	switch p {
	case ProtocolCMPP:
		v = CMPPVersion(vv)
	case ProtocolSMPP:
		v = SMPPVersion(vv)
	case ProtocolSGIP:
		v = SGIPVersion(vv)
	case ProtocolSMGP:
		v = SMGPVersion(vv)
	default:
		v = UnknownVersion(0)
	}
	return
}

// VersionString 标准协议版本 Stringer。cmpp2.0，返回 V20
func VersionString(u Version) string {
	return VersionIntString(uint8(u.ToInt8()))
}

func VersionIntString(u uint8) string {
	return fmt.Sprintf("V%x", u)
}

// VersionFromString ...
func VersionFromString(s string) (version int8) {
	_, _ = fmt.Sscanf(s, "V%x", &version)
	return
}

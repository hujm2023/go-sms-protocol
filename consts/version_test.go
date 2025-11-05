package consts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocolVersionString(t *testing.T) {
	p, v := ProtocolCMPP, CMPPVersion2_0
	s := ProtocolVersionString(p, v)
	t.Log(s)
	p1, v1 := ProtocolVersionFromString(s)
	assert.Equal(t, p, p1)
	assert.Equal(t, v, v1)

	p2, v2 := ProtocolSGIP, SGIPVersion1_2
	s = ProtocolVersionString(p2, v2)
	t.Log(s)
	p3, v3 := ProtocolVersionFromString(s)
	assert.Equal(t, p2, p3)
	assert.Equal(t, v2, v3)

	p4, v4 := ProtocolSMGP, SMGPVersion3_0
	s = ProtocolVersionString(p4, v4)
	t.Log(s)
	p5, v5 := ProtocolVersionFromString(s)
	assert.Equal(t, p4, p5)
	assert.Equal(t, v4, v5)
}

func TestVersionString(t *testing.T) {
	assert.Equal(t, "V20", VersionString(CMPPVersion2_0))
	assert.Equal(t, "V21", VersionString(CMPPVersion2_1))
	assert.Equal(t, "V30", VersionString(CMPPVersion3_0))
	assert.Equal(t, "V34", VersionString(SMPPVersion3_4))
	assert.Equal(t, "V12", VersionString(SGIPVersion1_2))
	assert.Equal(t, "V34", VersionIntString(uint8(SMPPVersion3_4.ToInt8())))
	assert.Equal(t, "V20", VersionIntString(uint8(CMPPVersion2_0.ToInt8())))
	assert.Equal(t, "V30", VersionIntString(uint8(CMPPVersion3_0.ToInt8())))
	assert.Equal(t, "V21", VersionIntString(uint8(CMPPVersion2_1.ToInt8())))
	assert.Equal(t, "V12", VersionIntString(uint8(SGIPVersion1_2.ToInt8())))
}

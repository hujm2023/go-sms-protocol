package smpp34

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/smpp"
)

type GenericNackTestSuite struct {
	suite.Suite

	genericNackBytes []byte
}

func (g *GenericNackTestSuite) SetupTest() {
	g.genericNackBytes = []byte{
		0, 0, 0, 16, 128, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0,
	}
}

func (g *GenericNackTestSuite) TestGenericNack_IEncode() {
	genericNack := GenericNack{
		Header: smpp.Header{
			ID:       smpp.GENERIC_NACK,
			Status:   smpp.ESME_RINVCMDLEN,
			Sequence: 0,
			Length:   16,
		},
	}

	data, err := genericNack.IEncode()
	g.Nil(err)
	g.Equal(g.genericNackBytes, data)
}

func (g *GenericNackTestSuite) TestGenericNack_IDecode() {
	genericNack := new(GenericNack)
	g.Nil(genericNack.IDecode(g.genericNackBytes))
	g.Equal(genericNack.Header.ID, smpp.GENERIC_NACK)
	g.Equal(genericNack.Header.Status, smpp.ESME_RINVCMDLEN)
}

func (g *GenericNackTestSuite) TestGenericNack_SetSequenceID() {
	genericNack := new(GenericNack)
	g.Nil(genericNack.IDecode(g.genericNackBytes))
	g.Equal(genericNack.GetSequenceID(), uint32(0))
	genericNack.SetSequenceID(12345)
	g.Equal(genericNack.GetSequenceID(), uint32(12345))
}

func (g *GenericNackTestSuite) TestGenericNack_GetSequenceID() {
	genericNack := new(GenericNack)
	g.Nil(genericNack.IDecode(g.genericNackBytes))
	g.Equal(genericNack.GetSequenceID(), uint32(0))
}

func (g *GenericNackTestSuite) TestGenericNack_GetCommand() {
	genericNack := new(GenericNack)
	g.Nil(genericNack.IDecode(g.genericNackBytes))
	g.Equal(genericNack.GetCommand(), smpp.GENERIC_NACK)
}

func (g *GenericNackTestSuite) TestGenericNack_GenEmptyResponse() {
	genericNack := new(GenericNack)
	g.Nil(genericNack.IDecode(g.genericNackBytes))

	assert.Nil(g.T(), genericNack.GenEmptyResponse())
}

func TestGeneticNackSuite(t *testing.T) {
	suite.Run(t, new(GenericNackTestSuite))
}

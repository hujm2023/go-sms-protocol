package datacoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatin1(t *testing.T) {
	s := `Lorem ipsum ÅÆÇÈÉÊËÌÍÎÏDxÐÑÒÓÔÕÖ×ØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïFxðñòóôõö÷øùúûüýþÿ¡¢£¤¥¦§¨©ª«¬SHY®¯°±²³´µ¶·¸¹º» ¼ ½ ¾ ¿`

	data, err := Latin1(s).Encode()
	assert.Nil(t, err)
	t.Log(data)

	ss, err := Latin1(data).Decode()
	assert.Nil(t, err)
	t.Log(ss)

	assert.Equal(t, s, string(ss))
}

package datacoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGB18030(t *testing.T) {
	s := `[ByteDance] 我的头发长，天下我为王 12345 "ahaha" @^*((^`

	data, err := GB18030(s).Encode()
	assert.Nil(t, err)
	t.Log(data)

	ss, err := GB18030(data).Decode()
	assert.Nil(t, err)
	t.Log(ss)

	assert.Equal(t, s, string(ss))
}

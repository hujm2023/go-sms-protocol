package sgip12

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBind(t *testing.T) {
	type args struct {
		account string
		passwd  string
		nodeID  uint32
		seqID   uint32
	}
	tests := []struct {
		name     string
		args     args
		wantName string
	}{
		{
			name: "TestNewBind",
			args: args{
				account: "testAccount",
				passwd:  "testPasswd",
				nodeID:  1,
				seqID:   1,
			},
			wantName: "testAccount",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantName, NewBind(tt.args.account, tt.args.passwd, tt.args.nodeID, tt.args.seqID).Name, "NewBind(%v, %v, %v, %v)", tt.args.account, tt.args.passwd, tt.args.nodeID, tt.args.seqID)
		})
	}
}

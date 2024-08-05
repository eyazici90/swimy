package swim

import (
	"bytes"
)

const (
	pingMsgType = 1
)

type msg interface {
	bytes() []byte
}

type pingMsg struct {
	from string
}

func (p pingMsg) bytes() []byte {
	var buff bytes.Buffer
	buff.WriteByte(pingMsgType)
	buff.WriteString(p.from)
	return buff.Bytes()
}

type MembershipInfo struct {
}

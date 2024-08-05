package swim

import (
	"bytes"
	"net"
)

const (
	pingMsgType = 1
)

type pingMsg struct {
	from net.Addr
}

func (p pingMsg) bytes() []byte {
	var buff bytes.Buffer
	buff.WriteByte(pingMsgType)
	buff.WriteString(p.from.String())
	return buff.Bytes()
}

type MembershipInfo struct {
}

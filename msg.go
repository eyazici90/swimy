package swim

import (
	"bytes"
	"net"
)

const (
	pingMsgType    = 1
	joinReqMsgType = 2
)

type joinReq struct {
	from net.Addr
}

func (j joinReq) bytes() []byte {
	var buff bytes.Buffer
	buff.WriteByte(joinReqMsgType)
	buff.WriteString(j.from.String())
	return buff.Bytes()
}

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

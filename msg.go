package swim

import (
	"fmt"
	"net"
	"time"
)

var allMsgTypes = [...]string{
	unknownMsgType:  "unknown",
	pingMsgType:     "ping",
	joinReqMsgType:  "join-req",
	leaveReqMsgType: "leave-req",
}

const (
	unknownMsgType = iota
	pingMsgType
	joinReqMsgType
	leaveReqMsgType
)

type (
	joinReq struct {
		from net.Addr
	}
	leaveReq struct {
		from net.Addr
	}
	pingMsg struct {
		from net.Addr
	}
)

var (
	pingMsg  = []byte{pingMsgType}
	joinReq  = []byte{joinReqMsgType}
	leaveReq = []byte{leaveReqMsgType}
)

func (ms *Membership) stream(conn net.Conn) error {
	addr := conn.RemoteAddr()
	var msgType byte
	defer func() {
		ms.observer.received(allMsgTypes[msgType], addr.String())
	}()

	const sizeOfMsg uint8 = 16
	msg := make([]byte, sizeOfMsg)
	if _, err := conn.Read(msg[:]); err != nil {
		return fmt.Errorf("read from conn: %w", err)
	}

	msgType = msg[0]
	switch msgType {
	case pingMsgType:
		ms.setState(alive, addr)
	case joinReqMsgType:
		m := &Member{
			addr:  addr,
			state: alive,
			since: time.Now().UTC(),
		}
		ms.becomeMembers(m)
		ms.observer.onJoin(m)
	case leaveReqMsgType:
		ms.setState(left, addr)
		ms.observer.onLeave(nil)
	default:
		return fmt.Errorf("unknown msg type: %d", msgType)
	}
	return nil
}

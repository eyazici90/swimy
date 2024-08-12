package swim

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

func (j joinReq) encode() []byte {
	var buf bytes.Buffer
	buf.WriteByte(joinReqMsgType)
	buf.WriteString(j.from.String())
	return buf.Bytes()
}

func (l leaveReq) encode() []byte {
	var buf bytes.Buffer
	buf.WriteByte(leaveReqMsgType)
	buf.WriteString(l.from.String())
	return buf.Bytes()
}

func (p pingMsg) encode() []byte {
	var buf bytes.Buffer
	buf.WriteByte(pingMsgType)
	buf.WriteString(p.from.String())
	return buf.Bytes()
}

func (ms *Membership) ping(ctx context.Context, addr net.Addr) error {
	out := pingMsg{from: ms.me.Addr()}
	if err := sendToTCP(ctx, addr, out.encode()); err != nil {
		return fmt.Errorf("send to: %w", err)
	}
	ms.observer.pinged()
	return nil
}

func (ms *Membership) joinReq(ctx context.Context, addr net.Addr) error {
	out := joinReq{from: ms.me.Addr()}
	if err := sendToTCP(ctx, addr, out.encode()); err != nil {
		return fmt.Errorf("send to: %w", err)
	}
	return nil
}

func (ms *Membership) stream(rw io.ReadWriter) error {
	var (
		msgType byte
		addr    net.Addr
	)
	defer func() {
		ms.observer.received(allMsgTypes[msgType], addr.String())
	}()

	const sizeOfMsg = 16
	msg := [sizeOfMsg]byte{}
	if _, err := rw.Read(msg[:]); err != nil {
		return fmt.Errorf("read from conn: %w", err)
	}

	msgType = msg[0]
	addr, err := net.ResolveTCPAddr("tcp", string(msg[1:]))
	if err != nil {
		return fmt.Errorf("resolve tcp addr: %w", err)
	}

	switch msgType {
	case pingMsgType:
		ms.setAliveAddrs(addr)
	case joinReqMsgType:
		m := &Member{
			addr:  addr,
			state: alive,
			since: time.Now().UTC(),
		}
		ms.becomeMembers(m)
		ms.observer.onJoin(m)
	case leaveReqMsgType:
		ms.setLeaveAddr(addr)
		ms.observer.onLeave(nil)
	default:
		return fmt.Errorf("unkown msg type: %d", msgType)
	}
	return nil
}

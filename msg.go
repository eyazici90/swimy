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

func (ms *Membership) ping(ctx context.Context, addr net.Addr) error {
	conn, err := dial(ctx, addr)
	if err != nil {
		return fmt.Errorf("dial to addr: %w", err)
	}
	defer func() { _ = conn.Close() }()

	out := pingMsg{from: ms.me.Addr()}
	if err = writeTo(ctx, conn, out.bytes()); err != nil {
		return fmt.Errorf("write msg to conn: %w", err)
	}
	ms.observer.pinged()
	return nil
}

func (ms *Membership) joinReq(ctx context.Context, addr net.Addr) error {
	conn, err := dial(ctx, addr)
	if err != nil {
		return fmt.Errorf("dial to addr: %w", err)
	}
	defer func() { _ = conn.Close() }()

	out := joinReq{from: ms.me.Addr()}
	if err = writeTo(ctx, conn, out.bytes()); err != nil {
		return fmt.Errorf("write msg to conn: %w", err)
	}
	return nil
}

func (ms *Membership) stream(rw io.ReadWriter) error {
	var (
		msgType byte
		addr    net.Addr
	)
	defer func() {
		ms.observer.received(msgType, addr.String())
	}()

	msg := [16]byte{}
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
	default:
		return fmt.Errorf("unkown msg type: %d", msgType)
	}
	return nil
}

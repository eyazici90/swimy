package swim

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"
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

func (ms *Membership) ping(ctx context.Context, addr net.Addr) error {
	conn, err := dial(ctx, addr)
	if err != nil {
		return fmt.Errorf("dial to addr: %w", err)
	}
	defer func() { _ = conn.Close() }()

	msg := pingMsg{from: ms.me.Addr()}
	if err = writeTo(ctx, conn, msg.bytes()); err != nil {
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
	req := joinReq{from: ms.me.Addr()}
	if err = writeTo(ctx, conn, req.bytes()); err != nil {
		return fmt.Errorf("write msg to conn: %w", err)
	}
	return nil
}

func (ms *Membership) stream(rw io.ReadWriter) error {
	msg := [16]byte{}
	if _, err := rw.Read(msg[:]); err != nil {
		return fmt.Errorf("read from conn: %w", err)
	}

	msgType := msg[0]
	defer ms.observer.received()

	switch msgType {
	case pingMsgType:
		log.Printf("received ping from: %s", string(msg[1:]))
	case joinReqMsgType:
		addr, err := net.ResolveTCPAddr("tcp", string(msg[1:]))
		if err != nil {
			return fmt.Errorf("resolve tcp addr: %w", err)
		}
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

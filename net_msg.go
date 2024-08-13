package swim

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"time"
)

var allMsgTypes = [...]string{
	unknownMsgType:  "unknown",
	pingMsgType:     "ping",
	joinReqMsgType:  "join-req",
	leaveReqMsgType: "leave-req",
	errMsgType:      "error",
}

const (
	unknownMsgType = iota
	pingMsgType
	joinReqMsgType
	leaveReqMsgType
	errMsgType
)

type (
	joinReq struct {
		sender net.Addr
	}
	leaveReq struct {
		sender net.Addr
	}
	pingMsg struct {
		sender net.Addr
	}
	errMsg struct {
		sender net.Addr
		target net.Addr
	}
)

func (j joinReq) encode() []byte {
	var buf bytes.Buffer
	buf.WriteByte(joinReqMsgType)
	buf.WriteString(j.sender.String())
	return buf.Bytes()
}

func (l leaveReq) encode() []byte {
	var buf bytes.Buffer
	buf.WriteByte(leaveReqMsgType)
	buf.WriteString(l.sender.String())
	return buf.Bytes()
}

func (p pingMsg) encode() []byte {
	var buf bytes.Buffer
	buf.WriteByte(pingMsgType)
	buf.WriteString(p.sender.String())
	return buf.Bytes()
}

func (e errMsg) encode() []byte {
	var buf bytes.Buffer
	buf.WriteByte(errMsgType)
	buf.WriteString(e.sender.String())
	buf.WriteString(e.target.String())
	return buf.Bytes()
}

func (ms *Membership) ping(ctx context.Context, addr net.Addr) error {
	out := pingMsg{sender: ms.me.Addr()}
	if err := sendToTCP(ctx, addr, out.encode()); err != nil {
		return fmt.Errorf("send to: %w", err)
	}
	ms.observer.pinged()
	return nil
}

func (ms *Membership) joinReq(ctx context.Context, addr net.Addr) error {
	out := joinReq{sender: ms.me.Addr()}
	if err := sendToTCP(ctx, addr, out.encode()); err != nil {
		return fmt.Errorf("send to: %w", err)
	}
	return nil
}

func (ms *Membership) stream(ctx context.Context, conn net.Conn) error {
	var (
		addr    net.Addr
		msgType byte
	)
	defer func() {
		ms.observer.received(allMsgTypes[msgType], addr.String())
	}()

	bufConn := bufio.NewReader(conn)
	msgType, err := bufConn.ReadByte()
	if err != nil {
		return fmt.Errorf("read msg-type: %w", err)
	}

	const bufSize uint8 = 15
	buff := make([]byte, bufSize)
	n, err := bufConn.Read(buff)
	if err != nil {
		return fmt.Errorf("bufcon read: %w", err)
	}

	addr, err = net.ResolveTCPAddr("tcp", string(buff[:n]))
	if err != nil {
		return fmt.Errorf("resolve tcp addr: %w", err)
	}

	switch msgType {
	case pingMsgType:
		ms.setState(alive, addr)
	case joinReqMsgType:
		if ms.isAware(addr) {
			return nil
		}
		// broadcast join req to all others
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
	case errMsgType:
		if n, err = bufConn.Read(buff); err != nil {
			return fmt.Errorf("bufcon read: %w", err)
		}
		tAddr, err := net.ResolveTCPAddr("tcp", string(buff[:n]))
		if err != nil {
			return fmt.Errorf("resolve tcp addr: %w", err)
		}
		ms.setState(dead, tAddr)
	default:
		return fmt.Errorf("unknown msg type: %d", msgType)
	}
	return nil
}
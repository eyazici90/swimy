package swimy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"unsafe"
)

var allMsgTypes = [...]string{
	unknownMsgType:          "unknown",
	ackRespMsgType:          "ack-resp",
	pingMsgType:             "ping",
	joinReqMsgType:          "join-req",
	joinReqBroadcastMsgType: "join-req-broadcast",
	leaveReqMsgType:         "leave-req",
	errMsgType:              "error",
}

const (
	unknownMsgType byte = iota
	ackRespMsgType
	pingMsgType
	joinReqMsgType
	joinReqBroadcastMsgType
	leaveReqMsgType
	errMsgType
)

type (
	joinReq struct {
		sender net.Addr
	}
	joinReqBroadcast struct {
		target net.Addr
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

func (j joinReqBroadcast) encode() []byte {
	var buf bytes.Buffer
	buf.WriteByte(joinReqBroadcastMsgType)
	buf.WriteString(j.target.String())
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

func (ms *Membership) pingACK(ctx context.Context, addr net.Addr) error {
	msg := pingMsg{sender: ms.me.Addr()}
	resp := make([]byte, 1)
	if err := sendReceiveTCP(ctx, addr, msg.encode(), resp); err != nil {
		return fmt.Errorf("send & wait ack: %w", err)
	}
	if resp[0] != ackRespMsgType {
		return fmt.Errorf("received byte is not ack")
	}
	ms.observer.pinged()
	return nil
}

func (ms *Membership) ack(w io.Writer) error {
	if _, err := w.Write([]byte{ackRespMsgType}); err != nil {
		return fmt.Errorf("write ack-resp: %w", err)
	}
	return nil
}

func (ms *Membership) joinReq(ctx context.Context, addr net.Addr) error {
	out := joinReq{sender: ms.me.Addr()}
	if err := sendTCP(ctx, addr, out.encode()); err != nil {
		return fmt.Errorf("send to: %w", err)
	}
	return nil
}

func (ms *Membership) stream(ctx context.Context, conn io.ReadWriter) error {
	var (
		sender  net.Addr
		msgType byte
	)
	defer func() {
		ms.observer.received(ctx, allMsgTypes[msgType], sender)
	}()

	bufConn := bufio.NewReader(conn)
	msgType, err := bufConn.ReadByte()
	if err != nil {
		return fmt.Errorf("read msg-type: %w", err)
	}

	const bufSize uint8 = 15
	buff := make([]byte, bufSize)
	sender, err = parseSender(bufConn, buff)
	if err != nil {
		return fmt.Errorf("parse sender: %w", err)
	}
	switch msgType {
	case pingMsgType:
		ms.setState(statusAlive, sender)
		if err := ms.ack(conn); err != nil {
			return err
		}
	case joinReqMsgType:
		m := newAliveMember(sender)
		ms.becomeMembers(m)
		ms.observer.onJoin(ctx, sender)
		msg := joinReqBroadcast{target: sender}
		if err = ms.broadCastToLives(ctx, msg.encode(), sender); err != nil {
			return fmt.Errorf("broadcast join-req:%w", err)
		}
	case joinReqBroadcastMsgType:
		m := newAliveMember(sender)
		ms.becomeMembers(m)
		ms.observer.onJoin(ctx, sender)
	case leaveReqMsgType:
		ms.setState(statusLeft, sender)
		ms.observer.onLeave(ctx, sender)
	case errMsgType:
		deadAddr, err := parseDeadAddr(bufConn, buff)
		if err != nil {
			return fmt.Errorf("parse dead addr: %w", err)
		}
		ms.setState(statusDead, deadAddr)
		ms.observer.onLeave(ctx, deadAddr)
	default:
		return fmt.Errorf("unknown msg type: %d", msgType)
	}
	return nil
}

func parseSender(r io.Reader, buff []byte) (net.Addr, error) {
	n, err := r.Read(buff)
	if err != nil {
		return nil, fmt.Errorf("bufcon read: %w", err)
	}
	b := buff[:n]
	str := *(*string)(unsafe.Pointer(&b))

	sender, err := net.ResolveTCPAddr("tcp", str)
	if err != nil {
		return nil, fmt.Errorf("resolve tcp addr: %w", err)
	}
	return sender, nil
}

func parseDeadAddr(r io.Reader, buff []byte) (net.Addr, error) {
	n, err := r.Read(buff)
	if err != nil {
		return nil, fmt.Errorf("bufcon read: %w", err)
	}
	b := buff[:n]
	str := *(*string)(unsafe.Pointer(&b))

	addr, err := net.ResolveTCPAddr("tcp", str)
	if err != nil {
		return nil, fmt.Errorf("resolve dead tcp addr: %w", err)
	}
	return addr, nil
}

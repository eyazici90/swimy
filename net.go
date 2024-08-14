package swim

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const (
	networkTCP            = "tcp"
	defaultTCPConnTimeout = time.Millisecond * 100
)

type netTCP struct {
	tcpLn  *net.TCPListener
	stream func(context.Context, io.ReadWriter) error
}

func newNetTCP(port uint16, stream func(context.Context, io.ReadWriter) error) (*netTCP, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, fmt.Errorf("resolve tcp addr: %w", err)
	}
	tcpLn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen tcp: %w", err)
	}

	return &netTCP{
		tcpLn:  tcpLn,
		stream: stream,
	}, nil
}

func (nt *netTCP) listen(ctx context.Context) error {
	defer func() {
		_ = nt.tcpLn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("listen tcp: %w", ctx.Err())
		default:
			conn, err := nt.tcpLn.AcceptTCP()
			if err != nil {
				return fmt.Errorf("accept tcp: %w", err)
			}
			go nt.handleConn(ctx, conn)
		}
	}
}

func (nt *netTCP) handleConn(ctx context.Context, conn net.Conn) {
	defer func() {
		if rvr := recover(); rvr != nil {
			log.Printf("recover from: %s", rvr)
		}
	}()
	defer func() {
		_ = conn.Close()
	}()

	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	default:
		if err = conn.SetDeadline(until100ms()); err != nil {
			break
		}
		err = nt.stream(ctx, conn)
	}
	if err != nil {
		log.Printf("handle conn: %s", err)
	}
}

func sendTCP(ctx context.Context, addr net.Addr, out []byte) error {
	nd := &net.Dialer{Timeout: defaultTCPConnTimeout}
	conn, err := nd.DialContext(ctx, networkTCP, addr.String())
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err = conn.SetWriteDeadline(until100ms()); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}
	select {
	case <-ctx.Done():
		return fmt.Errorf("send msg: %w", ctx.Err())
	default:
		if _, err = conn.Write(out); err != nil {
			return fmt.Errorf("write to conn: %w", err)
		}
	}
	return nil
}

func sendReceiveTCP(ctx context.Context, addr net.Addr, out, in []byte) error {
	nd := &net.Dialer{Timeout: defaultTCPConnTimeout}
	conn, err := nd.DialContext(ctx, networkTCP, addr.String())
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err = conn.SetDeadline(until100ms()); err != nil {
		return fmt.Errorf("set deadline: %w", err)
	}
	select {
	case <-ctx.Done():
		return fmt.Errorf("send rreceive msg: %w", ctx.Err())
	default:
		if _, err = conn.Write(out); err != nil {
			return fmt.Errorf("write to conn: %w", err)
		}
		if _, err = conn.Read(in); err != nil {
			return fmt.Errorf("read conn: %w", err)
		}
	}
	return nil
}

func until100ms() time.Time {
	return time.Now().Add(defaultTCPConnTimeout) // make it configurable later
}

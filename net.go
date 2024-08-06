package swim

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
)

type netTCP struct {
	protocolVersion uint8
	listener        *net.TCPListener
	stream          func(rw io.ReadWriter) error
}

func newNetTCP(sr func(rw io.ReadWriter) error) (*netTCP, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("resolve tcp addr: %w", err)
	}
	tcpLn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen tcp: %w", err)
	}

	return &netTCP{
		listener: tcpLn,
		stream:   sr,
	}, nil
}

func (nt *netTCP) listen(ctx context.Context) error {
	defer func() {
		_ = nt.listener.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("listen tcp: %w", ctx.Err())
		default:
			conn, err := nt.listener.AcceptTCP()
			if err != nil {
				return fmt.Errorf("accepting tcp: %w", err)
			}
			go handleConn(ctx, conn, nt.stream)
		}
	}
}

func handleConn(ctx context.Context, conn net.Conn, sr func(reader io.ReadWriter) error) {
	defer func() {
		_ = conn.Close()
	}()

	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	default:
		err = sr(conn)
	}
	if err != nil {
		log.Printf("handle conn: %s", ctx.Err())
	}
}

func dial(ctx context.Context, addr net.Addr) (net.Conn, error) {
	var nd net.Dialer
	return nd.DialContext(ctx, "tcp", addr.String())
}

func writeTo(ctx context.Context, conn net.Conn, msg []byte) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("write msg: %w", ctx.Err())
	default:
		if _, err := conn.Write(msg); err != nil {
			return fmt.Errorf("write to conn: %w", err)
		}
	}
	return nil
}

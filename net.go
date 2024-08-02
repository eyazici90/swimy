package swim

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

func newTCPln() (*net.TCPListener, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("resolve tcp addr: %w", err)
	}
	tcpLn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen tcp: %w", err)
	}
	return tcpLn, nil
}

func listen(ctx context.Context, tcpLn *net.TCPListener) error {
	defer func() {
		_ = tcpLn.Close()
	}()

	_ = tcpLn.SetDeadline(time.Now().Add(time.Millisecond * 50))
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("listen tcp: %w", ctx.Err())
		default:
			conn, err := tcpLn.AcceptTCP()
			if err != nil {
				return fmt.Errorf("accepting tcp: %w", err)
			}
			go streamConn(ctx, conn)
		}
	}
}

func streamConn(ctx context.Context, conn net.Conn) {
	for {
		var buff []byte
		select {
		case <-ctx.Done():
			log.Printf("streaming on conn: %s", ctx.Err())
		default:
			if _, err := conn.Read(buff); err != nil {
				log.Printf("read from conn: %s", err)
			}
		}
	}
}

func dial(ctx context.Context, addr net.Addr) (net.Conn, error) {
	var dr net.Dialer
	return dr.DialContext(ctx, "tcp", addr.String())
}

func writeMsg(ctx context.Context, conn net.Conn, msg []byte) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("send msg: %w", ctx.Err())
	default:
		if _, err := conn.Write(msg); err != nil {
			return fmt.Errorf("write to conn: %w", err)
		}
	}
	return nil
}

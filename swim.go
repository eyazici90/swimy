package swim

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Membership struct {
	cfg     *Config
	metrics Metrics

	membersMu sync.RWMutex
	me        *Member
	others    []*Member

	stop func()
}

func New(cfg *Config) (*Membership, error) {
	tcpLn, err := newTCPln()
	if err != nil {
		return nil, fmt.Errorf("initializing tcp listener: %w", err)
	}

	var ms Membership
	ms.me = &Member{
		addr:  tcpLn.Addr(),
		state: alive,
	}

	ctx, cancel := context.WithCancel(context.Background())
	ms.stop = cancel

	go func() {
		_ = ms.schedule(ctx, time.Millisecond*10, ms.gossip)
	}()
	go func() {
		_ = listen(ctx, tcpLn, ms.readStream)
	}()
	return &ms, nil
}

func (ms *Membership) Join(ctx context.Context, existing ...string) error {
	for _, exist := range existing {
		addr, err := net.ResolveTCPAddr("tcp", exist)
		if err != nil {
			return fmt.Errorf("resolve tcp addr: %w", err)
		}
		if err = ms.sendPing(ctx, addr); err != nil {
			return fmt.Errorf("send ping: %w", err)
		}

		m := &Member{
			addr:  addr,
			state: alive,
			since: time.Now().UTC(),
		}
		ms.membersMu.Lock()
		ms.others = append(ms.others, m)
		ms.membersMu.Unlock()
	}
	return nil
}

func (ms *Membership) Leave() error {
	return nil
}

func (ms *Membership) Stop() {
	ms.stop()
}

func (ms *Membership) Me() *Member {
	return ms.me
}

func (ms *Membership) Metrics() Metrics {
	return ms.metrics
}

func (ms *Membership) schedule(ctx context.Context, interval time.Duration, fn func(ctx context.Context) error) error {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := fn(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return fmt.Errorf("schedule :%w", ctx.Err())
		}
	}
}

func (ms *Membership) sendPing(ctx context.Context, addr net.Addr) error {
	conn, err := dial(ctx, addr)
	if err != nil {
		return fmt.Errorf("dial to addr: %w", err)
	}
	if err = writeMsg(ctx, conn, []byte{pingMsg}); err != nil {
		return fmt.Errorf("write msg to conn: %w", err)
	}
	atomic.AddUint32(&ms.metrics.SentNum, 1)
	return nil
}

func (ms *Membership) readStream(r io.Reader) error {
	buffR := bufio.NewReader(r)
	buff := [1]byte{}
	if _, err := io.ReadFull(buffR, buff[:]); err != nil {
		return fmt.Errorf("read from conn: %w", err)
	}
	atomic.AddUint32(&ms.metrics.ReceivedNum, 1)
	return nil
}

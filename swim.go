package swim

import (
	"context"
	"fmt"
	"io"
	"log"
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
	var ms Membership
	nTCP, err := newNetTCP(ms.stream)
	if err != nil {
		return nil, fmt.Errorf("initializing tcp listener: %w", err)
	}

	ms.me = &Member{
		addr:  nTCP.listener.Addr(),
		state: alive,
	}

	ctx, cancel := context.WithCancel(context.Background())
	ms.stop = cancel
	setDefaults(&cfg)
	go func() {
		if err := ms.schedule(ctx, cfg.GossipInterval, ms.gossip); err != nil {
			log.Print(err)
		}
	}()
	go func() {
		if err := nTCP.listen(ctx); err != nil {
			log.Print(err)
		}
	}()
	return &ms, nil
}

func (ms *Membership) Join(ctx context.Context, existing ...string) error {
	for _, exist := range existing {
		addr, err := net.ResolveTCPAddr("tcp", exist)
		if err != nil {
			return fmt.Errorf("resolve tcp addr: %w", err)
		}
		if err = ms.ping(ctx, addr); err != nil {
			return fmt.Errorf("send ping: %w", err)
		}

		m := &Member{
			addr:  addr,
			state: alive,
			since: time.Now().UTC(),
		}
		ms.becomeMembers(m)
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

func (ms *Membership) ping(ctx context.Context, addr net.Addr) error {
	conn, err := dial(ctx, addr)
	if err != nil {
		return fmt.Errorf("dial to addr: %w", err)
	}
	msg := pingMsg{from: ms.me.Addr()}
	if err = writeMsg(ctx, conn, msg.bytes()); err != nil {
		return fmt.Errorf("write msg to conn: %w", err)
	}
	atomic.AddUint32(&ms.metrics.SentNum, 1)
	return nil
}

func (ms *Membership) stream(rw io.ReadWriter) error {
	msg := [16]byte{}
	if _, err := rw.Read(msg[:]); err != nil {
		return fmt.Errorf("read from conn: %w", err)
	}

	switch msg[0] {
	case pingMsgType:
		log.Printf("received ping from: %s", string(msg[1:]))
	default:
		log.Print("unknown msg type!!!")
	}
	atomic.AddUint32(&ms.metrics.ReceivedNum, 1)
	return nil
}

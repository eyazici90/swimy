package swim

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Membership struct {
	cfg      *Config
	observer observation

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
	setDefaults(&cfg)
	ms.observer = observation{
		onJoinCallback:  cfg.OnJoin,
		onLeaveCallback: cfg.OnLeave,
	}

	ctx, cancel := context.WithCancel(context.Background())
	ms.stop = cancel
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
		if err = ms.joinReq(ctx, addr); err != nil {
			return fmt.Errorf("send ping: %w", err)
		}

		m := &Member{
			addr:  addr,
			state: alive,
			since: time.Now().UTC(),
		}
		ms.becomeMembers(m)
		ms.observer.onJoin(m)
	}
	return nil
}

func (ms *Membership) Leave(ctx context.Context) error {
	out := leaveReq{from: ms.me.Addr()}
	if err := ms.broadCast(ctx, out.encode()); err != nil {
		return fmt.Errorf("broadcasting :%w", err)
	}
	ms.observer.onLeave(ms.Me())
	return nil
}

func (ms *Membership) Stop() {
	ms.stop()
}

func (ms *Membership) Me() *Member {
	return ms.me
}

func (ms *Membership) Metrics() Metrics {
	return ms.observer.metrics
}

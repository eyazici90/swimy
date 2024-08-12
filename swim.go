package swim

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
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
	setDefaults(&cfg)

	var ms Membership
	ms.cfg = cfg
	nTCP, err := newNetTCP(cfg.Port, ms.stream)
	if err != nil {
		return nil, fmt.Errorf("initializing tcp listener: %w", err)
	}
	ms.me = &Member{
		addr:  nTCP.tcpLn.Addr(),
		state: alive,
	}
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
		if err = sendToTCP(ctx, addr, joinReq); err != nil {
			return fmt.Errorf("send to: %w", err)
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
	if err := ms.broadCast(ctx, leaveReq); err != nil {
		return fmt.Errorf("broadcast :%w", err)
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
	return Metrics{
		ActiveMembers: atomic.LoadUint32(&ms.observer.metrics.ActiveMembers),
		SentNum:       atomic.LoadUint32(&ms.observer.metrics.SentNum),
		ReceivedNum:   atomic.LoadUint32(&ms.observer.metrics.ReceivedNum),
	}
}

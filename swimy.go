package swimy

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Membership struct {
	cfg      *Config
	observer *defaultObserver

	membersMu sync.RWMutex
	me        Member
	others    map[net.Addr]Member

	stop func()
}

func New(cfg *Config) (*Membership, error) {
	setDefaults(&cfg)
	setUpSlog(os.Stdout)
	ms := Membership{
		cfg:    cfg,
		others: make(map[net.Addr]Member),
	}

	nTCP, err := newNetTCP(cfg.Port, ms.stream)
	if err != nil {
		return nil, fmt.Errorf("new tcp listener: %w", err)
	}
	me := Member{
		addr:  nTCP.tcpLn.Addr(),
		state: statusAlive,
	}
	ms.me = me
	ms.observer = &defaultObserver{
		me:              me.Addr(),
		onJoinCallback:  ms.cfg.OnJoin,
		onLeaveCallback: ms.cfg.OnLeave,
	}

	ctx, cancel := context.WithCancel(context.Background())
	ms.stop = cancel
	go func() {
		if err := ms.schedule(ctx, cfg.GossipInterval, ms.gossip); err != nil {
			slog.ErrorContext(ctx, err.Error())
		}
	}()
	go func() {
		if err := nTCP.listen(ctx); err != nil {
			slog.ErrorContext(ctx, err.Error())
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
			return fmt.Errorf("join req: %w", err)
		}

		m := Member{
			addr:  addr,
			state: statusAlive,
			since: time.Now().UTC(),
		}
		ms.becomeMembers(m)
	}
	return nil
}

func (ms *Membership) Leave(ctx context.Context) error {
	req := leaveReq{sender: ms.me.Addr()}
	if err := ms.broadCastToLives(ctx, req.encode()); err != nil {
		return fmt.Errorf("broadcast leave-req :%w", err)
	}
	return nil
}

func (ms *Membership) Stop() {
	ms.stop()
	ms.observer.onStop()
}

func (ms *Membership) Members() []Member {
	lives := ms.alives()
	lives = append(lives, ms.me)
	return lives
}

func (ms *Membership) Me() Member {
	return ms.me
}

func (ms *Membership) Metrics() Metrics {
	return Metrics{
		ActiveMembers: atomic.LoadUint32(&ms.observer.metrics.ActiveMembers),
		SentNum:       atomic.LoadUint32(&ms.observer.metrics.SentNum),
		ReceivedNum:   atomic.LoadUint32(&ms.observer.metrics.ReceivedNum),
	}
}

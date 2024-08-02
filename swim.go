package swim

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Membership struct {
	cfg *Config

	memberMu sync.RWMutex
	me       member
	others   []member

	stop func()
}

func New() (*Membership, error) {
	tcpLn, err := newTCPln()
	if err != nil {
		return nil, fmt.Errorf("initializing tcp listener: %w", err)
	}

	var mms Membership
	mms.me = member{
		addr:  tcpLn.Addr(),
		state: alive,
	}

	ctx, cancel := context.WithCancel(context.Background())
	mms.stop = cancel

	go func() {
		_ = mms.schedule(ctx, time.Millisecond*10, mms.gossip)
	}()
	go func() {
		_ = listen(ctx, tcpLn)
	}()
	return &mms, nil
}

func (ms *Membership) Join(existing ...string) error {
	return nil
}

func (ms *Membership) Leave() error {
	return nil
}

func (ms *Membership) Stop() {
	ms.stop()
}

func (ms *Membership) schedule(ctx context.Context, interval time.Duration, fn func() error) error {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := fn(); err != nil {
				return err
			}
		case <-ctx.Done():
			return fmt.Errorf("schedule :%w", ctx.Err())
		}
	}
}

func (ms *Membership) gossip() error {
	return nil
}

func (ms *Membership) broadCast(failure any) error {
	return nil
}

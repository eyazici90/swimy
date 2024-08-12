package swim

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"
)

func (ms *Membership) gossip(ctx context.Context) error {
	target, ok := ms.randomSelect()
	if !ok {
		return nil
	}
	if err := ms.ping(ctx, target.addr); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	ms.setAlives(target)
	return nil
}

func (ms *Membership) randomSelect() (*Member, bool) {
	ms.membersMu.RLock()
	defer ms.membersMu.RUnlock()

	if len(ms.others) <= 0 {
		return nil, false
	}

	var toSend []*Member
	for _, m := range ms.others {
		if m.state == alive || m.state == suspect {
			toSend = append(toSend, m)
		}
	}

	num := rand.Int() % len(toSend) // rnd choice
	return toSend[num], true
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

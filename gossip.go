package swim

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"
)

func (ms *Membership) gossip(ctx context.Context) error {
	target, ok := ms.rndTargetSelect()
	if !ok {
		return nil
	}
	if err := ms.ping(ctx, target.Addr()); err != nil {
		// add suspect mechanism here to reduce false positives
		// mark as suspect,
		// forward it to someone else to send ping
		// if still no, mark as dead & disseminate
		// ms.setState(suspect, target.Addr())
		ms.setState(dead, target.Addr())
		out := errMsg{sender: ms.Me().Addr(), target: target.Addr()}
		if berr := ms.broadCastToLives(ctx, out.encode()); berr != nil {
			log.Printf("Error: broadcasting dead member: %s from: %s", berr, ms.Me().Addr())
		}
		return nil
	}
	ms.observer.pinged()
	ms.setAlives(target)
	return nil
}

func (ms *Membership) rndTargetSelect() (*Member, bool) {
	ms.membersMu.RLock()
	defer ms.membersMu.RUnlock()

	if len(ms.others) == 0 {
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

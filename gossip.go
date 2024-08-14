package swim

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"sync"
	"time"
)

func (ms *Membership) gossip(ctx context.Context) error {
	targets, found := ms.rndTargets()
	if !found {
		return nil
	}
	n := len(targets)
	errCh := make(chan error, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for target := range targets {
		target := target
		go func() {
			defer wg.Done()
			if err := ms.ping(ctx, target.Addr()); err != nil {
				if errors.Is(err, context.Canceled) {
					errCh <- err
					return
				}
				// add suspect mechanism here to reduce false positives
				// mark as suspect,
				// forward it to someone else to send indirect ping
				// if still no, mark as dead & disseminate
				// ms.setState(suspect, target.Addr())
				ms.setState(dead, target.Addr())
				out := errMsg{sender: ms.Me().Addr(), target: target.Addr()}
				if berr := ms.broadCastToLives(ctx, out.encode()); berr != nil {
					log.Printf("Error: broadcasting dead member: %s me: %s", errors.Join(err, berr), ms.Me().Addr())
				}
				return
			}
			ms.observer.pinged()
			ms.setAlives(target)
		}()
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (ms *Membership) rndTargets() (map[*Member]struct{}, bool) {
	ms.membersMu.RLock()
	defer ms.membersMu.RUnlock()

	var possibles []*Member
	for _, m := range ms.others {
		if m.state == alive || m.state == suspect {
			possibles = append(possibles, m)
		}
	}
	if len(possibles) == 0 {
		return nil, false
	}

	const percentage = 100
	total := uint32(math.Ceil(float64(ms.cfg.GossipRatio) / float64(percentage)))
	targets := make(map[*Member]struct{}, total)
	for total > 0 {
		num := rand.Int() % len(possibles) // rnd choice
		selected := possibles[num]
		if _, ok := targets[selected]; ok {
			continue
		}
		targets[selected] = struct{}{}
		total--
	}
	return targets, true
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

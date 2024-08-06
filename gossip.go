package swim

import (
	"context"
	"fmt"
	"math/rand/v2"
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

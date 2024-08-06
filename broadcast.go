package swim

import (
	"context"
	"errors"
	"sync"
)

func (ms *Membership) broadCast(ctx context.Context, msg []byte) error {
	ms.membersMu.RLock()
	others := ms.others
	ms.membersMu.RUnlock()

	var errs []error
	var wg sync.WaitGroup
	n := len(others)
	wg.Add(n)
	for _, other := range others {
		other := other
		go func() {
			defer wg.Done()
			if err := sendToTCP(ctx, other.Addr(), msg); err != nil {
				errs = append(errs, err)
			}
		}()

	}
	wg.Wait()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

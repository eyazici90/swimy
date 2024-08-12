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

	n := len(others)
	errCh := make(chan error, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for _, other := range others {
		other := other
		go func() {
			defer wg.Done()
			if err := sendToTCP(ctx, other.Addr(), msg); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

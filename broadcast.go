package swim

import (
	"context"
	"errors"
	"sync"
)

func (ms *Membership) broadCastToLives(ctx context.Context, msg []byte) error {
	alives := ms.alives()
	n := len(alives)
	errCh := make(chan error, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for _, m := range alives {
		m := m
		go func() {
			defer wg.Done()
			if err := sendToTCP(ctx, m.Addr(), msg); err != nil {
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

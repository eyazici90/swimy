package swimy

import (
	"context"
	"errors"
	"net"
	"sync"
)

func (ms *Membership) broadCastToLives(ctx context.Context, msg []byte, excludes ...net.Addr) error {
	lives := ms.alives(excludes...)
	n := len(lives)
	errCh := make(chan error, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for _, m := range lives {
		m := m
		go func() {
			defer wg.Done()
			if err := sendTCP(ctx, m.Addr(), msg); err != nil {
				errCh <- err
			}
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

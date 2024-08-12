package swim

import (
	"context"
	"errors"
	"net"
	"sync"
)

func (ms *Membership) broadCast(ctx context.Context, msg []byte) error {
	alives := ms.alives()
	n := len(alives)
	errCh := make(chan error, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for _, m := range alives {
		go func(addr net.Addr) {
			defer wg.Done()
			if err := sendToTCP(ctx, addr, msg); err != nil {
				errCh <- err
			}
		}(m.Addr())
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

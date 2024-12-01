package swimy_test

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/eyazici90/swimy"
	"github.com/stretchr/testify/require"
)

func TestSwimy_Join(t *testing.T) {
	ctx := context.Background()

	var joinNum atomic.Uint32
	cfg := swimy.DefaultConfig()
	cfg.OnJoin = func(addr net.Addr) {
		joinNum.Add(1)
	}
	ms1, err := swimy.New(cfg)
	require.NoError(t, err)
	defer ms1.Stop()

	ms2, err := swimy.New(nil)
	require.NoError(t, err)
	defer ms2.Stop()

	err = ms2.Join(ctx, ms1.Me().Addr().String())
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return joinNum.Load() == 1
	}, time.Millisecond*150, time.Millisecond*20)
}

func TestSwimy_Leave(t *testing.T) {
	ctx := context.Background()

	var leftNum atomic.Uint32
	cfg := swimy.DefaultConfig()
	cfg.OnLeave = func(addr net.Addr) {
		leftNum.Add(1)
	}
	ms1, err := swimy.New(cfg)
	require.NoError(t, err)
	defer ms1.Stop()

	ms2, err := swimy.New(nil)
	require.NoError(t, err)
	defer ms2.Stop()

	err = ms2.Join(ctx, ms1.Me().Addr().String())
	require.NoError(t, err)

	err = ms2.Leave(ctx)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return leftNum.Load() == 1
	}, time.Millisecond*150, time.Millisecond*20)
}

func TestSwimy_Dead(t *testing.T) {
	ctx := context.Background()

	ms1, err := swimy.New(nil)
	require.NoError(t, err)
	defer ms1.Stop()

	ms2, err := swimy.New(nil)
	require.NoError(t, err)

	err = ms2.Join(ctx, ms1.Me().Addr().String())
	require.NoError(t, err)
	defer ms2.Stop()

	ms3, err := swimy.New(nil)
	require.NoError(t, err)
	err = ms3.Join(ctx, ms1.Me().Addr().String())
	require.NoError(t, err)
	<-time.After(time.Millisecond * 20)
	ms3.Stop()

	<-time.After(time.Millisecond * 150)
}

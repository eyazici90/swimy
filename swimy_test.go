package swimy_test

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/eyazici90/swimy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwimy_Join(t *testing.T) {
	ctx := context.Background()

	var joined atomic.Bool
	cfg := swimy.DefaultConfig()
	cfg.OnJoin = func(addr net.Addr) {
		joined.Store(true)
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
		return joined.Load() && len(ms1.Members()) > 1
	}, time.Millisecond*150, time.Millisecond*20)
}

func TestSwimy_Leave(t *testing.T) {
	ctx := context.Background()

	var left atomic.Bool
	cfg := swimy.DefaultConfig()
	cfg.OnLeave = func(addr net.Addr) {
		left.Store(true)
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
		return left.Load()
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
	assert.GreaterOrEqual(t, len(ms1.Members()), 3)
	<-time.After(time.Millisecond * 20)
	ms3.Stop()
	<-time.After(time.Millisecond * 150)
	assert.LessOrEqual(t, len(ms1.Members()), 2)
}

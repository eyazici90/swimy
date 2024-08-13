package swim_test

import (
	"context"
	"testing"
	"time"

	"github.com/eyazici90/swim"
	"github.com/stretchr/testify/require"
)

func TestSwim_Join(t *testing.T) {
	ctx := context.Background()

	ms1, err := swim.New(nil)
	require.NoError(t, err)
	defer ms1.Stop()

	ms2, err := swim.New(nil)
	require.NoError(t, err)
	defer ms2.Stop()

	err = ms2.Join(ctx, ms1.Me().Addr().String())
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return ms1.Metrics().ReceivedNum > uint32(0) && ms2.Metrics().SentNum > uint32(0)
	}, time.Millisecond*150, time.Millisecond*20)
}

func TestSwim_Leave(t *testing.T) {
	ctx := context.Background()

	ms1, err := swim.New(nil)
	require.NoError(t, err)
	defer ms1.Stop()

	ms2, err := swim.New(nil)
	require.NoError(t, err)
	defer ms2.Stop()

	err = ms2.Join(ctx, ms1.Me().Addr().String())
	require.NoError(t, err)

	err = ms2.Leave(ctx)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return ms1.Metrics().ActiveMembers == 0 && ms2.Metrics().ActiveMembers == 0
	}, time.Millisecond*150, time.Millisecond*20)
}

func TestSwim_Dead(t *testing.T) {
	ctx := context.Background()

	ms1, err := swim.New(nil)
	require.NoError(t, err)
	defer ms1.Stop()

	ms2, err := swim.New(nil)
	require.NoError(t, err)

	err = ms2.Join(ctx, ms1.Me().Addr().String())
	require.NoError(t, err)
	defer ms2.Stop()

	ms3, err := swim.New(nil)
	require.NoError(t, err)
	err = ms3.Join(ctx, ms1.Me().Addr().String())
	require.NoError(t, err)
	<-time.After(time.Millisecond * 20)
	ms3.Stop()

	<-time.After(time.Millisecond * 150)
}

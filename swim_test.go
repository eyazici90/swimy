package swim_test

import (
	"context"
	"testing"
	"time"

	"github.com/eyazici90/swim"
	"github.com/stretchr/testify/assert"
)

func TestSwim_New(t *testing.T) {
	ctx := context.Background()

	ms1, err := swim.New(nil)
	assert.NoError(t, err)

	ms2, err := swim.New(nil)
	err = ms2.Join(ctx, ms1.Me().Addr().String())
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 10)

	assert.Greater(t, ms1.Metrics().ReceivedNum, uint32(0))
	assert.Greater(t, ms2.Metrics().SentNum, uint32(0))
}

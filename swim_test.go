package swim_test

import (
	"testing"

	"github.com/eyazici90/swim"
	"github.com/stretchr/testify/assert"
)

func TestSwim_New(t *testing.T) {
	ms, err := swim.New()

	assert.NoError(t, err)
	assert.NotNil(t, ms)
}

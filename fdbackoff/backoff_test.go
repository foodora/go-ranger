package fdbackoff_test

import (
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
	"github.com/stretchr/testify/assert"
)

func TestConstant(t *testing.T) {
	constBackoff := fdbackoff.Constant(2 * time.Second)

	assert.Equal(t, 0*time.Second, constBackoff(0))
	assert.Equal(t, 2*time.Second, constBackoff(1))
	assert.Equal(t, 2*time.Second, constBackoff(2))
	assert.Equal(t, 2*time.Second, constBackoff(3))
}

func TestLinear(t *testing.T) {
	linearBackoff := fdbackoff.Linear(2 * time.Second)

	assert.Equal(t, 0*time.Second, linearBackoff(0))
	assert.Equal(t, 2*time.Second, linearBackoff(1))
	assert.Equal(t, 4*time.Second, linearBackoff(2))
	assert.Equal(t, 6*time.Second, linearBackoff(3))
}

func TestExponential(t *testing.T) {
	expBackoff := fdbackoff.Exponential(2 * time.Second)

	assert.Equal(t, 0*time.Second, expBackoff(0))
	assert.Equal(t, 2*time.Second, expBackoff(1))
	assert.Equal(t, 4*time.Second, expBackoff(2))
	assert.Equal(t, 8*time.Second, expBackoff(3))
}

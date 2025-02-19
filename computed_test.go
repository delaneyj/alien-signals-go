package alien_test

import (
	"testing"

	alien "github.com/delaneyj/alien-signals-go"
	"github.com/stretchr/testify/assert"
)

// from README
func TestBasics(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})
	count := alien.Signal(rs, 1)
	doubleCount := alien.Computed(rs, func(oldValue int) int {
		return count.Value() * 2
	})

	stopEffect := alien.Effect(rs, func() error {
		return nil
	})
	defer stopEffect()

	assert.Equal(t, 2, doubleCount.Value())

	count.SetValue(2)

	assert.Equal(t, 4, doubleCount.Value())
}

// should correctly propagate changes through computed signals
func TestComputed(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})

	src := alien.Signal(rs, 0)
	c1 := alien.Computed(rs, func(oldValue int) int {
		return src.Value() % 2
	})
	c2 := alien.Computed(rs, func(oldValue int) int {
		return c1.Value()
	})
	c3 := alien.Computed(rs, func(oldValue int) int {
		return c2.Value()
	})

	assert.Equal(t, 0, c1.Value())
	src.SetValue(1)                // c1 -> dirty, c2 -> toCheckDirty, c3 -> toCheckDirty
	assert.Equal(t, 1, c2.Value()) // c1 -> none, c2 -> none
	src.SetValue(3)                // c1 -> dirty, c2 -> toCheckDirty

	assert.Equal(t, 1, c3.Value())
}

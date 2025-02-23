package alien_test

import (
	"log"
	"testing"

	alien "github.com/delaneyj/alien-signals-go"
	"github.com/stretchr/testify/assert"
)

// from README
func TestBasicUsage(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	count := alien.Signal(rs, 1)
	doubleCount := alien.Computed(rs, func(oldValue int) int {
		return count.Value() * 2
	})

	stopEffect := alien.Effect(rs, func() error {
		log.Printf("Count is: %d", count.Value())
		return nil
	})
	defer stopEffect()

	assert.Equal(t, 2, doubleCount.Value())
	count.SetValue(2)
	assert.Equal(t, 4, doubleCount.Value())
}

// from README
func TestBasicEffect(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	count := alien.Signal(rs, 1)

	stopScope := alien.EffectScope(rs, func() error {
		alien.Effect(rs, func() error {
			log.Printf("Count in scope: %d", count.Value())
			return nil
		}) // Console: Count in scope: 1
		count.SetValue(2) // Console: Count in scope: 2

		return nil
	})

	stopScope()
	count.SetValue(3) // No console output
}

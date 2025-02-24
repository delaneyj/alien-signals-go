package alien_test

import (
	"log"
	"testing"

	alien "github.com/delaneyj/alien-signals-go"
	"github.com/stretchr/testify/assert"
)

// should clear subscriptions when untracked by all subscribers
func TestEffectClearSubsWhenUntracked(t *testing.T) {
	bRunTimes := 0

	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.SignalInt(rs, 1)
	b := alien.Computed(rs, func(oldValue alien.Int) alien.Int {
		bRunTimes++
		return a.Value() * 2
	})
	stopEffect := alien.Effect(rs, func() error {
		b.Value()
		return nil
	})

	assert.Equal(t, 1, bRunTimes)
	a.SetValue(2)
	assert.Equal(t, 2, bRunTimes)
	stopEffect()
	a.SetValue(3)
	assert.Equal(t, 2, bRunTimes)
}

// should not run untracked inner effect
func TestShouldNotRunUntrackedInnerEffect(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.SignalInt(rs, 3)
	b := alien.Computed(rs, func(oldValue alien.Bool) alien.Bool {
		return a.Value() > 0
	})

	alien.Effect(rs, func() error {
		if b.Value() {
			alien.Effect(rs, func() error {
				if a.Value() == 0 {
					assert.Fail(t, "bad")
				}
				return nil
			})
		}
		return nil
	})

	decrement := func() {
		a.SetValue(a.Value() - 1)
	}
	decrement()
	decrement()
	decrement()
}

// should run outer effect first
func TestShouldRunOuterEffectFirst(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.SignalInt(rs, 1)
	b := alien.SignalInt(rs, 1)

	alien.Effect(rs, func() error {
		aV := a.Value()
		if aV != 0 {
			alien.Effect(rs, func() error {
				aV, bV := a.Value(), b.Value()
				log.Printf("aV: %d, bV: %d", aV, bV)
				if aV == 0 {
					assert.Fail(t, "bad")
				}
				return nil
			})
		}
		return nil
	})

	rs.StartBatch()
	a.SetValue(0)
	b.SetValue(0)
	rs.EndBatch()
}

// should not trigger inner effect when resolve maybe dirty
func TestShouldNotTriggerInnerEffectWhenResolveMaybeDirty(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.SignalInt(rs, 0)
	b := alien.Computed(rs, func(oldValue alien.Bool) alien.Bool {
		return a.Value()%2 == 0
	})

	innerTriggerTimes := 0

	alien.Effect(rs, func() error {
		alien.Effect(rs, func() error {
			b.Value()
			innerTriggerTimes++
			if innerTriggerTimes >= 2 {
				assert.Fail(t, "bad")
			}
			return nil
		})
		return nil
	})

	a.SetValue(2)
}

// should trigger inner effects in sequence
func TestShouldTriggerInnerEffectsInSequence(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.SignalInt(rs, 0)
	b := alien.SignalInt(rs, 0)
	c := alien.Computed(rs, func(oldValue alien.Int) alien.Int {
		return a.Value() - b.Value()
	})
	order := []string{}

	alien.Effect(rs, func() error {
		c.Value()

		alien.Effect(rs, func() error {
			order = append(order, "first inner")
			a.Value()
			return nil
		})

		alien.Effect(rs, func() error {
			order = append(order, "last inner")
			a.Value()
			b.Value()
			return nil
		})

		return nil
	})

	order = order[:0]
	rs.StartBatch()
	b.SetValue(1)
	a.SetValue(1)
	rs.EndBatch()

	assert.Equal(t, []string{"first inner", "last inner"}, order)
}

// should trigger inner effects in sequence in effect scope
func TestShouldTriggerInnerEffectsInSequenceInEffectScope(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.SignalInt(rs, 0)
	b := alien.SignalInt(rs, 0)
	order := []string{}

	alien.EffectScope(rs, func() error {
		alien.Effect(rs, func() error {
			order = append(order, "first inner")
			a.Value()
			return nil
		})

		alien.Effect(rs, func() error {
			order = append(order, "last inner")
			a.Value()
			b.Value()
			return nil
		})

		return nil
	})

	order = order[:0]
	rs.StartBatch()
	b.SetValue(1)
	a.SetValue(1)
	rs.EndBatch()

	assert.Equal(t, []string{"first inner", "last inner"}, order)
}

// should custom effect support batch
func TestShouldCustomEffectSupportBatch(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})

	batchEffect := func(fn func() error) alien.ErrFn {
		return alien.Effect(rs, func() error {
			rs.StartBatch()
			defer rs.EndBatch()
			return fn()
		})
	}

	logs := []string{}
	a := alien.SignalInt(rs, 0)
	b := alien.SignalInt(rs, 0)

	aa := alien.Computed(rs, func(oldValue alien.Int) alien.Int {
		logs = append(logs, "aa-0")
		aV := a.Value()
		if aV == 0 {
			b.SetValue(1)
		}
		logs = append(logs, "aa-1")
		return 0
	})

	bb := alien.Computed(rs, func(oldValue alien.Int) alien.Int {
		logs = append(logs, "bb")
		bV := b.Value()
		return bV
	})

	batchEffect(func() error {
		bb.Value()
		return nil
	})

	batchEffect(func() error {
		aa.Value()
		return nil
	})

	assert.Equal(t, []string{"bb", "aa-0", "aa-1", "bb"}, logs)
}

// should not trigger after stop
func TestShouldNotTriggerAfterStop(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})

	count := alien.SignalInt(rs, 0)

	triggers := 0

	stopScope := alien.EffectScope(rs, func() error {
		alien.Effect(rs, func() error {
			triggers++
			count.Value()
			return nil
		})
		return nil
	})

	assert.Equal(t, 1, triggers)
	count.SetValue(2)
	assert.Equal(t, 2, triggers)
	stopScope()
	count.SetValue(3)
	assert.Equal(t, 2, triggers)
}

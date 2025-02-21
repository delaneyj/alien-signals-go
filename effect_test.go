package alien_test

import (
	"sync"
	"testing"

	alien "github.com/delaneyj/alien-signals-go"
	"github.com/stretchr/testify/assert"
)

// should clear subscriptions when untracked by all subscribers
func TestEffectClearSubsWhenUntracked(t *testing.T) {
	bRunTimes := 0

	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.Signal(rs, 1)
	b := alien.Computed(rs, func(oldValue int) int {
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
	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.Signal(rs, 3)
	b := alien.Computed(rs, func(oldValue bool) bool {
		return a.Value() > 0
	})

	alien.Effect(rs, func() error {
		if b.Value() {
			go alien.Effect(rs, func() error {
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
	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.Signal(rs, 1)
	b := alien.Signal(rs, 1)

	alien.Effect(rs, func() error {
		aV := a.Value()
		if aV != 0 {
			go alien.Effect(rs, func() error {
				aV, bV := a.Value(), b.Value()
				if aV == 0 || bV == 0 {
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
	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.Signal(rs, 0)
	b := alien.Computed(rs, func(oldValue bool) bool {
		return a.Value()%2 == 0
	})

	innerTriggerTimes := 0

	alien.Effect(rs, func() error {
		go alien.Effect(rs, func() error {
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
	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.Signal(rs, 0)
	b := alien.Signal(rs, 0)
	c := alien.Computed(rs, func(oldValue int) int {
		return a.Value() - b.Value()
	})
	values := map[string]struct{}{}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	alien.Effect(rs, func() error {
		c.Value()

		go func() {
			alien.Effect(rs, func() error {
				defer wg.Done()
				values["A"] = struct{}{}
				a.Value()
				return nil
			})
		}()

		go func() {
			alien.Effect(rs, func() error {
				defer wg.Done()
				values["B"] = struct{}{}
				a.Value()
				b.Value()
				return nil
			})
		}()

		return nil
	})

	rs.StartBatch()
	b.SetValue(1)
	a.SetValue(1)
	rs.EndBatch()

	wg.Wait()
	assert.Contains(t, values, "A")
	assert.Contains(t, values, "B")
}

// should trigger inner effects in sequence in effect scope
func TestShouldTriggerInnerEffectsInSequenceInEffectScope(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.Signal(rs, 0)
	b := alien.Signal(rs, 0)
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
	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})

	logs := map[string]struct{}{}
	a := alien.Signal(rs, 0)
	b := alien.Signal(rs, 0)

	aa := alien.Computed(rs, func(oldValue int) int {
		logs["aa-0"] = struct{}{}
		aV := a.Value()
		if aV == 0 {
			b.SetValue(1)
		}
		logs["aa-1"] = struct{}{}
		return 0
	})

	bb := alien.Computed(rs, func(oldValue int) int {
		logs["bb"] = struct{}{}
		bV := b.Value()
		return bV
	})

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go alien.Effect(rs, func() error {
		defer wg.Done()

		rs.StartBatch()
		defer rs.EndBatch()

		bb.Value()
		return nil
	})
	go alien.Effect(rs, func() error {
		defer wg.Done()

		rs.StartBatch()
		defer rs.EndBatch()

		aa.Value()
		return nil
	})
	wg.Wait()

	assert.Contains(t, logs, "bb")
	assert.Contains(t, logs, "aa-0")
	assert.Contains(t, logs, "aa-1")
}

// should not trigger after stop
func TestShouldNotTriggerAfterStop(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})

	count := alien.Signal(rs, 0)

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

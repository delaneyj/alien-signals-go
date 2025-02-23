package alien_test

import (
	"fmt"
	"testing"
	"time"

	alien "github.com/delaneyj/alien-signals-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopologyDropAbaUpdates(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})

	//     A
	//   / |
	//  B  | <- Looks like a flag doesn't it? :D
	//   \ |
	//     C
	//     |
	//     D
	a := alien.Signal(rs, 2)
	b := alien.Computed(rs, func(oldValue int) int {
		return a.Value() - 1
	})
	c := alien.Computed(rs, func(oldValue int) int {
		return a.Value() + b.Value()
	})
	callCount := 0
	d := alien.Computed(rs, func(oldValue string) string {
		callCount++
		return fmt.Sprintf("d: %d", c.Value())
	})

	// Trigger read
	dActual := d.Value()
	assert.Equal(t, "d: 3", dActual)
	assert.Equal(t, 1, callCount)

	a.SetValue(4)
	d.Value()
	assert.Equal(t, 2, callCount)
}

func TestShouldOnlyUpdateEverySignalOnceDiamond(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})

	// In this scenario "D" should only update once when "A" receives
	// an update. This is sometimes referred to as the "diamond" scenario.
	//     A
	//   /   \
	//  B     C
	//   \   /
	//     D

	a := alien.Signal(rs, "a")
	b := alien.Computed(rs, func(oldValue string) string {
		return a.Value()
	})
	c := alien.Computed(rs, func(oldValue string) string {
		return a.Value()
	})

	callCount := 0
	d := alien.Computed(rs, func(oldValue string) string {
		callCount++
		return b.Value() + " " + c.Value()
	})

	assert.Equal(t, "a a", d.Value())
	assert.Equal(t, 1, callCount)
	callCount = 0

	a.SetValue("aa")
	assert.Equal(t, "aa aa", d.Value())
	assert.Equal(t, 1, callCount)
}

func TestShouldOnlyUpdateEverySignalOnceDiamondTail(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})

	// "E" will be likely updated twice if our mark+sweep logic is buggy.
	//     A
	//   /   \
	//  B     C
	//   \   /
	//     D
	//     |
	//     E

	a := alien.Signal(rs, "a")
	b := alien.Computed(rs, func(oldValue string) string {
		return a.Value()
	})
	c := alien.Computed(rs, func(oldValue string) string {
		return a.Value()
	})
	d := alien.Computed(rs, func(oldValue string) string {
		return b.Value() + " " + c.Value()
	})

	eCallCount := 0
	e := alien.Computed(rs, func(oldValue string) string {
		eCallCount++
		return d.Value()
	})

	assert.Equal(t, "a a", e.Value())
	assert.Equal(t, 1, eCallCount)

	a.SetValue("aa")
	assert.Equal(t, "aa aa", e.Value())
	assert.Equal(t, 2, eCallCount)
}

func TestBailOutIfResultIsTheSame(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})

	// Bail out if value of "B" never changes
	// A->B->C
	a := alien.Signal(rs, "a")
	b := alien.Computed(rs, func(oldValue string) string {
		a.Value()
		return "foo"
	})

	callCount := 0
	c := alien.Computed(rs, func(oldValue string) string {
		callCount++
		return b.Value()
	})

	assert.Equal(t, "foo", c.Value())
	assert.Equal(t, 1, callCount)

	a.SetValue("aa")
	assert.Equal(t, "foo", c.Value())
	assert.Equal(t, 1, callCount)
}

func TestShouldOnlyUpdateEverySignalOnceJaggedDiamondTails(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})

	// "F" and "G" will be likely updated twice if our mark+sweep logic is buggy.
	//     A
	//   /   \
	//  B     C
	//  |     |
	//  |     D
	//   \   /
	//     E
	//   /   \
	//  F     G

	a := alien.Signal(rs, "a")
	b := alien.Computed(rs, func(oldValue string) string {
		return a.Value()
	})
	c := alien.Computed(rs, func(oldValue string) string {
		return a.Value()
	})
	d := alien.Computed(rs, func(oldValue string) string {
		return c.Value()
	})

	eCallCount, eTime := 0, time.Time{}
	e := alien.Computed(rs, func(oldValue string) string {
		bV, dV := b.Value(), d.Value()
		eV := bV + " " + dV
		eCallCount++
		eTime = time.Now()
		return eV
	})

	fCallCount, fTime := 0, time.Time{}
	f := alien.Computed(rs, func(oldValue string) string {
		ev := e.Value()
		fCallCount++
		fTime = time.Now()
		return ev
	})

	gCallCount, gTime := 0, time.Time{}
	g := alien.Computed(rs, func(oldValue string) string {
		ev := e.Value()
		gCallCount++
		gTime = time.Now()
		return ev
	})

	require.Equal(t, "a a", f.Value())
	require.Equal(t, 1, fCallCount)
	require.Equal(t, "a a", g.Value())
	require.Equal(t, 1, gCallCount)
	eCallCount, fCallCount, gCallCount = 0, 0, 0

	a.SetValue("b")
	require.Equal(t, "b b", e.Value())
	require.Equal(t, 1, eCallCount)
	require.Equal(t, "b b", f.Value())
	require.Equal(t, 1, fCallCount)
	require.Equal(t, "b b", g.Value())
	require.Equal(t, 1, gCallCount)
	eCallCount, fCallCount, gCallCount = 0, 0, 0

	a.SetValue("c")
	require.Equal(t, "c c", e.Value())
	require.Equal(t, 1, eCallCount)
	require.Equal(t, "c c", f.Value())
	require.Equal(t, 1, fCallCount)
	require.Equal(t, "c c", g.Value())
	require.Equal(t, 1, gCallCount)

	// top to bottom
	assert.True(t, eTime.Before(fTime))
	// left to right
	assert.True(t, fTime.Before(gTime))

}

func TestShouldOnlySubscribeToSignalsListenedTo(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})

	//    *A
	//   /   \
	// *B     C <- we don't listen to C
	a := alien.Signal(rs, "a")
	b := alien.Computed(rs, func(oldValue string) string {
		return a.Value()
	})
	callCount := 0
	alien.Computed(rs, func(oldValue string) string {
		callCount++
		return a.Value()
	})

	assert.Equal(t, "a", b.Value())
	assert.Equal(t, 0, callCount)

	a.SetValue("aa")
	assert.Equal(t, "aa", b.Value())
	assert.Equal(t, 0, callCount)
}

func TestShouldOnlySubscribeToSignalsListenedToII(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})

	// Here both "B" and "C" are active in the beginning, but
	// "B" becomes inactive later. At that point it should
	// not receive any updates anymore.
	//    *A
	//   /   \
	// *B     D <- we don't listen to C
	//  |
	// *C
	a := alien.Signal(rs, "a")
	bCallCount := 0
	b := alien.Computed(rs, func(oldValue string) string {
		bCallCount++
		return a.Value()
	})
	cCallCount := 0
	c := alien.Computed(rs, func(oldValue string) string {
		cCallCount++
		return b.Value()
	})
	d := alien.Computed(rs, func(oldValue string) string {
		return a.Value()
	})

	result := ""
	unsub := alien.Effect(rs, func() error {
		result = c.Value()
		return nil
	})

	assert.Equal(t, "a", result)
	assert.Equal(t, "a", d.Value())

	bCallCount, cCallCount = 0, 0
	unsub()

	a.SetValue("aa")
	assert.Equal(t, 0, bCallCount)
	assert.Equal(t, 0, cCallCount)
	assert.Equal(t, "aa", d.Value())
}

func TestShouldEnsureSubsUpdate(t *testing.T) {
	// In this scenario "C" always returns the same value. When "A"
	// changes, "B" will update, then "C" at which point its update
	// to "D" will be unmarked. But "D" must still update because
	// "B" marked it. If "D" isn't updated, then we have a bug.
	//     A
	//   /   \
	//  B     *C <- returns same value every time
	//   \   /
	//     D
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.Signal(rs, "a")
	b := alien.Computed(rs, func(oldValue string) string {
		return a.Value()
	})
	c := alien.Computed(rs, func(oldValue string) string {
		a.Value()
		return "c"
	})
	dCallCount := 0
	d := alien.Computed(rs, func(oldValue string) string {
		dCallCount++
		return b.Value() + " " + c.Value()
	})

	assert.Equal(t, "a c", d.Value())
	assert.Equal(t, 1, dCallCount)

	a.SetValue("aa")
	assert.Equal(t, "aa c", d.Value())
}

func TestShouldEnsureSubsUpdateEvenIfTwoDepsUnmarkIt(t *testing.T) {
	// In this scenario both "C" and "D" always return the same
	// value. But "E" must still update because "A" marked it.
	// If "E" isn't updated, then we have a bug.
	//     A
	//   / | \
	//  B *C *D
	//   \ | /
	//     E
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.Signal(rs, "a")
	b := alien.Computed(rs, func(oldValue string) string {
		return a.Value()
	})
	c := alien.Computed(rs, func(oldValue string) string {
		a.Value()
		return "c"
	})
	d := alien.Computed(rs, func(oldValue string) string {
		a.Value()
		return "d"
	})
	eCallCount := 0
	e := alien.Computed(rs, func(oldValue string) string {
		eCallCount++
		return b.Value() + " " + c.Value() + " " + d.Value()
	})

	assert.Equal(t, "a c d", e.Value())
	assert.Equal(t, 1, eCallCount)

	a.SetValue("aa")
	assert.Equal(t, "aa c d", e.Value())
	assert.Equal(t, 2, eCallCount)
}

func TestShouldEnsureSubsUpdateEvenIfAllDepsUnmarkIt(t *testing.T) {
	// In this scenario "B" and "C" always return the same value. When "A"
	// changes, "D" should not update.
	//     A
	//   /   \
	// *B     *C
	//   \   /
	//     D
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		assert.FailNow(t, err.Error())
	})
	a := alien.Signal(rs, "a")
	b := alien.Computed(rs, func(oldValue string) string {
		a.Value()
		return "b"
	})
	c := alien.Computed(rs, func(oldValue string) string {
		a.Value()
		return "c"
	})
	dCallCount := 0
	d := alien.Computed(rs, func(oldValue string) string {
		dCallCount++
		return b.Value() + " " + c.Value()
	})

	assert.Equal(t, "b c", d.Value())
	assert.Equal(t, 1, dCallCount)
	dCallCount = 0

	a.SetValue("aa")
	assert.Equal(t, 0, dCallCount)
}

func TestShouldKeepGraphConsistentOnActivationErrors(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		t.Error(err)
	})

	a := alien.Signal(rs, 0)
	b := alien.Computed(rs, func(oldValue int) int {
		panic("fail")
	})

	assert.Panics(t, func() {
		b.Value()
	})

	a.SetValue(1)
	assert.Equal(t, 1, a.Value())
}

func TestShouldKeepGraphConsistentOnComputedErrors(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		t.Error(err)
	})

	a := alien.Signal(rs, 0)
	b := alien.Computed(rs, func(oldValue int) int {
		panic("fail")
	})
	c := alien.Computed(rs, func(oldValue int) int {
		return a.Value()
	})

	assert.Panics(t, func() {
		b.Value()
	})

	a.SetValue(1)
	assert.Equal(t, 1, c.Value())
}

package alien_test

import (
	"testing"

	alien "github.com/delaneyj/alien-signals-go"
	"github.com/stretchr/testify/assert"
)

// should pause tracking
func TestShouldPauseTracking(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		t.FailNow()
	})

	src := alien.SignalInt(rs, 0)
	c := alien.Computed(rs, func(oldValue alien.Int) alien.Int {
		rs.PauseTracking()
		value := src.Value()
		rs.ResumeTracking()
		return value
	})
	actualC := c.Value()
	assert.Equal(t, 0, actualC.Int())

	src.SetValue(1)
	actualC = c.Value()
	assert.Equal(t, 0, actualC.Int())
}

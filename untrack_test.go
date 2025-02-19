package alien_test

import (
	"testing"

	alien "github.com/delaneyj/alien-signals-go"
	"github.com/stretchr/testify/assert"
)

// should pause tracking
func TestShouldPauseTracking(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(err error) {
		t.FailNow()
	})

	src := alien.Signal(rs, 0)
	c := alien.Computed(rs, func(oldValue int) int {
		rs.PauseTracking()
		value := src.Value()
		rs.ResumeTracking()
		return value
	})
	actualC := c.Value()
	assert.Equal(t, 0, actualC)

	src.SetValue(1)
	actualC = c.Value()
	assert.Equal(t, 0, actualC)
}

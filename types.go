package alien

import (
	"slices"
)

type subscriberFlags uint16

const (
	fComputed subscriberFlags = 1 << iota
	fEffect
	fTracking
	fNotified
	fRecursed
	fDirty
	fPendingComputed
	fPendingEffect
	fEffectScope
	fPropagated subscriberFlags = fDirty | fPendingComputed | fPendingEffect
)

type link struct {
	dep dependency
	sub subscriber
	// reused to link the previous stack in updateDirtyFlags
	// reused to link the previous stack in propagate
	prevSub *link
	nextSub *link
	// reused to link the notify effect in queueEffects
	nextDep *link
}

type subscriber interface {
	flags() subscriberFlags
	setFlags(subscriberFlags)
	deps() *link
	setDeps(*link)
	depsTail() *link
	setDepsTail(*link)
}

type baseSubscriber struct {
	_flags           subscriberFlags
	_deps, _depsTail *link
}

func (s *baseSubscriber) flags() subscriberFlags {
	return s._flags
}
func (s *baseSubscriber) setFlags(flags subscriberFlags) {
	s._flags = flags
}
func (s *baseSubscriber) deps() *link {
	return s._deps
}
func (s *baseSubscriber) setDeps(deps *link) {
	s._deps = deps
}
func (s *baseSubscriber) depsTail() *link {
	return s._depsTail
}
func (s *baseSubscriber) setDepsTail(depsTail *link) {
	s._depsTail = depsTail
}

type dependency interface {
	subs() *link
	setSubs(*link)
	subsTail() *link
	setSubsTail(*link)
}

type baseDependency struct {
	_subs, _subsTail *link
}

func (d *baseDependency) subs() *link {
	return d._subs
}

func (d *baseDependency) setSubs(subs *link) {
	d._subs = subs
}

func (d *baseDependency) subsTail() *link {
	return d._subsTail
}

func (d *baseDependency) setSubsTail(subsTail *link) {
	d._subsTail = subsTail
}

type dependencyAndSubscriber interface {
	dependency
	subscriber
}

type Equality[T any] interface {
	Equals(other T) bool
}

type String string

func (s String) Equals(other String) bool {
	return s.String() == other.String()
}

func (s String) String() string {
	return string(s)
}

func FromString(s string) String {
	return String(s)
}

type Int int

func (i Int) Equals(other Int) bool {
	return i == other
}

func (i Int) Int() int {
	return int(i)
}

func FromInt(i int) Int {
	return Int(i)
}

type Float64 float64

func (f Float64) Equals(other Float64) bool {
	return f == other
}

func (f Float64) Float64() float64 {
	return float64(f)
}

func FromFloat64(f float64) Float64 {
	return Float64(f)
}

type Bool bool

func (b Bool) Equals(other Bool) bool {
	return b == other
}

func (b Bool) Bool() bool {
	return bool(b)
}

func FromBool(b bool) Bool {
	return Bool(b)
}

type Slice[T comparable] []T

func (s Slice[T]) Equals(other Slice[T]) bool {
	return slices.Equal(s, other)
}

func (s Slice[T]) Slice() []T {
	return []T(s)
}

func FromSlice[T comparable](s []T) Slice[T] {
	return Slice[T](s)
}

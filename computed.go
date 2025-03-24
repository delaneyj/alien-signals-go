package alien

type ReadonlySignal[T comparable] struct {
	baseSubscriber
	baseDependency

	rs     *ReactiveSystem
	value  T
	getter func(oldValue T) T
}

func (s *ReadonlySignal[T]) dep() *baseDependency {
	return &s.baseDependency
}

func (s *ReadonlySignal[T]) sub() *baseSubscriber {
	return &s.baseSubscriber
}

func (s *ReadonlySignal[T]) isSignalAware() {}

func (s *ReadonlySignal[T]) Value() T {
	flags := s.flags
	if flags&(fDirty|fPendingComputed) != 0 {
		processComputedUpdate(s.rs, s, flags)
	}
	if s.rs.activeSub != nil {
		s.rs.link(s, s.rs.activeSub)
	} else if s.rs.activeScope != nil {
		s.rs.link(s, s.rs.activeScope)
	}

	return s.value
}

func (s *ReadonlySignal[T]) cas() bool {
	oldValue := s.value
	newValue := s.getter(oldValue)
	s.value = newValue
	return oldValue != newValue
}

func Computed[T comparable](rs *ReactiveSystem, getter func(oldValue T) T) *ReadonlySignal[T] {
	c := &ReadonlySignal[T]{
		rs:     rs,
		getter: getter,
		baseSubscriber: baseSubscriber{
			flags: fComputed | fDirty,
		},
	}
	return c
}

type computedAny interface {
	dependency
	subscriber
	cas() (wasDifferent bool)
}

func updateComputed(rs *ReactiveSystem, sub subscriber) bool {
	prevSub := rs.activeSub
	rs.activeSub = sub
	sub2 := sub.sub()
	rs.startTracking(sub2)

	defer func() {
		rs.activeSub = prevSub
		rs.endTracking(sub2)
	}()

	computed := sub.(computedAny)
	return computed.cas()
}

// Updates the computed subscriber if necessary before its value is accessed.
//
// If the subscriber is marked Dirty or PendingComputed, this function runs
// the provided updateComputed logic and triggers a shallowPropagate for any
// downstream subscribers if an actual update occurs.
//
// @param computed - The computed subscriber to update.
// @param flags - The current flag set for this subscriber.
func processComputedUpdate(rs *ReactiveSystem, computed computedAny, flags subscriberFlags) {
	dirty := flags&fDirty != 0
	if !dirty {
		sub := computed.sub()
		dirty = rs.checkDirty(sub.deps)
		sub.flags = flags & ^fPendingComputed
	}
	if dirty {
		if updateComputed(rs, computed) {
			subs := computed.dep().subs
			if subs != nil {
				rs.shallowPropagate(subs)
			}
		}
	}
}

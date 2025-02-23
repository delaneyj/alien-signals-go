package alien

type ReadonlySignal[T comparable] struct {
	baseSubscriber
	baseDependency

	rs     *ReactiveSystem
	value  T
	getter func(oldValue T) T
}

func (s *ReadonlySignal[T]) isSignalAware() {}

func (s *ReadonlySignal[T]) Value() T {
	flags := s._flags
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

func (s *ReadonlySignal[T]) cas() (wasDifferent bool) {
	oldValue := s.value
	newValue := s.getter(oldValue)
	wasDifferent = oldValue != newValue
	s.value = newValue
	return wasDifferent
}

func Computed[T comparable](rs *ReactiveSystem, getter func(oldValue T) T) *ReadonlySignal[T] {
	c := &ReadonlySignal[T]{
		rs:     rs,
		getter: getter,
		baseSubscriber: baseSubscriber{
			_flags: fComputed | fDirty,
		},
	}
	return c
}

type computedAny interface {
	dependencyAndSubscriber
	cas() (wasDifferent bool)
}

func updateComputed(rs *ReactiveSystem, c any) bool {
	computed, ok := c.(computedAny)
	if !ok {
		panic("not a computedAny")
	}

	prevSub := rs.activeSub
	rs.activeSub = computed
	rs.startTracking(computed)

	defer func() {
		rs.activeSub = prevSub
		rs.endTracking(computed)
	}()

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
	if flags&fDirty != 0 || func() bool {
		isDirty := rs.checkDirty(computed.deps())
		if isDirty {
			return true
		}
		computed.setFlags(flags & ^fPendingComputed)
		return false
	}() {
		if updateComputed(rs, computed) {
			subs := computed.subs()
			if subs != nil {
				rs.shallowPropagate(subs)
			}
		}
	}
}

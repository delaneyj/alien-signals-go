package alien

type ReadonlySignal[T comparable] struct {
	signal

	rs     *ReactiveSystem
	value  T
	getter func(oldValue T) T
}

func (s *ReadonlySignal[T]) isSignalAware() {}

func (s *ReadonlySignal[T]) Value() T {
	flags := s.flags
	signal := &s.signal
	if flags&(fDirty|fPendingComputed) != 0 {
		processComputedUpdate(s.rs, signal, flags)
	}
	if s.rs.activeSub != nil {
		s.rs.link(signal, s.rs.activeSub)
	} else if s.rs.activeScope != nil {
		s.rs.link(signal, s.rs.activeScope)
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
		signal: signal{
			flags: fComputed | fDirty,
		},
	}
	signal := &c.signal
	signal.ref = c
	return c
}

type computedAny interface {
	cas() (wasDifferent bool)
}

func updateComputed(rs *ReactiveSystem, signal *signal) bool {
	prevSub := rs.activeSub
	rs.activeSub = signal
	rs.startTracking(signal)

	defer func() {
		rs.activeSub = prevSub
		rs.endTracking(signal)
	}()

	return signal.ref.(computedAny).cas()
}

// Updates the computed subscriber if necessary before its value is accessed.
//
// If the subscriber is marked Dirty or PendingComputed, this function runs
// the provided updateComputed logic and triggers a shallowPropagate for any
// downstream subscribers if an actual update occurs.
//
// @param computed - The computed subscriber to update.
// @param flags - The current flag set for this subscriber.
func processComputedUpdate(rs *ReactiveSystem, signal *signal, flags subscriberFlags) {
	if flags&fDirty != 0 || rs.checkDirty(signal.deps) {
		if updateComputed(rs, signal) {
			subs := signal.subs
			if subs != nil {
				rs.shallowPropagate(subs)
			}
		}
	} else {
		signal.flags = flags & ^fPendingComputed
	}
}

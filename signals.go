package alien

type WriteableSignal[T comparable] struct {
	signal
	rs    *ReactiveSystem
	value T
}

func (s *WriteableSignal[T]) isSignalAware() {}

func (s *WriteableSignal[T]) Value() T {
	if s.rs.activeSub != nil {
		s.rs.link(&s.signal, s.rs.activeSub)
	}
	return s.value
}

func (s *WriteableSignal[T]) SetValue(v T) {
	if s.value == v {
		return
	}
	s.value = v
	subs := s.signal.subs
	if subs != nil {
		s.rs.propagate(subs)
		if s.rs.batchDepth == 0 {
			s.rs.processEffectNotifications()
		}
	}
}

func Signal[T comparable](rs *ReactiveSystem, initialValue T) *WriteableSignal[T] {
	s := &WriteableSignal[T]{
		rs:     rs,
		value:  initialValue,
		signal: signal{},
	}
	signal := &s.signal
	signal.ref = s
	return s
}

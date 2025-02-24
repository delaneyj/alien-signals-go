package alien

type WriteableSignal[T Equality[T]] struct {
	baseDependency
	rs    *ReactiveSystem
	value T
}

func (s *WriteableSignal[T]) isSignalAware() {}

func (s *WriteableSignal[T]) Value() T {
	if s.rs.activeSub != nil {
		s.rs.link(s, s.rs.activeSub)
	}
	return s.value
}

func (s *WriteableSignal[T]) SetValue(v T) {
	if s.value.Equals(v) {
		return
	}
	s.value = v
	subs := s._subs
	if subs != nil {
		s.rs.propagate(subs)
		if s.rs.batchDepth == 0 {
			s.rs.processEffectNotifications()
		}
	}
}

func Signal[T Equality[T]](rs *ReactiveSystem, initialValue T) *WriteableSignal[T] {
	s := &WriteableSignal[T]{
		rs:    rs,
		value: initialValue,
	}
	return s
}

func SignalInt(rs *ReactiveSystem, initialValue int) *WriteableSignal[Int] {
	return Signal(rs, Int(initialValue))
}

func SignalString(rs *ReactiveSystem, initialValue string) *WriteableSignal[String] {
	return Signal(rs, String(initialValue))
}

func SignalBool(rs *ReactiveSystem, initialValue bool) *WriteableSignal[Bool] {
	return Signal(rs, Bool(initialValue))
}

func SignalSlice[T comparable](rs *ReactiveSystem, initialValue ...T) *WriteableSignal[Slice[T]] {
	return Signal(rs, Slice[T](initialValue))
}

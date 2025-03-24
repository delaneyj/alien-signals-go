package alien

type ErrFn func() error

func Effect(rs *ReactiveSystem, fn ErrFn) ErrFn {
	e := &EffectRunner{
		fn: fn,
		signal: signal{
			flags: fEffect,
		},
	}
	signal := &e.signal
	signal.ref = e

	if rs.activeSub != nil {
		rs.link(signal, rs.activeSub)
	} else if rs.activeScope != nil {
		rs.link(signal, rs.activeScope)
	}
	rs.runEffect(e, signal)

	return func() error {
		rs.startTracking(signal)
		rs.endTracking(signal)
		return nil
	}
}

func (rs *ReactiveSystem) runEffect(e *EffectRunner, signal *signal) {
	prevSub := rs.activeSub
	rs.activeSub = signal
	rs.startTracking(signal)
	if err := e.fn(); err != nil {
		if rs.onError != nil {
			rs.onError(e, err)
		}
	}
	rs.endTracking(signal)
	rs.activeSub = prevSub
}

func (rs *ReactiveSystem) notifyEffect(signal *signal) bool {
	flags := signal.flags
	if flags&fEffectScope != 0 {
		if flags&fPendingEffect != 0 {
			rs.processPendingInnerEffects(signal, flags)
			return true
		}
		return false
	}

	if flags&fDirty != 0 ||
		(flags&fPendingComputed != 0 && rs.updateDirtyFlag(signal, flags)) {
		rs.runEffect(signal.ref.(*EffectRunner), signal)
	} else {
		rs.processPendingInnerEffects(signal, flags)
	}
	return true
}

func EffectScope(rs *ReactiveSystem, scopedFn ErrFn) (stopScope ErrFn) {
	e := &EffectRunner{
		signal: signal{
			flags: fEffect | fEffectScope,
		},
	}
	signal := &e.signal
	signal.ref = e
	rs.runEffectScope(e, signal, scopedFn)
	return func() error {
		rs.startTracking(signal)
		rs.endTracking(signal)
		return nil
	}
}

type EffectRunner struct {
	signal
	fn ErrFn
}

func (e *EffectRunner) isSignalAware() {}

func (rs *ReactiveSystem) runEffectScope(e *EffectRunner, signal *signal, scopedFn ErrFn) {
	prevSub := rs.activeSub
	rs.activeScope = signal
	rs.startTracking(signal)

	if err := scopedFn(); err != nil {
		if rs.onError != nil {
			rs.onError(e, err)
		}
	}

	rs.activeScope = prevSub
	rs.endTracking(signal)
}

// Ensures all pending internal effects for the given subscriber are processed.
//
// This should be called after an effect decides not to re-run itself but may still
// have dependencies flagged with PendingEffect. If the subscriber is flagged with
// PendingEffect, this function clears that flag and invokes `notifyEffect` on any
// related dependencies marked as Effect and Propagated, processing pending effects.
//
// @param sub - The subscriber which may have pending effects.
// @param flags - The current flags on the subscriber to check.
func (rs *ReactiveSystem) processPendingInnerEffects(sub *signal, flags subscriberFlags) {
	if flags&fPendingEffect != 0 {
		sub.flags = flags & ^fPendingEffect
		link := sub.deps
		for {
			dep := link.dep
			flags = dep.flags
			if flags&fEffect != 0 && flags&fPropagated != 0 {
				rs.notifyEffect(dep)
			}
			link = link.nextDep

			if link == nil {
				break
			}
		}
	}
}

// Processes queued effect notifications after a batch operation finishes.
//
// Iterates through all queued effects, calling notifyEffect on each.
// If an effect remains partially handled, its flags are updated, and future
// notifications may be triggered until fully handled.
func (rs *ReactiveSystem) processEffectNotifications() {
	for rs.queuedEffects != nil {
		effect := rs.queuedEffects.target
		rs.queuedEffects = rs.queuedEffects.linked
		if rs.queuedEffects == nil {
			rs.queuedEffectsTail = nil
		}
		if !rs.notifyEffect(effect) {
			effect.flags = effect.flags & ^fNotified
		}
	}
}

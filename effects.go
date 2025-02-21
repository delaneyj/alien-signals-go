package alien

type ErrFn func() error

func Effect(rs *ReactiveSystem, fn ErrFn) ErrFn {
	if !rs.lockAlreadyActive {
		rs.mu.Lock()
		rs.lockAlreadyActive = true
		defer func() {
			rs.lockAlreadyActive = false
			rs.mu.Unlock()
		}()
	}

	e := &effect{
		fn: fn,
		baseSubscriber: baseSubscriber{
			_flags: fEffect,
		},
	}

	if rs.activeSub != nil {
		rs.link(e, rs.activeSub)
	} else if rs.activeScope != nil {
		rs.link(e, rs.activeScope)
	}
	rs.runEffect(e)

	return func() error {
		rs.startTracking(e)
		rs.endTracking(e)
		return nil
	}
}

func (rs *ReactiveSystem) runEffect(e *effect) {
	prevSub := rs.activeSub
	rs.activeSub = e
	rs.startTracking(e)
	if err := e.fn(); err != nil {
		if rs.onError != nil {
			rs.onError(err)
		}
	}
	rs.endTracking(e)
	rs.activeSub = prevSub
}

func (rs *ReactiveSystem) notifyEffect(e *effect) bool {
	flags := e.flags()
	if flags&fEffectScope != 0 {
		flags := e.flags()
		if flags&fPendingEffect != 0 {
			rs.processPendingInnerEffects(e, e.flags())
			return true
		}
		return false
	}

	if flags&fDirty != 0 ||
		(flags&fPendingComputed != 0 && rs.updateDirtyFlag(e, flags)) {
		rs.runEffect(e)
	} else {
		rs.processPendingInnerEffects(e, flags)
	}
	return true
}

func EffectScope(rs *ReactiveSystem, scopedFn ErrFn) (stopScope ErrFn) {
	e := &effect{
		baseSubscriber: baseSubscriber{
			_flags: fEffect | fEffectScope,
		},
	}
	rs.runEffectScope(e, scopedFn)
	return func() error {
		rs.startTracking(e)
		rs.endTracking(e)
		return nil
	}
}

type effect struct {
	baseSubscriber
	baseDependency
	fn ErrFn
}

func (rs *ReactiveSystem) runEffectScope(e *effect, scopedFn ErrFn) {
	prevSub := rs.activeSub
	rs.activeScope = e
	rs.startTracking(e)

	if err := scopedFn(); err != nil {
		if rs.onError != nil {
			rs.onError(err)
		}
	}

	rs.activeScope = prevSub
	rs.endTracking(e)
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
func (rs *ReactiveSystem) processPendingInnerEffects(sub subscriber, flags subscriberFlags) {
	if flags&fPendingEffect != 0 {
		sub.setFlags(flags & ^fPendingEffect)
		link := sub.deps()
		for {
			dep := link.dep
			depSub, ok := dep.(dependencyAndSubscriber)
			if ok {
				flags := depSub.flags()
				if flags&fEffect != 0 && flags&fPropagated != 0 {
					effect, ok := depSub.(*effect)
					if !ok {
						panic("not an effect")
					}
					rs.notifyEffect(effect)
				}
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
		e := rs.queuedEffects
		depsTail := e.depsTail()
		queueNext := depsTail.nextDep
		if queueNext != nil {
			depsTail.nextDep = nil
			effect, ok := queueNext.sub.(*effect)
			if !ok {
				panic("not an effect")
			}
			rs.queuedEffects = effect
		} else {
			rs.queuedEffects = nil
			rs.queuedEffectsTail = nil
		}
		if !rs.notifyEffect(e) {
			e.setFlags(e.flags() & ^fNotified)
		}
	}
}

func mustEffect(sub subscriber) *effect {
	e, ok := sub.(*effect)
	if !ok {
		panic("not an effect")
	}
	return e
}

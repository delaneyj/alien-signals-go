package alien

type ErrFn func() error

func Effect(rs *ReactiveSystem, fn ErrFn) ErrFn {
	e := &EffectRunner{
		fn:             fn,
		baseDependency: baseDependency{},
		baseSubscriber: baseSubscriber{
			flags: fEffect,
		},
	}
	sub := e.sub()

	if rs.activeSub != nil {
		rs.link(e, rs.activeSub)
	} else if rs.activeScope != nil {
		rs.link(e, rs.activeScope)
	}
	rs.runEffect(e)

	return func() error {
		rs.startTracking(sub)
		rs.endTracking(sub)
		return nil
	}
}

func (rs *ReactiveSystem) runEffect(e *EffectRunner) {
	prevSub := rs.activeSub
	rs.activeSub = e
	sub := e.sub()
	rs.startTracking(sub)
	if err := e.fn(); err != nil {
		if rs.onError != nil {
			rs.onError(e, err)
		}
	}
	rs.endTracking(sub)
	rs.activeSub = prevSub
}

func (rs *ReactiveSystem) notifyEffect(sub subscriber) bool {
	e := mustEffect(sub)
	sub2 := sub.sub()
	flags := sub2.flags
	if flags&fEffectScope != 0 {
		if flags&fPendingEffect != 0 {
			rs.processPendingInnerEffects(sub2, flags)
			return true
		}
		return false
	}

	if flags&fDirty != 0 ||
		(flags&fPendingComputed != 0 && rs.updateDirtyFlag(sub2, flags)) {
		rs.runEffect(e)
	} else {
		rs.processPendingInnerEffects(sub2, flags)
	}
	return true
}

func EffectScope(rs *ReactiveSystem, scopedFn ErrFn) (stopScope ErrFn) {
	e := &EffectRunner{
		baseSubscriber: baseSubscriber{
			flags: fEffect | fEffectScope,
		},
	}
	sub := e.sub()
	rs.runEffectScope(e, scopedFn)
	return func() error {
		rs.startTracking(sub)
		rs.endTracking(sub)
		return nil
	}
}

type EffectRunner struct {
	baseSubscriber
	baseDependency
	fn ErrFn
}

func (s *EffectRunner) dep() *baseDependency {
	return &s.baseDependency
}

func (s *EffectRunner) sub() *baseSubscriber {
	return &s.baseSubscriber
}

func (e *EffectRunner) isSignalAware() {}

func (rs *ReactiveSystem) runEffectScope(e *EffectRunner, scopedFn ErrFn) {
	prevSub := rs.activeSub
	rs.activeScope = e
	sub := e.sub()
	rs.startTracking(sub)

	if err := scopedFn(); err != nil {
		if rs.onError != nil {
			rs.onError(e, err)
		}
	}

	rs.activeScope = prevSub
	rs.endTracking(sub)
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
func (rs *ReactiveSystem) processPendingInnerEffects(sub *baseSubscriber, flags subscriberFlags) {
	if flags&fPendingEffect != 0 {
		sub.flags = flags & ^fPendingEffect
		link := sub.deps
		for {
			dep := link.dep
			depSub, ok := dep.(subscriber)
			if ok {
				depSub2 := depSub.sub()
				flags := depSub2.flags
				if flags&fEffect != 0 && flags&fPropagated != 0 {
					rs.notifyEffect(depSub)
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
		effect := rs.queuedEffects.target
		rs.queuedEffects = rs.queuedEffects.linked
		if rs.queuedEffects == nil {
			rs.queuedEffectsTail = nil
		}
		if !rs.notifyEffect(effect) {
			sub := effect.sub()
			sub.flags = sub.flags & ^fNotified
		}
	}
}

func mustEffect(sub subscriber) *EffectRunner {
	e, ok := sub.(*EffectRunner)
	if !ok {
		panic("not an effect")
	}
	return e
}

package alien

type OnErrorFunc func(from SignalAware, err error)

type ReactiveSystem struct {
	batchDepth        int
	activeSub         subscriber
	queuedEffects     *EffectRunner
	queuedEffectsTail *EffectRunner

	activeScope subscriber
	onError     OnErrorFunc
	pauseStack  []subscriber
}

type SignalAware interface {
	isSignalAware()
}

func CreateReactiveSystem(onError OnErrorFunc) *ReactiveSystem {
	rs := &ReactiveSystem{onError: onError}

	return rs
}

func (rs *ReactiveSystem) StartBatch() {
	rs.batchDepth++
}

func (rs *ReactiveSystem) EndBatch() {
	rs.batchDepth--
	if rs.batchDepth == 0 {
		rs.processEffectNotifications()
	}
}

func (rs *ReactiveSystem) Batch(cb func()) {
	rs.StartBatch()
	defer rs.EndBatch()
	cb()
}

func (rs *ReactiveSystem) PauseTracking() {
	rs.pauseStack = append(rs.pauseStack, rs.activeSub)
	rs.activeSub = nil
}

func (rs *ReactiveSystem) ResumeTracking() {
	lastIdx := len(rs.pauseStack) - 1
	rs.activeSub = rs.pauseStack[lastIdx]
	rs.pauseStack = rs.pauseStack[:lastIdx]
}

// Updates the dirty flag for the given subscriber based on its dependencies.
//
// If the subscriber has any pending computeds, this function sets the Dirty flag
// and returns `true`. Otherwise, it clears the PendingComputed flag and returns `false`.
//
// @param sub - The subscriber to update.
// @param flags - The current flag set for this subscriber.
// @returns `true` if the subscriber is marked as Dirty; otherwise `false`.
func (rs *ReactiveSystem) updateDirtyFlag(sub subscriber, flags subscriberFlags) bool {
	if rs.checkDirty(sub.deps()) {
		sub.setFlags(flags | fDirty)
		return true
	}
	sub.setFlags(flags & ^fPendingComputed)
	return false
}

// Recursively checks and updates all computed subscribers marked as pending.
//
// It traverses the linked structure using a stack mechanism. For each computed
// subscriber in a pending state, updateComputed is called and shallowPropagate
// is triggered if a value changes. Returns whether any updates occurred.
//
// @param link - The starting link representing a sequence of pending computeds.
// @returns `true` if a computed was updated, otherwise `false`.
func (rs *ReactiveSystem) checkDirty(link *link) (dirty bool) {
	stack := 0

top:
	for {
		dirty = false
		dep := link.dep
		dependencySubscriber, isDependencySubscriber := dep.(dependencyAndSubscriber)
		if isDependencySubscriber {
			depFlags := dependencySubscriber.flags()
			if depFlags&(fComputed|fDirty) == fComputed|fDirty {
				if updateComputed(rs, dep) {
					subs := dep.subs()
					if subs.nextSub != nil {
						rs.shallowPropagate(subs)
					}
					dirty = true
				}
			} else if depFlags&(fComputed|fPendingComputed) == fComputed|fPendingComputed {
				depSubs := dep.subs()
				if depSubs.nextSub != nil {
					depSubs.prevSub = link
				}
				link = dependencySubscriber.deps()
				stack++
				continue
			}
		}

		if !dirty && link.nextDep != nil {
			link = link.nextDep
			continue
		}

		if stack != 0 {
			s := link.sub
			sub, ok := s.(dependencyAndSubscriber)
			if !ok {
				panic("not a dependencyAndSubscriber")
			}
			for {
				if stack == 0 {
					break
				}

				stack--
				subSubs := sub.subs()

				if dirty {
					if updateComputed(rs, sub) {
						link = subSubs.prevSub
						if link != nil {
							subSubs.prevSub = nil
							rs.shallowPropagate(sub.subs())
							sub = link.sub.(dependencyAndSubscriber)
						} else {
							sub = subSubs.sub.(dependencyAndSubscriber)
						}
						continue
					}
				} else {
					sub.setFlags(sub.flags() & ^fPendingComputed)
				}

				link = subSubs.prevSub
				if link != nil {
					subSubs.prevSub = nil
					if link.nextDep != nil {
						link = link.nextDep
						continue top
					}
					sub = link.sub.(dependencyAndSubscriber)
				} else {
					link = subSubs.nextDep
					if link != nil {
						continue top
					}
					sub = subSubs.sub.(dependencyAndSubscriber)
				}
				dirty = false
			}
		}

		return dirty
	}
}

// Links a given dependency and subscriber if they are not already linked.
//
// @param dep - The dependency to be linked.
// @param sub - The subscriber that depends on this dependency.
// @returns The newly created link object if the two are not already linked; otherwise `undefined`.
func (rs *ReactiveSystem) link(dep dependency, sub subscriber) *link {
	currentDep := sub.depsTail()
	if currentDep != nil && currentDep.dep == dep {
		return nil
	}

	var nextDep *link
	if currentDep != nil {
		nextDep = currentDep.nextDep
	} else {
		nextDep = sub.deps()
	}
	if nextDep != nil && nextDep.dep == dep {
		sub.setDepsTail(nextDep)
		return nil
	}

	depLastSub := dep.subsTail()
	if depLastSub != nil && depLastSub.sub == sub && rs.isValidLink(depLastSub, sub) {
		return nil
	}

	return rs.linkNewDep(dep, sub, nextDep, currentDep)

}

// Verifies whether the given link is valid for the specified subscriber.
//
// It iterates through the subscriber's link list (from sub.deps to sub.depsTail)
// to determine if the provided link object is part of that chain.
//
// @param checkLink - The link object to validate.
// @param sub - The subscriber whose link list is being checked.
// @returns `true` if the link is found in the subscriber's list; otherwise `false`.
func (rs *ReactiveSystem) isValidLink(checkLink *link, sub subscriber) bool {
	depsTails := sub.depsTail()
	if depsTails != nil {
		link := sub.deps()
		for {
			if link == checkLink {
				return true
			}
			if link == depsTails {
				break
			}
			link = link.nextDep

			if link == nil {
				break
			}
		}
	}
	return false
}

// Creates and attaches a new link between the given dependency and subscriber.
//
// Reuses a link object from the linkPool if available. The newly formed link
// is added to both the dependency's linked list and the subscriber's linked list.
//
// @param dep - The dependency to link.
// @param sub - The subscriber to be attached to this dependency.
// @param nextDep - The next link in the subscriber's chain.
// @param depsTail - The current tail link in the subscriber's chain.
// @returns The newly created link object.
func (rs *ReactiveSystem) linkNewDep(dep dependency, sub subscriber, nextDep, depsTail *link) *link {
	newLink := &link{
		dep:     dep,
		sub:     sub,
		nextDep: nextDep,
	}

	if depsTail == nil {
		sub.setDeps(newLink)
	} else {
		depsTail.nextDep = newLink
	}

	if dep.subs() == nil {
		dep.setSubs(newLink)
	} else {
		oldTail := dep.subsTail()
		newLink.prevSub = oldTail
		oldTail.nextSub = newLink
	}

	sub.setDepsTail(newLink)
	dep.setSubsTail(newLink)

	return newLink
}

// Traverses and marks subscribers starting from the provided link.
//
// It sets flags (e.g., Dirty, PendingComputed, PendingEffect) on each subscriber
// to indicate which ones require re-computation or effect processing.
// This function should be called after a signal's value changes.
//
// @param link - The starting link from which propagation begins.
func (rs *ReactiveSystem) propagate(link *link) {
	targetFlag := fDirty
	subs := link
	stack := 0

top:
	for {
		sub := link.sub
		subFlags := sub.flags()

		if (subFlags&(fTracking|fRecursed|fPropagated) == 0 &&
			func() bool {
				sub.setFlags(subFlags | targetFlag | fNotified)
				return true
			}()) ||
			(subFlags&fRecursed != 0 && subFlags&fTracking == 0 && func() bool {
				sub.setFlags(subFlags&^fRecursed | targetFlag | fNotified)
				return true
			}()) ||
			(subFlags&fPropagated == 0 && rs.isValidLink(link, sub) && func() bool {
				sub.setFlags(subFlags | fRecursed | targetFlag | fNotified)
				subDep, ok := sub.(dependencyAndSubscriber)
				return ok && subDep.subs() != nil
			}()) {
			subDep, ok := sub.(dependencyAndSubscriber)
			if !ok {
				panic("not a dependencyAndSubscriber")
			}
			subSubs := subDep.subs()
			if subSubs != nil {
				if subSubs.nextSub != nil {
					subSubs.prevSub = subs
					link = subSubs
					subs = subSubs
					targetFlag = fPendingComputed
					stack++
				} else {
					link = subSubs
					if subFlags&fEffect != 0 {
						targetFlag = fPendingEffect
					} else {
						targetFlag = fPendingComputed
					}
				}
				continue
			}
			if subFlags&fEffect != 0 {
				effect := mustEffect(sub)
				if rs.queuedEffectsTail != nil {
					rs.queuedEffectsTail.depsTail().nextDep = sub.deps()
				} else {
					rs.queuedEffects = effect
				}
				rs.queuedEffectsTail = effect
			}
		} else if subFlags&(fTracking|targetFlag) == 0 {
			sub.setFlags(subFlags | targetFlag | fNotified)
			if subFlags&(fEffect|fNotified) == fEffect {
				effect := mustEffect(sub)
				if rs.queuedEffectsTail != nil {
					rs.queuedEffectsTail.depsTail().nextDep = sub.deps()
				} else {
					rs.queuedEffects = effect
				}
				rs.queuedEffectsTail = effect
			}
		} else if subFlags&targetFlag == 0 &&
			subFlags&fPropagated != 0 &&
			rs.isValidLink(link, sub) {
			sub.setFlags(subFlags | targetFlag)
		}

		if link = subs.nextSub; link != nil {
			subs = link
			if stack != 0 {
				targetFlag = fPendingComputed
			} else {
				targetFlag = fDirty
			}
			continue
		}

		for stack != 0 {
			stack--
			dep := subs.dep
			depSubs := dep.subs()
			subs = depSubs.prevSub
			depSubs.prevSub = nil

			link = subs.nextSub
			if link != nil {
				subs = link
				if stack != 0 {
					targetFlag = fPendingComputed
				} else {
					targetFlag = fDirty
				}
				continue top
			}
		}
		break
	}
}

// Quickly propagates PendingComputed status to Dirty for each subscriber in the chain.
//
// If the subscriber is also marked as an effect, it is added to the queuedEffects list
// for later processing.
//
// @param link - The head of the linked list to process.
func (rs *ReactiveSystem) shallowPropagate(link *link) {
	for {
		sub := link.sub
		subFlags := sub.flags()
		justPendingDirty := subFlags & (fPendingComputed | fDirty)
		if justPendingDirty == fPendingComputed {
			sub.setFlags(subFlags | fDirty | fNotified)
			if subFlags&(fEffect|fNotified) == fEffect {
				effect := mustEffect(sub)
				if rs.queuedEffectsTail != nil {
					rs.queuedEffectsTail.depsTail().nextDep = sub.deps()
				} else {
					rs.queuedEffects = effect
				}

				rs.queuedEffectsTail = effect
			}
		}
		link = link.nextSub

		if link == nil {
			break
		}
	}
}

// Prepares the given subscriber to track new dependencies.
//
// It resets the subscriber's internal pointers (e.g., depsTail) and
// sets its flags to indicate it is now tracking dependency links.
//
// @param sub - The subscriber to start tracking.
func (rs *ReactiveSystem) startTracking(sub subscriber) {
	sub.setDepsTail(nil)
	flags := sub.flags()
	revised := flags & ^(fNotified|fRecursed|fPropagated) | fTracking
	sub.setFlags(revised)
}

// Concludes tracking of dependencies for the specified subscriber.
//
// It clears or unlinks any tracked dependency information, then
// updates the subscriber's flags to indicate tracking is complete.
//
// @param sub - The subscriber whose tracking is ending.
func (rs *ReactiveSystem) endTracking(sub subscriber) {
	depsTail := sub.depsTail()
	if depsTail != nil {
		nextDep := depsTail.nextDep
		if nextDep != nil {
			rs.clearTracking(nextDep)
			depsTail.nextDep = nil
		}
	} else {
		deps := sub.deps()
		if deps != nil {
			rs.clearTracking(deps)
		}
		sub.setDeps(nil)
	}
	sub.setFlags(sub.flags() & ^fTracking)
}

// Clears dependency-subscription relationships starting at the given link.
//
// Detaches the link from both the dependency and subscriber, then continues
// to the next link in the chain. The link objects are returned to linkPool for reuse.
//
// @param link - The head of a linked chain to be cleared.
func (rs *ReactiveSystem) clearTracking(link *link) {
	for {
		dep := link.dep
		nextDep := link.nextDep
		nextSub := link.nextSub
		prevSub := link.prevSub

		if nextSub != nil {
			nextSub.prevSub = prevSub
		} else {
			dep.setSubsTail(prevSub)
		}

		if prevSub != nil {
			prevSub.nextSub = nextSub
		} else {
			dep.setSubs(nextSub)
		}

		depSub, ok := dep.(dependencyAndSubscriber)
		ss := dep.subs()
		if ss == nil && ok {
			depFlags := depSub.flags()
			if depFlags&fDirty == 0 {
				depSub.setFlags(depFlags | fDirty)
			}

			depDeps := depSub.deps()
			if depDeps != nil {
				link = depDeps
				dt := depSub.depsTail()
				dt.nextDep = nextDep
				depSub.setDeps(nil)
				depSub.setDepsTail(nil)
				continue
			}
		}
		link = nextDep

		if link == nil {
			break
		}
	}
}

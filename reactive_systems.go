package alien

type OnErrorFunc func(from SignalAware, err error)

type ReactiveSystem struct {
	batchDepth        int
	activeSub         subscriber
	queuedEffects     *OneWayLink_subscriber
	queuedEffectsTail *OneWayLink_subscriber

	activeScope subscriber
	onError     OnErrorFunc
	pauseStack  []subscriber
}

type SignalAware interface {
	isSignalAware()
}

type OneWayLink_subscriber struct {
	target subscriber
	linked *OneWayLink_subscriber
}

type OneWayLink_link struct {
	target *link
	linked *OneWayLink_link
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
func (rs *ReactiveSystem) updateDirtyFlag(sub *baseSubscriber, flags subscriberFlags) bool {
	if rs.checkDirty(sub.deps) {
		sub.flags = flags | fDirty
		return true
	}
	sub.flags = flags & ^fPendingComputed
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
func (rs *ReactiveSystem) checkDirty(current *link) bool {
	prevLinks := (*OneWayLink_link)(nil)
	checkDepth := 0

top:
	for {
		dep := current.dep
		dep2 := dep.dep()
		dependencySubscriber, isDependencySubscriber := dep.(subscriber)
		if isDependencySubscriber {
			sub := dependencySubscriber.sub()
			depFlags := sub.flags
			if depFlags&(fComputed|fDirty) == fComputed|fDirty {
				if updateComputed(rs, dep) {
					subs := dep2.subs
					if subs.nextSub != nil {
						rs.shallowPropagate(subs)
					}
					for checkDepth != 0 {
						checkDepth--
						computed := current.sub.(dependency)
						firstSub := computed.dep().subs

						if updateComputed(rs, computed) {
							if firstSub.nextSub != nil {
								rs.shallowPropagate(firstSub)
								current = prevLinks.target
								prevLinks = prevLinks.linked
							} else {
								current = firstSub
							}
							continue
						}

						if firstSub.nextSub != nil {
							if current = prevLinks.target.nextDep; current == nil {
								return false
							}
							prevLinks = prevLinks.linked
							continue top
						}

						return false
					}
				}
			} else if depFlags&(fComputed|fPendingComputed) == fComputed|fPendingComputed {
				depSubs := dep2.subs
				if depSubs.nextSub != nil {
					depSubs.prevSub = current
				}
				current = sub.deps
				checkDepth++
				continue
			}
		}

		if current = current.nextDep; current == nil {
			return false
		}
	}
}

// Links a given dependency and subscriber if they are not already linked.
//
// @param dep - The dependency to be linked.
// @param sub - The subscriber that depends on this dependency.
// @returns The newly created link object if the two are not already linked; otherwise `undefined`.
func (rs *ReactiveSystem) link(dep dependency, sub subscriber) *link {
	dep2 := dep.dep()
	sub2 := sub.sub()
	currentDep := sub2.depsTail
	if currentDep != nil && currentDep.dep == dep {
		return nil
	}

	var nextDep *link
	if currentDep != nil {
		nextDep = currentDep.nextDep
	} else {
		nextDep = sub2.deps
	}
	if nextDep != nil && nextDep.dep == dep {
		sub2.depsTail = nextDep
		return nil
	}

	depLastSub := dep2.subsTail
	if depLastSub != nil && depLastSub.sub == sub && rs.isValidLink(depLastSub, sub2) {
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
func (rs *ReactiveSystem) isValidLink(checkLink *link, sub *baseSubscriber) bool {
	depsTails := sub.depsTail
	if depsTails != nil {
		link := sub.deps
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
	dep2 := dep.dep()
	sub2 := sub.sub()
	newLink := &link{
		dep:     dep,
		sub:     sub,
		nextDep: nextDep,
	}

	if depsTail == nil {
		sub2.deps = newLink
	} else {
		depsTail.nextDep = newLink
	}

	if dep2.subs == nil {
		dep2.subs = newLink
	} else {
		oldTail := dep2.subsTail
		newLink.prevSub = oldTail
		oldTail.nextSub = newLink
	}

	sub2.depsTail = newLink
	dep2.subsTail = newLink

	return newLink
}

// Traverses and marks subscribers starting from the provided link.
//
// It sets flags (e.g., Dirty, PendingComputed, PendingEffect) on each subscriber
// to indicate which ones require re-computation or effect processing.
// This function should be called after a signal's value changes.
//
// @param link - The starting link from which propagation begins.
func (rs *ReactiveSystem) propagate(current *link) {
	next := current.nextSub
	branchs := (*OneWayLink_link)(nil)
	branchDepth := 0
	targetFlag := fDirty

top:
	for {
		sub := current.sub
		sub2 := sub.sub()
		subFlags := sub2.flags
		shouldNotify := false

		if subFlags&(fTracking|fRecursed|fPropagated) == 0 {
			sub2.flags = subFlags | targetFlag | fNotified
			shouldNotify = true
		} else if subFlags&fRecursed != 0 && subFlags&fTracking == 0 {
			sub2.flags = subFlags&^fRecursed | targetFlag | fNotified
			shouldNotify = true
		} else if subFlags&fPropagated == 0 && rs.isValidLink(current, sub2) {
			sub2.flags = subFlags | fRecursed | targetFlag | fNotified
			subDep, ok := sub.(dependency)
			shouldNotify = ok && subDep.dep().subs != nil
		}

		if shouldNotify {
			subDep, ok := sub.(dependency)
			if !ok {
				panic("not a dependency")
			}
			subSubs := subDep.dep().subs
			if subSubs != nil {
				current = subSubs
				if subSubs.nextSub != nil {
					branchs = &OneWayLink_link{target: current, linked: branchs}
					branchDepth++
					next = current.nextSub
					targetFlag = fPendingComputed
				} else {
					if subFlags&fEffect != 0 {
						targetFlag = fPendingEffect
					} else {
						targetFlag = fPendingComputed
					}
				}
				continue
			}
			if subFlags&fEffect != 0 {
				if rs.queuedEffectsTail != nil {
					rs.queuedEffectsTail.linked = &OneWayLink_subscriber{target: sub}
					rs.queuedEffectsTail = rs.queuedEffectsTail.linked
				} else {
					rs.queuedEffects = &OneWayLink_subscriber{target: sub}
					rs.queuedEffectsTail = rs.queuedEffects
				}
			}
		} else if subFlags&(fTracking|targetFlag) == 0 {
			sub2.flags = subFlags | targetFlag | fNotified
			if subFlags&(fEffect|fNotified) == fEffect {
				if rs.queuedEffectsTail != nil {
					rs.queuedEffectsTail.linked = &OneWayLink_subscriber{target: sub}
					rs.queuedEffectsTail = rs.queuedEffectsTail.linked
				} else {
					rs.queuedEffects = &OneWayLink_subscriber{target: sub}
					rs.queuedEffectsTail = rs.queuedEffects
				}
			}
		} else if subFlags&targetFlag == 0 &&
			subFlags&fPropagated != 0 &&
			rs.isValidLink(current, sub2) {
			sub2.flags = subFlags | targetFlag
		}

		if current = next; current != nil {
			next = current.nextSub
			if branchDepth != 0 {
				targetFlag = fPendingComputed
			} else {
				targetFlag = fDirty
			}
			continue
		}

		for branchDepth != 0 {
			branchDepth--
			current = branchs.target
			branchs = branchs.linked
			if current != nil {
				next = current.nextSub
				if branchDepth != 0 {
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
		sub2 := sub.sub()
		subFlags := sub2.flags
		justPendingDirty := subFlags & (fPendingComputed | fDirty)
		if justPendingDirty == fPendingComputed {
			sub2.flags = subFlags | fDirty | fNotified
			if subFlags&(fEffect|fNotified) == fEffect {
				if rs.queuedEffectsTail != nil {
					rs.queuedEffectsTail.linked = &OneWayLink_subscriber{target: sub}
					rs.queuedEffectsTail = rs.queuedEffectsTail.linked
				} else {
					rs.queuedEffects = &OneWayLink_subscriber{target: sub}
					rs.queuedEffectsTail = rs.queuedEffects
				}
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
func (rs *ReactiveSystem) startTracking(sub *baseSubscriber) {
	sub.depsTail = nil
	flags := sub.flags
	revised := flags & ^(fNotified|fRecursed|fPropagated) | fTracking
	sub.flags = revised
}

// Concludes tracking of dependencies for the specified subscriber.
//
// It clears or unlinks any tracked dependency information, then
// updates the subscriber's flags to indicate tracking is complete.
//
// @param sub - The subscriber whose tracking is ending.
func (rs *ReactiveSystem) endTracking(sub *baseSubscriber) {
	depsTail := sub.depsTail
	if depsTail != nil {
		nextDep := depsTail.nextDep
		if nextDep != nil {
			rs.clearTracking(nextDep)
			depsTail.nextDep = nil
		}
	} else {
		deps := sub.deps
		if deps != nil {
			rs.clearTracking(deps)
		}
		sub.deps = nil
	}
	sub.flags = sub.flags & ^fTracking
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
		dep2 := dep.dep()
		nextDep := link.nextDep
		nextSub := link.nextSub
		prevSub := link.prevSub

		if nextSub != nil {
			nextSub.prevSub = prevSub
		} else {
			dep2.subsTail = prevSub
		}

		if prevSub != nil {
			prevSub.nextSub = nextSub
		} else {
			dep2.subs = nextSub
		}

		depSub, ok := dep.(subscriber)
		ss := dep2.subs
		if ss == nil && ok {
			depSub2 := depSub.sub()
			depFlags := depSub2.flags
			if depFlags&fDirty == 0 {
				depSub2.flags = depFlags | fDirty
			}

			depDeps := depSub2.deps
			if depDeps != nil {
				link = depDeps
				dt := depSub2.depsTail
				dt.nextDep = nextDep
				depSub2.deps = nil
				depSub2.depsTail = nil
				continue
			}
		}
		link = nextDep

		if link == nil {
			break
		}
	}
}

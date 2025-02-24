package alien

type subscriberFlags uint16

const (
	fComputed subscriberFlags = 1 << iota
	fEffect
	fTracking
	fNotified
	fRecursed
	fDirty
	fPendingComputed
	fPendingEffect
	fEffectScope
	fPropagated subscriberFlags = fDirty | fPendingComputed | fPendingEffect
)

type link struct {
	dep dependency
	sub subscriber
	// reused to link the previous stack in updateDirtyFlags
	// reused to link the previous stack in propagate
	prevSub *link
	nextSub *link
	// reused to link the notify effect in queueEffects
	nextDep *link
}

type subscriber interface {
	flags() subscriberFlags
	setFlags(subscriberFlags)
	deps() *link
	setDeps(*link)
	depsTail() *link
	setDepsTail(*link)
}

type baseSubscriber struct {
	_flags           subscriberFlags
	_deps, _depsTail *link
}

func (s *baseSubscriber) flags() subscriberFlags {
	return s._flags
}
func (s *baseSubscriber) setFlags(flags subscriberFlags) {
	s._flags = flags
}
func (s *baseSubscriber) deps() *link {
	return s._deps
}
func (s *baseSubscriber) setDeps(deps *link) {
	s._deps = deps
}
func (s *baseSubscriber) depsTail() *link {
	return s._depsTail
}
func (s *baseSubscriber) setDepsTail(depsTail *link) {
	s._depsTail = depsTail
}

type dependency interface {
	subs() *link
	setSubs(*link)
	subsTail() *link
	setSubsTail(*link)
}

type baseDependency struct {
	_subs, _subsTail *link
}

func (d *baseDependency) subs() *link {
	return d._subs
}

func (d *baseDependency) setSubs(subs *link) {
	d._subs = subs
}

func (d *baseDependency) subsTail() *link {
	return d._subsTail
}

func (d *baseDependency) setSubsTail(subsTail *link) {
	d._subsTail = subsTail
}

type dependencyAndSubscriber interface {
	dependency
	subscriber
}

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
	dep     dependency
	sub     subscriber
	prevSub *link
	nextSub *link
	nextDep *link
}

type subscriber interface {
	sub() *baseSubscriber
}

type dependency interface {
	dep() *baseDependency
}

type baseSubscriber struct {
	flags          subscriberFlags
	deps, depsTail *link
}

type baseDependency struct {
	subs, subsTail *link
}

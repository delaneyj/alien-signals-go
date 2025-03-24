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
	dep     *signal
	sub     *signal
	prevSub *link
	nextSub *link
	nextDep *link
}

type signal struct {
	ref                            interface{}
	flags                          subscriberFlags
	deps, depsTail, subs, subsTail *link
}

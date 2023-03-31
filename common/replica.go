package common

type iReplica interface {
	process(i *invocation)
}

type replica struct {
	*godes.Runner
	arrivalCond    		*godes.BooleanControl
	arrivalQueue		*godes.FIFOQueue
	frp					*resourceProvisioner
	replicaID			string
	appID              	string
	funcID             	string
	idlenessDeadline 	time.Duration
}

func newReplica(frp *resourceProvisioner, rid, aid, fid string) *replica {
	return &replica{
		Runner:			&godes.Runner{},
		arrivalCond:    godes.NewBooleanControl(),
		arrivalQueue:	godes.NewFIFOQueue("arrival"),
		frp:			frp,
		replicaID		rid,
		appID:			aid,
		funcID:			fid,
	}
}

func (r *replica) process(i *invocation) {
	r.arrivalQueue.Place(i)
	r.arrivalCond.Set(true)
}

func (r *replica) tailLatency() float64 {
	return 0
}

func (r *replica) terminate() {
	r.terminated = true
	r.arrivalCond.Set(true)
}

func (r *replica) Run() {
	for {
		r.arrivalCond.Wait(true)
		i := r.arrivalQueue.Get().(*invocation)
		i.updateHops(r.replicaID)
		dur := i.getDuration() + r.tailLatency()
		godes.Advance(dur)

		i.setProcessedTs(godes.GetSystemTime())
		i.updateHopResponse(dur)
		r.frp.setAvailable(r)
		if r.terminated { break }
		r.arrivalCond.Set(false)
	}
}
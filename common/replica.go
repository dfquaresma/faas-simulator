package common

import (
	"time"
	"strconv"

	"github.com/agoussia/godes"
)

type iReplica interface {
	process(i *invocation)
	getOutPut() [][]string
	terminate()
}

type replica struct {
	*godes.Runner
	arrivalCond      *godes.BooleanControl
	arrivalQueue     *godes.FIFOQueue
	frp              *resourceProvisioner
	replicaID        string
	appID            string
	funcID           string
	idlenessDeadline time.Duration
	terminated       bool
	busyTime         float64
	upTime           float64
	tailLatency      float64
	tailLatencyProb  float64
	reqsProcessed    int
}

func newReplica(frp *resourceProvisioner, rid, aid, fid string, tl, tlp float64) *replica {
	return &replica{
		Runner:          &godes.Runner{},
		arrivalCond:     godes.NewBooleanControl(),
		arrivalQueue:    godes.NewFIFOQueue("arrival"),
		frp:             frp,
		replicaID:       rid,
		appID:           aid,
		funcID:          fid,
		tailLatency:     tl,
		tailLatencyProb: tlp,
	}
}

func (r *replica) process(i *invocation) {
	r.arrivalQueue.Place(i)
	r.arrivalCond.Set(true)
}

func (r *replica) getTailLatency() float64 {
	return r.tailLatency
}

func (r *replica) terminate() {
	r.terminated = true
	r.arrivalCond.Set(true)
}

func (r *replica) Run() {
	start := godes.GetSystemTime()
	for {
		r.arrivalCond.Wait(true)
		if r.arrivalQueue.Len() > 0 {
			i := r.arrivalQueue.Get().(*invocation)
			i.updateHops(r.replicaID)
			dur := i.getDuration() + r.getTailLatency()
			godes.Advance(dur)
			r.busyTime += dur

			i.setProcessedTs(godes.GetSystemTime())
			i.updateHopResponse(dur)
			r.frp.setAvailable(r)
			r.reqsProcessed += 1
		}
		r.arrivalCond.Set(false)
		if r.terminated {
			r.upTime = godes.GetSystemTime() - start
			break
		}
	}
}

func (r *replica) getOutPut() []string {
	return []string{
		r.replicaID,
		r.frp.frpID,
		r.appID,
		r.funcID,
		strconv.FormatFloat(r.busyTime, 'f', -1, 64),
		strconv.FormatFloat(r.upTime, 'f', -1, 64),		
		strconv.Itoa(r.reqsProcessed),
	}
}

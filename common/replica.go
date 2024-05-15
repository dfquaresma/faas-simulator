package common

import (
	"math"
	"strconv"

	"github.com/agoussia/godes"
)

type replica struct {
	*godes.Runner
	arrivalCond      *godes.BooleanControl
	terminatedCond   *godes.BooleanControl
	arrivalQueue     *godes.FIFOQueue
	rp               *resourceProvisioner
	replicaID        string
	appID            string
	funcID           string
	idlenessDeadline float64
	startTS          float64
	shutdownTS       float64
	lastWorkTS       float64
	busyTime         float64
	upTime           float64
	reqsProcessed    int
}

func newReplica(rp *resourceProvisioner, rid, aid, fid string, idl float64) *replica {
	return &replica{
		Runner:           &godes.Runner{},
		arrivalCond:      godes.NewBooleanControl(),
		terminatedCond:   godes.NewBooleanControl(),
		arrivalQueue:     godes.NewFIFOQueue("arrival"),
		rp:               rp,
		replicaID:        rid,
		appID:            aid,
		funcID:           fid,
		idlenessDeadline: idl,
	}
}

func (r *replica) process(i *invocation) {
	r.arrivalQueue.Place(i)
	r.arrivalCond.Set(true)
}

func (r *replica) Run() {
	r.startTS = godes.GetSystemTime()
	for {
		r.arrivalCond.Wait(true)
		if r.arrivalQueue.Len() > 0 {
			i := r.arrivalQueue.Get().(*invocation)

			forwardLatency := r.rp.cfg.ForwardLatency
			godes.Advance(forwardLatency)
			i.updateHopResponse(forwardLatency)
			i.updateHops(r.replicaID)

			if i.isTailLatency() {
				tailLatency := i.getDuration() - i.getP95()
				shouldSkipReq, timeToWaste := r.rp.warnReqLatency(i, tailLatency)
				if shouldSkipReq {
					godes.Advance(timeToWaste)
					r.busyTime += timeToWaste
					r.lastWorkTS = godes.GetSystemTime()
					r.rp.setAvailable(r)
					continue
				}
			}

			dur := i.getDuration()
			godes.Advance(dur)
			r.busyTime += dur

			r.lastWorkTS = godes.GetSystemTime()
			i.addProcessedTs(r.lastWorkTS)
			i.updateHopResponse(dur)
			r.rp.response(i)
			r.rp.setAvailable(r)
			r.reqsProcessed += 1
		}
		if r.arrivalQueue.Len() == 0 {
			r.arrivalCond.Set(false)
			if r.terminatedCond.GetState() {
				r.shutdownTS = godes.GetSystemTime()
				if r.idlenessDeadline >= 0 {
					r.shutdownTS = math.Min(godes.GetSystemTime(), r.lastWorkTS+r.idlenessDeadline)
				}
				r.upTime = r.shutdownTS - r.startTS
				break
			}
		}
	}
}

func (r *replica) terminate() {
	r.terminatedCond.Set(true)
	r.arrivalCond.Set(true)
}

func (r *replica) getOutPut() []string {
	return []string{
		r.replicaID,
		r.rp.rpID,
		r.appID,
		r.funcID,
		strconv.FormatFloat(r.busyTime, 'f', -1, 64),
		strconv.FormatFloat(r.upTime, 'f', -1, 64),
		strconv.Itoa(r.reqsProcessed),
		strconv.FormatFloat(r.lastWorkTS, 'f', -1, 64),
		strconv.FormatFloat(r.shutdownTS, 'f', -1, 64),
	}
}

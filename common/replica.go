package common

import (
	"math"
	"math/rand"
	"strconv"
	"time"

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
	terminated       bool
	idlenessDeadline float64
	startTS          float64
	lastWorkTS       float64
	busyTime         float64
	upTime           float64
	tailLatency      float64
	tailLatencyProb  float64
	reqsProcessed    int
}

func newReplica(frp *resourceProvisioner, rid, aid, fid string, tl, tlp, idl float64) *replica {
	return &replica{
		Runner:           &godes.Runner{},
		arrivalCond:      godes.NewBooleanControl(),
		arrivalQueue:     godes.NewFIFOQueue("arrival"),
		frp:              frp,
		replicaID:        rid,
		appID:            aid,
		funcID:           fid,
		tailLatency:      tl,
		tailLatencyProb:  tlp,
		idlenessDeadline: idl,
	}
}

func (r *replica) process(i *invocation) {
	r.arrivalQueue.Place(i)
	r.arrivalCond.Set(true)
}

func (r *replica) getTailLatency() float64 {
	tailLatency := 0.0
	rand.Seed(time.Now().UnixNano())
	if r.tailLatencyProb >= rand.Float64() {
		tailLatency = r.tailLatency
	}
	return tailLatency
}

func (r *replica) terminate() {
	r.terminated = true
	r.arrivalCond.Set(true)
}

func (r *replica) Run() {
	r.startTS = godes.GetSystemTime()
	for {
		r.arrivalCond.Wait(true)
		if r.arrivalQueue.Len() > 0 {
			i := r.arrivalQueue.Get().(*invocation)

			forwardLatency := r.frp.cfg.ForwardLatency
			godes.Advance(forwardLatency)
			i.updateHopResponse(forwardLatency)

			i.updateHops(r.replicaID)
			tailLatency := r.getTailLatency()

			shouldSkipReq, timeToWaste := r.frp.warnReqLatency(i, tailLatency)
			if shouldSkipReq {
				godes.Advance(timeToWaste)
				r.busyTime += timeToWaste
				r.lastWorkTS = godes.GetSystemTime()
				r.frp.setAvailable(r)
				continue
			}

			dur := i.getDuration() + tailLatency
			godes.Advance(dur)
			r.busyTime += dur

			r.lastWorkTS = godes.GetSystemTime()
			i.addProcessedTs(r.lastWorkTS)
			i.updateHopResponse(dur)
			r.frp.response(i)
			r.frp.setAvailable(r)
			r.reqsProcessed += 1
		}
		r.arrivalCond.Set(false)
		if r.terminated {
			shutdownTS := math.Min(godes.GetSystemTime(), r.lastWorkTS+r.idlenessDeadline)
			r.upTime = shutdownTS - r.startTS
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

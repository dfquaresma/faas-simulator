package common

import (
	"math"
	"strconv"

	"github.com/agoussia/godes"
	"github.com/dfquaresma/faas-simulator/model"
)

type replica struct {
	*godes.Runner
	arrivalCond    *godes.BooleanControl
	terminatedCond *godes.BooleanControl
	arrivalQueue   *godes.FIFOQueue
	rp             *resourceProvisioner
	replicaID      string
	appID          string
	funcID         string
	cfg            model.Config
	startTS        float64
	shutdownTS     float64
	lastWorkTS     float64
	busyTime       float64
	upTime         float64
	coldstart      float64
	reqsProcessed  int
}

func newReplica(rp *resourceProvisioner, rid, aid, fid string, cfg model.Config, coldstart float64) *replica {
	return &replica{
		Runner:         &godes.Runner{},
		arrivalCond:    godes.NewBooleanControl(),
		terminatedCond: godes.NewBooleanControl(),
		arrivalQueue:   godes.NewFIFOQueue(rid),
		rp:             rp,
		replicaID:      rid,
		appID:          aid,
		funcID:         fid,
		cfg:            cfg,
		coldstart:      coldstart,
	}
}

func (r *replica) process(i *model.Invocation) {
	r.arrivalQueue.Place(i)
	r.arrivalCond.Set(true)
}

func (r *replica) Run() {
	r.startTS = godes.GetSystemTime()
	for {
		r.arrivalCond.Wait(true)
		if r.arrivalQueue.Len() > 0 {
			i := r.arrivalQueue.Get().(*model.Invocation)

			forwardLatency := r.cfg.ForwardLatency
			godes.Advance(forwardLatency)
			i.UpdateHopResponse(forwardLatency)
			i.UpdateHops(r.replicaID)

			if r.reqsProcessed == 0 {
				// first Req of this replica
				i.SetDuration(r.coldstart)
				i.SetAsColdStart()

			} else if i.IsTailLatency() {
				tailLatency := i.GetDuration() - i.GetTailLatencyThreshold()
				shouldSkipReq, timeToWaste := r.rp.warnReqLatency(i, tailLatency)
				if shouldSkipReq {
					r.busyTime += timeToWaste
					r.lastWorkTS = godes.GetSystemTime() + timeToWaste
					r.rp.setAvailable(r)
					continue
				}
			}

			dur := i.GetDuration()
			godes.Advance(dur)
			r.busyTime += dur

			r.lastWorkTS = godes.GetSystemTime()
			i.AddProcessedTs(r.lastWorkTS)
			i.UpdateHopResponse(dur)
			r.rp.response(i)
			r.reqsProcessed += 1
		}
		if r.arrivalQueue.Len() == 0 {
			r.arrivalCond.Set(false)
			if r.terminatedCond.GetState() {
				r.shutdownTS = godes.GetSystemTime()
				if r.cfg.Idletime >= 0 {
					r.shutdownTS = math.Min(r.shutdownTS, r.lastWorkTS+r.cfg.Idletime)
				}
				r.upTime = r.shutdownTS - r.startTS
				break
			}
			r.rp.setAvailable(r)
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

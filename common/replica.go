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
	isBusy         *godes.BooleanControl
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
	reqsProcessed  int
}

func newReplica(rp *resourceProvisioner, rid, aid, fid string, cfg model.Config) *replica {
	return &replica{
		Runner:         &godes.Runner{},
		arrivalCond:    godes.NewBooleanControl(),
		terminatedCond: godes.NewBooleanControl(),
		isBusy:         godes.NewBooleanControl(),
		arrivalQueue:   godes.NewFIFOQueue(rid),
		rp:             rp,
		replicaID:      rid,
		appID:          aid,
		funcID:         fid,
		cfg:            cfg,
	}
}

func (r *replica) process(i *model.Invocation) {
	r.arrivalQueue.Place(i)
	r.arrivalCond.Set(true)
}

func (r *replica) IsBusy() bool {
	return r.isBusy.GetState()
}

func (r *replica) SetBusy() {
	r.isBusy.Set(true)
}

func (r *replica) Run() {
	r.startTS = godes.GetSystemTime()
	for {
		r.arrivalCond.Wait(true)
		if r.arrivalQueue.Len() > 0 {
			r.isBusy.Set(true)
			i := r.arrivalQueue.Get().(*model.Invocation)

			forwardLatency := r.cfg.ForwardLatency
			godes.Advance(forwardLatency)
			i.UpdateHopResponse(forwardLatency)
			i.UpdateHops(r.replicaID)

			dur := i.GetDuration()
			if r.reqsProcessed == 0 {
				// first Req of this replica
				i.SetDuration(i.GetP100())
				i.SetAsColdStart()
				dur = i.GetP100()
			}

			if i.IsTailLatency() {
				tailLatencyThreshold := i.GetTailLatencyThreshold()
				tailLatency := i.GetDuration() - tailLatencyThreshold
				switch r.cfg.Technique {
				case "GCI":
					r.rp.warnReqLatency(i)
					godes.Advance(tailLatencyThreshold)
					r.busyTime += tailLatencyThreshold
					r.lastWorkTS = godes.GetSystemTime()
					r.isBusy.Set(false)
					r.rp.setAvailable(r)
					continue

				case "RequestHedgingOpt":
					godes.Advance(tailLatencyThreshold)
					r.busyTime += tailLatencyThreshold
					dur = tailLatency
					r.rp.warnReqLatency(i)
				}
			}

			godes.Advance(dur)
			r.busyTime += dur

			r.lastWorkTS = godes.GetSystemTime()
			i.AddProcessedTs(r.lastWorkTS)
			i.UpdateHopResponse(i.GetDuration())
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
			r.isBusy.Set(false)
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

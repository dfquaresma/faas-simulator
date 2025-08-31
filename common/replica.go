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
	r.isBusy.Set(true)
	r.arrivalCond.Set(true)
}

func (r *replica) Run() {
	r.startTS = godes.GetSystemTime()
	r.rp.notifyReadyness(r.funcID, godes.GetSystemTime())
	for {
		r.arrivalCond.Wait(true)
		if r.arrivalQueue.Len() > 0 {
			i := r.arrivalQueue.Get().(*model.Invocation)

			forwardLatency := r.cfg.ForwardLatency
			godes.Advance(forwardLatency)
			i.UpdateResponse(forwardLatency, r.replicaID)

			dur := i.GetDuration()
			if r.reqsProcessed == 0 {
				i.SetAsColdStart()
				dur = i.GetColdStart()
			}

			if i.IsTailLatency() {
				tailLatencyThreshold := i.GetTailLatencyThreshold()
				i.UpdateResponse(tailLatencyThreshold, r.replicaID)
				dur = i.GetDuration() - tailLatencyThreshold // dur is now the surplus latency after the threshold
				switch r.cfg.Technique {
				case "GCI":
					if !i.IsColdStart() {
						r.rp.warnReqLatency(i)
						godes.Advance(tailLatencyThreshold)
						r.busyTime += tailLatencyThreshold
						r.lastWorkTS = godes.GetSystemTime()

						if r.terminatedCond.GetState() {
							r.setUptimeStats()
							break
						}

						r.isBusy.Set(false)
						r.rp.setAvailable(r)
						continue
					}

				case "RequestHedgingOpt":
					if !i.IsCopy() {
						godes.Advance(tailLatencyThreshold)
						r.busyTime += tailLatencyThreshold
						r.rp.warnReqLatency(i)
					}
				}
			}

			godes.Advance(dur)
			i.UpdateResponse(dur, r.replicaID)

			r.busyTime += dur
			r.lastWorkTS = godes.GetSystemTime()
			i.AddProcessedTs(r.lastWorkTS)

			r.rp.response(i)
			r.reqsProcessed += 1
		}

		if r.arrivalQueue.Len() == 0 {
			if r.terminatedCond.GetState() {
				r.setUptimeStats()
				r.rp.notifyTermination(r.funcID, godes.GetSystemTime())
				break
			}
			r.arrivalCond.Set(false)
			r.isBusy.Set(false)
			r.rp.setAvailable(r)
		}
	}
}

func (r *replica) setUptimeStats() {
	r.shutdownTS = godes.GetSystemTime()
	if r.cfg.Idletime >= 0 {
		r.shutdownTS = math.Min(r.shutdownTS, r.lastWorkTS+r.cfg.Idletime)
	}
	r.upTime = r.shutdownTS - r.startTS
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
		strconv.FormatFloat(r.startTS, 'f', -1, 64),
		strconv.FormatFloat(r.shutdownTS, 'f', -1, 64),
	}
}

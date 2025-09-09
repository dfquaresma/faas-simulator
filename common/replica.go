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
	provisioner    *provisioner
	replicaID      string
	appID          string
	funcID         string
	cfg            model.Config
	startTS        float64
	shutdownTS     float64
	lastWorkTS     float64
	busyTime       float64
	upTime         float64
	reqsCount      int
}

func newReplica(p *provisioner, rid, aid, fid string, cfg model.Config) *replica {
	return &replica{
		Runner:         &godes.Runner{},
		arrivalCond:    godes.NewBooleanControl(),
		terminatedCond: godes.NewBooleanControl(),
		isBusy:         godes.NewBooleanControl(),
		arrivalQueue:   godes.NewFIFOQueue(rid),
		provisioner:    p,
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
	r.provisioner.notifyReadyness(r.funcID, godes.GetSystemTime())
	for {
		r.arrivalCond.Wait(true)
		if r.arrivalQueue.Len() > 0 {
			i := r.arrivalQueue.Get().(*model.Invocation)
			if r.reqsCount == 0 {
				i.SetAsColdStart()
			}

			forwardLatency := r.cfg.ForwardLatency
			if forwardLatency != 0 {
				godes.Advance(forwardLatency)
			}
			i.UpdateResponse(forwardLatency, r.replicaID)

			delay := r.provisioner.getTechniqueDelay(i)
			if delay != 0 {
				godes.Advance(delay)
				r.busyTime += delay
				i.UpdateResponse(delay, r.replicaID)
			}
			dur := i.GetDuration() - delay // dur is now the surplus latency after delay to warn TLTT

			shouldTimeout, timeToTimeout := r.provisioner.triggerTechnique(i)
			if shouldTimeout {
				godes.Advance(timeToTimeout)
				r.busyTime += timeToTimeout
				r.lastWorkTS = godes.GetSystemTime()

				if r.terminatedCond.GetState() {
					r.setUptimeStats()
					break
				}

				r.isBusy.Set(false)
				r.provisioner.setAvailable(r)
				continue
			}

			godes.Advance(dur)
			i.UpdateResponse(dur, r.replicaID)

			r.busyTime += dur
			r.lastWorkTS = godes.GetSystemTime()
			i.AddProcessedTs(r.lastWorkTS)

			r.provisioner.response(i)
			r.reqsCount += 1
		}

		if r.arrivalQueue.Len() == 0 {
			if r.terminatedCond.GetState() {
				r.setUptimeStats()
				r.provisioner.notifyTermination(r.funcID, godes.GetSystemTime())
				break
			}
			r.arrivalCond.Set(false)
			r.isBusy.Set(false)
			r.provisioner.setAvailable(r)
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

func (r *replica) getRequestCount() int {
	return r.reqsCount
}

func (r *replica) getOutPut() []string {
	return []string{
		r.replicaID,
		r.provisioner.rpID,
		r.appID,
		r.funcID,
		strconv.FormatFloat(r.busyTime, 'f', -1, 64),
		strconv.FormatFloat(r.upTime, 'f', -1, 64),
		strconv.Itoa(r.reqsCount),
		strconv.FormatFloat(r.lastWorkTS, 'f', -1, 64),
		strconv.FormatFloat(r.startTS, 'f', -1, 64),
		strconv.FormatFloat(r.shutdownTS, 'f', -1, 64),
	}
}

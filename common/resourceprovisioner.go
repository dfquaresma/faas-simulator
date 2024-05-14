package common

import (
	"time"

	"github.com/agoussia/godes"
)

type iResourceProvisioner interface {
	forward(i *invocation)
	response(i *invocation, d float64)
	setAvailable(r *replica)
	warnReqLatency(i *invocation, tl float64) (bool, float64)
	terminate()
	getOutPut() [][]string
}

type resourceProvisioner struct {
	*godes.Runner
	arrivalCond       *godes.BooleanControl
	terminatedCond    *godes.BooleanControl
	availableCond     *godes.BooleanControl
	arrivalQueue      *godes.FIFOQueue
	availableReplicas *godes.LIFOQueue
	appID             string
	funcID            string
	rpID              string
	cfg               Config
	replicas          []*replica
	technique         *technique
}

func newResourceProvisioner(aid, fid string, cfg Config) *resourceProvisioner {
	rp := &resourceProvisioner{
		Runner:            &godes.Runner{},
		arrivalCond:       godes.NewBooleanControl(),
		terminatedCond:    godes.NewBooleanControl(),
		availableCond:     godes.NewBooleanControl(),
		arrivalQueue:      godes.NewFIFOQueue("arrival"),
		availableReplicas: godes.NewLIFOQueue("available"),
		appID:             aid,
		funcID:            fid,
		rpID:              aid + "-" + fid,
		replicas:          make([]*replica, 0),
		cfg:               cfg,
	}
	rp.technique = newTechnique(rp, cfg.Technique)
	return rp
}

func (rp *resourceProvisioner) forward(i *invocation) {
	rp.technique.forward(rp, i)
}

func (rp *resourceProvisioner) response(i *invocation) {
	rp.technique.processResponse(i)
}

func (rp *resourceProvisioner) setAvailable(r *replica) {
	rp.availableReplicas.Place(r)
}

func (rp *resourceProvisioner) getAvailableReplica() *replica {
	for rp.availableReplicas.Len() > 0 {
		r := rp.availableReplicas.Get().(*replica)
		if rp.cfg.Idletime < 0 {
			return r
		}
		if godes.GetSystemTime()-r.lastWorkTS < rp.cfg.Idletime {
			return r
		}
		r.terminate()
	}
	replica := newReplica(rp, time.Now().String(), rp.appID, rp.funcID, rp.cfg.TailLatency, rp.cfg.TailLatencyProb, rp.cfg.Idletime)
	godes.AddRunner(replica)
	rp.replicas = append(rp.replicas, replica)
	return replica
}

func (rp *resourceProvisioner) warnReqLatency(i *invocation, tl float64) (bool, float64) {
	return rp.technique.processWarning(i, tl)
}

func (rp *resourceProvisioner) Run() {
	for {
		rp.arrivalCond.Wait(true)
		if rp.arrivalQueue.Len() > 0 {
			i := rp.arrivalQueue.Get().(*invocation)
			r := rp.getAvailableReplica()
			r.process(i)
			continue
		}
		if rp.arrivalQueue.Len() == 0 {
			rp.arrivalCond.Set(false)
			if rp.terminatedCond.GetState() {
				break
			}
		}
	}
}

func (rp *resourceProvisioner) terminate() {
	rp.terminatedCond.Set(true)
	rp.arrivalCond.Set(true)
	rp.arrivalCond.Wait(false)
	for _, r := range rp.replicas {
		r.terminate()
	}
}

func (rp *resourceProvisioner) getOutPut() [][]string {
	res := [][]string{}
	for _, r := range rp.replicas {
		res = append(res, r.getOutPut())
	}
	return res
}

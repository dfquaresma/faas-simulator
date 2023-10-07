package common

import (
	"fmt"

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
	availableCond     *godes.BooleanControl
	arrivalQueue      *godes.FIFOQueue
	availableReplicas *godes.LIFOQueue
	appID             string
	funcID            string
	frpID             string
	terminated        bool
	cfg               Config
	replicas          []*replica
	technique         *technique
}

func newResourceProvisioner(aid, fid string, cfg Config) *resourceProvisioner {
	frp := &resourceProvisioner{
		Runner:            &godes.Runner{},
		arrivalCond:       godes.NewBooleanControl(),
		availableCond:     godes.NewBooleanControl(),
		arrivalQueue:      godes.NewFIFOQueue("arrival"),
		availableReplicas: godes.NewLIFOQueue("available"),
		appID:             aid,
		funcID:            fid,
		frpID:             aid + "-" + fid,
		replicas:          make([]*replica, 0),
		cfg:               cfg,
	}
	frp.technique = newTechnique(frp, cfg.Technique)
	return frp
}

func (frp *resourceProvisioner) forward(i *invocation) {
	frp.arrivalQueue.Place(i)
	frp.arrivalCond.Set(true)
}

func (frp *resourceProvisioner) response(i *invocation, dur float64) {
	frp.technique.processResponse(i, dur)
}

func (frp *resourceProvisioner) setAvailable(r *replica) {
	frp.availableReplicas.Place(r)
}

func (frp *resourceProvisioner) getAvailableReplica() *replica {
	if frp.availableReplicas.Len() > 0 {
		return frp.availableReplicas.Get().(*replica)
	}
	rid := fmt.Sprintf("%d", len(frp.replicas))
	replica := newReplica(frp, rid, frp.appID, frp.funcID, frp.cfg.TailLatency, frp.cfg.TailLatencyProb)
	godes.Advance(frp.cfg.ColdstartLatency)
	godes.AddRunner(replica)
	frp.replicas = append(frp.replicas, replica)
	return replica
}

func (frp *resourceProvisioner) terminate() {
	frp.terminated = true
	frp.arrivalCond.Set(true)
	for _, r := range frp.replicas {
		r.terminate()
	}
}

func (frp *resourceProvisioner) warnReqLatency(i *invocation, tl float64) (bool, float64) {
	return frp.technique.processWarning(i, tl)
}

func (frp *resourceProvisioner) Run() {
	for {
		frp.arrivalCond.Wait(true)
		if frp.arrivalQueue.Len() > 0 {
			i := frp.arrivalQueue.Get().(*invocation)
			r := frp.getAvailableReplica()
			godes.Advance(frp.cfg.ForwardLatency)
			r.process(i)
			continue
		}
		frp.arrivalCond.Set(false)
		if frp.terminated {
			break
		}
	}
}

func (frp *resourceProvisioner) getOutPut() [][]string {
	res := [][]string{}
	for _, r := range frp.replicas {
		res = append(res, r.getOutPut())
	}
	return res
}

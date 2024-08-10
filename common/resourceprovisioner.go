package common

import (
	"time"

	"github.com/agoussia/godes"
	"github.com/dfquaresma/faas-simulator/model"
)

type resourceProvisioner struct {
	*godes.Runner
	arrivalCond       *godes.BooleanControl
	terminatedCond    *godes.BooleanControl
	arrivalQueue      *godes.FIFOQueue
	availableReplicas *godes.LIFOQueue
	appID             string
	funcID            string
	rpID              string
	cfg               model.Config
	replicas          []*replica
	technique         *technique
	lp                *latencyProcessor
}

func newResourceProvisioner(aid, fid string, cfg model.Config) *resourceProvisioner {
	rp := &resourceProvisioner{
		Runner:            &godes.Runner{},
		arrivalCond:       godes.NewBooleanControl(),
		terminatedCond:    godes.NewBooleanControl(),
		arrivalQueue:      godes.NewFIFOQueue("arrival"),
		availableReplicas: godes.NewLIFOQueue("available"),
		appID:             aid,
		funcID:            fid,
		rpID:              aid + "-" + fid,
		replicas:          make([]*replica, 0),
		cfg:               cfg,
	}
	rp.technique = newTechnique(rp, cfg.Technique)
	rp.lp = newLatencyProcessor(rp)
	return rp
}

func (rp *resourceProvisioner) forward(i *model.Invocation) {
	rp.technique.forward(rp, i)
	rp.arrivalCond.Set(true)
}

func (rp *resourceProvisioner) response(i *model.Invocation) {
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
	replica := newReplica(rp, time.Now().String(), rp.appID, rp.funcID, rp.cfg.Idletime)
	godes.AddRunner(replica)
	rp.replicas = append(rp.replicas, replica)
	return replica
}

func (rp *resourceProvisioner) warnReqLatency(i *model.Invocation, tl float64) (bool, float64) {
	return rp.technique.processWarning(i, tl)
}

func (rp *resourceProvisioner) Run() {
	for {
		rp.arrivalCond.Wait(true)
		if rp.arrivalQueue.Len() > 0 {
			i := rp.arrivalQueue.Get().(*model.Invocation)
			r := rp.getAvailableReplica()

			if !rp.cfg.HasOracle {
				newThreshould, err := rp.lp.getCurrTLThreshould(i)
				if err != nil {
					panic(err)
				}
				i.SetTailLatencieThreshold(newThreshould)
			}

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
	rp.arrivalCond.Clear()
	rp.terminatedCond.Clear()
	rp.arrivalQueue.Clear()
	rp.availableReplicas.Clear()
}

func (rp *resourceProvisioner) getOutPut() [][]string {
	res := [][]string{}
	for _, r := range rp.replicas {
		res = append(res, r.getOutPut())
	}
	return res
}

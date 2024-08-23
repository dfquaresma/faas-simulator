package common

import (
	"strconv"
	"time"

	"github.com/agoussia/godes"
	"github.com/dfquaresma/faas-simulator/model"
)

type resourceProvisioner struct {
	*godes.Runner
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
		Runner:   &godes.Runner{},
		appID:    aid,
		funcID:   fid,
		rpID:     aid + "-" + fid,
		replicas: make([]*replica, 0),
		cfg:      cfg,
	}
	rp.technique = newTechnique(rp, cfg.Technique)
	rp.availableReplicas = godes.NewLIFOQueue(rp.rpID)

	p, err := strconv.Atoi(cfg.TailLatencyProb[1:])
	if err != nil {
		panic("Error parsing tailLatencyProb")
	}
	rp.lp = newLatencyProcessor(rp, p)

	return rp
}

func (rp *resourceProvisioner) forward(i *model.Invocation) {
	if !rp.cfg.HasOracle {
		i.SetTailLatencieThreshold(rp.lp.getCurrTLThreshould(i))
	}
	rp.getAvailableReplica().process(i)
	rp.technique.forward(i)
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
	replica := newReplica(rp, time.Now().String(), rp.appID, rp.funcID, rp.cfg)
	godes.AddRunner(replica)
	rp.replicas = append(rp.replicas, replica)
	return replica
}

func (rp *resourceProvisioner) warnReqLatency(i *model.Invocation, tl float64) (bool, float64) {
	return rp.technique.processWarning(i, tl)
}

func (rp *resourceProvisioner) terminate() {
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

package common

import (
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
}

func newResourceProvisioner(aid, fid string, cfg model.Config, mu, sigma float64) *resourceProvisioner {
	rp := &resourceProvisioner{
		Runner:   &godes.Runner{},
		appID:    aid,
		funcID:   fid,
		rpID:     aid + "-" + fid,
		replicas: make([]*replica, 0),
		cfg:      cfg,
	}
	rp.technique = newTechnique(rp, cfg.Technique, mu, sigma)
	rp.availableReplicas = godes.NewLIFOQueue(rp.rpID)

	return rp
}

func (rp *resourceProvisioner) forward(i *model.Invocation) {
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
		if r.terminatedCond.GetState() {
			continue
		}
		if rp.cfg.Idletime < 0 || rp.cfg.Idletime > godes.GetSystemTime()-r.lastWorkTS {
			return r
		}
		r.terminate()
	}
	replica := newReplica(rp, time.Now().String(), rp.appID, rp.funcID, rp.cfg)
	godes.AddRunner(replica)
	rp.replicas = append(rp.replicas, replica)
	return replica
}

func (rp *resourceProvisioner) warnReqLatency(i *model.Invocation) {
	rp.technique.processWarning(i)
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

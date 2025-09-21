package common

import (
	"time"

	"github.com/agoussia/godes"
	"github.com/dfquaresma/faas-simulator/trace_model/model"
)

type provisioner struct {
	*godes.Runner
	availableReplicas *godes.LIFOQueue
	appID             string
	funcID            string
	rpID              string
	cfg               model.Config
	replicas          []*replica
	technique         *technique
	router            *router
}

func newProvisioner(aid, fid string, cfg model.Config, r *router) *provisioner {
	p := &provisioner{
		Runner:   &godes.Runner{},
		appID:    aid,
		funcID:   fid,
		rpID:     aid + "-" + fid,
		replicas: make([]*replica, 0),
		cfg:      cfg,
		router:   r,
	}
	p.technique = newTechnique(p, cfg.Technique, r)
	p.availableReplicas = godes.NewLIFOQueue(p.rpID)

	return p
}

func (p *provisioner) forward(i *model.Invocation) {
	p.getAvailableReplica().process(i)
}

func (p *provisioner) response(i *model.Invocation) {
	p.technique.processResponse(i)
}

func (p *provisioner) setAvailable(r *replica) {
	p.availableReplicas.Place(r)
}

func (p *provisioner) getAvailableReplica() *replica {
	for p.availableReplicas.Len() > 0 {
		r := p.availableReplicas.Get().(*replica)
		if r.terminatedCond.GetState() {
			continue
		}
		if p.cfg.Idletime < 0 || p.cfg.Idletime > godes.GetSystemTime()-r.lastWorkTS {
			return r
		}
		r.terminate()
	}
	replica := newReplica(p, time.Now().String(), p.appID, p.funcID, p.cfg)
	godes.AddRunner(replica)
	p.replicas = append(p.replicas, replica)
	return replica
}

func (p *provisioner) notifyReadyness(funcID string, timestamp float64) {
	p.router.registerReplicaScaling(funcID, 1, timestamp)
}

func (p *provisioner) notifyTermination(funcID string, timestamp float64) {
	p.router.registerReplicaScaling(funcID, -1, timestamp)
}

func (p *provisioner) triggerTechnique(i *model.Invocation) (bool, float64) {
	return p.technique.trigger(i)
}

func (p *provisioner) getTechniqueDelay(i *model.Invocation) float64 {
	return p.technique.getTechniqueDelay(i)
}

func (p *provisioner) terminate() {
	for _, r := range p.replicas {
		r.terminate()
	}
}

func (p *provisioner) getOutPut() [][]string {
	res := [][]string{}
	for _, r := range p.replicas {
		res = append(res, r.getOutPut())
	}
	return res
}

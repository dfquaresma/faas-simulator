package common

import (
	"fmt"

	"github.com/agoussia/godes"
)

type iResourceProvisioner interface {
	forward(i *invocation)
	setAvailable(r *replica)
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
	replicas          []*replica
}

func newResourceProvisioner(aid, fid string) *resourceProvisioner {
	return &resourceProvisioner{
		Runner:            &godes.Runner{},
		arrivalCond:       godes.NewBooleanControl(),
		availableCond:     godes.NewBooleanControl(),
		arrivalQueue:      godes.NewFIFOQueue("arrival"),
		availableReplicas: godes.NewLIFOQueue("available"),
		appID:             aid,
		funcID:            fid,
		frpID:             aid + "-" + fid,
		replicas:          make([]*replica, 0),
	}
}

func (frp *resourceProvisioner) forward(i *invocation) {
	frp.arrivalQueue.Place(i)
	frp.arrivalCond.Set(true)
}

func (frp *resourceProvisioner) setAvailable(r *replica) {
	frp.availableReplicas.Place(r)
}

func (frp *resourceProvisioner) getAvailableReplica() *replica {
	if frp.availableReplicas.Len() > 0 {
		return frp.availableReplicas.Get().(*replica)
	}
	rid := fmt.Sprintf("%d", len(frp.replicas))
	replica := newReplica(frp, rid, frp.appID, frp.funcID)
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

func (frp *resourceProvisioner) Run() {
	for {
		frp.arrivalCond.Wait(true)
		if frp.arrivalQueue.Len() > 0 {
			i := frp.arrivalQueue.Get().(*invocation)
			frp.getAvailableReplica().process(i)
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

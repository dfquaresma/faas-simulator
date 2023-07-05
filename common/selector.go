package common

import (
	"github.com/agoussia/godes"
)

type iSelector interface {
	forward(i *invocation)
	terminate()
	getOutPut() [][]string
}

type selector struct {
	*godes.Runner
	arrivalCond  *godes.BooleanControl
	arrivalQueue *godes.FIFOQueue
	provisioners map[string]*resourceProvisioner
	terminated   bool
}

func NewSelector() *selector {
	return &selector{
		Runner:       &godes.Runner{},
		arrivalCond:  godes.NewBooleanControl(),
		arrivalQueue: godes.NewFIFOQueue("arrival"),
		provisioners: make(map[string]*resourceProvisioner),
	}
}

func (fs *selector) getProvisioner(aid, fid string) (*resourceProvisioner) {
	frp := fs.provisioners[aid + fid]
	if frp == nil {
		frp = fs.newProvisioner(aid, fid)
	}
	return frp
}

func (fs *selector) newProvisioner(aid, fid string) *resourceProvisioner {
	frp := newResourceProvisioner(aid, fid)
	fs.provisioners[aid + fid] = frp
	godes.AddRunner(frp)
	return frp
}

func (fs *selector) forward(i *invocation) {
	fs.arrivalQueue.Place(i)
	fs.arrivalCond.Set(true)
}

func (fs *selector) Run() {
	for {
		fs.arrivalCond.Wait(true)
		if fs.arrivalQueue.Len() > 0 {
			i := fs.arrivalQueue.Get().(*invocation)
			frp := fs.getProvisioner(i.getAppID(), i.getFuncID())
			frp.forward(i)
			continue
		}
		fs.arrivalCond.Set(false)
		if fs.terminated {
			break
		}
	}
}

func (fs *selector) terminate() {
	fs.terminated = true
	fs.arrivalCond.Set(true)
	for _, frp := range fs.provisioners {
		frp.terminate()
	}
}

func (fs *selector) GetOutPut() [][]string {
	res := [][]string{}
	header := []string{"replicaID", "frpID", "appID", "funcID", "busyTime", "upTime", "reqsProcessed"}
	res = append(res, header)
	for _, frp := range fs.provisioners {
		for _, o := range frp.getOutPut() {
			res = append(res, o)
		} 
	}
	return res
}

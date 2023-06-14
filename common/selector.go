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
	provisioners map[string]*resourceProvisioner
}

func NewSelector() *selector {
	return &selector{
		provisioners: make(map[string]*resourceProvisioner),
	}
}

func (fs *selector) getProvisioner(fid string) (*resourceProvisioner, bool) {
	frp := fs.provisioners[fid]
	bo := frp != nil
	return frp, bo
}

func (fs *selector) newProvisioner(aid, fid string) *resourceProvisioner {
	frp := newResourceProvisioner(aid, fid)
	fs.provisioners[fid] = frp
	godes.AddRunner(frp)
	return frp
}

func (fs *selector) forward(i *invocation) {
	frp, exist := fs.getProvisioner(i.getFuncID())
	if !exist {
		frp = fs.newProvisioner(i.getAppID(), i.getFuncID())
	}
	frp.forward(i)
}

func (fs *selector) terminate() {
	for _, frp := range fs.provisioners {
		frp.terminate()
	}
}

func (fs *selector) GetOutPut() [][]string {
	res := [][]string{}
	for _, frp := range fs.provisioners {
		res = append(res, frp.getOutPut())
	}
	return res
}

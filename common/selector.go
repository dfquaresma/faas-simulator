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
	provisioners	map[string]iResourceProvisioner
}

func newSelector() *selector {
	return &selector{
		provisioners:	make(map[string]iResourceProvisioner),
	}
}

func (fs *selector) getProvisioner(fid string) (*resourceProvisioner, bool) {
	frp := provisioners[fid]
	bo := frp != nil
	return frp, bo
}

func (fs *selector) newProvisioner(aid, fid string) *resourceProvisioner {
	frp := newResourceProvisioner(aid, fid)
	provisioners[fid] = frp
	godes.AddRunner(frp)
	return frp
}

func (fs *selector) forward(i *invocation) {
	frp, exist := getProvisioner(i.getFuncID())
	if !exist {
		frp = newProvisioner(i.getAppID(), i.getFuncID())
	}
	frp.forward(i)
}

func (fs *selector) terminate(i *invocation) {
	for _, frp := range fr.provisioners {
		frp.terminate()
	}
}

func (fs *selector) getOutPut() [][]string {
	res := [][]string{} 
	for _, frp := range fs.provisioners {
		res = append(res, frp.getOutPut())
	}
	return res
}

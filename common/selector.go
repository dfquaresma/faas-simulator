package common

type iSelector interface {
	forward(inv iInvocation)
}

type selector struct {
	provisioners	map[string]IResourceProvisioner
}

func newSelector() *selector {
	return &selector{
		provisioners:	make(map[string]IResourceProvisioner),
	}
}

func (fs *selector) forward(inv iInvocation) *ResourceProvisioner {
	frp, exist := getProvisioner(inv.getFuncID())
	if !exist {
		frp = newProvisioner(inv.getAppID(), inv.getFuncID())
	}
	frp.forward(inv)
}

func (fs *selector) getProvisioner(funcID string) *ResourceProvisioner, bool {
	frp := provisioners[funcID]
	bo := frp != nil
	return frp, bo
}

func (fs *selector) newProvisioner(appId, funcID string) *ResourceProvisioner {
	frp := newResourceProvisioner(appId, funcID)
	provisioners[funcID] = frp
	return frp
}

package common

type iSelector interface {
	forward(i *invocation)
	terminate()
	getOutPut() [][]string
}

type selector struct {
	*godes.Runner
	provisioners	map[string]IResourceProvisioner
}

func newSelector() *selector {
	return &selector{
		provisioners:	make(map[string]IResourceProvisioner),
	}
}

func (fs *selector) getProvisioner(fid string) (*ResourceProvisioner, bool) {
	frp := provisioners[fid]
	bo := frp != nil
	return frp, bo
}

func (fs *selector) newProvisioner(aid, fid string) *ResourceProvisioner {
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

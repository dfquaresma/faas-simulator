package common

import (
	"github.com/dfquaresma/faas-simulator/model"
)

type selector struct {
	provisioners map[string]*resourceProvisioner
	cfg          model.Config
}

func NewSelector(cfg model.Config) *selector {
	return &selector{
		provisioners: make(map[string]*resourceProvisioner),
		cfg:          cfg,
	}
}

func (s *selector) getProvisioner(i *model.Invocation) *resourceProvisioner {
	rp := s.provisioners[i.GetAppID()+i.GetFuncID()]
	if rp == nil {
		rp = s.newProvisioner(i.GetAppID(), i.GetFuncID(), i.GetMU(), i.GetSigma())
	}
	return rp
}

func (s *selector) newProvisioner(aid, fid string, mu, sigma float64) *resourceProvisioner {
	rp := newResourceProvisioner(aid, fid, s.cfg, mu, sigma)
	s.provisioners[aid+fid] = rp
	return rp
}

func (s *selector) forward(i *model.Invocation) {
	s.getProvisioner(i).forward(i)
}

func (s *selector) terminate() {
	for _, rp := range s.provisioners {
		rp.terminate()
	}
}

func (s *selector) GetOutPut() [][]string {
	res := [][]string{}
	header := []string{"replicaID", "rpID", "appID", "funcID", "busyTime", "upTime", "reqsProcessed", "lastWorkTS", "shutdownTS"}
	res = append(res, header)
	for _, rp := range s.provisioners {
		for _, o := range rp.getOutPut() {
			res = append(res, o)
		}
	}
	return res
}

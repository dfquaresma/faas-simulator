package common

import (
	"strconv"

	"github.com/dfquaresma/faas-simulator/model"
)

type selector struct {
	provisioners map[string]*resourceProvisioner
	dataset      *model.Dataset
	cfg          model.Config
	register     [][]string
	replicas     int64
}

func NewSelector(dataset *model.Dataset, cfg model.Config) *selector {
	return &selector{
		provisioners: make(map[string]*resourceProvisioner),
		dataset:      dataset,
		cfg:          cfg,
		register:     [][]string{},
	}
}

func (s *selector) getDataSet() *model.Dataset {
	return s.dataset
}

func (s *selector) getProvisioner(i *model.Invocation) *resourceProvisioner {
	rp := s.provisioners[i.GetAppID()+i.GetFuncID()]
	if rp == nil {
		rp = s.newProvisioner(i.GetAppID(), i.GetFuncID())
	}
	return rp
}

func (s *selector) newProvisioner(aid, fid string) *resourceProvisioner {
	rp := newResourceProvisioner(aid, fid, s.cfg, s)
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

func (s *selector) registerReplicaScaling(amount int64, timestamp float64) {
	s.replicas += amount
	replicasStr := strconv.FormatInt(s.replicas, 10)
	timestampStr := strconv.FormatFloat(timestamp, 'f', -1, 64)
	s.register = append(s.register, []string{replicasStr, timestampStr})
}

func (s *selector) GetOutPut() ([][]string, [][]string) {
	rp_res := [][]string{}
	header := []string{"replicaID", "rpID", "appID", "funcID", "busyTime", "upTime", "reqsProcessed", "lastWorkTS", "startTS", "shutdownTS"}
	rp_res = append(rp_res, header)
	for _, rp := range s.provisioners {
		rp_res = append(rp_res, rp.getOutPut()...)
	}

	sel_res := [][]string{}
	header = []string{"replica_amount", "timestamp"}
	sel_res = append(sel_res, header)
	sel_res = append(sel_res, s.register...)

	return rp_res, sel_res
}

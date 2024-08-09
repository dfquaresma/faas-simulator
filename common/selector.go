package common

import (
	"github.com/agoussia/godes"
)

type Config struct {
	ForwardLatency  float64
	Idletime        float64
	TailLatency     float64
	TailLatencyProb string
	Technique       string
	HasOracle       bool
}

type selector struct {
	*godes.Runner
	arrivalCond    *godes.BooleanControl
	terminatedCond *godes.BooleanControl
	arrivalQueue   *godes.FIFOQueue
	provisioners   map[string]*resourceProvisioner
	cfg            Config
}

func NewSelector(cfg Config) *selector {
	return &selector{
		Runner:         &godes.Runner{},
		arrivalCond:    godes.NewBooleanControl(),
		terminatedCond: godes.NewBooleanControl(),
		arrivalQueue:   godes.NewFIFOQueue("arrival"),
		provisioners:   make(map[string]*resourceProvisioner),
		cfg:            cfg,
	}
}

func (s *selector) getProvisioner(aid, fid string) *resourceProvisioner {
	rp := s.provisioners[aid+fid]
	if rp == nil {
		rp = s.newProvisioner(aid, fid)
	}
	return rp
}

func (s *selector) newProvisioner(aid, fid string) *resourceProvisioner {
	rp := newResourceProvisioner(aid, fid, s.cfg)
	s.provisioners[aid+fid] = rp
	godes.AddRunner(rp)
	return rp
}

func (s *selector) forward(i *invocation) {
	s.arrivalQueue.Place(i)
	s.arrivalCond.Set(true)
}

func (s *selector) Run() {
	for {
		s.arrivalCond.Wait(true)
		if s.arrivalQueue.Len() > 0 {
			i := s.arrivalQueue.Get().(*invocation)
			rp := s.getProvisioner(i.getAppID(), i.getFuncID())
			rp.forward(i)
			continue
		}
		if s.arrivalQueue.Len() == 0 {
			s.arrivalCond.Set(false)
			if s.terminatedCond.GetState() {
				break
			}
		}
	}
}

func (s *selector) terminate() {
	s.terminatedCond.Set(true)
	s.arrivalCond.Set(true)
	s.arrivalCond.Wait(false)
	for _, rp := range s.provisioners {
		rp.terminate()
	}
	s.arrivalCond.Clear()
	s.terminatedCond.Clear()
	s.arrivalQueue.Clear()
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

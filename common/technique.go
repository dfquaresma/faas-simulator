package common

import (
	"time"

	"github.com/agoussia/godes"
	"github.com/dfquaresma/faas-simulator/model"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

type technique struct {
	*godes.Runner
	rp      *resourceProvisioner
	s       *selector
	config  string
	uniform distuv.Uniform
}

func newTechnique(rp *resourceProvisioner, t string, s *selector) *technique {
	return &technique{
		Runner: &godes.Runner{},
		rp:     rp,
		s:      s,
		config: t,
		uniform: distuv.Uniform{
			Min: 0,
			Max: 1,
			Src: rand.NewSource(uint64(time.Now().Nanosecond())),
		},
	}
}

func (t *technique) newLatency(id string) float64 {
	latencies := t.s.getDataSet().GetLatenciesOf(id)
	return latencies[int(t.uniform.Rand()*float64(len(latencies)))]
}

func (t *technique) forward(i *model.Invocation) {
	t.rp.getAvailableReplica().process(i)
	if t.config == "RequestHedgingDefault" {
		iCopy := model.CopyInvocation(i)
		iCopy.SetDuration(t.newLatency(i.GetAppID() + i.GetFuncID()))
		i.SetForwardedTs(godes.GetSystemTime())
		t.rp.getAvailableReplica().process(iCopy)
	}
}

func (t *technique) processWarning(i *model.Invocation) {
	switch t.config {
	case "GCI":
		if !i.IsColdStart() {
			i.SetDuration(t.newLatency(i.GetAppID() + i.GetFuncID()))
			i.SetForwardedTs(godes.GetSystemTime())
			t.rp.getAvailableReplica().process(i)
		}

	case "RequestHedgingOpt":
		if !i.IsCopy() {
			iCopy := model.CopyInvocation(i)
			iCopy.SetDuration(t.newLatency(i.GetAppID() + i.GetFuncID()))
			iCopy.SetForwardedTs(godes.GetSystemTime())
			t.rp.getAvailableReplica().process(iCopy)
		}
	}
}

func (t *technique) processResponse(i *model.Invocation) {
	if i.IsCopy() {
		iRef := i.GetSrcInvoc()
		iRef.UpdateTechniqueResponseTime(i.GetResponseTime())
	}
}

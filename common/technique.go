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
	provisioner *provisioner
	router      *router
	config      string
	uniform     distuv.Uniform
}

func newTechnique(p *provisioner, t string, r *router) *technique {
	return &technique{
		Runner:      &godes.Runner{},
		provisioner: p,
		router:      r,
		config:      t,
		uniform: distuv.Uniform{
			Min: 0,
			Max: 1,
			Src: rand.NewSource(uint64(time.Now().Nanosecond())),
		},
	}
}

func (t *technique) newLatency(id string) float64 {
	latencies := t.router.getDataSet().GetLatenciesOf(id)
	return latencies[int(t.uniform.Rand()*float64(len(latencies)))]
}

func (t *technique) trigger(i *model.Invocation) (bool, float64) {
	switch t.config {
	case "simplehr", "delayedhr":
		if !i.IsCopy() {
			iCopy := model.CopyInvocation(i)
			iCopy.SetForwardedTs(godes.GetSystemTime())
			iCopy.SetDuration(t.newLatency(i.GetAppID() + i.GetFuncID()))
			t.provisioner.getAvailableReplica().process(iCopy)
		}

	case "gci":
		if !i.IsColdStart() && i.IsTailLatency() {
			iCopy := model.CopyInvocation(i)
			iCopy.SetForwardedTs(godes.GetSystemTime())
			iCopy.SetDuration(t.newLatency(i.GetAppID() + i.GetFuncID()))
			if i.IsCopy() {
				// we may shedding multiple times, thus fix source invocation ref
				iCopy.UpdateSource(i)
			}
			t.provisioner.getAvailableReplica().process(iCopy)
			return true, i.GetTailLatencyThreshold()
		}
	}
	return false, 0
}

func (t *technique) getDelay(i *model.Invocation) float64 {
	switch t.config {
	case "DelayedHR":
		return i.GetTailLatencyThreshold()

	default:
		return 0
	}
}

func (t *technique) processResponse(i *model.Invocation) {
	if i.IsCopy() {
		srcInvoc := i.GetSrcInvoc()
		srcInvoc.UpdateTechniqueResponseTime(i)
	}
}

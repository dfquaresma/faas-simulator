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
	iCopy := model.CopyInvocation(i)
	copyReplicaIsWarm := false
	switch t.config {
	case "hedge_nodelay_nocancel", "delayed_hedge_nocancel", "hedged_request":
		if !i.IsCopy() {
			iCopy.SetForwardedTs(godes.GetSystemTime())
			iCopy.SetDuration(t.newLatency(i.GetAppID() + i.GetFuncID()))
			replica := t.provisioner.getAvailableReplica()
			copyReplicaIsWarm = replica.getRequestCount() != 0
			replica.process(iCopy)
		}

	case "gci":
		if !i.IsColdStart() && i.IsTailLatency() {
			iCopy.SetForwardedTs(godes.GetSystemTime())
			iCopy.SetDuration(t.newLatency(i.GetAppID() + i.GetFuncID()))
			if i.IsCopy() {
				// we may shed multiple times, thus fix the source invocation ref to the latest shed
				iCopy.UpdateSource(i)
			}
			i.UpdateShedTimes()
			t.provisioner.getAvailableReplica().process(iCopy)
		}
	}

	switch t.config {
	case "hedged_request":
		switch i.IsCopy() {
		// if it is the triggered copy, cancel it if it takes longer than the original
		case true:
			if i.GetDuration() > i.GetSrcInvoc().GetDuration() {
				return true, i.GetSrcInvoc().GetDuration()
			}
		// if it is the original, cancel it if it takes longer than the copy
		// but only if copy was sent to a warm replica and has no coldstart
		case false:
			if i.GetDuration() > iCopy.GetDuration() && copyReplicaIsWarm {
				return true, iCopy.GetDuration()
			}
		}

	case "gci":
		return true, i.GetDuration() - i.GetTailLatencyThreshold()
	}

	return false, 0
}

func (t *technique) getTechniqueDelay(i *model.Invocation) float64 {
	switch t.config {
	case "delayed_hedge_nocancel", "hedged_request":
		if !i.IsCopy() {
			return i.GetTailLatencyThreshold()
		}
	}
	return 0
}

func (t *technique) processResponse(i *model.Invocation) {
	if i.IsCopy() {
		srcInvoc := i.GetSrcInvoc()
		srcInvoc.UpdateTechniqueResponseTime(i)
	}
}

package common

import (
	"math/rand"
	"time"

	"github.com/agoussia/godes"
	"github.com/dfquaresma/faas-simulator/model"
)

type technique struct {
	*godes.Runner
	rp        *resourceProvisioner
	iIdAidFid map[string]*model.Invocation
	config    string
}

func newTechnique(rp *resourceProvisioner, t string) *technique {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return &technique{
		Runner:    &godes.Runner{},
		rp:        rp,
		iIdAidFid: make(map[string]*model.Invocation),
		config:    t,
	}
}

func (t *technique) forward(i *model.Invocation) {
	t.iIdAidFid[i.GetID()] = i
	if t.config == "RequestHedgingDefault" {
		iCopy := model.CopyInvocation(i)
		iCopy.SetDuration(t.newLatency(i.GetP100()))
		t.rp.getAvailableReplica().process(iCopy)
	}
}

func (t *technique) newLatency(p99 float64) float64 {
	return rand.Float64() * p99
}

func (t *technique) processWarning(i *model.Invocation, tl float64) (bool, float64) {
	switch t.config {
	case "GCI":
		// shed only requests that are not coldstart
		// don't shed the same invocation more than twice
		if !i.IsColdStart() && i.GetShedTimes() < 2 {
			i.IncrementShedTimes()
			i.SetDuration(t.newLatency(i.GetP100()))
			t.rp.getAvailableReplica().process(i)
			return true, tl
		}
		return false, 0

	case "RequestHedgingOpt":
		if !i.IsCopy() {
			iCopy := model.CopyInvocation(i)
			iCopy.SetDuration(t.newLatency(i.GetP100()))
			iCopy.ResetResponseTime()
			godes.Advance(i.GetTailLatencyThreshold())
			t.rp.getAvailableReplica().process(iCopy)
			return true, tl
		}
		return false, 0

	default:
		return false, 0
	}
}

func (t *technique) processResponse(i *model.Invocation) {
	if i.IsCopy() {
		iRef := t.iIdAidFid[i.GetID()]
		iRef.UpdateRhInvocationMetadata(
			i.GetForwardedTs(),
			i.GetProcessedTs(),
			i.GetResponseTime(),
			i.GetHops(),
			i.GetHopResponses(),
		)
	}
}

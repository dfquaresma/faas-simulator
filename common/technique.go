package common

import (
	"math/rand"
	"time"

	"github.com/dfquaresma/faas-simulator/model"
)

type technique struct {
	rp        *resourceProvisioner
	iIdAidFid map[string]*model.Invocation
	config    string
}

func newTechnique(rp *resourceProvisioner, t string) *technique {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return &technique{
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

func (t *technique) newLatency(p100 float64) float64 {
	return rand.Float64() * p100
}

func (t *technique) processWarning(i *model.Invocation, tl float64) (bool, float64) {
	switch t.config {
	case "GCI":
		// always shed requests since processWarning is always called from a tail latency request
		i.IncrementShedTimes()
		i.SetDuration(t.newLatency(i.GetP100()))
		t.rp.getAvailableReplica().process(i)
		return true, tl

	case "RequestHedgingOpt":
		if !i.IsCopy() {
			iCopy := model.CopyInvocation(i)
			iCopy.SetDuration(i.GetTailLatencyThreshold() + t.newLatency(i.GetP100()))
			iCopy.ResetResponseTime()
			t.rp.getAvailableReplica().process(iCopy)
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

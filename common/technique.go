package common

import (
	"math/rand"
	"time"

	"github.com/dfquaresma/faas-simulator/model"
)

type technique struct {
	rp        *resourceProvisioner
	iIdAidFid map[string]*model.Invocation
	processed map[string]bool
	config    string
}

func newTechnique(rp *resourceProvisioner, t string) *technique {
	return &technique{
		rp:        rp,
		iIdAidFid: make(map[string]*model.Invocation),
		processed: make(map[string]bool),
		config:    t,
	}
}

func (t *technique) forward(rp *resourceProvisioner, i *model.Invocation) {
	t.iIdAidFid[i.GetID()] = i
	rp.arrivalQueue.Place(i)
	if t.config == "RequestHedgingDefault" {
		iCopy := model.CopyInvocation(i)
		iCopy.SetDuration(t.newLatency(i.GetP100()))
		rp.arrivalQueue.Place(iCopy)
	}
}

func (t *technique) newLatency(p100 float64) float64 {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand.Float64() * p100
}

func (t *technique) processWarning(i *model.Invocation, tl float64) (bool, float64) {
	switch t.config {
	case "GCI":
		// always shed requests since processWarning is always called from a tail latency request
		i.IncrementShedTimes()
		i.SetDuration(t.newLatency(i.GetP100()))
		t.rp.arrivalQueue.Place(i)
		t.rp.arrivalCond.Set(true)
		return true, tl

	case "RequestHedgingOpt":
		iId := i.GetID()
		shouldHedge := !t.processed[iId]
		if shouldHedge {
			t.processed[iId] = true
			iCopy := model.CopyInvocation(i)
			iCopy.SetDuration(i.GetTailLatencyThreshold() + t.newLatency(i.GetP100()))
			iCopy.ResetResponseTime()
			t.rp.arrivalQueue.Place(iCopy)
			t.rp.arrivalCond.Set(true)
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

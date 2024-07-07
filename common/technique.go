package common

import (
	"math/rand"
	"time"
)

type technique struct {
	rp        *resourceProvisioner
	iIdAidFid map[string]*invocation
	processed map[string]bool
	config    string
}

func newTechnique(rp *resourceProvisioner, t string) *technique {
	return &technique{
		rp:        rp,
		iIdAidFid: make(map[string]*invocation),
		processed: make(map[string]bool),
		config:    t,
	}
}

func (t *technique) forward(rp *resourceProvisioner, i *invocation) {
	t.iIdAidFid[i.getID()+i.getAppID()+i.getFuncID()] = i
	rp.arrivalQueue.Place(i)
	if t.config == "RequestHedgingDefault" {
		iCopy := copyInvocation(i)
		iCopy.setDuration(t.getNewLatency(i.getP100()))
		rp.arrivalQueue.Place(iCopy)
	}
	rp.arrivalCond.Set(true)
}

func (t *technique) getNewLatency(p100 float64) float64 {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand.Float64() * p100
}

func (t *technique) processWarning(i *invocation, tl float64) (bool, float64) {
	switch t.config {
	case "GCI":
		// always shed requests since processWarning is always called from a tail latency request
		i.im.shed_times = i.im.shed_times + 1
		i.setDuration(t.getNewLatency(i.getP100()))
		t.rp.arrivalQueue.Place(i)
		t.rp.arrivalCond.Set(true)
		return true, tl

	case "RequestHedging95":
		iId := i.getID() + i.getAppID() + i.getFuncID()
		shouldHedge := !t.processed[iId]
		if shouldHedge {
			t.processed[iId] = true
			iCopy := copyInvocation(i)
			iCopy.setDuration(i.getTailLatencyThreshold() + t.getNewLatency(i.getP100()))
			iCopy.resetResponseTime()
			t.rp.arrivalQueue.Place(iCopy)
			t.rp.arrivalCond.Set(true)
		}
		return false, 0

	default:
		return false, 0
	}
}

func (t *technique) processResponse(i *invocation) {
	if i.im.is_copy {
		iRef := t.iIdAidFid[i.getID()+i.getAppID()+i.getFuncID()]
		iRef.updateRhInvocationMetadata(
			i.im.forwardedTs,
			i.im.processedTs,
			i.im.responseTime,
			i.im.hops,
			i.im.hopResponses,
		)
	}
}

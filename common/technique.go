package common

import (
	"reflect"
)

type iTechnique interface {
	processWarning(i *invocation, tl float64) (bool, float64)
	forward(frp *resourceProvisioner, i *invocation)
	processResponse(i *invocation, d float64)
}

type technique struct {
	frp       *resourceProvisioner
	iIdAidFid map[string]*invocation
	iAidFid   map[string][]*invocation
	processed map[string]bool
	config    string
}

func newTechnique(frp *resourceProvisioner, t string) *technique {
	return &technique{
		frp:       frp,
		iIdAidFid: make(map[string]*invocation),
		iAidFid:   make(map[string][]*invocation),
		processed: make(map[string]bool),
		config:    t,
	}
}

func (t *technique) forward(frp *resourceProvisioner, i *invocation) {
	t.iIdAidFid[i.getID()+i.getAppID()+i.getFuncID()] = i
	iAidFidList := t.iAidFid[i.getAppID()+i.getFuncID()]
	iAidFidList = append(iAidFidList, i)
	t.iAidFid[i.getAppID()+i.getFuncID()] = iAidFidList
	frp.arrivalQueue.Place(i)
	if t.config == "RequestHedgingDefault" {
		iCopy := copyInvocation(i)
		frp.arrivalQueue.Place(iCopy)
	}
	frp.arrivalCond.Set(true)
}

func (t *technique) processWarning(i *invocation, tl float64) (bool, float64) {
	switch t.config {
	case "GCI":
		shouldRedirectReq := tl > 0
		if shouldRedirectReq {
			r := t.frp.getAvailableReplica()
			r.process(i)
		}
		return shouldRedirectReq, tl

	case "RequestHedging95":
		p95 := i.getP95()
		iId := i.getID() + i.getAppID() + i.getFuncID()
		shouldHedge := tl+i.getDuration() > p95 && !t.processed[iId]
		if shouldHedge {
			t.processed[iId] = true
			iCopy := copyInvocation(i)
			iCopy.setDuration(p95 + i.getDuration())
			iCopy.resetResponseTime()
			rep := t.frp.getAvailableReplica()
			rep.process(iCopy)
		}
		return false, 0

	default:
		return false, 0
	}
}

func (t *technique) processResponse(i *invocation) {
	if t.config == "RequestHedgingDefault" || t.config == "RequestHedging95" {
		iRef := t.iIdAidFid[i.getID()+i.getAppID()+i.getFuncID()]
		if !reflect.DeepEqual(i, iRef) {
			iRef.updateRhInvocationMetadata(
				i.im.forwardedTs,
				i.im.processedTs,
				i.im.responseTime,
				i.im.hops,
				i.im.hopResponses,
			)
		}
	}
}

package common

import (
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/agoussia/godes"
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
	config    string
}

func newTechnique(frp *resourceProvisioner, t string) *technique {
	return &technique{
		frp:       frp,
		iIdAidFid: make(map[string]*invocation),
		iAidFid:   make(map[string][]*invocation),
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
			godes.Advance(t.frp.cfg.ForwardLatency)
			i.updateHopResponse(t.frp.cfg.ForwardLatency)
			r.process(i)
		}
		return shouldRedirectReq, tl

	case "RequestHedging95":
		iCopy := copyInvocation(i)
		iCopyDur := iCopy.getDuration()
		p95 := i.getP95()
		if iCopyDur > p95 {
			fInv := t.iAidFid[iCopy.getAppID()+iCopy.getFuncID()]
			r := rand.New(rand.NewSource(time.Now().Unix()))
			dur := fInv[r.Intn(len(fInv))].getDuration()
			iCopy.setDuration(p95 + dur)
			t.frp.arrivalQueue.Place(iCopy)
			t.frp.arrivalCond.Set(true)
		}
		return false, 0

	default:
		return false, 0
	}
}

func (t *technique) processResponse(inv *invocation, dur float64) {
	inv.updateHopResponse(dur)
	if t.config == "RequestHedgingDefault" || t.config == "RequestHedging95" {
		i := t.iIdAidFid[inv.getID()+inv.getAppID()+inv.getFuncID()]
		iHops := i.im.hops
		invCopyHops := inv.im.hops
		if len(iHops) == len(invCopyHops) && !reflect.DeepEqual(iHops, invCopyHops) {
			iLastHop, _ := strconv.ParseFloat(iHops[:len(iHops)-1][0], 64)
			i.updateHopResponse(math.Min(iLastHop, dur))
		}
	}
}

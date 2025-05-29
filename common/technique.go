package common

import (
	"time"

	"golang.org/x/exp/rand"

	"github.com/agoussia/godes"
	"github.com/dfquaresma/faas-simulator/model"
	"gonum.org/v1/gonum/stat/distuv"
)

type technique struct {
	*godes.Runner
	rp        *resourceProvisioner
	iIdAidFid map[string]*model.Invocation
	config    string
	seed      rand.Source
}

func newTechnique(rp *resourceProvisioner, t string) *technique {
	return &technique{
		Runner:    &godes.Runner{},
		rp:        rp,
		iIdAidFid: make(map[string]*model.Invocation),
		config:    t,
		seed:      rand.NewSource(uint64(time.Now().Nanosecond())),
	}
}

func (t *technique) forward(i *model.Invocation) {
	t.iIdAidFid[i.GetID()] = i
	if t.config == "RequestHedgingDefault" {
		iCopy := model.CopyInvocation(i)
		iCopy.SetDuration(t.newLatency(i.GetMU(), i.GetSigma()))
		t.rp.getAvailableReplica().process(iCopy)
	}
}

func (t *technique) newLatency(mu, sigma float64) float64 {
	ln := distuv.LogNormal{
		Mu:    mu,
		Sigma: sigma,
		Src:   t.seed,
	}
	return ln.Rand()
}

func (t *technique) processWarning(i *model.Invocation, tl float64) (bool, float64) {
	switch t.config {
	case "GCI":
		// shed only requests that are not coldstart
		// don't shed the same invocation more than twice
		if !i.IsColdStart() {
			i.IncrementShedTimes()
			i.SetDuration(t.newLatency(i.GetMU(), i.GetSigma()))
			t.rp.getAvailableReplica().process(i)
			return true, tl
		}
		return false, 0

	case "RequestHedgingOpt":
		if !i.IsCopy() {
			iCopy := model.CopyInvocation(i)
			iCopy.SetDuration(t.newLatency(i.GetMU(), i.GetSigma()))
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

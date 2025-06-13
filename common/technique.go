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
	ln        distuv.LogNormal
}

func newTechnique(rp *resourceProvisioner, t string, mu, sigma float64) *technique {
	return &technique{
		Runner:    &godes.Runner{},
		rp:        rp,
		iIdAidFid: make(map[string]*model.Invocation),
		config:    t,
		ln: distuv.LogNormal{
			Mu:    mu,
			Sigma: sigma,
			Src:   rand.NewSource(uint64(time.Now().Nanosecond())),
		},
	}
}

func (t *technique) newLatency(mu, sigma float64) float64 {
	return t.ln.Rand()
}

func (t *technique) forward(i *model.Invocation) {
	t.iIdAidFid[i.GetID()] = i
	if t.config == "RequestHedgingDefault" {
		iCopy := model.CopyInvocation(i)
		iCopy.SetDuration(t.newLatency(i.GetMU(), i.GetSigma()))
		t.rp.getAvailableReplica().process(iCopy)
	}
}

func (t *technique) processWarning(i *model.Invocation) {
	switch t.config {
	case "GCI":
		if !i.IsColdStart() {
			i.IncrementShedTimes()
			i.SetDuration(t.newLatency(i.GetMU(), i.GetSigma()))
			t.rp.getAvailableReplica().process(i)
		}

	case "RequestHedgingOpt":
		if !i.IsCopy() {
			iCopy := model.CopyInvocation(i)
			iCopy.SetDuration(t.newLatency(i.GetMU(), i.GetSigma()))
			t.rp.getAvailableReplica().process(iCopy)
		}
	}
}

func (t *technique) processResponse(i *model.Invocation) {
	if i.IsCopy() {
		iRef := t.iIdAidFid[i.GetID()]
		if i.GetResponseTime() < iRef.GetResponseTime() {
			iRef.UpdateRhInvocationMetadata(
				i.GetForwardedTs(),
				i.GetResponseTime(),
				i.GetProcessedTs(),
				i.GetHopResponses(),
				i.GetHops(),
			)
		}
	}
}

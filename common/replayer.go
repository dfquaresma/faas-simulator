package common

import (
	"github.com/agoussia/godes"
)

type replayer struct {
	*godes.Runner
	invocations *invocations
	selector    *selector
}

func NewReplayer(invocations *invocations, selector *selector) *replayer {
	return &replayer{
		Runner:      &godes.Runner{},
		invocations: invocations,
		selector:    selector,
	}
}

func (tr *replayer) Run() {
	godes.AddRunner(tr.selector)
	godes.Run()
	for i := tr.invocations.next(); i != nil; i = tr.invocations.next() {
		godes.Advance(i.getStartTS())
		i.setForwardedTs(godes.GetSystemTime())
		tr.selector.forward(i)
	}
	tr.selector.terminate()
	godes.WaitUntilDone()
}

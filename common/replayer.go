package common

import (
	"github.com/agoussia/godes"
)

type replayer struct {
	invocations     iInvocations
	selector 		iSelector
}

func newReplayer(invocations *iInvocations, selector iSelector) *replayer {
	return &replayer{
		Runner:  			&godes.Runner{},
		invocations: 		invocations,
		selector: 			selector,
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

package common

type replayer struct {
	*godes.Runner
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
	for i := tr.invocations.next(); i != nil; i = tr.invocations.next() {
		godes.Advance(i.getStartTS())
		i.setForwardedTs(godes.GetSystemTime())
		tr.selector.forward(i)
    }
	tr.selector.terminate()
}

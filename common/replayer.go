package common

type iReplayer interface {
	play()
}

type replayer struct {
	*godes.Runner
	invocations      iInvocations
	functionSelector iFunctionSelector
}

func newReplayer(invocations iInvocations, functionSelector iFunctionSelector) *replayer {
	return &replayer{
		Runner:  			&godes.Runner{},
		invocations: 		invocations,
		functionSelector: 	functionSelector,
	}
}

func (tr *replayer) play() {
	for invoc := tr.invocations.next(); invoc != nil; invoc = tr.invocations.next() {
		godes.Advance(invoc.getStartTS())
		invoc.setForwardedTs(godes.GetSystemTime())
		tr.functionSelector.forward(invoc)
    }
	tr.functionSelector.terminate()
}

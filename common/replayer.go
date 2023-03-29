package sim

type iReplayer interface {
	play() error
}

type replayer struct {
	*godes.Runner
	invocations      []iInvocation
	functionSelector iFunctionSelector
}

func newReplayer(invocations []iInvocation, functionSelector iFunctionSelector) *replayer {
	return &replayer{
		Runner:  			&godes.Runner{},
		invocations: 		invocations,
		functionSelector: 	functionSelector,
	}
}

func (tr *replayer) play() error {}

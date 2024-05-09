package common

import (
	"github.com/agoussia/godes"
	"github.com/schollz/progressbar/v3"
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

func (re *replayer) Run() {
	godes.AddRunner(re.selector)
	godes.Run()
	previousTs := 0.0
	bar := progressbar.Default(re.invocations.GetSize())
	for i := re.invocations.next(); i != nil; i = re.invocations.next() {
		currStartTs := i.getStartTS()
		godes.Advance(currStartTs - previousTs)
		previousTs = currStartTs
		i.setForwardedTs(godes.GetSystemTime())
		re.selector.forward(i)
		bar.Add(1)
	}
	re.selector.terminate()
	godes.WaitUntilDone()
}

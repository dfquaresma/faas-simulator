package common

import (
	"fmt"

	"github.com/agoussia/godes"
	"github.com/schollz/progressbar/v3"
)

type replayer struct {
	*godes.Runner
	invocations *invocations
	selector    *selector
	id          string
}

func NewReplayer(invocations *invocations, selector *selector, id string) *replayer {
	return &replayer{
		Runner:      &godes.Runner{},
		invocations: invocations,
		selector:    selector,
		id:          id,
	}
}

func (re *replayer) Run() {
	fmt.Println("Starting Replayer...")
	godes.AddRunner(re.selector)
	godes.Run()
	previousTs := 0.0

	bar := progressbar.Default(re.invocations.GetSize())
	for i := re.invocations.next(); i != nil; i = re.invocations.next() {
		currStartTs := i.getStartTS()
		if currStartTs-previousTs < 0 {
			fmt.Errorf("NEGATIVE ADVANCE DURING REPLAYER")
			panic(-1)
		}
		godes.Advance(currStartTs - previousTs)
		previousTs = currStartTs
		i.setForwardedTs(godes.GetSystemTime())
		re.selector.forward(i)
		bar.Add(1)
	}
	re.selector.terminate()
	godes.WaitUntilDone()
	godes.Clear()
}

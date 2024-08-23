package common

import (
	"fmt"

	"github.com/agoussia/godes"
	"github.com/dfquaresma/faas-simulator/model"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

type replayer struct {
	*godes.Runner
	dataset  *model.Dataset
	selector *selector
	id       string
	desc     string
}

func NewReplayer(dataset *model.Dataset, selector *selector, id, desc string) *replayer {
	return &replayer{
		Runner:   &godes.Runner{},
		dataset:  dataset,
		selector: selector,
		id:       id,
		desc:     desc,
	}
}

func (re *replayer) Run() {
	fmt.Println("Starting Replayer...")
	godes.Run()
	previousTs := 0.0

	bar := progressbar.NewOptions(re.dataset.GetSize(),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()), //you should install "github.com/k0kubun/go-ansi"
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionSetDescription(re.desc),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	progress := 0
	barjump := 1000
	for i := re.dataset.Next(); i != nil; i = re.dataset.Next() {
		currStartTs := i.GetStartTS()
		godes.Advance(currStartTs - previousTs)
		previousTs = currStartTs
		i.SetForwardedTs(godes.GetSystemTime())
		re.selector.forward(i)
		progress += 1
		if progress%barjump == 0 {
			bar.Add(barjump)
		}
	}
	re.selector.terminate()
	godes.WaitUntilDone()
	godes.Clear()
}

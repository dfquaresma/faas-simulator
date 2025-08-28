package runner

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dfquaresma/faas-simulator/common"
	"github.com/dfquaresma/faas-simulator/io"
	"github.com/dfquaresma/faas-simulator/model"
)

func Sim(tracePath, outputPath string, techniques, tailLatencyProbs []string, idletimes, forwardLatencies []int, skipColdStartOpts []bool) {
	start := time.Now()
	count := 1
	total := len(forwardLatencies) * len(idletimes) * len(tailLatencyProbs) * len(techniques) * len(skipColdStartOpts)

	io.WriteOutputHeaderRow(outputPath, "replayer-stats.csv", []string{"elapsedTime", "id"})
	for _, f := range forwardLatencies {
		fLatency := float64(f)
		for _, scs := range skipColdStartOpts {
			for _, p := range tailLatencyProbs {
				for _, i := range idletimes {
					idleTimeFloat := float64(i)
					for _, t := range techniques {
						replayerOut := simulate(tracePath, outputPath, p, t, fLatency, idleTimeFloat, count, total, scs)
						io.WriteOutputByRow(outputPath, "replayer-stats.csv", replayerOut)
						count++
					}
				}
			}
		}
	}
	fmt.Printf("Total Simulation Time: %s", time.Since(start))
}

func simulate(tracePath, outputPath, prob, technique string, fLatency, idleTimeFloat float64, count, total int, skipColdStart bool) []string {
	cfg := model.Config{
		ForwardLatency:  fLatency,
		Idletime:        idleTimeFloat,
		TailLatencyProb: prob,
		Technique:       technique,
		SkipColdStart:   skipColdStart,
	}
	fmt.Printf(
		"VALUES FOR CFG:\nForwardLatency: %f\nSkipColdStart: %t\nIdletime: %f\nTailLatencyProb: %s\nTechnique: %s\n\n",
		cfg.ForwardLatency,
		cfg.SkipColdStart,
		cfg.Idletime,
		cfg.TailLatencyProb,
		cfg.Technique,
	)

	rows := io.ReadInput(tracePath)
	invocations, err := model.NewDataSet(rows, cfg.TailLatencyProb)
	if err != nil {
		panic(err)
	}

	idleDesc := "INF"
	if idleTimeFloat >= 0 {
		idleDesc = fmt.Sprintf("%.1f", idleTimeFloat)
	}
	simulationName := fmt.Sprintf("%s_idletime%s_tlprob%s_skipColdStart%s", technique, idleDesc, prob, strconv.FormatBool(skipColdStart))
	fmt.Printf("SimulationName: %s\n", simulationName)
	fmt.Printf("OutputPath: %s\n\n", outputPath)

	selector := common.NewSelector(invocations, cfg)
	replayer := common.NewReplayer(invocations, selector, simulationName, fmt.Sprintf("[cyan][%d/%d][reset] Running simulation...", count, total))

	fmt.Print("Starting simulation...")
	replayer.Run()
	fmt.Println("\n..Simulation for " + simulationName + " is finished")

	techniqueOutputPath := outputPath + technique + "/"
	fmt.Println("Writing results at " + techniqueOutputPath)
	io.WriteOutput(techniqueOutputPath, simulationName+"-invocations.csv", invocations.GetOutPut())

	replicasOutput, selectorOutput := selector.GetOutPut()
	io.WriteOutput(techniqueOutputPath, simulationName+"-replicas.csv", replicasOutput)
	io.WriteOutput(techniqueOutputPath, simulationName+"-selector.csv", selectorOutput)
	fmt.Println("Results for " + simulationName + " was written\n")

	return replayer.GetOutPut()
}

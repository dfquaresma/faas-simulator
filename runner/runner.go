package runner

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dfquaresma/faas-simulator/common"
	"github.com/dfquaresma/faas-simulator/io"
	"github.com/dfquaresma/faas-simulator/model"
)

func Sim(tracePath, outputPath string, techniques, hasOracle, tailLatencyProbs []string, idletimes, forwardLatencies []int) {
	start := time.Now()
	count := 1
	total := len(forwardLatencies) * len(idletimes) * len(tailLatencyProbs) * len(hasOracle) * len(techniques)
	replayerStats := [][]string{{"id", "elapsedTime"}}
	for _, f := range forwardLatencies {
		fLatency := float64(f)
		for _, i := range idletimes {
			idleTimeFloat := float64(i)
			for _, o := range hasOracle {
				hasOracle, err := strconv.ParseBool(o)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error in conversion for oracle bool: %s\n", err)
					os.Exit(1)
				}
				for _, p := range tailLatencyProbs {
					for _, t := range techniques {
						replayerOut := simulate(tracePath, outputPath, p, t, fLatency, idleTimeFloat, hasOracle, count, total)
						replayerStats = append(replayerStats, replayerOut)
						count++
					}
				}
			}
		}
	}

	fmt.Printf("Total Simulation Time: %s", time.Since(start))
	io.WriteOutput(outputPath+"/", "replayer-stats.csv", replayerStats)
}

func simulate(tracePath, outputPath, prob, technique string, fLatency, idleTimeFloat float64, hasOracle bool, count, total int) []string {
	cfg := model.Config{
		ForwardLatency:  fLatency,
		Idletime:        idleTimeFloat,
		TailLatencyProb: prob,
		Technique:       technique,
		HasOracle:       hasOracle,
	}
	fmt.Printf(
		"VALUES FOR CFG:\nForwardLatency: %f\nIdletime: %f\nTailLatencyProb: %s\nTechnique: %s\nHasOracle: %t\n\n",
		cfg.ForwardLatency,
		cfg.Idletime,
		cfg.TailLatencyProb,
		cfg.Technique,
		cfg.HasOracle,
	)

	rows := io.ReadInput(tracePath)
	invocations, err := model.NewDataSet(rows, cfg.TailLatencyProb, cfg.HasOracle)
	if err != nil {
		panic(err)
	}

	idleDesc := "INF"
	if idleTimeFloat >= 0 {
		idleDesc = fmt.Sprintf("%.1f", idleTimeFloat)
	}
	simulationName := fmt.Sprintf("%s_hasOracle%v_idletime%s_tlprob%s", technique, hasOracle, idleDesc, prob)
	fmt.Printf("SimulationName: %s\n", simulationName)
	fmt.Printf("OutputPath: %s\n\n", outputPath)

	selector := common.NewSelector(cfg)
	replayer := common.NewReplayer(invocations, selector, simulationName, fmt.Sprintf("[cyan][%d/%d][reset] Running simulation...", count, total))

	fmt.Print("Starting simulation...")
	replayer.Run()
	fmt.Println("\n..Simulation for " + simulationName + " is finished")

	fmt.Println("Writing results at " + outputPath)
	io.WriteOutput(outputPath+"/"+technique+"/", simulationName+"-invocations.csv", invocations.GetOutPut())
	io.WriteOutput(outputPath+"/"+technique+"/", simulationName+"-replicas.csv", selector.GetOutPut())
	fmt.Println("Results for " + simulationName + " was written\n")

	return replayer.GetOutPut()
}

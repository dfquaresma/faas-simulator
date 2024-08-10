package runner

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dfquaresma/faas-simulator/common"
	"github.com/dfquaresma/faas-simulator/io"
	"github.com/dfquaresma/faas-simulator/model"
)

func Sim(tracePath, outputPath string, techniques, hasOracle, tailLatencyProbs []string, idletimes, forwardLatencies []int) {
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
						cfg := model.Config{
							ForwardLatency:  fLatency,
							Idletime:        idleTimeFloat,
							TailLatencyProb: p,
							Technique:       t,
							HasOracle:       hasOracle,
						}
						fmt.Printf(
							"VALUES FOR CFG:\nForwardLatency:%f\nIdletime:%f\nTailLatencyProb:%s\nTechnique:%s\nHasOracle:%t\n\n",
							cfg.ForwardLatency,
							cfg.Idletime,
							cfg.TailLatencyProb,
							cfg.Technique,
							cfg.HasOracle,
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
						simulationName := fmt.Sprintf("%s_hasOracle%s_idletime%s_tlprob%s", t, o, idleDesc, p)
						fmt.Printf("SimulationName: %s\n", simulationName)
						fmt.Printf("OutputPath: %s\n", outputPath)

						selector := common.NewSelector(cfg)
						replayer := common.NewReplayer(invocations, selector, simulationName)
						fmt.Print("Starting simulation...")
						replayer.Run()
						fmt.Println("\n..Simulation for " + simulationName + " is finished")

						fmt.Println("Writing results at " + outputPath)
						io.WriteOutput(outputPath+"/"+t+"/", simulationName+"-invocations.csv", invocations.GetOutPut())
						io.WriteOutput(outputPath+"/"+t+"/", simulationName+"-replicas.csv", selector.GetOutPut())
						fmt.Println("Results for " + simulationName + " was written\n")
					}
				}
			}
		}
	}

}

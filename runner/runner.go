package runner

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dfquaresma/faas-simulator/common"
	"github.com/dfquaresma/faas-simulator/io"
)

func Sim(tracePath, outputPath string, techniques, hasOracle, tailLatencyProbs []string, idletimes, forwardLatencies []int) {
	for _, t := range techniques {
		for _, p := range tailLatencyProbs {
			for _, f := range forwardLatencies {
				fLatency := float64(f)

				for _, i := range idletimes {
					idleTimeFloat := float64(i)

					for _, o := range hasOracle {
						hasOracle, err := strconv.ParseBool(o)
						if err != nil {
							fmt.Println(fmt.Errorf("Error in conversion: ", err))
							os.Exit(0)
						}

						rows := io.ReadInput(tracePath)
						cfg := common.Config{
							ForwardLatency:  fLatency,
							Idletime:        idleTimeFloat,
							TailLatencyProb: p,
							Technique:       t,
							HasOracle:       hasOracle,
						}

						fmt.Printf(
							"Values for cfg: ForwardLatency:%f Idletime:%f TailLatencyProb:%s Technique:%s HasOracle:%t\n",
							cfg.ForwardLatency,
							cfg.Idletime,
							cfg.TailLatencyProb,
							cfg.Technique,
							cfg.HasOracle,
						)

						invocations, err := common.NewInvocations(rows, cfg.TailLatencyProb)
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
						fmt.Println("Results for " + simulationName + " was written")
					}
				}
			}
		}
	}

}

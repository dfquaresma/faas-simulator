package runner

import (
	"fmt"
	"time"

	"github.com/dfquaresma/faas-simulator/common"
	"github.com/dfquaresma/faas-simulator/io"
	"github.com/dfquaresma/faas-simulator/model"
)

func Sim(tracePath, outputPath string, techniques, tailLatencyProbs []string, idletimes []int, forwardLatencies int) {
	start := time.Now()
	count := 1
	fLatency := float64(forwardLatencies)
	total := forwardLatencies * len(idletimes) * len(tailLatencyProbs) * len(techniques)
	io.WriteOutputHeaderRow(outputPath, "replayer-stats.csv", []string{"elapsedTime", "currentTime", "id"})
	for _, p := range tailLatencyProbs {
		for _, i := range idletimes {
			idleTimeFloat := float64(i)
			for _, t := range techniques {
				replayerOut := simulate(tracePath, outputPath, p, t, fLatency, idleTimeFloat, count, total)
				io.WriteOutputByRow(
					outputPath,
					"replayer-stats.csv",
					[]string{
						replayerOut[0],
						time.Now().Format("2006-01-02 15:04:05"),
						replayerOut[1],
					},
				)
				count++
			}
		}
	}
	fmt.Printf("Total Simulation Time: %s", time.Since(start))
}

func simulate(tracePath, outputPath, prob, technique string, fLatency, idleTimeFloat float64, count, total int) []string {
	cfg := model.Config{
		ForwardLatency:  fLatency,
		Idletime:        idleTimeFloat,
		TailLatencyProb: prob,
		Technique:       technique,
	}
	fmt.Printf(
		"VALUES FOR CFG:\nForwardLatency: %f\nIdletime: %f\nTailLatencyProb: %s\nTechnique: %s\n\n",
		cfg.ForwardLatency,
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
	simulationName := fmt.Sprintf("%s_idletime%s_tlprob%s", technique, idleDesc, prob)
	fmt.Printf("SimulationName: %s\n", simulationName)
	fmt.Printf("OutputPath: %s\n\n", outputPath)

	router := common.NewRouter(invocations, cfg)
	replayer := common.NewReplayer(invocations, router, simulationName, fmt.Sprintf("[cyan][%d/%d][reset] Running simulation...", count, total))

	fmt.Print("Starting simulation...")
	replayer.Run()
	fmt.Println("\n..Simulation for " + simulationName + " is finished")

	techniqueOutputPath := outputPath + technique + "/"
	fmt.Println("Writing results at " + techniqueOutputPath)
	io.WriteOutput(techniqueOutputPath, simulationName+"-invocations.csv", invocations.GetOutPut())

	replicasOutput, selectorOutput := router.GetOutPut()
	io.WriteOutput(techniqueOutputPath, simulationName+"-replicas.csv", replicasOutput)
	io.WriteOutput(techniqueOutputPath, simulationName+"-selector.csv", selectorOutput)
	fmt.Println("Results for " + simulationName + " was written\n")

	return replayer.GetOutPut()
}

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/dfquaresma/faas-simulator/common"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("config.json")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
		return
	}

	tracePath := viper.GetString("tracePath")
	techniques := viper.GetStringSlice("resourceProvisioner.technique")
	hasOracle := viper.GetStringSlice("resourceProvisioner.hasOracle")
	ForwardLatencies := viper.GetIntSlice("resourceProvisioner.forwardLatency")
	idletimes := viper.GetIntSlice("resourceProvisioner.idletime")
	tailLatencyProbs := viper.GetStringSlice("resourceProvisioner.tailLatencyProb")
	for _, t := range techniques {
		for _, p := range tailLatencyProbs {
			for _, f := range ForwardLatencies {
				fLatency := float64(f)

				for _, i := range idletimes {
					idleTimeFloat := float64(i)

					for _, o := range hasOracle {
						hasOracle, err := strconv.ParseBool(o)
						if err != nil {
							fmt.Println(fmt.Errorf("Error in conversion: ", err))
							os.Exit(0)
						}

						rows := readInput(tracePath)
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
						outputPath := viper.GetString("outputPath")
						fmt.Printf("SimulationName: %s\n", simulationName)
						fmt.Printf("OutputPath: %s\n", outputPath)

						selector := common.NewSelector(cfg)
						replayer := common.NewReplayer(invocations, selector, simulationName)
						fmt.Print("Starting simulation...")
						replayer.Run()
						fmt.Println("\n..Simulation for " + simulationName + " is finished")

						fmt.Println("Writing results at " + outputPath)
						writeOutput(outputPath+"/"+t+"/", simulationName+"-invocations.csv", invocations.GetOutPut())
						writeOutput(outputPath+"/"+t+"/", simulationName+"-replicas.csv", selector.GetOutPut())
						fmt.Println("Results for " + simulationName + " was written")
					}
				}
			}
		}
	}

}

func readInput(tracePath string) [][]string {
	input, err := os.Open(tracePath)
	if err != nil {
		panic(err)
	}
	defer input.Close()

	r := csv.NewReader(input)
	rows, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	return rows[1:]
}

func writeOutput(outputPath, simulationName string, data [][]string) {
	err := os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	filePath := outputPath + simulationName
	output, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer output.Close()

	writer := csv.NewWriter(output)
	defer writer.Flush()

	for _, record := range data {
		err := writer.Write(record)
		if err != nil {
			panic(err)
		}
	}
}

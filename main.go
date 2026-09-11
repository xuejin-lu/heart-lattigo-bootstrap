package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	configPath := flag.String("config", "", "bootstrap config JSON path; empty uses DefaultBootstrapConfig")
	outputPath := flag.String("out", "", "result JSON path; empty writes JSON to stdout")
	repetitions := flag.Int("repetitions", 0, "override config repetition count")
	warmup := flag.Int("warmup", -1, "override config warm-up count")
	stages := flag.Bool("stages", false, "measure bootstrap stages instead of only the full bootstrap")
	correctness := flag.Bool("correctness", false, "decode one complete bootstrap and record numerical correctness")
	flag.Parse()

	cfg, err := LoadBootstrapConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	if *repetitions > 0 {
		cfg.Repetitions = *repetitions
	}
	if *warmup >= 0 {
		cfg.Warmup = *warmup
	}

	primaryRoot, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	backendRoot := os.Getenv("LATTIGO_REPO")
	if backendRoot == "" {
		backendRoot = primaryRoot + "/../lattigo"
	}
	if *stages && *correctness {
		log.Fatal("-stages and -correctness cannot be used together")
	}
	if *correctness {
		result, err := RunCorrectnessExperiment(cfg, primaryRoot, backendRoot)
		if err != nil {
			log.Fatal(err)
		}
		if err := WriteCorrectnessResult(result, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *stages {
		result, err := RunStageExperiment(cfg, primaryRoot, backendRoot)
		if err != nil {
			log.Fatal(err)
		}
		if err := WriteStageExperimentResult(result, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else {
		result, err := RunExperiment(cfg, primaryRoot, backendRoot)
		if err != nil {
			log.Fatal(err)
		}
		if err := WriteExperimentResult(result, *outputPath); err != nil {
			log.Fatal(err)
		}
	}
	if *outputPath != "" {
		fmt.Printf("Wrote experiment result: %s\n", *outputPath)
	}
}

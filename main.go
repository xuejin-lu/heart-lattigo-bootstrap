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
	diagnostic := flag.Bool("diagnostic", false, "trace public bootstrap stages for numerical diagnosis")
	finalizationDiagnostic := flag.Bool("finalization-diagnostic", false, "diagnose the Fast public finalization boundary")
	standardReference := flag.String("standard-reference", "", "EXP-002-C-DIAG-P Standard result JSON used as the logical reference")
	q01Diagnostic := flag.Bool("q01-diagnostic", false, "diagnose q0/q1 intermediate projections")
	q01StandardReference := flag.String("q01-standard-reference", "", "EXP-002-C-DIAG-Q01 Standard result JSON used as the logical reference")
	fix001DiagPoly := flag.Bool("fix001-diag-poly", false, "replay formal Chebyshev powers for FIX-001 diagnosis")
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
	selectedModes := 0
	if *stages {
		selectedModes++
	}
	if *correctness {
		selectedModes++
	}
	if *diagnostic {
		selectedModes++
	}
	if *finalizationDiagnostic {
		selectedModes++
	}
	if *q01Diagnostic {
		selectedModes++
	}
	if *fix001DiagPoly {
		selectedModes++
	}
	if selectedModes > 1 {
		log.Fatal("bootstrap diagnostic modes cannot be used together")
	}
	if *q01Diagnostic {
		result, err := RunQ01DiagnosticExperiment(cfg, primaryRoot, backendRoot, *q01StandardReference)
		if err != nil {
			log.Fatal(err)
		}
		if err := WriteQ01DiagnosticResult(result, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001DiagPoly {
		if err := runFIX001DiagPoly(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *finalizationDiagnostic {
		result, err := RunFinalizationBoundaryExperiment(cfg, primaryRoot, backendRoot, *standardReference)
		if err != nil {
			log.Fatal(err)
		}
		if err := WriteFinalizationDiagnosticResult(result, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *diagnostic {
		result, err := RunDiagnosticExperiment(cfg, primaryRoot, backendRoot)
		if err != nil {
			log.Fatal(err)
		}
		if err := WriteDiagnosticResult(result, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *correctness {
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

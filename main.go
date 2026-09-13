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
	fix001DiagT3 := flag.Bool("fix001-diag-t3", false, "replay formal T3 correction alignment for FIX-001 diagnosis")
	fix001DiagT3Capacity := flag.Bool("fix001-diag-t3-capacity", false, "diagnose FIX-001 T3 q0/q1 capacity")
	fix001P3GlobalSemantics := flag.Bool("fix001-p3-global-semantics", false, "diagnose source-backed global PS semantics")
	fix001P3ChebyshevOracleBasis := flag.Bool("fix001-p3-chebyshev-oracle-basis", false, "resolve Chebyshev plaintext oracle basis semantics")
	fix001P3TargetScaleRestoration := flag.Bool("fix001-p3-target-scale-restoration", false, "validate canonical 2^91 public target-scale restoration")
	fix001P3TargetScaleCapacityDomain := flag.Bool("fix001-p3-target-scale-capacity-domain", false, "correct q0/q1 capacity oracle domain for canonical 2^91 restoration")
	fix001P3DoubleAngle := flag.Bool("fix001-p3-double-angle", false, "diagnose FIX-001 DoubleAngle rounds")
	fix001P3DoubleAngleMulAlias := flag.Bool("fix001-p3-double-angle-mul-alias", false, "isolate FIX-001 DoubleAngle square alias")
	fix001P3DesignNormalized := flag.Bool("fix001-p3-design-normalized-double-angle", false, "validate normalized FIX-001 DoubleAngle recurrence")
	fix001P3IntegrateLogN13Mod1 := flag.Bool("fix001-p3-integrate-logn13-mod1", false, "verify production normalized LogN13 Fast Mod1")
	fix001P3DiagPostMod1S2C := flag.Bool("fix001-p3-diag-logn13-post-mod1-s2c", false, "validate production LogN13 post-Mod1 SlotsToCoeffs core output")
	fix001P3ValidateLogN13E2E := flag.Bool("fix001-p3-validate-logn13-e2e", false, "validate the production LogN13 Fast Bootstrap public boundary")
	fix001P3DiagLogN13SemanticBisect := flag.Bool("fix001-p3-diag-logn13-semantic-bisect", false, "bisect LogN13 Fast semantic divergence against a genuine Standard control")
	fix001P3DiagLogN13EvalModCausal := flag.Bool("fix001-p3-diag-logn13-evalmod-causal", false, "localize LogN13 EvalMod divergence with matched semantic inputs")
	fix001P3DiagLogN13C2SPrecision := flag.Bool("fix001-p3-diag-logn13-c2s-precision", false, "localize LogN13 C2S precision loss against a downstream budget")
	fix001P3DiagLogN13C2SGroup0LinearAlias := flag.Bool("fix001-p3-diag-logn13-c2s-group0-linear-alias", false, "localize LogN13 group 0 LinearTransform q0/q1 alias")
	fix001P3DesignLogN13C2SGroup0CompressedLinear := flag.Bool("fix001-p3-design-logn13-c2s-group0-compressed-linear", false, "validate diagnostic compressed LogN13 group 0 LinearTransform")
	fix001P3DiagLogN13C2SGroup1LinearAlias := flag.Bool("fix001-p3-diag-logn13-c2s-group1-linear-alias", false, "localize LogN13 group 1 LinearTransform q0/q1 alias")
	fix001P3DesignLogN13C2SGroup1CompressedLinear := flag.Bool("fix001-p3-design-logn13-c2s-group1-compressed-linear", false, "validate diagnostic compressed LogN13 group 1 LinearTransform")
	fix001P3DiagLogN13C2SFinalSplitProbeFix := flag.Bool("fix001-p3-diag-logn13-c2s-final-split-probe-fix", false, "replay LogN13 C2S final split without reapplying DFT factors")
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
	if *fix001DiagT3 {
		selectedModes++
	}
	if *fix001DiagT3Capacity {
		selectedModes++
	}
	if *fix001P3GlobalSemantics {
		selectedModes++
	}
	if *fix001P3ChebyshevOracleBasis {
		selectedModes++
	}
	if *fix001P3TargetScaleRestoration {
		selectedModes++
	}
	if *fix001P3TargetScaleCapacityDomain {
		selectedModes++
	}
	if *fix001P3DoubleAngle {
		selectedModes++
	}
	if *fix001P3DoubleAngleMulAlias {
		selectedModes++
	}
	if *fix001P3DesignNormalized {
		selectedModes++
	}
	if *fix001P3IntegrateLogN13Mod1 {
		selectedModes++
	}
	if *fix001P3DiagPostMod1S2C {
		selectedModes++
	}
	if *fix001P3ValidateLogN13E2E {
		selectedModes++
	}
	if *fix001P3DiagLogN13SemanticBisect {
		selectedModes++
	}
	if *fix001P3DiagLogN13EvalModCausal {
		selectedModes++
	}
	if *fix001P3DiagLogN13C2SPrecision {
		selectedModes++
	}
	if *fix001P3DiagLogN13C2SGroup0LinearAlias {
		selectedModes++
	}
	if *fix001P3DesignLogN13C2SGroup0CompressedLinear {
		selectedModes++
	}
	if *fix001P3DiagLogN13C2SGroup1LinearAlias {
		selectedModes++
	}
	if *fix001P3DesignLogN13C2SGroup1CompressedLinear {
		selectedModes++
	}
	if *fix001P3DiagLogN13C2SFinalSplitProbeFix {
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
	} else if *fix001DiagT3 {
		if err := runFIX001T3(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001DiagT3Capacity {
		if err := runFIX001T3Capacity(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3GlobalSemantics {
		if err := runFIX001P3GlobalSemantics(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3ChebyshevOracleBasis {
		if err := runFIX001P3ChebyshevOracleBasis(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3TargetScaleRestoration {
		if err := runFIX001P3TargetScaleRestoration(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3TargetScaleCapacityDomain {
		if err := runFIX001P3TargetScaleCapacityDomain(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DoubleAngle {
		if err := runFIX001P3DoubleAngle(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DoubleAngleMulAlias {
		if err := runFIX001P3DoubleAngleMulAlias(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignNormalized {
		if err := runFIX001P3DesignNormalizedRecurrence(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3IntegrateLogN13Mod1 {
		if err := runFIX001P3IntegrateLogN13Mod1(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagPostMod1S2C {
		if err := runFIX001P3DiagLogN13PostMod1S2C(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3ValidateLogN13E2E {
		if err := runFIX001P3ValidateLogN13E2E(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13SemanticBisect {
		if err := runFIX001P3DiagLogN13SemanticBisect(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13EvalModCausal {
		if err := runFIX001P3DiagLogN13EvalModCausal(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13C2SPrecision {
		if err := runFIX001P3DiagLogN13C2SPrecision(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13C2SGroup0LinearAlias {
		if err := runFIX001P3DiagLogN13C2SGroup0LinearAlias(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13C2SGroup0CompressedLinear {
		if err := runFIX001P3DesignLogN13C2SGroup0CompressedLinear(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13C2SGroup1LinearAlias {
		if err := runFIX001P3DiagLogN13C2SGroup1LinearAlias(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13C2SGroup1CompressedLinear {
		if err := runFIX001P3DesignLogN13C2SGroup1CompressedLinear(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13C2SFinalSplitProbeFix {
		if err := runFIX001P3DiagLogN13C2SFinalSplitProbeFix(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
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

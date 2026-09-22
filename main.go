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
	fix001P3DiagLogN13EvalModMatchedNormalizedOracle := flag.Bool("fix001-p3-diag-logn13-evalmod-matched-normalized-oracle", false, "diagnose LogN13 EvalMod with matched normalized full-RNS oracle")
	fix001P3IntegrateLogN13C2SCompression := flag.Bool("fix001-p3-integrate-logn13-c2s-compression", false, "integrate and validate production LogN13 Fast C2S compression")
	fix001P3DiagLogN13EvalModS2CAmplification := flag.Bool("fix001-p3-diag-logn13-evalmod-s2c-amplification", false, "diagnose LogN13 EvalMod semantic amplification through S2C")
	fix001P3DesignLogN13EvalModPrecisionScaleSweep := flag.Bool("fix001-p3-design-logn13-evalmod-precision-scale-sweep", false, "sweep diagnostic Fast EvalMod common plan scales")
	fix001P3DesignLogN13GuardedPowerPrecision := flag.Bool("fix001-p3-design-logn13-guarded-power-precision", false, "sweep diagnostic guarded Chebyshev power precision")
	fix001P3DiagLogN13PolynomialPrecisionDecomposition := flag.Bool("fix001-p3-diag-logn13-polynomial-precision-decomposition", false, "decompose LogN13 polynomial precision into power semantics and Fast PS arithmetic")
	fix001P3DesignLogN13PSOracleScalePrecision := flag.Bool("fix001-p3-design-logn13-ps-oracle-scale-precision", false, "sweep LogN13 Fast PS oracle common plan scales")
	fix001P3DesignLogN13PSMixedScalePrecision := flag.Bool("fix001-p3-design-logn13-ps-mixed-scale-precision", false, "design LogN13 Fast PS mixed per-block oracle scales")
	fix001P3DesignLogN13PSRescaleGuardPrecision := flag.Bool("fix001-p3-design-logn13-ps-rescale-guard-precision", false, "design LogN13 Fast PS pre-rescale guard factors")
	fix001P3DesignLogN13Q056RescaleGuardFeasibility := flag.Bool("fix001-p3-design-logn13-q0-56-rescale-guard-feasibility", false, "test LogN13 q0=56 rescale-guard feasibility")
	fix001P3DesignLogN13Q056F0TwoBitGuard := flag.Bool("fix001-p3-design-logn13-q0-56-f0-two-bit-guard", false, "test LogN13 q0=56 F0 two-bit guard")
	fix001P3DesignLogN13F0LocalQ2RescaleFeasibility := flag.Bool("fix001-p3-design-logn13-f0-local-q2-rescale-feasibility", false, "test LogN13 F0 local q2 rescale feasibility")
	fix001P3DesignLogN13F0LocalQ2GuardSweep := flag.Bool("fix001-p3-design-logn13-f0-local-q2-guard-sweep", false, "sweep LogN13 F0 local q2 guard bits")
	fix001P3DiagLogN13K3DownstreamSufficiency := flag.Bool("fix001-p3-diag-logn13-k3-downstream-sufficiency", false, "diagnose LogN13 k3 downstream sufficiency")
	fix001P3DesignLogN13DAR0LocalQ2Feasibility := flag.Bool("fix001-p3-design-logn13-da-r0-local-q2-feasibility", false, "test LogN13 DA round0 local q2 feasibility")
	fix001P3DesignLogN13DAAllRoundsLocalQ2 := flag.Bool("fix001-p3-design-logn13-da-all-rounds-local-q2", false, "test LogN13 all-rounds DA local q2 feasibility")
	fix001P3DesignLogN13G0LocalQ2GuardSweep := flag.Bool("fix001-p3-design-logn13-g0-local-q2-guard-sweep", false, "sweep LogN13 G0 local q2 guard bits")
	fix001P3DiagLogN13S2CErrorDecomposition := flag.Bool("fix001-p3-diag-logn13-s2c-error-decomposition", false, "decompose LogN13 S2C error for the accepted G0=2 candidate")
	fix001P3DiagLogN13EvalModResidualDecomposition := flag.Bool("fix001-p3-diag-logn13-evalmod-residual-error-decomposition", false, "decompose LogN13 EvalMod residual error")
	fix001P3DiagLogN13HighPrecisionOracleRecovery := flag.Bool("fix001-p3-diag-logn13-high-precision-oracle-recovery", false, "recover LogN13 high-precision EvalMod oracle")
	fix001P3DiagLogN13QuantizationAwareEvalModDecomposition := flag.Bool("fix001-p3-diag-logn13-quantization-aware-evalmod-decomposition", false, "decompose LogN13 EvalMod with canonical CKKS quantization")
	fix001P3DiagLogN13PSResidualLocalization := flag.Bool("fix001-p3-diag-logn13-ps-residual-localization", false, "localize LogN13 PS residual with checkpoint resets")
	fix001P3DiagLogN13G0MergeAttribution := flag.Bool("fix001-p3-diag-logn13-g0-merge-attribution", false, "attribute LogN13 G0 merge residual")
	fix001P3DiagLogN13B4BabyStepAttribution := flag.Bool("fix001-p3-diag-logn13-b4-baby-step-attribution", false, "attribute LogN13 B4 baby-step residual")
	fix001P3DesignLogN13B4T2LocalScalarGuard := flag.Bool("fix001-p3-design-logn13-b4-t2-local-scalar-guard", false, "test LogN13 B4 T2 local scalar guard feasibility")
	fix001P3IntegrateLogN13B4T2OneBitScalarGuard := flag.Bool("fix001-p3-integrate-logn13-b4-t2-one-bit-scalar-guard", false, "integrate and validate the production LogN13 B4 T2 one-bit scalar guard")
	fix001P3DiagLogN13GeneratedPowerReentry := flag.Bool("fix001-p3-diag-logn13-generated-power-reentry", false, "diagnose LogN13 generated-power reentry under the accepted oracle stack")
	fix001P3DesignLogN13GeneratedPowerMultiplyFirstLocalQ2 := flag.Bool("fix001-p3-design-logn13-generated-power-multiply-first-local-q2", false, "design LogN13 generated-power multiply-first temporary-q2 feasibility")
	fix001P3DesignLogN13GeneratedPowerNativeQ01MultiplyFirst := flag.Bool("fix001-p3-design-logn13-generated-power-native-q01-multiply-first", false, "design LogN13 native q0/q1 generated-power multiply-first feasibility")
	fix001P3DesignLogN13GeneratedPowerNativeQ012MultiplyFirst := flag.Bool("fix001-p3-design-logn13-generated-power-native-q012-multiply-first", false, "design LogN13 native q0/q1/q2 generated-power multiply-first feasibility")
	fix001P3DesignLogN13D0ProductionPreconditionClosure := flag.Bool("fix001-p3-design-logn13-d0-production-precondition-closure", false, "close LogN13 D0 production preconditions")
	fix001P3DesignLogN13Q055Plan92T2LocalQ2Guard := flag.Bool("fix001-p3-design-logn13-q055-plan92-t2-local-q2-guard-feasibility", false, "evaluate q0=55 plan92 final-parent T2 temporary-q2 guard feasibility")
	fix001P3DiagLogN13Q055Plan92PSFirstDivergence := flag.Bool("fix001-p3-diag-logn13-q055-plan92-ps-first-divergence", false, "localize the first LogN13 q0=55 plan92 PS divergence")
	fix001P3DiagLogN13Q055Plan92PSDependencyWindowClosure := flag.Bool("fix001-p3-diag-logn13-q055-plan92-ps-dependency-window-closure", false, "close the LogN13 q0=55 plan92 PS dependency q012 window")
	fix001P3DiagLogN13Q012G0F0SourceFaithfulClosure := flag.Bool("fix001-p3-diag-logn13-q012-g0-f0-source-faithful-closure", false, "close LogN13 q012 G0-F0 source-faithful dependency path")
	fix001P3DesignLogN13PSWideQ012Q056ScaleSweep := flag.Bool("fix001-p3-design-logn13-ps-wide-q012-q056-scale-sweep", false, "sweep diagnostic PS-wide Q012 q0=56 scales")
	fix001P3DiagP93EvalModFirstDivergence := flag.Bool("fix001-p3-diag-p93-evalmod-first-divergence", false, "locate first fresh P93 reference versus production EvalMod divergence")
	fix001P3DiagP93S2CAttribution := flag.Bool("fix001-p3-diag-p93-s2c-attribution", false, "reconcile accepted P93 reference with dirty production S2C")
	fix001P3DiagP93ProductionC2SFirstMaterialDivergence := flag.Bool("fix001-p3-diag-p93-production-c2s-first-material-divergence", false, "locate first P93 production C2S material divergence")
	fix001P3DiagP93ActualEvalModVsForced := flag.Bool("fix001-p3-diag-p93-actual-evalmod-vs-forced", false, "compare actual public Fast EvalMod with forced P93 Fast path")
	fix001P3DiagP93EvalModPublicScaleResetCausalProof := flag.Bool("fix001-p3-diag-p93-evalmod-public-scale-reset-causal-proof", false, "prove causal effect of the public Fast EvalMod scale reset")
	fix001P3DiagP93PostS2CScaleCollapseCausalProof := flag.Bool("fix001-p3-diag-p93-post-s2c-scale-collapse-causal-proof", false, "prove the first downstream post-S2C scale convergence boundary")
	fix001P3DiagP93GenuineStandardVsMatchedReconciliation := flag.Bool("fix001-p3-diag-p93-genuine-standard-vs-matched-reconciliation", false, "reconcile Genuine Standard and matched diagnostic reference lineages")
	fix001P3DiagP93FastVsGenuineStandardInternalBisect := flag.Bool("fix001-p3-diag-p93-fast-vs-genuine-standard-evalmod-internal-bisect", false, "bisect source-faithful Fast vs Genuine Standard EvalMod internals")
	fix001P3DiagP93EvalModLockstepLogicalSemantics := flag.Bool("fix001-p3-diag-p93-evalmod-lockstep-logical-semantics-trace", false, "trace source-faithful Fast and Genuine Standard EvalMod in canonical logical coordinates")
	fix001P3DiagP93FinalMaintainedRestoreFactor := flag.Bool("fix001-p3-diag-p93-final-maintained-restore-factor-causal-proof", false, "prove the causal effect of the final Fast maintained restore factor")
	fix001P3DiagP93DoubleAnglePlaintextDecomposition := flag.Bool("fix001-p3-diag-p93-doubleangle-plaintext-causal-decomposition", false, "decompose plaintext DoubleAngle propagation from Fast implementation error")
	fix001P3DiagP93PolynomialReferenceReconciliation := flag.Bool("fix001-p3-diag-p93-polynomial-reference-reconciliation", false, "reconcile fresh historical P93 polynomial reference with Genuine Standard and localize PS regression")
	fix001P3DiagP93T3CausalDecomposition := flag.Bool("fix001-p3-diag-p93-t3-generated-power-causal-decomposition", false, "decompose P93 T3 generated-power regression into input propagation and implementation effect")
	fix001P3DiagP93T3PrimitiveSemanticResidualLocalization := flag.Bool("fix001-p3-diag-p93-t3-primitive-semantic-residual-localization", false, "localize P93 T3 primitive semantic residual at stable checkpoints")
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
	if *fix001P3DiagLogN13EvalModMatchedNormalizedOracle {
		selectedModes++
	}
	if *fix001P3IntegrateLogN13C2SCompression {
		selectedModes++
	}
	if *fix001P3DiagLogN13EvalModS2CAmplification {
		selectedModes++
	}
	if *fix001P3DesignLogN13EvalModPrecisionScaleSweep {
		selectedModes++
	}
	if *fix001P3DesignLogN13GuardedPowerPrecision {
		selectedModes++
	}
	if *fix001P3DiagLogN13PolynomialPrecisionDecomposition {
		selectedModes++
	}
	if *fix001P3DesignLogN13PSOracleScalePrecision {
		selectedModes++
	}
	if *fix001P3DesignLogN13PSMixedScalePrecision {
		selectedModes++
	}
	if *fix001P3DesignLogN13PSRescaleGuardPrecision {
		selectedModes++
	}
	if *fix001P3DesignLogN13Q056RescaleGuardFeasibility {
		selectedModes++
	}
	if *fix001P3DesignLogN13Q056F0TwoBitGuard {
		selectedModes++
	}
	if *fix001P3DesignLogN13F0LocalQ2RescaleFeasibility {
		selectedModes++
	}
	if *fix001P3DesignLogN13F0LocalQ2GuardSweep {
		selectedModes++
	}
	if *fix001P3DiagLogN13K3DownstreamSufficiency {
		selectedModes++
	}
	if *fix001P3DesignLogN13DAR0LocalQ2Feasibility {
		selectedModes++
	}
	if *fix001P3DesignLogN13DAAllRoundsLocalQ2 {
		selectedModes++
	}
	if *fix001P3DesignLogN13G0LocalQ2GuardSweep {
		selectedModes++
	}
	if *fix001P3DiagLogN13S2CErrorDecomposition {
		selectedModes++
	}
	if *fix001P3DiagLogN13EvalModResidualDecomposition {
		selectedModes++
	}
	if *fix001P3DiagLogN13HighPrecisionOracleRecovery {
		selectedModes++
	}
	if *fix001P3DiagLogN13QuantizationAwareEvalModDecomposition {
		selectedModes++
	}
	if *fix001P3DiagLogN13PSResidualLocalization {
		selectedModes++
	}
	if *fix001P3DiagLogN13G0MergeAttribution {
		selectedModes++
	}
	if *fix001P3DiagLogN13B4BabyStepAttribution {
		selectedModes++
	}
	if *fix001P3DesignLogN13B4T2LocalScalarGuard {
		selectedModes++
	}
	if *fix001P3IntegrateLogN13B4T2OneBitScalarGuard {
		selectedModes++
	}
	if *fix001P3DiagLogN13GeneratedPowerReentry {
		selectedModes++
	}
	if *fix001P3DesignLogN13GeneratedPowerMultiplyFirstLocalQ2 {
		selectedModes++
	}
	if *fix001P3DesignLogN13GeneratedPowerNativeQ01MultiplyFirst {
		selectedModes++
	}
	if *fix001P3DesignLogN13GeneratedPowerNativeQ012MultiplyFirst {
		selectedModes++
	}
	if *fix001P3DesignLogN13D0ProductionPreconditionClosure {
		selectedModes++
	}
	if *fix001P3DesignLogN13Q055Plan92T2LocalQ2Guard {
		selectedModes++
	}
	if *fix001P3DiagLogN13Q055Plan92PSFirstDivergence {
		selectedModes++
	}
	if *fix001P3DiagLogN13Q055Plan92PSDependencyWindowClosure {
		selectedModes++
	}
	if *fix001P3DiagP93ProductionC2SFirstMaterialDivergence {
		selectedModes++
	}
	if *fix001P3DiagP93ActualEvalModVsForced {
		selectedModes++
	}
	if *fix001P3DiagP93EvalModPublicScaleResetCausalProof {
		selectedModes++
	}
	if *fix001P3DiagP93PostS2CScaleCollapseCausalProof {
		selectedModes++
	}
	if *fix001P3DiagP93GenuineStandardVsMatchedReconciliation {
		selectedModes++
	}
	if *fix001P3DiagP93FastVsGenuineStandardInternalBisect {
		selectedModes++
	}
	if *fix001P3DiagP93EvalModLockstepLogicalSemantics {
		selectedModes++
	}
	if *fix001P3DiagP93FinalMaintainedRestoreFactor {
		selectedModes++
	}
	if *fix001P3DiagP93DoubleAnglePlaintextDecomposition {
		selectedModes++
	}
	if *fix001P3DiagP93PolynomialReferenceReconciliation {
		selectedModes++
	}
	if *fix001P3DiagP93T3CausalDecomposition {
		selectedModes++
	}
	if *fix001P3DiagP93T3PrimitiveSemanticResidualLocalization {
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
	} else if *fix001P3DiagLogN13EvalModMatchedNormalizedOracle {
		if err := runFIX001P3DiagLogN13EvalModMatchedNormalizedOracle(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3IntegrateLogN13C2SCompression {
		if err := runFIX001P3IntegrateLogN13C2SCompression(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13EvalModS2CAmplification {
		if err := runFIX001P3DiagLogN13EvalModS2CAmplification(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13EvalModPrecisionScaleSweep {
		if err := runFIX001P3DesignLogN13EvalModPrecisionScaleSweep(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13GuardedPowerPrecision {
		if err := runFIX001P3DesignLogN13GuardedPowerPrecision(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13PolynomialPrecisionDecomposition {
		if err := runFIX001P3DiagLogN13PolynomialPrecisionDecomposition(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13PSOracleScalePrecision {
		if err := runFIX001P3DesignLogN13PSOracleScalePrecision(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13PSMixedScalePrecision {
		if err := runFIX001P3DesignLogN13PSMixedScalePrecision(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13PSRescaleGuardPrecision {
		if err := runFIX001P3DesignLogN13PSRescaleGuardPrecision(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13Q056RescaleGuardFeasibility {
		if err := runFIX001P3DesignLogN13Q056RescaleGuardFeasibility(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13Q056F0TwoBitGuard {
		if err := runFIX001P3DesignLogN13Q056F0TwoBitGuard(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13F0LocalQ2RescaleFeasibility {
		if err := runFIX001P3DesignLogN13F0LocalQ2RescaleFeasibility(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13F0LocalQ2GuardSweep {
		if err := runFIX001P3DesignLogN13F0LocalQ2GuardSweep(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13K3DownstreamSufficiency {
		if err := runFIX001P3DiagLogN13K3DownstreamSufficiency(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13DAR0LocalQ2Feasibility {
		if err := runFIX001P3DesignLogN13DAR0LocalQ2Feasibility(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13DAAllRoundsLocalQ2 {
		if err := runFIX001P3DesignLogN13DAAllRoundsLocalQ2(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13G0LocalQ2GuardSweep {
		if err := runFIX001P3DesignLogN13G0LocalQ2GuardSweep(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13S2CErrorDecomposition {
		if err := runFIX001P3DiagLogN13S2CErrorDecomposition(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13EvalModResidualDecomposition {
		if err := runFIX001P3DiagLogN13EvalModResidualDecomposition(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13HighPrecisionOracleRecovery {
		if err := runFIX001P3DiagLogN13HighPrecisionOracleRecovery(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13QuantizationAwareEvalModDecomposition {
		if err := runFIX001P3DiagLogN13QuantizationAwareEvalModDecomposition(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13PSResidualLocalization {
		if err := runFIX001P3DiagLogN13PSResidualLocalization(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13G0MergeAttribution {
		if err := runFIX001P3DiagLogN13G0MergeAttribution(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13B4BabyStepAttribution {
		if err := runFIX001P3DiagLogN13B4BabyStepAttribution(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13B4T2LocalScalarGuard {
		if err := runFIX001P3DesignLogN13B4T2LocalScalarGuard(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3IntegrateLogN13B4T2OneBitScalarGuard {
		if err := runFIX001P3IntegrateLogN13B4T2OneBitScalarGuard(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13GeneratedPowerReentry {
		if err := runFIX001P3DiagLogN13GeneratedPowerReentry(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13GeneratedPowerMultiplyFirstLocalQ2 {
		if err := runFIX001P3DesignLogN13GeneratedPowerMultiplyFirstLocalQ2(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13GeneratedPowerNativeQ01MultiplyFirst {
		if err := runFIX001P3DesignLogN13GeneratedPowerNativeQ01MultiplyFirst(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13GeneratedPowerNativeQ012MultiplyFirst {
		if err := runFIX001P3DesignLogN13GeneratedPowerNativeQ012MultiplyFirst(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13D0ProductionPreconditionClosure {
		if err := runFIX001P3DesignLogN13D0ProductionPreconditionClosure(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13Q055Plan92T2LocalQ2Guard {
		if err := runFIX001P3DesignLogN13Q055Plan92T2LocalQ2GuardFeasibility(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13Q055Plan92PSFirstDivergence {
		if err := runFIX001P3DiagLogN13Q055Plan92PSFirstDivergence(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13Q055Plan92PSDependencyWindowClosure {
		if err := runFIX001P3DiagLogN13Q055Plan92PSDependencyWindowClosure(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagLogN13Q012G0F0SourceFaithfulClosure {
		if err := runFIX001P3DiagLogN13Q012G0F0SourceFaithfulClosure(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DesignLogN13PSWideQ012Q056ScaleSweep {
		if err := runFIX001P3DesignLogN13PSWideQ012Q056ScaleSweep(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93EvalModFirstDivergence {
		if err := runFIX001P3DiagP93EvalModFirstDivergence(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93S2CAttribution {
		if err := runFIX001P3DiagP93S2CAttribution(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93ProductionC2SFirstMaterialDivergence {
		if err := runFIX001P3DiagP93ProductionC2SFirstMaterialDivergence(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93ActualEvalModVsForced {
		if err := runFIX001P3DiagP93ActualEvalModVsForced(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93EvalModPublicScaleResetCausalProof {
		if err := runFIX001P3DiagP93EvalModPublicScaleResetCausalProof(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93PostS2CScaleCollapseCausalProof {
		if err := runFIX001P3DiagP93PostS2CScaleCollapseCausalProof(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93GenuineStandardVsMatchedReconciliation {
		if err := runFIX001P3DiagP93GenuineStandardVsMatchedReconciliation(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93FastVsGenuineStandardInternalBisect {
		if err := runFIX001P3DiagP93FastVsGenuineStandardEvalModInternalBisect(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93EvalModLockstepLogicalSemantics {
		if err := runFIX001P3DiagP93EvalModLockstepLogicalSemantics(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93FinalMaintainedRestoreFactor {
		if err := runFIX001P3DiagP93FinalMaintainedRestoreFactorCausalProof(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93DoubleAnglePlaintextDecomposition {
		if err := runFIX001P3DiagP93DoubleAnglePlaintextCausalDecomposition(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93PolynomialReferenceReconciliation {
		if err := runFIX001P3DiagP93PolynomialReferenceReconciliation(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93T3CausalDecomposition {
		if err := runFIX001P3DiagP93T3CausalDecomposition(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
			log.Fatal(err)
		}
	} else if *fix001P3DiagP93T3PrimitiveSemanticResidualLocalization {
		if err := runFIX001P3DiagP93T3PrimitiveSemanticResidualLocalization(cfg, primaryRoot, backendRoot, *outputPath); err != nil {
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

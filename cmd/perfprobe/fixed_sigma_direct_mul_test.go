package main

import (
	"testing"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

func TestFixedSigmaDirectMulOptionsRequireVersionedFixtureAndSeededArm(t *testing.T) {
	backendCommit := publicStandardSHA
	if publicBackendName() == "fast" {
		backendCommit = publicFixedSigmaChainFastSHA
	}
	valid := cliOptions{
		mode: "public-native", profile: "logn13", backendCommit: backendCommit,
		mulRescaleChain: true, fixedSigmaWorkload: perfmeasure.PublicDirectMulWorkloadID,
		fixedSigmaAblationArm: "OFF/OFF", fixedSigmaCVariant: "direct-base",
		fixedSigmaSeedProvided: true, bootstrapBudget: 0,
	}
	if err := validateFixedSigmaNoiseOptions(valid); err != nil {
		t.Fatalf("valid direct-ctct baseline options rejected: %v", err)
	}
	for _, mutate := range []func(*cliOptions){
		func(opts *cliOptions) { opts.fixedSigmaAblationArm = "" },
		func(opts *cliOptions) { opts.fixedSigmaCVariant = "canonical" },
		func(opts *cliOptions) { opts.fixedSigmaWorkload = "legacy-add-fixture" },
		func(opts *cliOptions) { opts.mode = "legacy-diagnostic" },
		func(opts *cliOptions) { opts.bootstrapBudget = 1 },
	} {
		candidate := valid
		mutate(&candidate)
		if err := validateFixedSigmaNoiseOptions(candidate); err == nil {
			t.Fatalf("invalid direct-ctct options accepted: %+v", candidate)
		}
	}
}

func TestFixedSigmaDirectMulPropagationDecompositionUsesMatchedC(t *testing.T) {
	xOff := []complex128{complex(0.25, -0.5), complex(0.75, 0.125)}
	delta := []complex128{complex(0.01, -0.02), complex(-0.03, 0.04)}
	xOn := addComplexVectors(xOff, delta)
	actualDelta := subtractComplexVectors(xOn, xOff)
	c := []complex128{complex(0.5, 0.125), complex(-0.25, 0.75)}
	r2b := []complex128{complex(0.001, -0.002), complex(-0.003, 0.004)}
	r2a := []complex128{complex(0.002, -0.001), complex(-0.004, 0.003)}
	yOff := addComplexVectors(multiplyComplexVectors(xOff, c), r2b)
	yOn := addComplexVectors(multiplyComplexVectors(xOn, c), r2a)
	rows := map[string]string{"A.c0.q0": "same", "B.c0.q0": "same", "C_q4.c0.q0": "same"}
	makeRun := func(x, y []complex128, helperDelta []complex128) loadedFixedSigmaAblationArtifact {
		stages := map[string]fixedSigmaChainStageVectors{}
		if helperDelta != nil {
			stages["stage1"] = fixedSigmaChainStageVectors{DecodedNoiseDelta: encodeVector(helperDelta)}
		}
		return loadedFixedSigmaAblationArtifact{
			doc: publicNativeDocument{FixedSigmaChain: &fixedSigmaChainEvidence{FastInputRowsSHA256: rows}},
			vectors: publicVectorDocument{
				Checkpoints: map[string][]complexValue{
					"chain_stage1_output_level4": encodeVector(x),
					"chain_stage2_output_level3": encodeVector(y),
					"encrypt_c_q4_level4":        encodeVector(c),
				},
				FixedSigmaChain: &fixedSigmaChainVectors{Stages: stages},
			},
		}
	}
	runs := map[string]loadedFixedSigmaAblationArtifact{
		"OFF/OFF": makeRun(xOff, yOff, nil),
		"ON/OFF":  makeRun(xOn, yOn, actualDelta),
		"OFF/ON":  makeRun(xOff, yOff, nil),
		"ON/ON":   makeRun(xOn, yOn, nil),
	}
	result, err := computeFixedSigmaAblationPropagation(runs)
	if err != nil {
		t.Fatal(err)
	}
	if !result.InjectedDeltaEvidenceMatched || !result.Stage2OperandMatchedAllArms || !result.FastActiveInputRowsMatched ||
		result.Stage2OperandDecodedSHA256 != perfmeasure.Fingerprint(c) || result.DecompositionToleranceFails != 0 {
		t.Fatalf("direct-mul propagation did not validate its matched C decomposition: %+v", result)
	}
}

func TestFixedSigmaDirectMulTransferScreenAppliesFrozenTenPercentBand(t *testing.T) {
	metric := func(rms float64) fixedSigmaVectorAnalysis {
		return fixedSigmaVectorAnalysis{Statistics: numericalmetrics.ComplexVectorStats{RMS: rms, Status: "FINITE"}}
	}
	makePair := func(seed int64, onOnStage1, onOnStage2 float64) fixedSigmaAblationSeedPairResult {
		standard := fixedSigmaAblationArmAnalysis{Stage1Incremental: metric(1), Stage2Incremental: metric(2)}
		arms := make(map[string]fixedSigmaAblationArmAnalysis, 4)
		for _, arm := range []string{"OFF/OFF", "ON/OFF", "OFF/ON"} {
			arms[arm] = fixedSigmaAblationArmAnalysis{Stage1Incremental: metric(1.02), Stage2Incremental: metric(1.98)}
		}
		arms["ON/ON"] = fixedSigmaAblationArmAnalysis{Stage1Incremental: metric(onOnStage1), Stage2Incremental: metric(onOnStage2)}
		return fixedSigmaAblationSeedPairResult{Stage1Seed: seed, Stage2Seed: seed + 1, Standard: standard, Arms: arms}
	}
	workload := fixedSigmaAblationWorkloadResult{SeedPairs: []fixedSigmaAblationSeedPairResult{
		makePair(20261101, 1.05, 2.10), makePair(20261103, 0.91, 1.82),
	}}
	screen := computeFixedSigmaDirectMulTransferScreen(workload)
	if !screen.AllOnOnWithinBand || len(screen.SeedPairs) != 2 || !screen.SeedPairs[0].Arms["ON/ON"].Stage1.WithinTenPercentBand {
		t.Fatalf("within-band ON/ON arms did not pass the descriptive screen: %+v", screen)
	}
	workload.SeedPairs[1].Arms["ON/ON"] = fixedSigmaAblationArmAnalysis{Stage1Incremental: metric(1.11), Stage2Incremental: metric(2)}
	if screen = computeFixedSigmaDirectMulTransferScreen(workload); screen.AllOnOnWithinBand {
		t.Fatalf("an ON/ON stage outside the frozen ±10%% screen was accepted: %+v", screen)
	}
}

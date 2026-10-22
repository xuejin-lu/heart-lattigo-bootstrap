package main

import (
	"path/filepath"
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

func TestFixedSigmaDirectMulAuthorizedF9F13BaselineReuseEvidence(t *testing.T) {
	captures := filepath.Join("..", "..", "results", "FIXED-SIGMA-DIRECT-MUL-004", "captures")
	load := func(stem string) loadedFixedSigmaAblationArtifact {
		t.Helper()
		artifact, err := loadFixedSigmaAblationArtifact(fixedSigmaAblationArtifactRef{
			Result: filepath.Join(captures, stem+".result.json"), Vectors: filepath.Join(captures, stem+".vectors.json"),
		})
		if err != nil {
			t.Fatalf("load %s evidence: %v", stem, err)
		}
		return artifact
	}
	f9 := load("fast-direct-c4-s1-off-off")
	f13 := load("fast-direct-c4-s2-off-off")
	reuse := fixedSigmaAblationBaselineReuse{
		OriginSlot: "F13", OriginStage1Seed: 20261107, OriginStage2Seed: 20261108,
		WitnessSlot: "F9",
	}
	evidence, err := validateFixedSigmaDirectMulF9F13BaselineReuse(f13, f9, "direct-c4", 20261105, reuse)
	if err != nil {
		t.Fatalf("reviewed F9/F13 OFF/OFF baseline equivalence was rejected: %v", err)
	}
	if evidence.OriginSlot != "F13" || evidence.OriginStageSeeds != [2]int64{20261107, 20261108} ||
		evidence.RequestedStageSeeds != [2]int64{20261105, 20261106} ||
		evidence.OriginBinarySHA256 != fixedSigmaDirectMulF13BinarySHA256 ||
		evidence.WitnessSlot != "F9" || evidence.WitnessStageSeeds != [2]int64{20261105, 20261106} ||
		evidence.WitnessBinarySHA256 != fixedSigmaDirectMulF9BinarySHA256 || evidence.NoiseInjectionCount != 0 ||
		!evidence.DecodedVectorDocumentsMatch || !evidence.CheckpointStatesMatch || !evidence.FastInputRowsMatch {
		t.Fatalf("baseline reuse output omits true origin or equality evidence: %+v", evidence)
	}
}

func TestFixedSigmaDirectMulBaselineReuseFailsClosed(t *testing.T) {
	captures := filepath.Join("..", "..", "results", "FIXED-SIGMA-DIRECT-MUL-004", "captures")
	load := func(stem string) loadedFixedSigmaAblationArtifact {
		t.Helper()
		artifact, err := loadFixedSigmaAblationArtifact(fixedSigmaAblationArtifactRef{
			Result: filepath.Join(captures, stem+".result.json"), Vectors: filepath.Join(captures, stem+".vectors.json"),
		})
		if err != nil {
			t.Fatalf("load %s evidence: %v", stem, err)
		}
		return artifact
	}
	base := load("fast-direct-c4-s2-off-off")
	witness := load("fast-direct-c4-s1-off-off")
	validReuse := fixedSigmaAblationBaselineReuse{
		OriginSlot: "F13", OriginStage1Seed: 20261107, OriginStage2Seed: 20261108,
		WitnessSlot: "F9",
	}
	check := func(name string, candidateBase, candidateWitness loadedFixedSigmaAblationArtifact, variant string, seed int64, reuse fixedSigmaAblationBaselineReuse) {
		t.Helper()
		if _, err := validateFixedSigmaDirectMulF9F13BaselineReuse(candidateBase, candidateWitness, variant, seed, reuse); err == nil {
			t.Errorf("baseline reuse accepted %s mismatch", name)
		}
	}

	wrongSeed := validReuse
	check("target seed", base, witness, "direct-c4", 20261103, wrongSeed)
	wrongOrigin := validReuse
	wrongOrigin.OriginSlot = "F9"
	check("origin slot", base, witness, "direct-c4", 20261105, wrongOrigin)
	check("workload variant", base, witness, "direct-base", 20261105, validReuse)

	noiseWitness := witness
	noiseChain := *witness.doc.FixedSigmaChain
	noiseChain.LogicalNoiseInjections = 1
	noiseWitness.doc.FixedSigmaChain = &noiseChain
	check("nonzero injection count", base, noiseWitness, "direct-c4", 20261105, validReuse)

	rowWitness := witness
	rowChain := *witness.doc.FixedSigmaChain
	rowChain.FastInputRowsSHA256 = make(map[string]string, len(witness.doc.FixedSigmaChain.FastInputRowsSHA256))
	for key, value := range witness.doc.FixedSigmaChain.FastInputRowsSHA256 {
		rowChain.FastInputRowsSHA256[key] = value
	}
	rowChain.FastInputRowsSHA256["A.c0.q0"] = "mismatch"
	rowWitness.doc.FixedSigmaChain = &rowChain
	check("input active-row hash", base, rowWitness, "direct-c4", 20261105, validReuse)

	checkpointWitness := witness
	checkpointWitness.doc.Checkpoints = append([]publicCheckpoint(nil), witness.doc.Checkpoints...)
	checkpointWitness.doc.Checkpoints[0].State.Level++
	check("checkpoint state", base, checkpointWitness, "direct-c4", 20261105, validReuse)

	vectorWitness := witness
	vectorWitness.vectors.Checkpoints = map[string][]complexValue{}
	check("decoded vectors", base, vectorWitness, "direct-c4", 20261105, validReuse)
}

func TestFixedSigmaAblationRunCountDeduplicatesReusedBaselineAndCountsWitness(t *testing.T) {
	ids := make(map[string]struct{})
	firstPair := fixedSigmaAblationSeedPairResult{
		ArtifactSHA256: map[string]map[string]string{
			"standard": {"result": "standard-result", "vectors": "standard-vectors"},
			"OFF/OFF":  {"result": "f13-result", "vectors": "f13-vectors"},
			"ON/OFF":   {"result": "f10-result", "vectors": "f10-vectors"},
			"OFF/ON":   {"result": "f11-result", "vectors": "f11-vectors"},
			"ON/ON":    {"result": "f12-result", "vectors": "f12-vectors"},
		},
		BaselineReuse: &fixedSigmaAblationBaselineReuseEvidence{
			WitnessResultSHA256: "f9-result", WitnessVectorsSHA256: "f9-vectors",
		},
	}
	secondPair := fixedSigmaAblationSeedPairResult{
		ArtifactSHA256: map[string]map[string]string{
			"standard": {"result": "standard-result", "vectors": "standard-vectors"},
			"OFF/OFF":  {"result": "f13-result", "vectors": "f13-vectors"},
			"ON/OFF":   {"result": "f14-result", "vectors": "f14-vectors"},
			"OFF/ON":   {"result": "f15-result", "vectors": "f15-vectors"},
			"ON/ON":    {"result": "f16-result", "vectors": "f16-vectors"},
		},
	}
	addFixedSigmaAblationFastArtifactIDs(ids, firstPair)
	addFixedSigmaAblationFastArtifactIDs(ids, secondPair)
	if len(ids) != 8 {
		t.Fatalf("expected four F10-F13 arms, the F9 equivalence witness, and four F13-F16 arms deduplicated at F13; got %d unique captures", len(ids))
	}
}

package main

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const (
	fixedSigmaDirectMulSchema       = "fast-standard-fixed-sigma-direct-ctct.v1"
	fixedSigmaDirectMulVectorSchema = "fast-standard-fixed-sigma-direct-ctct-vectors.v1"
	fixedSigmaDirectMulChainSchema  = "fixed-sigma-direct-ctct-chain.v1"
)

var fixedSigmaDirectMulCheckpointNames = []string{
	"encrypt_a_level5", "encrypt_b_q5_level5", "chain_stage1_mulrelin_level5",
	"chain_stage1_output_level4", "encrypt_c_q4_level4", "chain_stage2_mulrelin_level4",
	"chain_stage2_output_level3",
}

func validatePublicFixedSigmaDirectMulArtifact(doc publicNativeDocument, vectors publicVectorDocument, backend string) error {
	if doc.SchemaVersion != fixedSigmaDirectMulSchema || doc.Mode != "fixed-sigma-direct-ctct" ||
		doc.Status != "PASS_ZERO_BOOTSTRAP_DIRECT_CTCT_CHAIN" || doc.Backend != backend || doc.BuildTag != "perf_"+backend ||
		doc.BootstrapBudget != 0 || doc.BootstrapCalls != 0 || len(doc.BootstrapResults) != 0 ||
		doc.EphemeralSecretWeight != 32 || doc.InputSlots != 1<<12 || doc.Profile != "logn13" ||
		doc.ProfileKind != publicProfileKindFrozen || doc.CapacityPolicy != publicCapacityPolicyStrict ||
		doc.FixedSigmaChain == nil || doc.FixedSigmaChain.SchemaVersion != fixedSigmaDirectMulChainSchema ||
		doc.FixedSigmaChain.WorkloadID != perfmeasure.PublicDirectMulWorkloadID ||
		doc.FixedSigmaChain.Classification != "DIAGNOSTIC_ONLY" || doc.FixedSigmaChain.LogicalCtCtOperations != 2 ||
		len(doc.FixedSigmaChain.Operations) != 2 || !validSHA256Hex(doc.BinarySHA256) || !validSHA256Hex(doc.PrimarySourceSHA256) {
		return errors.New("direct-ctct artifact provenance, scope, or operation count is incomplete")
	}
	if doc.ConfigSHA256 != publicConfigSHA || doc.QPSHA256 != publicQPSHA ||
		doc.CapacityEvidenceVersion != "public-direct-mul-qprefix-capacity.v1" ||
		doc.ObservedNativeOracleStatus != "PASS_SEVEN_DIRECT_CTCT_CHECKPOINT_PLAINTEXT_ORACLES" ||
		doc.NumericalGate != publicNumericalGate || !doc.PrimaryClean || !doc.BackendClean ||
		doc.FixedSigmaChain.PrimaryBuildCommit != doc.PrimaryCommit || !doc.FixedSigmaChain.PrimaryBuildClean {
		return errors.New("direct-ctct artifact does not bind to the clean LogN13/E32 public profile")
	}
	if !directMulArtifactHasNoAddPhase(doc) {
		return errors.New("direct-ctct execution phases contain a forbidden Add operation")
	}
	profile, err := publicProfileForNativeDocument(doc)
	if err != nil || profile.Name != "logn13" || profile.Kind != publicProfileKindFrozen ||
		doc.Parameters.LogN != 13 || doc.Parameters.LogSlots != 12 || doc.Parameters.RingN != 1<<13 ||
		doc.Parameters.InputSlots != 1<<12 || doc.Parameters.EphemeralSecretWeight != 32 ||
		doc.Parameters.DefaultScale != fixedSigmaChainScaleText(new(big.Int).Lsh(big.NewInt(1), 45)) {
		return errors.New("direct-ctct effective parameters differ from the frozen LogN13/E32/Δ profile")
	}
	if err := validatePublicFixedSigmaChainBackendPin(backend, doc.BackendCommit, profile.Name); err != nil {
		return err
	}
	if backend == "standard" && doc.FixedSigmaChain.AblationArm != "OFF/OFF" {
		return errors.New("genuine Standard direct-ctct reference must be OFF/OFF")
	}
	if _, _, err := fixedSigmaAblationInjectionStages(doc.FixedSigmaChain.AblationArm); err != nil {
		return err
	}
	if doc.FixedSigmaChain.CVariant != "direct-base" && doc.FixedSigmaChain.CVariant != "direct-c4" {
		return errors.New("direct-ctct artifact has an unsupported workload variant")
	}
	if doc.FixedSigmaChain.InputRandomnessStatus == "" {
		return errors.New("direct-ctct input randomness provenance is missing")
	}
	if backend == "fast" {
		if doc.FixedSigmaChain.FastC1Invariant == "" || doc.FixedSigmaChain.FastDormantRowPolicy == "" || doc.FixedSigmaChain.FastMetadataInvariant == "" {
			return errors.New("Fast direct-ctct active-row, c1, dormant-row and metadata evidence is incomplete")
		}
	} else if len(doc.FixedSigmaChain.FastInputRowsSHA256) != 0 || doc.FixedSigmaChain.FastC1Invariant != "" ||
		doc.FixedSigmaChain.FastDormantRowPolicy != "" || doc.FixedSigmaChain.FastMetadataInvariant != "" {
		return errors.New("Standard direct-ctct artifact must not claim Fast-specific row evidence")
	}

	if vectors.SchemaVersion != fixedSigmaDirectMulVectorSchema || vectors.Mode != doc.Mode ||
		vectors.Backend != backend || vectors.BackendCommit != doc.BackendCommit || vectors.PrimaryCommit != doc.PrimaryCommit ||
		vectors.SourceSHA256 != doc.PrimarySourceSHA256 || vectors.ConfigSHA256 != doc.ConfigSHA256 ||
		vectors.QPSHA256 != doc.QPSHA256 || vectors.InputSHA256 != doc.InputSHA256 ||
		vectors.WorkloadSHA256 != doc.WorkloadSHA256 || vectors.E != 32 ||
		vectors.FixedSigmaChain == nil || len(vectors.FixedSigmaChain.Stages) != 2 || len(vectors.FixedSigmaChain.Inputs) != 3 {
		return errors.New("direct-ctct vector artifact has missing or mismatched provenance")
	}
	if !samePublicCheckpointNames(doc.Checkpoints, fixedSigmaDirectMulCheckpointNames) ||
		len(vectors.Checkpoints) != len(fixedSigmaDirectMulCheckpointNames) {
		return errors.New("direct-ctct checkpoint coverage includes a missing or extra operation boundary")
	}
	inputs := make(map[string][]complex128, 3)
	for _, name := range []string{"A", "B", "C"} {
		encoded, ok := vectors.FixedSigmaChain.Inputs[name]
		if !ok || len(encoded) != doc.InputSlots {
			return fmt.Errorf("direct-ctct original input %s is missing or has the wrong slot count", name)
		}
		values := decodeComplexValues(encoded)
		if !allFinite(values) || perfmeasure.Fingerprint(values) != doc.FixedSigmaChain.InputMessageSHA256[name] {
			return fmt.Errorf("direct-ctct original input %s fingerprint is invalid", name)
		}
		inputs[name] = values
	}
	if len(doc.FixedSigmaChain.InputMessageSHA256) != 3 {
		return errors.New("direct-ctct message fingerprint map must contain exactly A, B and C")
	}
	wantA, wantB, wantC, err := perfmeasure.PublicDirectMulWorkload(doc.InputSlots, doc.FixedSigmaChain.CVariant)
	if err != nil || !complexVectorsEqual(inputs["A"], wantA) || !complexVectorsEqual(inputs["B"], wantB) || !complexVectorsEqual(inputs["C"], wantC) {
		return errors.New("direct-ctct artifact messages differ from the frozen direct-base/direct-c4 inputs")
	}
	inputHash, err := perfmeasure.PublicDirectMulInputsFingerprint(inputs["A"], inputs["B"], inputs["C"])
	if err != nil || inputHash != doc.InputSHA256 {
		return errors.New("direct-ctct combined input identity does not recompute from exact A/B/C messages")
	}
	q, err := directMulEffectivePrimes(doc.Parameters.QPrimes)
	if err != nil || len(q) <= 5 || len(doc.Parameters.PPrimes) == 0 {
		return errors.New("direct-ctct effective Q/P chain is incomplete")
	}
	p, err := directMulEffectivePrimes(doc.Parameters.PPrimes)
	if err != nil || perfmeasure.QPFingerprint(q, p) != doc.QPSHA256 {
		return errors.New("direct-ctct effective Q/P fingerprint does not recompute")
	}
	q4, q5 := new(big.Int).SetUint64(q[4]), new(big.Int).SetUint64(q[5])
	capacity, err := perfmeasure.BoundPublicDirectMulChain(inputs["A"], inputs["B"], inputs["C"], q, doc.Parameters.RingN)
	if err != nil || doc.DirectMulCapacity == nil || !reflect.DeepEqual(capacity, *doc.DirectMulCapacity) || capacity.Status != "STRICT_QPREFIX_FIT" ||
		doc.ConservativeBoundStatus != capacity.Status {
		return errors.New("direct-ctct two-stage strict Q-prefix capacity plan does not recompute")
	}
	workloadHash, err := perfmeasure.PublicDirectMulWorkloadFingerprint(
		inputs["A"], inputs["B"], inputs["C"], doc.FixedSigmaChain.CVariant,
		new(big.Int).Lsh(big.NewInt(1), 45).String(), q5.String(), q4.String(), doc.QPSHA256,
	)
	if err != nil || workloadHash != doc.WorkloadSHA256 {
		return errors.New("direct-ctct workload identity does not recompute from messages, schedule and effective Q/P")
	}

	wantStage1Oracle := publicMul(inputs["A"], inputs["B"])
	wantStage2Oracle := publicMul(wantStage1Oracle, inputs["C"])
	wantCheckpointOracles := map[string][]complex128{
		"encrypt_a_level5": inputs["A"], "encrypt_b_q5_level5": inputs["B"],
		"chain_stage1_mulrelin_level5": wantStage1Oracle, "chain_stage1_output_level4": wantStage1Oracle,
		"encrypt_c_q4_level4": inputs["C"], "chain_stage2_mulrelin_level4": wantStage2Oracle,
		"chain_stage2_output_level3": wantStage2Oracle,
	}
	byName := make(map[string]publicCheckpoint, len(doc.Checkpoints))
	for _, checkpoint := range doc.Checkpoints {
		encoded, ok := vectors.Checkpoints[checkpoint.Name]
		if !ok || len(encoded) != doc.InputSlots || !checkpoint.OraclePass || checkpoint.Oracle.MaxComplexDifference > publicNumericalGate ||
			checkpoint.DecodedSHA256 != perfmeasure.Fingerprint(decodeComplexValues(encoded)) {
			return fmt.Errorf("direct-ctct checkpoint %s is missing or failed its plaintext oracle", checkpoint.Name)
		}
		if checkpoint.CiphertextDomain != "NTT_NON_MONTGOMERY" {
			return fmt.Errorf("direct-ctct checkpoint %s representation is %q; want NTT non-Montgomery", checkpoint.Name, checkpoint.CiphertextDomain)
		}
		wantCompact := backend == "fast" && (checkpoint.Name == "chain_stage1_mulrelin_level5" ||
			checkpoint.Name == "chain_stage1_output_level4" || checkpoint.Name == "chain_stage2_mulrelin_level4" ||
			checkpoint.Name == "chain_stage2_output_level3")
		if checkpoint.FastCompactPrefix != wantCompact || len(checkpoint.PhysicalRowLengths) != 2 || len(checkpoint.RowSHA256) != 2 {
			return fmt.Errorf("direct-ctct checkpoint %s physical row evidence does not match backend layout", checkpoint.Name)
		}
		for component := 0; component < 2; component++ {
			if len(checkpoint.RowSHA256[component]) != 4 || len(checkpoint.PhysicalRowLengths[component]) < 4 {
				return fmt.Errorf("direct-ctct checkpoint %s component %d lacks four active row hashes or physical row lengths", checkpoint.Name, component)
			}
			for qRow := 0; qRow < 4; qRow++ {
				if checkpoint.PhysicalRowLengths[component][qRow] != doc.Parameters.RingN || !validSHA256Hex(checkpoint.RowSHA256[component][qRow]) {
					return fmt.Errorf("direct-ctct checkpoint %s component %d q%d active row evidence is invalid", checkpoint.Name, component, qRow)
				}
			}
			if wantCompact {
				for qRow := 4; qRow < len(checkpoint.PhysicalRowLengths[component]); qRow++ {
					if checkpoint.PhysicalRowLengths[component][qRow] != 0 {
						return fmt.Errorf("direct-ctct Fast compact checkpoint %s retains dormant component %d q%d coefficients", checkpoint.Name, component, qRow)
					}
				}
			}
		}
		if backend == "standard" {
			if !checkpoint.C1Nonzero || checkpoint.FastCompactPrefix || checkpoint.DecodePath != "rlwe.NewDecryptor(...).DecryptNew -> ckks.NewEncoder.Decode" {
				return fmt.Errorf("Standard direct-ctct checkpoint %s lacks native decrypt provenance", checkpoint.Name)
			}
		} else if checkpoint.C1Nonzero || checkpoint.DecodePath != "Fast c0 projection to the authoritative Q-prefix plaintext (zero-secret measurement adapter)" {
			return fmt.Errorf("Fast direct-ctct checkpoint %s violates zero-secret Q-prefix decode provenance", checkpoint.Name)
		}
		byName[checkpoint.Name] = checkpoint
		actual := decodeComplexValues(encoded)
		metrics := compareVectors(wantCheckpointOracles[checkpoint.Name], actual, publicNumericalGate)
		if metrics != checkpoint.Oracle {
			return fmt.Errorf("direct-ctct checkpoint %s oracle metrics do not recompute", checkpoint.Name)
		}
	}
	if backend == "standard" && !doc.InputC1Nonzero || backend == "fast" && doc.InputC1Nonzero {
		return errors.New("direct-ctct input c1 provenance does not match Standard/Fast mode")
	}
	wantLevels := map[string]int{
		"encrypt_a_level5": 5, "encrypt_b_q5_level5": 5, "chain_stage1_mulrelin_level5": 5,
		"chain_stage1_output_level4": 4, "encrypt_c_q4_level4": 4,
		"chain_stage2_mulrelin_level4": 4, "chain_stage2_output_level3": 3,
	}
	deltaScale := fixedSigmaChainScaleText(new(big.Int).Lsh(big.NewInt(1), 45))
	wantScales := map[string]string{
		"encrypt_a_level5": deltaScale, "encrypt_b_q5_level5": fixedSigmaChainScaleText(q5),
		"chain_stage1_mulrelin_level5": fixedSigmaChainScaleText(new(big.Int).Mul(new(big.Int).Lsh(big.NewInt(1), 45), q5)),
		"chain_stage1_output_level4":   deltaScale, "encrypt_c_q4_level4": fixedSigmaChainScaleText(q4),
		"chain_stage2_mulrelin_level4": fixedSigmaChainScaleText(new(big.Int).Mul(new(big.Int).Lsh(big.NewInt(1), 45), q4)),
		"chain_stage2_output_level3":   deltaScale,
	}
	for name, level := range wantLevels {
		checkpoint := byName[name]
		if checkpoint.State.Level != level || checkpoint.State.QPrefixRows != 4 || checkpoint.State.Degree != 1 || checkpoint.State.Scale != wantScales[name] {
			return fmt.Errorf("direct-ctct checkpoint %s state is Level%d/rows%d/Scale%s; want Level%d/rows4/Scale%s", name,
				checkpoint.State.Level, checkpoint.State.QPrefixRows, checkpoint.State.Scale, level, wantScales[name])
		}
	}
	if len(doc.FixedSigmaChain.Operations) != 2 || doc.FixedSigmaChain.Operations[0].RescaleDivisor != q5.String() ||
		doc.FixedSigmaChain.Operations[1].RescaleDivisor != q4.String() {
		return errors.New("direct-ctct Rescale divisors do not match effective q5 and q4")
	}
	wantRefs := [][3]string{
		{"encrypt_a_level5", "encrypt_b_q5_level5", "chain_stage1_output_level4"},
		{"chain_stage1_output_level4", "encrypt_c_q4_level4", "chain_stage2_output_level3"},
	}
	productNames := []string{"chain_stage1_mulrelin_level5", "chain_stage2_mulrelin_level4"}
	for index, operation := range doc.FixedSigmaChain.Operations {
		stageName := fmt.Sprintf("stage%d", index+1)
		refs := wantRefs[index]
		productName := productNames[index]
		if operation.Name != stageName || operation.InputXCheckpoint != refs[0] || operation.InputYCheckpoint != refs[1] ||
			operation.ProductCheckpoint != productName || operation.OutputCheckpoint != refs[2] ||
			!reflect.DeepEqual(operation.InputXState, byName[refs[0]].State) || !reflect.DeepEqual(operation.InputYState, byName[refs[1]].State) ||
			!reflect.DeepEqual(operation.OutputState, byName[refs[2]].State) || operation.MulRelinElapsedNS <= 0 || operation.RescaleElapsedNS <= 0 ||
			operation.Residual.ElapsedNS != operation.MulRelinElapsedNS+operation.RescaleElapsedNS || operation.Residual.LogicalCtCtMulCount != 1 ||
			operation.Residual.EncryptionMode != "sk" || operation.Residual.PKStatus != "UNSUPPORTED_FAST_ZERO_SECRET_MULRELIN_REQUIRES_C1_ZERO" ||
			operation.Residual.SchemaVersion != "fixed-sigma-mul-relin-rescale-residual.v1" || operation.Residual.Classification != "DIAGNOSTIC_ONLY" ||
			operation.Residual.DispatchEvidence != fixedSigmaChainDispatchEvidence(backend) || operation.MultiplicationScale != wantScales[productName] {
			return fmt.Errorf("direct-ctct %s call, state, or dispatch evidence is invalid", stageName)
		}
		wantStage1, wantStage2, _ := fixedSigmaAblationInjectionStages(doc.FixedSigmaChain.AblationArm)
		wantInjection := backend == "fast" && (index == 0 && wantStage1 || index == 1 && wantStage2)
		if (operation.Residual.LogicalNoiseInjections == 1) != wantInjection {
			return fmt.Errorf("direct-ctct %s injection count does not match ablation arm", stageName)
		}
		if wantInjection && (operation.Residual.SamplingSeed == nil || *operation.Residual.SamplingSeed != doc.FixedSigmaChain.StageSeeds[index]) {
			return fmt.Errorf("direct-ctct %s sampler seed does not match its frozen stage seed", stageName)
		}
		stageVectors, ok := vectors.FixedSigmaChain.Stages[stageName]
		if !ok || len(stageVectors.IncrementalResidual) != doc.InputSlots || len(stageVectors.TotalError) != doc.InputSlots ||
			len(stageVectors.MulRelinOracle) != doc.InputSlots || len(stageVectors.PlaintextOracle) != doc.InputSlots {
			return fmt.Errorf("direct-ctct %s vectors are incomplete", stageName)
		}
		inputX := decodeComplexValues(vectors.Checkpoints[refs[0]])
		inputY := decodeComplexValues(vectors.Checkpoints[refs[1]])
		output := decodeComplexValues(vectors.Checkpoints[refs[2]])
		incremental := subtractComplexVectors(output, multiplyComplexVectors(inputX, inputY))
		total := subtractComplexVectors(output, decodeComplexValues(stageVectors.PlaintextOracle))
		stageOracle := wantStage1Oracle
		wantOutputOracle := wantStage1Oracle
		if index == 1 {
			stageOracle, wantOutputOracle = wantStage2Oracle, wantStage2Oracle
		}
		if !complexVectorsEqual(incremental, decodeComplexValues(stageVectors.IncrementalResidual)) ||
			!complexVectorsEqual(total, decodeComplexValues(stageVectors.TotalError)) ||
			!complexVectorsEqual(decodeComplexValues(stageVectors.MulRelinOracle), stageOracle) ||
			!complexVectorsEqual(decodeComplexValues(stageVectors.PlaintextOracle), wantOutputOracle) {
			return fmt.Errorf("direct-ctct %s incremental or total error vectors do not recompute", stageName)
		}
		residualStats := numericalmetrics.SummarizeComplexVector(incremental)
		totalStats := numericalmetrics.SummarizeComplexVector(total)
		if !sameComplexVectorStats(residualStats, operation.Residual.Statistics) || !sameComplexVectorStats(totalStats, operation.TotalError) ||
			operation.Residual.InputXSHA256 != perfmeasure.Fingerprint(inputX) || operation.Residual.InputYSHA256 != perfmeasure.Fingerprint(inputY) ||
			operation.Residual.OutputSHA256 != perfmeasure.Fingerprint(output) {
			return fmt.Errorf("direct-ctct %s residual statistics or decoded input hashes do not recompute", stageName)
		}
		if err := validateDirectMulInjectionEvidence(backend, operation, stageVectors, inputX, inputY, output, capacity.Q0123, doc.InputSlots); err != nil {
			return fmt.Errorf("direct-ctct %s injection: %w", stageName, err)
		}
	}
	if backend == "fast" {
		wantRows := make(map[string]bool)
		for _, name := range []string{"A", "B", "C_q4"} {
			for component := 0; component < 2; component++ {
				for qRow := 0; qRow < 4; qRow++ {
					wantRows[fmt.Sprintf("%s.c%d.q%d", name, component, qRow)] = true
				}
			}
		}
		if len(doc.FixedSigmaChain.FastInputRowsSHA256) != len(wantRows) {
			return errors.New("Fast direct-ctct active input-row hashes must cover only A/B/C q0..q3")
		}
		for name := range wantRows {
			if !validSHA256Hex(doc.FixedSigmaChain.FastInputRowsSHA256[name]) {
				return fmt.Errorf("Fast direct-ctct active input row %s is missing or invalid", name)
			}
		}
		inputCheckpoints := map[string]string{
			"A": "encrypt_a_level5", "B": "encrypt_b_q5_level5", "C_q4": "encrypt_c_q4_level4",
		}
		for input, checkpointName := range inputCheckpoints {
			checkpoint := byName[checkpointName]
			for component := 0; component < 2; component++ {
				for qRow := 0; qRow < 4; qRow++ {
					rowKey := fmt.Sprintf("%s.c%d.q%d", input, component, qRow)
					if doc.FixedSigmaChain.FastInputRowsSHA256[rowKey] != checkpoint.RowSHA256[component][qRow] {
						return fmt.Errorf("Fast direct-ctct input row %s does not match its checkpoint", rowKey)
					}
				}
			}
		}
		if doc.FixedSigmaChain.FastC1Invariant != "PASS: Fast EncryptNew inputs and all captured checkpoints have c1=0; each enabled injection hashes c1 q0..q3 before and after commit." ||
			doc.FixedSigmaChain.FastDormantRowPolicy == "" || doc.FixedSigmaChain.FastMetadataInvariant == "" {
			return errors.New("Fast direct-ctct c1/dormant-row/metadata invariants were not recorded")
		}
	}
	seeds := doc.FixedSigmaChain.StageSeeds
	if seeds[0] == int64(^uint64(0)>>1) || seeds[1] != seeds[0]+1 ||
		doc.FixedSigmaChain.Stage2SeedDerivation != "stage2 seed is stage1 seed + 1; OFF stages do not invoke the sampler" {
		return errors.New("direct-ctct stage seeds are not the declared distinct deterministic pair")
	}
	wantStage1, wantStage2, _ := fixedSigmaAblationInjectionStages(doc.FixedSigmaChain.AblationArm)
	wantCount := 0
	if backend == "fast" {
		if wantStage1 {
			wantCount++
		}
		if wantStage2 {
			wantCount++
		}
	}
	if doc.FixedSigmaChain.LogicalNoiseInjections != wantCount {
		return errors.New("direct-ctct total injection count does not match the selected arm")
	}
	if wantCount == 0 {
		if doc.FixedSigmaChain.SigmaCoefficient != nil {
			return errors.New("direct-ctct zero-injection arm reports a sigma")
		}
	} else if doc.FixedSigmaChain.SigmaCoefficient == nil || *doc.FixedSigmaChain.SigmaCoefficient != fixedSigmaChainCoefficientSigma {
		return errors.New("direct-ctct injection does not use the frozen coefficient sigma")
	}
	return nil
}

func validateDirectMulInjectionEvidence(backend string, operation fixedSigmaChainOperation, vectors fixedSigmaChainStageVectors, inputX, inputY, output []complex128, qPrefixProduct string, slots int) error {
	injected := operation.Residual.LogicalNoiseInjections == 1
	if !injected {
		if operation.Residual.FixedSigmaNoise != nil || operation.PreNoiseResidual != nil || operation.DecodedNoiseDelta != nil ||
			len(vectors.PreNoiseOutput) != 0 || len(vectors.DecodedNoiseDelta) != 0 || operation.Residual.SamplingSeed != nil {
			return errors.New("OFF stage contains sampler or noise evidence")
		}
		return nil
	}
	if backend != "fast" || operation.Residual.FixedSigmaNoise == nil ||
		operation.Residual.FixedSigmaNoise.SigmaCoefficient != fixedSigmaChainCoefficientSigma ||
		operation.Residual.SamplingSeed == nil || operation.PreNoiseResidual == nil || operation.DecodedNoiseDelta == nil ||
		operation.Residual.DecodedNoiseDelta == nil || operation.Residual.PreNoiseStatistics == nil ||
		len(vectors.PreNoiseOutput) != slots || len(vectors.DecodedNoiseDelta) != slots {
		return errors.New("ON stage does not contain one complete Fast-only frozen-sigma injection")
	}
	evidence := operation.Residual.FixedSigmaNoise
	if evidence.SamplingSeed != *operation.Residual.SamplingSeed || evidence.CoefficientSampleCount != 1<<13 ||
		evidence.QPrefixRows != 4 || evidence.QPrefixCheckedCoefficients != 1<<13 || !evidence.QPrefixStrictFit ||
		evidence.QPrefixPrefixProduct != qPrefixProduct || !validSHA256Hex(evidence.CoefficientSHA256) ||
		!validFixedSigmaNoisePhaseMeasurements(evidence.PhaseMeasurements) || evidence.DispatchEvidence == "" {
		return errors.New("ON stage sampler or transactional Q-prefix capacity evidence is invalid")
	}
	preNoise := decodeComplexValues(vectors.PreNoiseOutput)
	noiseDelta := decodeComplexValues(vectors.DecodedNoiseDelta)
	preResidual := preResidualVector(inputX, inputY, preNoise)
	wantDelta := subtractComplexVectors(output, preNoise)
	deltaStats := numericalmetrics.SummarizeComplexVector(noiseDelta)
	preResidualStats := numericalmetrics.SummarizeComplexVector(preResidual)
	if !allFinite(preNoise) || !allFinite(noiseDelta) || !complexVectorsEqual(noiseDelta, wantDelta) ||
		!sameComplexVectorStats(deltaStats, *operation.DecodedNoiseDelta) ||
		!sameComplexVectorStats(deltaStats, *operation.Residual.DecodedNoiseDelta) ||
		!sameComplexVectorStats(preResidualStats, *operation.PreNoiseResidual) ||
		!sameComplexVectorStats(preResidualStats, *operation.Residual.PreNoiseStatistics) ||
		operation.Residual.PreNoiseOutputSHA256 != perfmeasure.Fingerprint(preNoise) {
		return errors.New("ON stage decoded pre-noise or injection-delta evidence does not recompute")
	}
	return nil
}

func preResidualVector(inputX, inputY, output []complex128) []complex128 {
	return subtractComplexVectors(output, multiplyComplexVectors(inputX, inputY))
}

func directMulEffectivePrimes(values []string) ([]uint64, error) {
	primes := make([]uint64, len(values))
	for index, value := range values {
		parsed, ok := new(big.Int).SetString(value, 10)
		if !ok || !parsed.IsUint64() || parsed.Sign() <= 0 || parsed.String() != value {
			return nil, fmt.Errorf("effective prime q/p[%d]=%q is not a canonical positive uint64", index, value)
		}
		primes[index] = parsed.Uint64()
	}
	return primes, nil
}

func validatePublicDirectMulPairArtifacts(standard, fast publicNativeDocument, standardVectors, fastVectors publicVectorDocument) error {
	if standard.Mode != "fixed-sigma-direct-ctct" || fast.Mode != standard.Mode ||
		standard.SchemaVersion != fixedSigmaDirectMulSchema || fast.SchemaVersion != standard.SchemaVersion ||
		standard.Status != "PASS_ZERO_BOOTSTRAP_DIRECT_CTCT_CHAIN" || fast.Status != standard.Status {
		return errors.New("paired artifacts are not successful direct-ctct chain runs")
	}
	if err := validatePublicFixedSigmaDirectMulArtifact(standard, standardVectors, "standard"); err != nil {
		return fmt.Errorf("Standard direct-ctct artifact: %w", err)
	}
	if err := validatePublicFixedSigmaDirectMulArtifact(fast, fastVectors, "fast"); err != nil {
		return fmt.Errorf("Fast direct-ctct artifact: %w", err)
	}
	if standard.FixedSigmaChain.AblationArm != "OFF/OFF" || standard.FixedSigmaChain.CVariant != fast.FixedSigmaChain.CVariant ||
		standard.EphemeralSecretWeight != 32 || fast.EphemeralSecretWeight != 32 || standard.BootstrapBudget != 0 || fast.BootstrapBudget != 0 ||
		standard.BootstrapCalls != 0 || fast.BootstrapCalls != 0 || standard.InputSlots != fast.InputSlots ||
		!standard.PrimaryClean || !fast.PrimaryClean || !standard.BackendClean || !fast.BackendClean ||
		standard.PrimaryCommit == "" || standard.PrimaryCommit != fast.PrimaryCommit ||
		standard.PrimarySourceSHA256 == "" || standard.PrimarySourceSHA256 != fast.PrimarySourceSHA256 ||
		standard.ConfigSHA256 != fast.ConfigSHA256 || standard.QPSHA256 != fast.QPSHA256 ||
		standard.InputSHA256 != fast.InputSHA256 || standard.WorkloadSHA256 != fast.WorkloadSHA256 ||
		!reflect.DeepEqual(standard.Parameters, fast.Parameters) || !reflect.DeepEqual(standard.DirectMulCapacity, fast.DirectMulCapacity) ||
		!reflect.DeepEqual(standard.FixedSigmaChain.InputMessageSHA256, fast.FixedSigmaChain.InputMessageSHA256) {
		return errors.New("Standard/Fast direct-ctct pair differs in fixture, source, inputs, effective parameters or strict capacity")
	}
	if !validSHA256Hex(standard.BinarySHA256) || !validSHA256Hex(fast.BinarySHA256) || standard.BinarySHA256 == fast.BinarySHA256 ||
		standard.BuildTag != "perf_standard" || fast.BuildTag != "perf_fast" {
		return errors.New("direct-ctct Standard/Fast binaries do not have distinct source-bound build provenance")
	}
	for _, name := range fixedSigmaDirectMulCheckpointNames {
		left, right := standard.Checkpoints[indexOfCheckpoint(standard.Checkpoints, name)], fast.Checkpoints[indexOfCheckpoint(fast.Checkpoints, name)]
		if left.Name != right.Name || left.State.Level != right.State.Level || left.State.Degree != right.State.Degree || left.State.Scale != right.State.Scale {
			return fmt.Errorf("paired direct-ctct checkpoint %s has mismatched Level/Scale/Degree", name)
		}
	}
	wantStage1, wantStage2, err := fixedSigmaAblationInjectionStages(fast.FixedSigmaChain.AblationArm)
	if err != nil {
		return err
	}
	wantCount := 0
	if wantStage1 {
		wantCount++
	}
	if wantStage2 {
		wantCount++
	}
	if fast.FixedSigmaChain.LogicalNoiseInjections != wantCount || standard.FixedSigmaChain.LogicalNoiseInjections != 0 {
		return errors.New("paired direct-ctct Fast injection count does not match its ablation arm")
	}
	return nil
}

func indexOfCheckpoint(checkpoints []publicCheckpoint, name string) int {
	for index := range checkpoints {
		if checkpoints[index].Name == name {
			return index
		}
	}
	return -1
}

func directMulWorkloadIDFromDocument(doc publicNativeDocument) bool {
	return doc.FixedSigmaChain != nil && doc.FixedSigmaChain.WorkloadID == perfmeasure.PublicDirectMulWorkloadID
}

func directMulArtifactHasNoAddPhase(doc publicNativeDocument) bool {
	for _, phase := range doc.ExecutionPhases {
		name := strings.ToLower(phase.Phase)
		if strings.Contains(name, "add_ab") || strings.Contains(name, "addnew") {
			return false
		}
	}
	return true
}

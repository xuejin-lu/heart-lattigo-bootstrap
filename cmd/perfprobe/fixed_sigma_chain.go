package main

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const fixedSigmaChainCoefficientSigma = 4.004280012097432

func runPublicFixedSigmaChain(
	opts cliOptions,
	configHash string,
	residual ckks.Parameters,
	params bootstrapping.Parameters,
	effective perfmeasure.EffectiveParameters,
	profile publicProfileContract,
) error {
	if opts.bootstrapBudget != 0 || opts.publicBootstrap || opts.publicRepeatability || opts.preflightPair != "" || opts.traceFixtureOut != "" {
		return errors.New("fixed-sigma two-step chain requires an explicit zero-Bootstrap public-native run")
	}
	if profile.Name != "logn13" || profile.Kind != publicProfileKindFrozen || configHash != profile.ConfigSHA256 ||
		opts.ephemeralSecretWeight != 32 || effective.InputSlots != 1<<12 || effective.RingN != 1<<13 {
		return errors.New("fixed-sigma two-step chain is diagnostic-only for the accepted LogN13/E32 profile")
	}
	if opts.fixedSigmaNoise && opts.fixedSigmaNoiseSigma != fixedSigmaChainCoefficientSigma {
		return fmt.Errorf("fixed-sigma two-step chain freezes sigma at %.15g", fixedSigmaChainCoefficientSigma)
	}
	if err := verifyCompiledBackendSource(opts.secondaryRoot); err != nil {
		return err
	}
	primaryRoot, primaryCommit, err := cleanRepositoryState(".")
	if err != nil {
		return fmt.Errorf("Primary provenance: %w", err)
	}
	primaryBuildCommit, primaryBuildClean, err := fixedSigmaChainBinaryProvenance(primaryCommit)
	if err != nil {
		return fmt.Errorf("Primary binary provenance: %w", err)
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	binarySHA256, err := sha256File(executable)
	if err != nil {
		return fmt.Errorf("hash perfprobe executable: %w", err)
	}
	secondaryRoot, secondaryCommit, secondaryRef, err := cleanSecondaryState(opts.secondaryRoot, opts.backendCommit)
	if err != nil {
		return fmt.Errorf("backend provenance: %w", err)
	}
	if err := validatePublicFixedSigmaChainBackendPin(publicBackendName(), secondaryCommit, profile.Name); err != nil {
		return fmt.Errorf("UNSUPPORTED_BACKEND_PROFILE: %s for %s: %w", publicBackendName(), profile.Name, err)
	}
	sourceHash, err := primaryMeasurementSourceFingerprint(primaryRoot)
	if err != nil {
		return err
	}
	valuesA, valuesB, valuesC, err := perfmeasure.PublicMulRescaleWorkload(effective.InputSlots)
	if err != nil {
		return err
	}
	inputHash := perfmeasure.Fingerprint(valuesA)
	if profile.InputSHA256 != "" && inputHash != profile.InputSHA256 {
		return fmt.Errorf("canonical fixed-sigma chain %s input fingerprint mismatch: %s", profile.Name, inputHash)
	}
	full := params.BootstrappingParameters
	q, p := full.Q(), full.P()
	if len(q) <= 5 || len(q) != full.MaxLevel()+1 {
		return fmt.Errorf("fixed-sigma chain requires constructed q0..q5; got %d Q rows at max Level %d", len(q), full.MaxLevel())
	}
	qpHash := perfmeasure.QPFingerprint(q, p)
	if profile.QPSHA256 != "" && qpHash != profile.QPSHA256 {
		return fmt.Errorf("canonical fixed-sigma chain %s Q/P fingerprint mismatch: %s", profile.Name, qpHash)
	}
	defaultScale := residual.DefaultScale()
	q5Scale, q4Scale := rlwe.NewScale(q[5]), rlwe.NewScale(q[4])
	workloadHash, err := perfmeasure.PublicFixedSigmaChainFingerprint(
		valuesA, valuesB, valuesC,
		defaultScale.BigInt().String(), q5Scale.BigInt().String(), q4Scale.BigInt().String(),
	)
	if err != nil {
		return err
	}
	capacity, err := assessPublicCapacityForProfile(profile, valuesA, valuesB, valuesC, q, full.N())
	if err != nil {
		return fmt.Errorf("original public fixture capacity preflight: %w", err)
	}
	if full.LogN() != 13 || full.N() != 1<<13 || effective.LogSlots != 12 || effective.InputSlots != 1<<effective.LogSlots {
		return errors.New("constructed fixed-sigma chain parameters do not match the accepted LogN13/4096-slot profile")
	}
	backend, err := newPublicBackend(params, residual)
	if err != nil {
		return err
	}
	if backend.name != publicBackendName() {
		return fmt.Errorf("build tag/backend mismatch: adapter=%s selected=%s", backend.name, publicBackendName())
	}
	phases := append([]phaseTiming(nil), backend.construction...)
	checkpoints := make([]publicCheckpoint, 0, 10)
	checkpointVectors := make(map[string][]complexValue, 10)
	decoded := make(map[string][]complex128, 10)
	addCheckpoint := func(name string, ct *rlwe.Ciphertext, expected []complex128, compact bool) ([]complex128, error) {
		var decodePath string
		values, phase, decodeErr := measurePublicPhase("checkpoint_decode_"+name, func() ([]complex128, error) {
			result, path, err := backend.decode(ct)
			decodePath = path
			return result, err
		})
		phases = append(phases, phase)
		if decodeErr != nil {
			return nil, fmt.Errorf("decode checkpoint %s: %w", name, decodeErr)
		}
		cp, err := makePublicCheckpoint(backend, name, ct, values, expected, q, compact)
		if err != nil {
			return nil, err
		}
		cp.DecodePath = decodePath
		if !cp.OraclePass {
			return nil, fmt.Errorf("%s cleartext oracle max %.12g exceeds %.12g", name, cp.Oracle.MaxComplexDifference, publicNumericalGate)
		}
		checkpoints = append(checkpoints, cp)
		checkpointVectors[name] = encodeVector(values)
		decoded[name] = append([]complex128(nil), values...)
		return values, nil
	}

	ctA, phase, err := backend.encryptAtLevel(valuesA, defaultScale, 5, effective.LogSlots)
	phases = append(phases, setPhaseName(phase, "chain_encrypt_a_level5"))
	if err != nil {
		return fmt.Errorf("EncryptNew(A): %w", err)
	}
	ctB, phase, err := backend.encryptAtLevel(valuesB, defaultScale, 5, effective.LogSlots)
	phases = append(phases, setPhaseName(phase, "chain_encrypt_b_level5"))
	if err != nil {
		return fmt.Errorf("EncryptNew(B): %w", err)
	}
	ctCq5, phase, err := backend.encryptAtLevel(valuesC, q5Scale, 5, effective.LogSlots)
	phases = append(phases, setPhaseName(phase, "chain_encrypt_c_q5_level5"))
	if err != nil {
		return fmt.Errorf("EncryptNew(C,q5 scale): %w", err)
	}
	for _, input := range []struct {
		name  string
		ct    *rlwe.Ciphertext
		scale rlwe.Scale
	}{{"A", ctA, defaultScale}, {"B", ctB, defaultScale}, {"C/q5", ctCq5, q5Scale}} {
		if input.ct == nil || input.ct.Level() != 5 || input.ct.Degree() != 1 || !input.ct.IsNTT || input.ct.IsMontgomery || !input.ct.Scale.Equal(input.scale) || input.ct.LogDimensions.Cols != effective.LogSlots {
			return fmt.Errorf("EncryptNew(%s) violates the Level5/Scale/domain/dimension contract", input.name)
		}
		c1Nonzero, err := isCiphertextComponentNonzero(input.ct, 1)
		if err != nil {
			return err
		}
		if backend.name == "fast" && c1Nonzero {
			return fmt.Errorf("Fast zero-secret EncryptNew(%s) produced nonzero c1", input.name)
		}
		if backend.name == "standard" && !c1Nonzero {
			return fmt.Errorf("genuine Standard EncryptNew(%s) produced zero c1", input.name)
		}
	}
	if _, err := addCheckpoint("encrypt_a_level5", ctA, valuesA, false); err != nil {
		return err
	}
	if _, err := addCheckpoint("encrypt_b_level5", ctB, valuesB, false); err != nil {
		return err
	}
	if _, err := addCheckpoint("encrypt_c_q5_level5", ctCq5, valuesC, false); err != nil {
		return err
	}
	ctX, phase, err := measurePublicPhase("chain_add_ab_level5", func() (*rlwe.Ciphertext, error) {
		return backend.evaluator.AddNew(ctA, ctB)
	})
	phases = append(phases, phase)
	if err != nil {
		return fmt.Errorf("AddNew(A,B): %w", err)
	}
	valuesX := publicAdd(valuesA, valuesB)
	if ctX.Level() != 5 || !ctX.Scale.Equal(defaultScale) {
		return errors.New("AddNew(A,B) changed the expected Level5/Scale2^45 state")
	}
	decodedX, err := addCheckpoint("add_ab_level5", ctX, valuesX, true)
	if err != nil {
		return err
	}
	decodedCq5 := decoded["encrypt_c_q5_level5"]
	ctCq4, phase, err := backend.encryptAtLevel(valuesC, q4Scale, 4, effective.LogSlots)
	phases = append(phases, setPhaseName(phase, "chain_encrypt_c_q4_level4"))
	if err != nil {
		return fmt.Errorf("EncryptNew(C,q4 scale): %w", err)
	}
	if ctCq4 == nil || ctCq4.Level() != 4 || ctCq4.Degree() != 1 || !ctCq4.IsNTT || ctCq4.IsMontgomery || !ctCq4.Scale.Equal(q4Scale) || ctCq4.LogDimensions.Cols != effective.LogSlots {
		return errors.New("EncryptNew(C,q4 scale) violates the Level4/q4-Scale/domain/dimension contract")
	}
	c1Nonzero, err := isCiphertextComponentNonzero(ctCq4, 1)
	if err != nil {
		return err
	}
	if backend.name == "fast" && c1Nonzero {
		return errors.New("Fast zero-secret EncryptNew(C,q4 scale) produced nonzero c1")
	}
	if backend.name == "standard" && !c1Nonzero {
		return errors.New("genuine Standard EncryptNew(C,q4 scale) produced zero c1")
	}
	decodedCq4, err := addCheckpoint("encrypt_c_q4_level4", ctCq4, valuesC, false)
	if err != nil {
		return err
	}

	stageOperations := make([]fixedSigmaChainOperation, 0, 2)
	stageVectors := make(map[string]fixedSigmaChainStageVectors, 2)
	runStage := func(
		name, inputXName, inputYName, productName, outputName string,
		ctInputX, ctInputY *rlwe.Ciphertext,
		inputX, inputY, outputOracle []complex128,
		rescaleDivisor uint64,
		outputLevel int,
		seed int64,
	) (*rlwe.Ciphertext, []complex128, error) {
		mulScale := rlwe.NewScale(new(big.Int).Mul(ctInputX.Scale.BigInt(), ctInputY.Scale.BigInt()))
		product, mulPhase, err := measurePublicPhase(name+"_MulRelinNew", func() (*rlwe.Ciphertext, error) {
			return backend.evaluator.MulRelinNew(ctInputX, ctInputY)
		})
		phases = append(phases, mulPhase)
		if err != nil {
			return nil, nil, fmt.Errorf("%s MulRelinNew: %w", name, err)
		}
		if product.Level() != ctInputX.Level() || !product.Scale.Equal(mulScale) {
			return nil, nil, fmt.Errorf("%s MulRelinNew state mismatch: Level=%d Scale=%s, want Level=%d Scale=%s", name, product.Level(), product.Scale.Value.Text('e', 80), ctInputX.Level(), mulScale.Value.Text('e', 80))
		}
		productOracle := append([]complex128(nil), outputOracle...)
		if _, err := addCheckpoint(productName, product, productOracle, true); err != nil {
			return nil, nil, err
		}
		rescaled, rescalePhase, err := measurePublicPhase(name+"_Rescale", func() (*rlwe.Ciphertext, error) {
			out := ckks.NewCiphertext(full, 1, outputLevel)
			if rescaleErr := backend.evaluator.Rescale(product, out); rescaleErr != nil {
				return nil, rescaleErr
			}
			return out, nil
		})
		phases = append(phases, rescalePhase)
		if err != nil {
			return nil, nil, fmt.Errorf("%s Rescale(q%d): %w", name, ctInputX.Level(), err)
		}
		if rescaled.Level() != outputLevel || !rescaled.Scale.Equal(defaultScale) {
			return nil, nil, fmt.Errorf("%s Rescale did not produce Level%d/Scale2^45", name, outputLevel)
		}
		preNoise, decodePhase, err := measurePublicPhase(name+"_decode_pre_injection", func() ([]complex128, error) {
			values, _, decodeErr := backend.decode(rescaled)
			return values, decodeErr
		})
		phases = append(phases, decodePhase)
		if err != nil {
			return nil, nil, fmt.Errorf("%s decode before optional injection: %w", name, err)
		}
		var injection *fixedSigmaNoiseEvidence
		var injectionElapsedNS int64
		if opts.fixedSigmaNoise {
			beforeState := perfmeasure.CiphertextStateOf(rescaled, q, outputLevel)
			beforeMetadata := *rescaled.MetaData
			beforeNTT, beforeMontgomery := rescaled.IsNTT, rescaled.IsMontgomery
			rows := min(outputLevel+1, 4)
			beforeC1Rows := make([]string, rows)
			for row := range beforeC1Rows {
				beforeC1Rows[row] = hashUint64Row(rescaled.Value[1].Coeffs[row])
			}
			evidence, phase, injectErr := measurePublicPhase(name+"_fixed_sigma_injection", func() (fixedSigmaNoiseEvidence, error) {
				return injectPublicFixedSigmaNoise(backend.evaluator, rescaled, opts.fixedSigmaNoiseSigma, seed)
			})
			phases = append(phases, phase)
			if injectErr != nil {
				return nil, nil, fmt.Errorf("%s fixed-sigma injection: %w", name, injectErr)
			}
			injectionElapsedNS = phase.ElapsedNS
			if evidence.SigmaCoefficient != opts.fixedSigmaNoiseSigma || evidence.SamplingSeed != seed ||
				evidence.CoefficientSampleCount != full.N() || evidence.QPrefixRows != rows ||
				evidence.QPrefixCheckedCoefficients != full.N() || !evidence.QPrefixStrictFit || !validSHA256Hex(evidence.CoefficientSHA256) {
				return nil, nil, fmt.Errorf("%s fixed-sigma helper returned incomplete strict Q-prefix evidence", name)
			}
			if perfmeasure.CiphertextStateOf(rescaled, q, outputLevel) != beforeState || !reflect.DeepEqual(*rescaled.MetaData, beforeMetadata) ||
				rescaled.IsNTT != beforeNTT || rescaled.IsMontgomery != beforeMontgomery {
				return nil, nil, fmt.Errorf("%s fixed-sigma injection changed Level/Scale/domain metadata", name)
			}
			for row := range beforeC1Rows {
				if hashUint64Row(rescaled.Value[1].Coeffs[row]) != beforeC1Rows[row] {
					return nil, nil, fmt.Errorf("%s fixed-sigma injection changed c1 q%d", name, row)
				}
			}
			injection = &evidence
		}
		outputDecoded, err := addCheckpoint(outputName, rescaled, outputOracle, true)
		if err != nil {
			return nil, nil, err
		}
		operationResidual, totalError, residualVector, totalVector, err := makePublicOperationResidual(inputX, inputY, outputDecoded, outputOracle)
		if err != nil {
			return nil, nil, fmt.Errorf("%s incremental residual: %w", name, err)
		}
		operationResidual.Classification = "DIAGNOSTIC_ONLY"
		operationResidual.EncryptionMode = "sk"
		operationResidual.DispatchEvidence = publicMulRelinRescaleDispatchEvidence()
		operationResidual.ElapsedNS = mulPhase.ElapsedNS + rescalePhase.ElapsedNS
		preStats := (*numericalmetrics.ComplexVectorStats)(nil)
		noiseDeltaStats := (*numericalmetrics.ComplexVectorStats)(nil)
		stageVector := fixedSigmaChainStageVectors{
			IncrementalResidual: residualVector,
			TotalError:          totalVector,
			MulRelinOracle:      encodeVector(productOracle),
			PlaintextOracle:     encodeVector(outputOracle),
		}
		if injection != nil {
			preResidual, _, _, _, preErr := makePublicOperationResidual(inputX, inputY, preNoise, outputOracle)
			if preErr != nil {
				return nil, nil, fmt.Errorf("%s pre-injection residual: %w", name, preErr)
			}
			noiseDelta := make([]complex128, len(outputDecoded))
			for index := range noiseDelta {
				noiseDelta[index] = outputDecoded[index] - preNoise[index]
			}
			noiseStats := numericalmetrics.SummarizeComplexVector(noiseDelta)
			if noiseStats.Status != "FINITE" || noiseStats.Count != effective.InputSlots {
				return nil, nil, fmt.Errorf("%s decoded injection delta is incomplete or non-finite", name)
			}
			seedValue, sigmaValue := injection.SamplingSeed, injection.SigmaCoefficient
			operationResidual.LogicalNoiseInjections = 1
			operationResidual.SigmaCoefficient = &sigmaValue
			operationResidual.SigmaStatus = "FROZEN_CP_B_CANDIDATE_DIAGNOSTIC_ONLY"
			operationResidual.SamplingSeedStatus = "SEEDED_GO_MATH_RAND_RECORDED"
			operationResidual.SamplingSeed = &seedValue
			operationResidual.FixedSigmaNoise = injection
			operationResidual.PreNoiseOutputSHA256 = perfmeasure.Fingerprint(preNoise)
			operationResidual.PreNoiseStatistics = &preResidual.Statistics
			operationResidual.DecodedNoiseDelta = &noiseStats
			operationResidual.InjectionElapsedNS = injectionElapsedNS
			preStats, noiseDeltaStats = &preResidual.Statistics, &noiseStats
			stageVector.PreNoiseOutput = encodeVector(preNoise)
			stageVector.DecodedNoiseDelta = encodeVector(noiseDelta)
		}
		stageOperations = append(stageOperations, fixedSigmaChainOperation{
			Name: name, InputXCheckpoint: inputXName, InputYCheckpoint: inputYName,
			InputXState:         stateOf(ctInputX, q, min(ctInputX.Level()+1, 4)),
			InputYState:         stateOf(ctInputY, q, min(ctInputY.Level()+1, 4)),
			MultiplicationScale: product.Scale.Value.Text('e', 80), RescaleDivisor: new(big.Int).SetUint64(rescaleDivisor).String(),
			ProductCheckpoint: productName, OutputCheckpoint: outputName,
			OutputState:       stateOf(rescaled, q, min(outputLevel+1, 4)),
			MulRelinElapsedNS: mulPhase.ElapsedNS, RescaleElapsedNS: rescalePhase.ElapsedNS,
			Residual: *operationResidual, TotalError: totalError,
			PreNoiseResidual: preStats, DecodedNoiseDelta: noiseDeltaStats,
		})
		stageVectors[name] = stageVector
		return rescaled, outputDecoded, nil
	}

	stage1Oracle := publicMul(valuesX, valuesC)
	stage1, decodedStage1, err := runStage("stage1", "add_ab_level5", "encrypt_c_q5_level5", "chain_stage1_mulrelin_level5", "chain_stage1_output_level4",
		ctX, ctCq5, decodedX, decodedCq5, stage1Oracle, q[5], 4, opts.fixedSigmaNoiseSeed)
	if err != nil {
		return err
	}
	stage2Oracle := publicMul(stage1Oracle, valuesC)
	stage2Seed := opts.fixedSigmaNoiseSeed
	if stage2Seed == int64(^uint64(0)>>1) {
		return errors.New("fixed-sigma chain seed cannot be incremented for the second stage")
	}
	stage2Seed++
	_, _, err = runStage("stage2", "chain_stage1_output_level4", "encrypt_c_q4_level4", "chain_stage2_mulrelin_level4", "chain_stage2_output_level3",
		stage1, ctCq4, decodedStage1, decodedCq4, stage2Oracle, q[4], 3, stage2Seed)
	if err != nil {
		return err
	}

	inputKind, constructor := publicInputContract(backend.name)
	inputC1Nonzero, err := isCiphertextComponentNonzero(ctA, 1)
	if err != nil {
		return err
	}
	inputMessageHashes := map[string]string{
		"A": perfmeasure.Fingerprint(valuesA),
		"B": perfmeasure.Fingerprint(valuesB),
		"C": perfmeasure.Fingerprint(valuesC),
	}
	chain := &fixedSigmaChainEvidence{
		SchemaVersion: "fixed-sigma-two-step-chain.v2", Classification: "DIAGNOSTIC_ONLY",
		PrimaryBuildCommit: primaryBuildCommit, PrimaryBuildClean: primaryBuildClean,
		LogicalCtCtOperations: 2, LogicalNoiseInjections: 0,
		InputMessageSHA256: inputMessageHashes, Operations: stageOperations,
		Limitations: []string{
			"Two explicit public ct×ct MulRelinNew→Rescale operations only; no Bootstrap, lazy/deferred Bootstrap chain, or generic Rescale interception was run.",
			"Stage 2 re-encodes the same original C message at the effective q4 Scale so the ordinary q4 Rescale returns to Scale2^45; this is a diagnostic CKKS schedule, not a general application workload claim.",
			"Standard uses genuine native SK encryption; Fast uses the pinned intentionally insecure zero-secret c1=0 mode. PK comparison remains unsupported.",
		},
	}
	if opts.fixedSigmaNoise {
		chain.LogicalNoiseInjections = 2
		sigma := opts.fixedSigmaNoiseSigma
		chain.SigmaCoefficient = &sigma
		chain.Stage2SeedDerivation = "stage2 seed is stage1 seed + 1; both actual seeds are recorded on their operation evidence"
	}
	mode, schema, vectorSchema := "fixed-sigma-two-step-chain", "fast-standard-fixed-sigma-two-step-chain.v2", "fast-standard-fixed-sigma-two-step-chain-vectors.v2"
	status := "PASS_ZERO_BOOTSTRAP_TWO_STEP_CHAIN"
	limitations := []string{
		"Diagnostic-only two-step MulRelin+Rescale evidence; no Bootstrap call, output, or performance claim is included.",
		"Same canonical messages and public CKKS call sequence are built against isolated genuine Standard and current Fast sources; keys and ciphertext coefficients are backend-native.",
	}
	if opts.fixedSigmaNoise {
		limitations = append(limitations, "Fast-only fixed-sigma injection is explicit after each complete MulRelin+Rescale pair; it is not automatic for arbitrary MulRelin or Rescale calls.")
	}
	observedOracleStatus := publicStrictOraclePassed
	doc := publicNativeDocument{
		SchemaVersion: schema, Mode: mode, Status: status, Timestamp: time.Now().UTC(),
		Profile: profile.Name, ProfileKind: profile.Kind, Backend: backend.name, BackendCommit: secondaryCommit,
		BackendRef: opts.backendRef, BackendCheckoutRef: secondaryRef, BackendPath: secondaryRoot, BackendClean: true,
		BinarySHA256: binarySHA256, PrimaryPath: primaryRoot, PrimaryCommit: primaryCommit, PrimaryClean: true,
		PrimarySourceSHA256: sourceHash, BuildTag: "perf_" + backend.name,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(),
		NumCPU: runtime.NumCPU(), GOMAXPROCS: runtime.GOMAXPROCS(0), GOGC: environmentSetting("GOGC"),
		GOMEMLIMIT: environmentSetting("GOMEMLIMIT"), GODEBUG: environmentSetting("GODEBUG"),
		ConfigPath: opts.config, ConfigSHA256: configHash, QPSHA256: qpHash, InputSHA256: inputHash, WorkloadSHA256: workloadHash,
		Parameters: effective, EphemeralSecretWeight: params.EphemeralSecretWeight, InputSlots: effective.InputSlots,
		InputKind: inputKind, InputConstructor: constructor, InputC1Nonzero: inputC1Nonzero,
		PublicEvaluatorDispatch: backend.dispatch, CapacityEvidenceVersion: publicCapacityEvidenceVersion,
		CapacityPolicy: profile.CapacityPolicy, ConservativeBoundStatus: capacity.TerminalQ0Status,
		ObservedNativeOracleStatus: observedOracleStatus, Capacity: capacity, NumericalGate: publicNumericalGate,
		BootstrapBudget: 0, BootstrapCalls: 0, ExecutionPhases: phases, Checkpoints: checkpoints, FixedSigmaChain: chain,
		Limitations: limitations,
	}
	vectorDoc := publicVectorDocument{
		SchemaVersion: vectorSchema, Mode: mode, Backend: backend.name,
		BackendCommit: secondaryCommit, PrimaryCommit: primaryCommit, SourceSHA256: sourceHash,
		ConfigSHA256: configHash, QPSHA256: qpHash, InputSHA256: inputHash, WorkloadSHA256: workloadHash,
		E: params.EphemeralSecretWeight, Checkpoints: checkpointVectors,
		FixedSigmaChain: &fixedSigmaChainVectors{
			Inputs: map[string][]complexValue{
				"A": encodeVector(valuesA), "B": encodeVector(valuesB), "C": encodeVector(valuesC),
			},
			Stages: stageVectors,
		},
	}
	if err := writeExclusiveJSON(opts.out, doc); err != nil {
		return err
	}
	if err := writeExclusiveJSON(opts.vectorsOut, vectorDoc); err != nil {
		return err
	}
	fmt.Printf("%s %s mode=%s result=%s operations=%d injections=%d bootstrap_calls=0\n", profile.Name, backend.name, mode, opts.out, len(stageOperations), chain.LogicalNoiseInjections)
	return nil
}

func setPhaseName(phase phaseTiming, name string) phaseTiming {
	phase.Phase = name
	return phase
}

var fixedSigmaChainCheckpointNames = []string{
	"encrypt_a_level5", "encrypt_b_level5", "encrypt_c_q5_level5", "add_ab_level5",
	"chain_stage1_mulrelin_level5", "chain_stage1_output_level4", "encrypt_c_q4_level4",
	"chain_stage2_mulrelin_level4", "chain_stage2_output_level3",
}

func validatePublicFixedSigmaChainPairArtifacts(standard, fast publicNativeDocument, standardVectors, fastVectors publicVectorDocument) error {
	if standard.Mode != "fixed-sigma-two-step-chain" || fast.Mode != standard.Mode ||
		!validFixedSigmaChainSchema(standard.SchemaVersion) || fast.SchemaVersion != standard.SchemaVersion ||
		standard.Status != "PASS_ZERO_BOOTSTRAP_TWO_STEP_CHAIN" || fast.Status != standard.Status {
		return errors.New("paired artifacts are not successful fixed-sigma two-step chain runs")
	}
	if err := validatePublicFixedSigmaChainArtifact(standard, standardVectors, "standard"); err != nil {
		return fmt.Errorf("Standard chain artifact: %w", err)
	}
	if err := validatePublicFixedSigmaChainArtifact(fast, fastVectors, "fast"); err != nil {
		return fmt.Errorf("Fast chain artifact: %w", err)
	}
	profile, err := publicProfileForNativeDocument(standard)
	if err != nil {
		return err
	}
	if standard.Profile != "logn13" || fast.Profile != standard.Profile || fast.ProfileKind != standard.ProfileKind ||
		standard.EphemeralSecretWeight != 32 || fast.EphemeralSecretWeight != 32 ||
		standard.BootstrapBudget != 0 || fast.BootstrapBudget != 0 || standard.BootstrapCalls != 0 || fast.BootstrapCalls != 0 ||
		standard.InputSlots != profile.InputSlots || fast.InputSlots != standard.InputSlots ||
		!standard.PrimaryClean || !fast.PrimaryClean || !standard.BackendClean || !fast.BackendClean {
		return errors.New("paired chain artifacts violate the pinned LogN13/E32 clean-provenance zero-Bootstrap contract")
	}
	if err := validatePublicFixedSigmaChainBackendPin("standard", standard.BackendCommit, profile.Name); err != nil {
		return err
	}
	if err := validatePublicFixedSigmaChainBackendPin("fast", fast.BackendCommit, profile.Name); err != nil {
		return err
	}
	if !validSHA256Hex(standard.BinarySHA256) || !validSHA256Hex(fast.BinarySHA256) || standard.BinarySHA256 == fast.BinarySHA256 ||
		standard.BuildTag != "perf_standard" || fast.BuildTag != "perf_fast" ||
		standard.PrimaryCommit == "" || standard.PrimaryCommit != fast.PrimaryCommit ||
		standard.PrimarySourceSHA256 == "" || standard.PrimarySourceSHA256 != fast.PrimarySourceSHA256 ||
		standard.ConfigSHA256 != profile.ConfigSHA256 || fast.ConfigSHA256 != standard.ConfigSHA256 ||
		standard.QPSHA256 != profile.QPSHA256 || fast.QPSHA256 != standard.QPSHA256 ||
		standard.InputSHA256 != profile.InputSHA256 || fast.InputSHA256 != standard.InputSHA256 ||
		standard.WorkloadSHA256 == "" || fast.WorkloadSHA256 != standard.WorkloadSHA256 ||
		!reflect.DeepEqual(standard.Parameters, fast.Parameters) || !reflect.DeepEqual(standard.Capacity, fast.Capacity) {
		return errors.New("paired chain artifacts have mismatched pins, source, binaries, config, effective parameters, capacity, input, or workload provenance")
	}
	if !reflect.DeepEqual(standard.FixedSigmaChain.InputMessageSHA256, fast.FixedSigmaChain.InputMessageSHA256) {
		return errors.New("paired chain artifacts do not use identical original plaintext inputs")
	}
	if standard.FixedSigmaChain.LogicalNoiseInjections != 0 ||
		(fast.FixedSigmaChain.LogicalNoiseInjections != 0 && fast.FixedSigmaChain.LogicalNoiseInjections != 2) {
		return errors.New("paired chain has an invalid Standard/Fast noise injection count")
	}
	for index := range fixedSigmaChainCheckpointNames {
		left, right := standard.Checkpoints[index], fast.Checkpoints[index]
		if left.Name != right.Name || left.State.Level != right.State.Level || left.State.Degree != right.State.Degree || left.State.Scale != right.State.Scale {
			return fmt.Errorf("paired chain checkpoint %s has mismatched Level/Scale/Degree", fixedSigmaChainCheckpointNames[index])
		}
	}
	return nil
}

func validatePublicFixedSigmaChainArtifact(doc publicNativeDocument, vectors publicVectorDocument, backend string) error {
	if doc.Backend != backend || doc.BuildTag != "perf_"+backend || doc.BootstrapBudget != 0 || doc.BootstrapCalls != 0 ||
		doc.EphemeralSecretWeight != 32 || doc.InputSlots != 1<<12 || doc.CapacityPolicy != publicCapacityPolicyStrict ||
		doc.FixedSigmaChain == nil || !validFixedSigmaChainEvidenceSchema(doc.FixedSigmaChain.SchemaVersion) ||
		doc.FixedSigmaChain.Classification != "DIAGNOSTIC_ONLY" || doc.FixedSigmaChain.LogicalCtCtOperations != 2 ||
		len(doc.FixedSigmaChain.Operations) != 2 || !validSHA256Hex(doc.BinarySHA256) || !validSHA256Hex(doc.PrimarySourceSHA256) {
		return errors.New("chain provenance, scope, or operation count is incomplete")
	}
	if doc.SchemaVersion == "fast-standard-fixed-sigma-two-step-chain.v2" &&
		(doc.FixedSigmaChain.SchemaVersion != "fixed-sigma-two-step-chain.v2" ||
			doc.FixedSigmaChain.PrimaryBuildCommit != doc.PrimaryCommit || !doc.FixedSigmaChain.PrimaryBuildClean) {
		return errors.New("chain binary was not built from the clean recorded Primary commit")
	}
	if vectors.SchemaVersion != fixedSigmaChainVectorSchema(doc.SchemaVersion) || vectors.Mode != doc.Mode ||
		vectors.Backend != backend || vectors.BackendCommit != doc.BackendCommit || vectors.PrimaryCommit != doc.PrimaryCommit ||
		vectors.SourceSHA256 != doc.PrimarySourceSHA256 || vectors.ConfigSHA256 != doc.ConfigSHA256 ||
		vectors.QPSHA256 != doc.QPSHA256 || vectors.InputSHA256 != doc.InputSHA256 ||
		vectors.WorkloadSHA256 != doc.WorkloadSHA256 || vectors.E != 32 ||
		vectors.FixedSigmaChain == nil || len(vectors.FixedSigmaChain.Stages) != 2 || len(vectors.FixedSigmaChain.Inputs) != 3 {
		return errors.New("chain vector artifact has missing or mismatched provenance")
	}
	if !samePublicCheckpointNames(doc.Checkpoints, fixedSigmaChainCheckpointNames) ||
		len(vectors.Checkpoints) != len(fixedSigmaChainCheckpointNames) {
		return errors.New("chain checkpoint coverage does not match the exact two-operation schedule")
	}
	inputs := make(map[string][]complex128, 3)
	for _, name := range []string{"A", "B", "C"} {
		encoded, ok := vectors.FixedSigmaChain.Inputs[name]
		if !ok || len(encoded) != doc.InputSlots {
			return fmt.Errorf("original plaintext input %s is missing or has the wrong slot count", name)
		}
		values := decodeComplexValues(encoded)
		if !allFinite(values) || perfmeasure.Fingerprint(values) != doc.FixedSigmaChain.InputMessageSHA256[name] {
			return fmt.Errorf("original plaintext input %s fingerprint is invalid", name)
		}
		inputs[name] = values
	}
	if len(doc.FixedSigmaChain.InputMessageSHA256) != 3 {
		return errors.New("chain message fingerprint map must contain exactly A, B, and C")
	}
	byName := make(map[string]publicCheckpoint, len(doc.Checkpoints))
	for _, checkpoint := range doc.Checkpoints {
		encoded, ok := vectors.Checkpoints[checkpoint.Name]
		if !ok || len(encoded) != doc.InputSlots || !checkpoint.OraclePass ||
			checkpoint.Oracle.MaxComplexDifference > publicNumericalGate ||
			checkpoint.DecodedSHA256 != perfmeasure.Fingerprint(decodeComplexValues(encoded)) {
			return fmt.Errorf("checkpoint %s is missing, non-finite, or failed its decoded plaintext oracle", checkpoint.Name)
		}
		if backend == "standard" {
			if !checkpoint.C1Nonzero || checkpoint.FastCompactPrefix || checkpoint.DecodePath != "rlwe.NewDecryptor(...).DecryptNew -> ckks.NewEncoder.Decode" {
				return fmt.Errorf("Standard checkpoint %s does not show genuine native decryption", checkpoint.Name)
			}
		} else if checkpoint.C1Nonzero || checkpoint.DecodePath != "Fast c0 projection to the authoritative Q-prefix plaintext (zero-secret measurement adapter)" {
			return fmt.Errorf("Fast checkpoint %s violates zero-secret Q-prefix decode provenance", checkpoint.Name)
		}
		byName[checkpoint.Name] = checkpoint
	}
	if backend == "standard" && !doc.InputC1Nonzero || backend == "fast" && doc.InputC1Nonzero {
		return errors.New("top-level input c1 provenance does not match the selected backend")
	}
	addOracle := publicAdd(inputs["A"], inputs["B"])
	stage1Oracle := publicMul(addOracle, inputs["C"])
	stage2Oracle := publicMul(stage1Oracle, inputs["C"])
	wantCheckpointOracles := map[string][]complex128{
		"encrypt_a_level5": inputs["A"], "encrypt_b_level5": inputs["B"],
		"encrypt_c_q5_level5": inputs["C"], "add_ab_level5": addOracle,
		"chain_stage1_mulrelin_level5": stage1Oracle, "chain_stage1_output_level4": stage1Oracle,
		"encrypt_c_q4_level4": inputs["C"], "chain_stage2_mulrelin_level4": stage2Oracle,
		"chain_stage2_output_level3": stage2Oracle,
	}
	for name, expected := range wantCheckpointOracles {
		actual := decodeComplexValues(vectors.Checkpoints[name])
		metrics := compareVectors(expected, actual, publicNumericalGate)
		if metrics != byName[name].Oracle {
			return fmt.Errorf("checkpoint %s oracle metrics do not recompute from the original message", name)
		}
	}
	wantLevels := map[string]int{
		"encrypt_a_level5": 5, "encrypt_b_level5": 5, "encrypt_c_q5_level5": 5, "add_ab_level5": 5,
		"chain_stage1_mulrelin_level5": 5, "chain_stage1_output_level4": 4, "encrypt_c_q4_level4": 4,
		"chain_stage2_mulrelin_level4": 4, "chain_stage2_output_level3": 3,
	}
	for name, level := range wantLevels {
		if byName[name].State.Level != level || byName[name].State.QPrefixRows != min(level+1, 4) {
			return fmt.Errorf("checkpoint %s has Level/Q-prefix state %d/%d, want %d/%d", name, byName[name].State.Level, byName[name].State.QPrefixRows, level, min(level+1, 4))
		}
	}
	if len(doc.Parameters.QPrimes) <= 5 || doc.FixedSigmaChain.Operations[0].RescaleDivisor != doc.Parameters.QPrimes[5] ||
		doc.FixedSigmaChain.Operations[1].RescaleDivisor != doc.Parameters.QPrimes[4] {
		return errors.New("chain Rescale divisors do not match the exact effective q5/q4 parameters")
	}
	q4, ok4 := new(big.Int).SetString(doc.Parameters.QPrimes[4], 10)
	q5, ok5 := new(big.Int).SetString(doc.Parameters.QPrimes[5], 10)
	if !ok4 || !ok5 {
		return errors.New("effective q4/q5 parameters are not canonical decimal integers")
	}
	delta := new(big.Int).Lsh(big.NewInt(1), 45)
	defaultScale := fixedSigmaChainScaleText(delta)
	q4Scale := fixedSigmaChainScaleText(q4)
	q5Scale := fixedSigmaChainScaleText(q5)
	stage1ProductScale := fixedSigmaChainScaleText(new(big.Int).Mul(new(big.Int).Set(delta), q5))
	stage2ProductScale := fixedSigmaChainScaleText(new(big.Int).Mul(new(big.Int).Set(delta), q4))
	wantScales := map[string]string{
		"encrypt_a_level5": defaultScale, "encrypt_b_level5": defaultScale,
		"encrypt_c_q5_level5": q5Scale, "add_ab_level5": defaultScale,
		"chain_stage1_mulrelin_level5": stage1ProductScale, "chain_stage1_output_level4": defaultScale,
		"encrypt_c_q4_level4": q4Scale, "chain_stage2_mulrelin_level4": stage2ProductScale,
		"chain_stage2_output_level3": defaultScale,
	}
	for name, scale := range wantScales {
		if byName[name].State.Scale != scale {
			return fmt.Errorf("checkpoint %s Scale=%s, want effective chain Scale %s", name, byName[name].State.Scale, scale)
		}
	}
	wantRefs := [][3]string{
		{"add_ab_level5", "encrypt_c_q5_level5", "chain_stage1_output_level4"},
		{"chain_stage1_output_level4", "encrypt_c_q4_level4", "chain_stage2_output_level3"},
	}
	for index, operation := range doc.FixedSigmaChain.Operations {
		stageName := fmt.Sprintf("stage%d", index+1)
		refs := wantRefs[index]
		productName := []string{"chain_stage1_mulrelin_level5", "chain_stage2_mulrelin_level4"}[index]
		if operation.Name != stageName || operation.InputXCheckpoint != refs[0] || operation.InputYCheckpoint != refs[1] ||
			operation.ProductCheckpoint != productName || operation.OutputCheckpoint != refs[2] ||
			!reflect.DeepEqual(operation.InputXState, byName[refs[0]].State) ||
			!reflect.DeepEqual(operation.InputYState, byName[refs[1]].State) ||
			!reflect.DeepEqual(operation.OutputState, byName[refs[2]].State) ||
			operation.MulRelinElapsedNS <= 0 || operation.RescaleElapsedNS <= 0 ||
			operation.Residual.ElapsedNS != operation.MulRelinElapsedNS+operation.RescaleElapsedNS ||
			operation.Residual.LogicalCtCtMulCount != 1 || operation.Residual.EncryptionMode != "sk" ||
			operation.Residual.PKStatus != "UNSUPPORTED_FAST_ZERO_SECRET_MULRELIN_REQUIRES_C1_ZERO" ||
			operation.Residual.SchemaVersion != "fixed-sigma-mul-relin-rescale-residual.v1" ||
			operation.Residual.Classification != "DIAGNOSTIC_ONLY" ||
			operation.Residual.DispatchEvidence != fixedSigmaChainDispatchEvidence(backend) {
			return fmt.Errorf("%s operation metadata or phase timing is incomplete", stageName)
		}
		productState := byName[productName].State
		if operation.MultiplicationScale != productState.Scale || productState.Level != operation.InputXState.Level {
			return fmt.Errorf("%s MulRelin checkpoint Level/Scale differs from operation metadata", stageName)
		}
		stageVectors, ok := vectors.FixedSigmaChain.Stages[stageName]
		if !ok || len(stageVectors.IncrementalResidual) != doc.InputSlots || len(stageVectors.TotalError) != doc.InputSlots ||
			len(stageVectors.MulRelinOracle) != doc.InputSlots || len(stageVectors.PlaintextOracle) != doc.InputSlots {
			return fmt.Errorf("%s stage vectors are incomplete", stageName)
		}
		inputX := decodeComplexValues(vectors.Checkpoints[refs[0]])
		inputY := decodeComplexValues(vectors.Checkpoints[refs[1]])
		output := decodeComplexValues(vectors.Checkpoints[refs[2]])
		residual := subtractComplexVectors(output, multiplyComplexVectors(inputX, inputY))
		totalError := subtractComplexVectors(output, decodeComplexValues(stageVectors.PlaintextOracle))
		if !complexVectorsEqual(residual, decodeComplexValues(stageVectors.IncrementalResidual)) ||
			!complexVectorsEqual(totalError, decodeComplexValues(stageVectors.TotalError)) ||
			!complexVectorsEqual(decodeComplexValues(stageVectors.MulRelinOracle), productOracleForStage(stageName, addOracle, inputs["C"])) ||
			!complexVectorsEqual(decodeComplexValues(stageVectors.PlaintextOracle), wantCheckpointOracles[refs[2]]) {
			return fmt.Errorf("%s residual or plaintext-oracle vectors do not match the recorded chain", stageName)
		}
		residualStats := numericalmetrics.SummarizeComplexVector(residual)
		totalStats := numericalmetrics.SummarizeComplexVector(totalError)
		if !sameComplexVectorStats(residualStats, operation.Residual.Statistics) || !sameComplexVectorStats(totalStats, operation.TotalError) ||
			operation.Residual.InputXSHA256 != perfmeasure.Fingerprint(inputX) ||
			operation.Residual.InputYSHA256 != perfmeasure.Fingerprint(inputY) ||
			operation.Residual.OutputSHA256 != perfmeasure.Fingerprint(output) {
			return fmt.Errorf("%s residual summary/hash does not recompute from its decoded checkpoints", stageName)
		}
		if operation.Residual.LogicalNoiseInjections == 0 {
			if operation.Residual.FixedSigmaNoise != nil || operation.PreNoiseResidual != nil || operation.DecodedNoiseDelta != nil ||
				len(stageVectors.PreNoiseOutput) != 0 || len(stageVectors.DecodedNoiseDelta) != 0 {
				return fmt.Errorf("%s records injection data despite a zero injection count", stageName)
			}
		} else {
			if operation.Residual.LogicalNoiseInjections != 1 || operation.Residual.FixedSigmaNoise == nil ||
				operation.Residual.FixedSigmaNoise.SigmaCoefficient != fixedSigmaChainCoefficientSigma ||
				operation.PreNoiseResidual == nil || operation.DecodedNoiseDelta == nil ||
				len(stageVectors.PreNoiseOutput) != doc.InputSlots || len(stageVectors.DecodedNoiseDelta) != doc.InputSlots {
				return fmt.Errorf("%s fixed-sigma injection evidence is incomplete", stageName)
			}
			preNoise := decodeComplexValues(stageVectors.PreNoiseOutput)
			noiseDelta := decodeComplexValues(stageVectors.DecodedNoiseDelta)
			preResidual := subtractComplexVectors(preNoise, multiplyComplexVectors(inputX, inputY))
			wantDelta := subtractComplexVectors(output, preNoise)
			if !complexVectorsEqual(noiseDelta, wantDelta) || !sameComplexVectorStats(numericalmetrics.SummarizeComplexVector(preResidual), *operation.PreNoiseResidual) ||
				!sameComplexVectorStats(numericalmetrics.SummarizeComplexVector(noiseDelta), *operation.DecodedNoiseDelta) ||
				operation.Residual.PreNoiseOutputSHA256 != perfmeasure.Fingerprint(preNoise) ||
				!validSHA256Hex(operation.Residual.FixedSigmaNoise.CoefficientSHA256) ||
				operation.Residual.FixedSigmaNoise.QPrefixRows != min(operation.OutputState.Level+1, 4) ||
				!operation.Residual.FixedSigmaNoise.QPrefixStrictFit {
				return fmt.Errorf("%s fixed-sigma residual, delta, or strict Q-prefix evidence does not recompute", stageName)
			}
		}
	}
	if backend == "standard" && doc.FixedSigmaChain.LogicalNoiseInjections != 0 ||
		backend == "fast" && doc.FixedSigmaChain.LogicalNoiseInjections != 0 && doc.FixedSigmaChain.LogicalNoiseInjections != 2 {
		return errors.New("chain injection count does not match the Standard/Fast diagnostic mode")
	}
	if doc.FixedSigmaChain.LogicalNoiseInjections == 2 {
		if backend != "fast" || doc.FixedSigmaChain.SigmaCoefficient == nil || *doc.FixedSigmaChain.SigmaCoefficient != fixedSigmaChainCoefficientSigma ||
			doc.FixedSigmaChain.Operations[0].Residual.SamplingSeed == nil || doc.FixedSigmaChain.Operations[1].Residual.SamplingSeed == nil ||
			*doc.FixedSigmaChain.Operations[0].Residual.SamplingSeed == int64(^uint64(0)>>1) ||
			*doc.FixedSigmaChain.Operations[1].Residual.SamplingSeed != *doc.FixedSigmaChain.Operations[0].Residual.SamplingSeed+1 ||
			doc.FixedSigmaChain.Stage2SeedDerivation != "stage2 seed is stage1 seed + 1; both actual seeds are recorded on their operation evidence" {
			return errors.New("two-step chain does not contain the same frozen sigma and distinct reproducible stage seeds")
		}
	} else if doc.FixedSigmaChain.SigmaCoefficient != nil {
		return errors.New("chain reports a sigma without two explicit injections")
	}
	return nil
}

func fixedSigmaChainScaleText(value *big.Int) string {
	scale := rlwe.NewScale(value)
	return scale.Value.Text('e', 80)
}

func fixedSigmaChainDispatchEvidence(backend string) string {
	switch backend {
	case "standard":
		return "ordinary ckks.Evaluator.MulRelinNew and Rescale on the isolated genuine Standard source"
	case "fast":
		return "ordinary ckks.Evaluator.MulRelinNew -> Fast zero-secret MulCore.ApplyRows; ordinary ckks.Evaluator.Rescale -> shared Q-prefix RescaleCore.ApplyRows; no Standard fallback"
	default:
		return ""
	}
}

func validFixedSigmaChainSchema(value string) bool {
	return value == "fast-standard-fixed-sigma-two-step-chain.v1" || value == "fast-standard-fixed-sigma-two-step-chain.v2"
}

func validFixedSigmaChainEvidenceSchema(value string) bool {
	return value == "fixed-sigma-two-step-chain.v1" || value == "fixed-sigma-two-step-chain.v2"
}

func fixedSigmaChainVectorSchema(value string) string {
	switch value {
	case "fast-standard-fixed-sigma-two-step-chain.v1":
		return "fast-standard-fixed-sigma-two-step-chain-vectors.v1"
	case "fast-standard-fixed-sigma-two-step-chain.v2":
		return "fast-standard-fixed-sigma-two-step-chain-vectors.v2"
	default:
		return ""
	}
}

func validateFixedSigmaChainBinaryProvenance(primaryCommit, buildCommit, vcsModified string) (bool, error) {
	if primaryCommit == "" || buildCommit == "" || buildCommit != primaryCommit {
		return false, fmt.Errorf("binary vcs.revision=%q does not match Primary HEAD=%q", buildCommit, primaryCommit)
	}
	if vcsModified != "false" {
		return false, fmt.Errorf("binary vcs.modified=%q, want false", vcsModified)
	}
	return true, nil
}

func fixedSigmaChainBinaryProvenance(primaryCommit string) (string, bool, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false, errors.New("Go build info is unavailable")
	}
	var buildCommit, vcsModified string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			buildCommit = setting.Value
		case "vcs.modified":
			vcsModified = setting.Value
		}
	}
	clean, err := validateFixedSigmaChainBinaryProvenance(primaryCommit, buildCommit, vcsModified)
	if err != nil {
		return buildCommit, false, err
	}
	return buildCommit, clean, nil
}

func validatePublicFixedSigmaChainBackendPin(backend, commit, profile string) error {
	if backend == "fast" && profile == "logn13" && commit == publicFixedSigmaChainFastSHA {
		return nil
	}
	return validatePublicBackendPin(backend, commit, profile)
}

func productOracleForStage(stage string, add, c []complex128) []complex128 {
	first := publicMul(add, c)
	if stage == "stage1" {
		return first
	}
	return publicMul(first, c)
}

func comparePublicFixedSigmaChainOperations(standard, fast publicNativeDocument, standardVectors, fastVectors publicVectorDocument) ([]publicPairOperationResidual, error) {
	operations := make([]publicPairOperationResidual, 0, 2)
	for index := range standard.FixedSigmaChain.Operations {
		standardOperation, fastOperation := standard.FixedSigmaChain.Operations[index], fast.FixedSigmaChain.Operations[index]
		stage := standardOperation.Name
		standardResidual := decodeComplexValues(standardVectors.FixedSigmaChain.Stages[stage].IncrementalResidual)
		fastResidual := decodeComplexValues(fastVectors.FixedSigmaChain.Stages[stage].IncrementalResidual)
		standardStats := numericalmetrics.SummarizeComplexVector(standardResidual)
		fastStats := numericalmetrics.SummarizeComplexVector(fastResidual)
		if !sameComplexVectorStats(standardStats, standardOperation.Residual.Statistics) ||
			!sameComplexVectorStats(fastStats, fastOperation.Residual.Statistics) {
			return nil, fmt.Errorf("%s paired incremental residual statistics do not match source vectors", stage)
		}
		varianceDelta := standardStats.Variance - fastStats.Variance
		varianceStatus := "NONNEGATIVE_SLOT_VARIANCE_DELTA_ONLY"
		if varianceDelta < 0 {
			varianceStatus = "UNMODELLED_NEGATIVE_SLOT_VARIANCE_DELTA"
		}
		paired := publicPairOperationResidual{
			SchemaVersion: "fixed-sigma-two-step-operation-pair.v1", StageName: stage,
			Standard: standardStats, Fast: fastStats,
			StandardTotalError: standardOperation.TotalError, FastTotalError: fastOperation.TotalError,
			StandardMinusFastVariance: varianceDelta, VarianceDeltaStatus: varianceStatus,
			CrossBackendCovariance: numericalmetrics.CompareComplexCovariance(standardResidual, fastResidual),
			CoefficientSigmaStatus: "NO_INJECTION_CP_B_BASELINE",
			Interpretation:         "Stagewise descriptive residual and total-oracle moments use each backend's decoded inputs; they are not a distribution-equivalence or security claim.",
		}
		if fastOperation.Residual.LogicalNoiseInjections == 1 {
			evidence := fastOperation.Residual.FixedSigmaNoise
			sigma, seed, coefficientRMS := evidence.SigmaCoefficient, evidence.SamplingSeed, evidence.CoefficientRMS
			paired.FastFixedSigmaCoefficient = &sigma
			paired.FastSamplingSeed = &seed
			paired.FastCoefficientRMS = &coefficientRMS
			paired.FastCoefficientSHA256 = evidence.CoefficientSHA256
			paired.FastDecodedNoiseDelta = fastOperation.DecodedNoiseDelta
			paired.FastInjectionElapsedNS = fastOperation.Residual.InjectionElapsedNS
			paired.CoefficientSigmaStatus = "FROZEN_CP_B_CANDIDATE_APPLIED_DIAGNOSTIC_ONLY"
			paired.Interpretation = "The frozen coefficient sigma was injected once after this complete Fast MulRelin+Rescale stage. RMS/variance/tail similarity remain separate observations."
		}
		operations = append(operations, paired)
	}
	return operations, nil
}

func multiplyComplexVectors(a, b []complex128) []complex128 {
	if len(a) != len(b) {
		return nil
	}
	out := make([]complex128, len(a))
	for index := range out {
		out[index] = a[index] * b[index]
	}
	return out
}

func subtractComplexVectors(a, b []complex128) []complex128 {
	if len(a) != len(b) {
		return nil
	}
	out := make([]complex128, len(a))
	for index := range out {
		out[index] = a[index] - b[index]
	}
	return out
}

func complexVectorsEqual(a, b []complex128) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if real(a[index]) != real(b[index]) || imag(a[index]) != imag(b[index]) {
			return false
		}
	}
	return true
}

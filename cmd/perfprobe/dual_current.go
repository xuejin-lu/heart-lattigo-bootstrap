package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const (
	dualCurrentStandardCommit = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
	dualCurrentFastCommit     = "2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f"
	dualCurrentInput13        = "d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285"
	dualCurrentMetadata13     = "9171dac4a79b97f7f916001d0ea5805f6137711527d364f8dfcd8e71ed2764f6"
	dualCurrentInput16        = "659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf"
	dualCurrentMetadata16     = "58b699119d1ca2d73cfe8812e6ae776f77d4267ff4c25e67229d7b140a6cf44f"
	dualCurrentConfig13       = "919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98"
	dualCurrentConfig16       = "16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02"
	dualCurrentQP13           = "1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b"
	dualCurrentQP16           = "6c82ca724cbfdce2083b2ef2c18d53f81e658a6f28ff283e9a671d276710ec5f"
)

type dualCurrentCheckpoint struct {
	Name               string          `json:"name"`
	State              ciphertextState `json:"state"`
	PhysicalRowLengths [][]int         `json:"physical_row_lengths_by_component"`
	ActiveRowSHA256    [][]string      `json:"active_row_sha256_by_component"`
	ActiveRowLabels    []string        `json:"active_q_rows"`
	IsNTT              bool            `json:"is_ntt"`
	IsMontgomery       bool            `json:"is_montgomery"`
	C1Nonzero          bool            `json:"c1_nonzero"`
	DecodedSHA256      string          `json:"decoded_sha256,omitempty"`
	DecodedValues      []complexValue  `json:"decoded_values,omitempty"`
	DecodeMethod       string          `json:"decode_method,omitempty"`
}

type dualCurrentC2SAudit struct {
	Applicable      bool                  `json:"applicable"`
	Reason          string                `json:"reason,omitempty"`
	MatrixGroupSize int                   `json:"first_group_matrix_count,omitempty"`
	Before          dualCurrentCheckpoint `json:"modup_before_first_c2s_group,omitempty"`
	AfterRescale    dualCurrentCheckpoint `json:"after_first_c2s_group_rescale,omitempty"`
	AfterNextMatrix dualCurrentCheckpoint `json:"after_next_c2s_matrix,omitempty"`
}

type dualCurrentDocument struct {
	SchemaVersion       string                          `json:"schema_version"`
	Mode                string                          `json:"mode"`
	Status              string                          `json:"status"`
	Timestamp           time.Time                       `json:"timestamp"`
	Profile             string                          `json:"profile"`
	Backend             string                          `json:"backend"`
	BackendCommit       string                          `json:"backend_commit"`
	BackendRef          string                          `json:"backend_ref"`
	BackendPath         string                          `json:"backend_path"`
	LattigoModulePath   string                          `json:"compiled_lattigo_module_path"`
	BackendClean        bool                            `json:"backend_clean"`
	PrimaryPath         string                          `json:"primary_path"`
	PrimaryCommit       string                          `json:"primary_commit"`
	PrimaryClean        bool                            `json:"primary_clean"`
	PrimarySourceSHA256 string                          `json:"primary_measurement_source_sha256"`
	BinarySHA256        string                          `json:"binary_sha256"`
	BuildTag            string                          `json:"build_tag"`
	GoVersion           string                          `json:"go_version"`
	OS                  string                          `json:"os"`
	Arch                string                          `json:"arch"`
	ConfigPath          string                          `json:"config_path"`
	ConfigSHA256        string                          `json:"config_sha256"`
	QPSHA256            string                          `json:"qp_sha256"`
	InputSHA256         string                          `json:"input_sha256"`
	InputMetadataSHA256 string                          `json:"input_metadata_sha256"`
	InputKind           string                          `json:"input_kind"`
	InputConstructor    string                          `json:"input_constructor"`
	InputSlots          int                             `json:"input_slots"`
	InputScale          string                          `json:"input_scale"`
	InputC1Nonzero      bool                            `json:"input_c1_nonzero"`
	PreBootstrapVsInput *numericalmetrics.SNR           `json:"pre_bootstrap_decoded_vs_original_input,omitempty"`
	Dispatch            string                          `json:"public_evaluator_dispatch"`
	Parameters          perfmeasure.EffectiveParameters `json:"effective_parameters"`
	BootstrapBudget     int                             `json:"bootstrap_budget"`
	BootstrapCalls      int                             `json:"actual_bootstrap_calls"`
	PreflightPairSHA256 string                          `json:"preflight_pair_sha256,omitempty"`
	C2SAudit            dualCurrentC2SAudit             `json:"standard_c2s_capacity_audit"`
	Checkpoints         []dualCurrentCheckpoint         `json:"checkpoints"`
	OutputVsInput       *vectorMetrics                  `json:"bootstrap_output_metrics_vs_original_input,omitempty"`
	OutputAgainstInput  *numericalmetrics.SNR           `json:"bootstrap_output_snr_vs_original_input,omitempty"`
	Limitations         []string                        `json:"limitations"`
}

type dualCurrentStageComparison struct {
	Name                       string               `json:"name"`
	LogicalStateMatched        bool                 `json:"level_scale_degree_dimensions_matched"`
	RepresentationFlagsMatched bool                 `json:"ntt_montgomery_flags_matched"`
	Comparable                 bool                 `json:"comparable"`
	DecodedBitwiseIdentical    bool                 `json:"decoded_float64_bitwise_identical"`
	Reason                     string               `json:"reason,omitempty"`
	Standard                   ciphertextState      `json:"standard_state"`
	Fast                       ciphertextState      `json:"fast_state"`
	StandardIsNTT              bool                 `json:"standard_is_ntt"`
	FastIsNTT                  bool                 `json:"fast_is_ntt"`
	StandardMontgomery         bool                 `json:"standard_is_montgomery"`
	FastMontgomery             bool                 `json:"fast_is_montgomery"`
	FastVsStd                  vectorMetrics        `json:"fast_vs_standard_error_metrics"`
	SNR                        numericalmetrics.SNR `json:"fast_vs_standard_snr"`
}

type dualCurrentBootstrapComparison struct {
	SchemaVersion           string               `json:"schema_version"`
	Mode                    string               `json:"mode"`
	Status                  string               `json:"status"`
	Timestamp               time.Time            `json:"timestamp"`
	Profile                 string               `json:"profile"`
	StandardCommit          string               `json:"standard_commit"`
	FastCommit              string               `json:"fast_commit"`
	InputSHA256             string               `json:"input_sha256"`
	ConfigSHA256            string               `json:"config_sha256"`
	SlotsCompared           int                  `json:"slots_compared"`
	ThresholdRMSE           float64              `json:"suggested_rmse_gate"`
	StandardVsInput         vectorMetrics        `json:"standard_vs_original_input"`
	StandardInputSNR        numericalmetrics.SNR `json:"standard_vs_original_input_snr"`
	FastVsInput             vectorMetrics        `json:"fast_vs_original_input"`
	FastInputSNR            numericalmetrics.SNR `json:"fast_vs_original_input_snr"`
	FastVsStandard          vectorMetrics        `json:"fast_vs_genuine_standard"`
	FastStandardSNR         numericalmetrics.SNR `json:"fast_vs_genuine_standard_snr"`
	DecodedBitwiseIdentical bool                 `json:"decoded_float64_bitwise_identical"`
	StandardState           ciphertextState      `json:"standard_output_state"`
	FastState               ciphertextState      `json:"fast_output_state"`
}

type dualCurrentPair struct {
	SchemaVersion       string                       `json:"schema_version"`
	Mode                string                       `json:"mode"`
	Status              string                       `json:"status"`
	Timestamp           time.Time                    `json:"timestamp"`
	Profile             string                       `json:"profile"`
	ConfigSHA256        string                       `json:"config_sha256"`
	InputSHA256         string                       `json:"input_sha256"`
	InputMetadataSHA256 string                       `json:"input_metadata_sha256"`
	QPSHA256            string                       `json:"qp_sha256"`
	PrimaryCommit       string                       `json:"primary_commit"`
	PrimarySourceSHA256 string                       `json:"primary_measurement_source_sha256"`
	StandardSHA256      string                       `json:"standard_preflight_sha256"`
	FastSHA256          string                       `json:"fast_preflight_sha256"`
	StandardCommit      string                       `json:"standard_commit"`
	FastCommit          string                       `json:"fast_commit"`
	Comparisons         []dualCurrentStageComparison `json:"stage_comparisons"`
	Limitations         []string                     `json:"limitations"`
}

type dualCurrentOptions struct {
	command, profile, config, backendCommit, secondaryRoot string
	out, standard, fast, preflightPair, attemptJournal     string
}

func runDualCurrent(args []string) error {
	if len(args) == 0 {
		return errors.New("expected preflight, compare-preflight, bootstrap-once, or compare-bootstrap")
	}
	opts := dualCurrentOptions{command: args[0]}
	flags := flag.NewFlagSet("dual-current "+opts.command, flag.ContinueOnError)
	flags.StringVar(&opts.profile, "profile", "", "logn13 or logn16")
	flags.StringVar(&opts.config, "config", "", "shared bootstrap config JSON")
	flags.StringVar(&opts.backendCommit, "backend-commit", "", "exact Secondary commit")
	flags.StringVar(&opts.secondaryRoot, "secondary-root", "", "clean isolated Secondary worktree")
	flags.StringVar(&opts.out, "out", "", "new result path")
	flags.StringVar(&opts.standard, "standard", "", "Standard input artifact")
	flags.StringVar(&opts.fast, "fast", "", "Fast input artifact")
	flags.StringVar(&opts.preflightPair, "preflight-pair", "", "passing compare-preflight artifact")
	flags.StringVar(&opts.attemptJournal, "attempt-journal", "", "exclusive append-only Bootstrap attempt journal")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	switch opts.command {
	case "preflight":
		return runDualCurrentPreflight(opts)
	case "compare-preflight":
		return compareDualCurrentPreflight(opts)
	case "bootstrap-once":
		return runDualCurrentBootstrapOnce(opts)
	case "compare-bootstrap":
		return compareDualCurrentBootstrap(opts)
	default:
		return fmt.Errorf("unknown dual-current command %q", opts.command)
	}
}

func runDualCurrentPreflight(opts dualCurrentOptions) error {
	if err := validateDualCurrentRunOptions(opts, false); err != nil {
		return err
	}
	residual, params, effective, values, configHash, qpHash, err := loadDualCurrentProfile(opts)
	if err != nil {
		return err
	}
	backend, provenance, err := newDualCurrentBackend(opts, params, residual, configHash, qpHash, effective)
	if err != nil {
		return err
	}
	input, err := encodeDualCurrentInput(residual, effective.LogSlots, values)
	if err != nil {
		return err
	}
	inputDecoded, _, err := decodeDualCurrentStage(backend, "input", input)
	if err != nil {
		return fmt.Errorf("decode diagnostic input: %w", err)
	}
	state, err := dualCurrentCheckpointOf(backend, "input", input, inputDecoded, true)
	if err != nil {
		return err
	}
	doc := provenance
	doc.Mode, doc.Status = "zero_call_public_stage_preflight", "RUNNING"
	doc.BootstrapBudget, doc.BootstrapCalls = 0, 0
	doc.InputSHA256 = perfmeasure.Fingerprint(values)
	doc.InputMetadataSHA256 = perfmeasure.CiphertextStateFingerprint(state.State)
	doc.InputKind = "diagnostic-encoded-c0-c1-zero-v1"
	doc.InputConstructor = "ckks.Encode at residual Level0; c0=encoded plaintext; c1=0; diagnostic-only"
	doc.InputSlots, doc.InputScale = len(values), input.Scale.Value.Text('e', 80)
	doc.InputC1Nonzero = state.C1Nonzero
	doc.Checkpoints = append(doc.Checkpoints, state)
	if err := validateDualCurrentInput(opts.profile, configHash, doc.InputSHA256, doc.InputMetadataSHA256, effective, input); err != nil {
		return err
	}

	packed, _, _, err := backend.bootstrapEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return fmt.Errorf("PackAndSwitchN1ToN2 replay: %w", err)
	}
	if len(packed) != 1 {
		return fmt.Errorf("PackAndSwitchN1ToN2 replay returned %d ciphertexts", len(packed))
	}
	scaled, _, err := backend.bootstrapEval.ScaleDown(&packed[0])
	if err != nil {
		return fmt.Errorf("ScaleDown replay: %w", err)
	}
	if err := appendDualCurrentStage(backend, &doc, "scale_down", scaled); err != nil {
		return err
	}
	modUp, err := backend.bootstrapEval.ModUp(scaled.CopyNew())
	if err != nil {
		return fmt.Errorf("ModUp replay: %w", err)
	}
	if err := appendDualCurrentStage(backend, &doc, "mod_up", modUp); err != nil {
		return err
	}
	doc.C2SAudit, err = auditDualCurrentC2SGroup(backend, modUp.CopyNew())
	if err != nil {
		return fmt.Errorf("zero-call C2S first-group audit: %w", err)
	}
	realCT, imagCT, err := backend.bootstrapEval.CoeffsToSlots(modUp.CopyNew())
	if err != nil {
		return fmt.Errorf("CoeffsToSlots replay: %w", err)
	}
	if err := appendDualCurrentStage(backend, &doc, "c2s_real", realCT); err != nil {
		return err
	}
	if imagCT == nil {
		return errors.New("canonical full-slot C2S returned no imaginary branch")
	}
	if err := appendDualCurrentStage(backend, &doc, "c2s_imag", imagCT); err != nil {
		return err
	}
	realCT, err = backend.bootstrapEval.EvalMod(realCT)
	if err != nil {
		return fmt.Errorf("EvalMod real replay: %w", err)
	}
	if err := appendDualCurrentStage(backend, &doc, "evalmod_real", realCT); err != nil {
		return err
	}
	imagCT, err = backend.bootstrapEval.EvalMod(imagCT)
	if err != nil {
		return fmt.Errorf("EvalMod imag replay: %w", err)
	}
	if err := appendDualCurrentStage(backend, &doc, "evalmod_imag", imagCT); err != nil {
		return err
	}
	s2c, err := backend.bootstrapEval.SlotsToCoeffs(realCT, imagCT)
	if err != nil {
		return fmt.Errorf("SlotsToCoeffs replay: %w", err)
	}
	if err := appendDualCurrentStage(backend, &doc, "s2c", s2c); err != nil {
		return err
	}
	doc.Status = "PREFLIGHT_PASS"
	doc.Limitations = []string{
		"Diagnostic-only encoded-c0/zero-c1 workload; not an E32 public lifecycle or security-equivalent Standard/Fast input.",
		"Public stage checkpoints are replay-derived and are not in-circuit traces.",
		"No Bootstrap was invoked in this preflight.",
	}
	return writeExclusiveJSON(opts.out, doc)
}

func compareDualCurrentPreflight(opts dualCurrentOptions) error {
	if opts.standard == "" || opts.fast == "" || opts.out == "" {
		return errors.New("compare-preflight requires --standard, --fast, and --out")
	}
	if err := ensureDistinctDualCurrentArtifacts(opts.standard, opts.fast); err != nil {
		return err
	}
	standard, standardBytes, err := readDualCurrentDocument(opts.standard)
	if err != nil {
		return fmt.Errorf("read Standard preflight: %w", err)
	}
	fast, fastBytes, err := readDualCurrentDocument(opts.fast)
	if err != nil {
		return fmt.Errorf("read Fast preflight: %w", err)
	}
	if err := validateDualCurrentPairInputs(standard, fast); err != nil {
		return err
	}
	if err := validateDualCurrentStandardCapacityAudit(standard); err != nil {
		return err
	}
	comparisons := compareDualCurrentCheckpoints(standard.Checkpoints, fast.Checkpoints)
	if len(standard.Checkpoints) != 8 || len(fast.Checkpoints) != 8 || len(comparisons) != 8 {
		return fmt.Errorf("expected 8 paired public checkpoints, got Standard=%d Fast=%d paired=%d", len(standard.Checkpoints), len(fast.Checkpoints), len(comparisons))
	}
	wantStages := []string{"input", "scale_down", "mod_up", "c2s_real", "c2s_imag", "evalmod_real", "evalmod_imag", "s2c"}
	for index, name := range wantStages {
		if standard.Checkpoints[index].Name != name || fast.Checkpoints[index].Name != name || comparisons[index].Name != name || !comparisons[index].Comparable {
			return fmt.Errorf("paired checkpoint %d is not the required %s Standard/Fast comparison", index, name)
		}
	}
	standardHash, fastHash := sha256.Sum256(standardBytes), sha256.Sum256(fastBytes)
	pair := dualCurrentPair{
		SchemaVersion: "fast-current-dual-backend-numerical-revalidation-002/v1", Mode: "zero_call_preflight_pair",
		Status: "PREFLIGHT_PASS", Timestamp: time.Now().UTC(), Profile: standard.Profile,
		ConfigSHA256: standard.ConfigSHA256, InputSHA256: standard.InputSHA256,
		InputMetadataSHA256: standard.InputMetadataSHA256, QPSHA256: standard.QPSHA256,
		PrimaryCommit: standard.PrimaryCommit, PrimarySourceSHA256: standard.PrimarySourceSHA256,
		StandardSHA256: hex.EncodeToString(standardHash[:]), FastSHA256: hex.EncodeToString(fastHash[:]),
		StandardCommit: standard.BackendCommit, FastCommit: fast.BackendCommit,
		Comparisons: comparisons,
		Limitations: []string{"Stage comparisons are replay-derived; this pair validates provenance/structure and does not claim numerical gates pass."},
	}
	return writeExclusiveJSON(opts.out, pair)
}

func runDualCurrentBootstrapOnce(opts dualCurrentOptions) error {
	if err := validateDualCurrentRunOptions(opts, true); err != nil {
		return err
	}
	if opts.preflightPair == "" || opts.attemptJournal == "" {
		return errors.New("bootstrap-once requires --preflight-pair and --attempt-journal")
	}
	pairBytes, err := os.ReadFile(opts.preflightPair)
	if err != nil {
		return err
	}
	var pair dualCurrentPair
	if err := json.Unmarshal(pairBytes, &pair); err != nil {
		return fmt.Errorf("decode preflight pair: %w", err)
	}
	if pair.Status != "PREFLIGHT_PASS" || pair.Mode != "zero_call_preflight_pair" || pair.Profile != opts.profile || pair.StandardCommit != dualCurrentStandardCommit || pair.FastCommit != dualCurrentFastCommit || pair.StandardSHA256 == "" || pair.FastSHA256 == "" {
		return errors.New("bootstrap-once requires a matching PREFLIGHT_PASS pair")
	}
	residual, params, effective, values, configHash, qpHash, err := loadDualCurrentProfile(opts)
	if err != nil {
		return err
	}
	if pair.ConfigSHA256 != configHash || pair.InputSHA256 != perfmeasure.Fingerprint(values) || pair.QPSHA256 != qpHash {
		return errors.New("preflight pair config/input fingerprint does not match this process")
	}
	backend, provenance, err := newDualCurrentBackend(opts, params, residual, configHash, qpHash, effective)
	if err != nil {
		return err
	}
	if backend.name == "standard" && pair.StandardCommit != opts.backendCommit || backend.name == "fast" && pair.FastCommit != opts.backendCommit {
		return errors.New("preflight pair does not contain this backend commit")
	}
	if provenance.PrimaryCommit != pair.PrimaryCommit || provenance.PrimarySourceSHA256 != pair.PrimarySourceSHA256 {
		return errors.New("Primary source changed after the paired preflight")
	}
	input, err := encodeDualCurrentInput(residual, effective.LogSlots, values)
	if err != nil {
		return err
	}
	inputState, err := dualCurrentCheckpointOf(backend, "input", input, nil, false)
	if err != nil {
		return err
	}
	inputMetadata := perfmeasure.CiphertextStateFingerprint(inputState.State)
	if inputMetadata != pair.InputMetadataSHA256 {
		return errors.New("Bootstrap input metadata fingerprint differs from passing paired preflight")
	}
	if err := ensureDualCurrentOutputAbsent(opts.out); err != nil {
		return err
	}
	preValues, _, err := decodeDualCurrentStage(backend, "input", input)
	if err != nil {
		return fmt.Errorf("decode pre-Bootstrap input: %w", err)
	}
	if len(preValues) != len(values) || !allFinite(preValues) {
		return errors.New("pre-Bootstrap input decode has invalid slot coverage or non-finite values")
	}
	journal, err := reserveDualCurrentAttempt(opts.attemptJournal, opts, pair, inputMetadata)
	if err != nil {
		return err
	}
	output, callErr := backend.bootstrapEval.Bootstrap(input.CopyNew())
	if callErr != nil {
		_ = finishDualCurrentAttempt(journal, "FAILED", callErr.Error())
		return fmt.Errorf("the single authorized public Bootstrap call failed and is consumed: %w", callErr)
	}
	if output == nil {
		_ = finishDualCurrentAttempt(journal, "FAILED", "Bootstrap returned nil ciphertext")
		return errors.New("the single authorized public Bootstrap call returned nil and is consumed")
	}
	decoded, _, err := decodeDualCurrentStage(backend, "final_public_output", output)
	if err != nil {
		_ = finishDualCurrentAttempt(journal, "FAILED", "decode output: "+err.Error())
		return fmt.Errorf("Bootstrap succeeded once but output decode failed (call consumed): %w", err)
	}
	for i, value := range decoded {
		if !finiteComplex(value) {
			_ = finishDualCurrentAttempt(journal, "FAILED", fmt.Sprintf("non-finite decoded slot %d", i))
			return fmt.Errorf("Bootstrap output contains non-finite slot %d (call consumed)", i)
		}
	}
	if len(decoded) != len(values) {
		_ = finishDualCurrentAttempt(journal, "FAILED", "decoded output slot count mismatch")
		return fmt.Errorf("Bootstrap output slots=%d, want %d (call consumed)", len(decoded), len(values))
	}
	final, err := dualCurrentCheckpointOf(backend, "bootstrap_output", output, decoded, true)
	if err != nil {
		_ = finishDualCurrentAttempt(journal, "FAILED", err.Error())
		return fmt.Errorf("Bootstrap output checkpoint invalid (call consumed): %w", err)
	}
	metric := numericalmetrics.Compare(values, decoded)
	outputMetrics := compareVectors(values, decoded, 1e-7)
	doc := provenance
	doc.Mode, doc.Status = "single_cold_public_bootstrap", "BOOTSTRAP_COMPLETE"
	doc.BootstrapBudget, doc.BootstrapCalls = 1, 1
	doc.InputSHA256, doc.InputMetadataSHA256 = pair.InputSHA256, inputMetadata
	doc.InputKind = "diagnostic-encoded-c0-c1-zero-v1"
	doc.InputConstructor = "ckks.Encode at residual Level0; c0=encoded plaintext; c1=0; diagnostic-only"
	doc.InputSlots, doc.InputScale, doc.InputC1Nonzero = len(values), input.Scale.Value.Text('e', 80), false
	preMetric := numericalmetrics.Compare(values, preValues)
	doc.PreBootstrapVsInput = &preMetric
	pairHash := sha256.Sum256(pairBytes)
	doc.PreflightPairSHA256 = hex.EncodeToString(pairHash[:])
	doc.Checkpoints = []dualCurrentCheckpoint{final}
	doc.OutputVsInput, doc.OutputAgainstInput = &outputMetrics, &metric
	doc.Limitations = []string{"Exactly one cold public Bootstrap call; no warmup or retry.", "Input is diagnostic-only encoded-c0/zero-c1, not a formal public E32 lifecycle."}
	if err := writeExclusiveJSON(opts.out, doc); err != nil {
		_ = finishDualCurrentAttempt(journal, "COMPLETE_OUTPUT_WRITE_FAILED", err.Error())
		return fmt.Errorf("Bootstrap was completed once but result output could not be created: %w", err)
	}
	return finishDualCurrentAttempt(journal, "COMPLETED", "")
}

func compareDualCurrentBootstrap(opts dualCurrentOptions) error {
	if opts.standard == "" || opts.fast == "" || opts.out == "" {
		return errors.New("compare-bootstrap requires --standard, --fast, and --out")
	}
	standard, _, err := readDualCurrentDocument(opts.standard)
	if err != nil {
		return err
	}
	fast, _, err := readDualCurrentDocument(opts.fast)
	if err != nil {
		return err
	}
	if err := validateDualCurrentBootstrapInputs(standard, fast, opts.standard, opts.fast); err != nil {
		return err
	}
	standardValues, fastValues := decodeDualValues(standard.Checkpoints[0].DecodedValues), decodeDualValues(fast.Checkpoints[0].DecodedValues)
	result := dualCurrentBootstrapComparison{
		SchemaVersion: "fast-current-dual-backend-numerical-revalidation-002/v1", Mode: "paired_single_cold_bootstrap",
		Status: "MEASURED", Timestamp: time.Now().UTC(), Profile: standard.Profile,
		StandardCommit: standard.BackendCommit, FastCommit: fast.BackendCommit,
		InputSHA256: standard.InputSHA256, ConfigSHA256: standard.ConfigSHA256,
		SlotsCompared: len(standardValues), ThresholdRMSE: 1e-7,
		StandardVsInput: *standard.OutputVsInput, StandardInputSNR: *standard.OutputAgainstInput,
		FastVsInput: *fast.OutputVsInput, FastInputSNR: *fast.OutputAgainstInput,
		FastVsStandard:          compareVectors(standardValues, fastValues, 1e-7),
		FastStandardSNR:         numericalmetrics.Compare(standardValues, fastValues),
		DecodedBitwiseIdentical: dualCurrentDecodedBitwiseEqual(standardValues, fastValues),
		StandardState:           standard.Checkpoints[0].State, FastState: fast.Checkpoints[0].State,
	}
	return writeExclusiveJSON(opts.out, result)
}

func validateDualCurrentRunOptions(opts dualCurrentOptions, bootstrap bool) error {
	if opts.profile != "logn13" && opts.profile != "logn16" {
		return errors.New("--profile must be logn13 or logn16")
	}
	if opts.config == "" || opts.out == "" || opts.backendCommit == "" || opts.secondaryRoot == "" {
		return errors.New("--config, --out, --backend-commit, and --secondary-root are required")
	}
	want := dualCurrentStandardCommit
	if opts.backendCommit == dualCurrentFastCommit {
		want = dualCurrentFastCommit
	} else if opts.backendCommit != dualCurrentStandardCommit {
		return fmt.Errorf("backend commit %s is not one of the pinned task sources", opts.backendCommit)
	}
	if opts.backendCommit != want {
		return errors.New("backend commit validation failed")
	}
	if bootstrap && (opts.preflightPair == "" || opts.attemptJournal == "") {
		return errors.New("bootstrap-once requires an explicit paired preflight and attempt journal")
	}
	return nil
}

func loadDualCurrentProfile(opts dualCurrentOptions) (ckks.Parameters, bootstrapping.Parameters, perfmeasure.EffectiveParameters, []complex128, string, string, error) {
	cfg, sum, err := perfmeasure.LoadConfig(opts.config)
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, perfmeasure.EffectiveParameters{}, nil, "", "", err
	}
	configHash := hex.EncodeToString(sum[:])
	wantConfig, wantLogN := dualCurrentConfig13, 13
	if opts.profile == "logn16" {
		wantConfig, wantLogN = dualCurrentConfig16, 16
	}
	if cfg.LogN != wantLogN || configHash != wantConfig {
		return ckks.Parameters{}, bootstrapping.Parameters{}, perfmeasure.EffectiveParameters{}, nil, "", "", fmt.Errorf("%s config identity mismatch: LogN=%d SHA256=%s", opts.profile, cfg.LogN, configHash)
	}
	residual, params, effective, err := perfmeasure.ParametersFromConfigWithE(cfg, 0)
	if err != nil {
		return ckks.Parameters{}, bootstrapping.Parameters{}, perfmeasure.EffectiveParameters{}, nil, "", "", err
	}
	values := perfmeasure.DeterministicInput(effective.InputSlots)
	inputHash := perfmeasure.Fingerprint(values)
	wantInput := dualCurrentInput13
	if opts.profile == "logn16" {
		wantInput = dualCurrentInput16
	}
	if inputHash != wantInput || len(values) != 1<<effective.LogSlots {
		return ckks.Parameters{}, bootstrapping.Parameters{}, perfmeasure.EffectiveParameters{}, nil, "", "", fmt.Errorf("%s canonical diagnostic input mismatch: SHA256=%s slots=%d", opts.profile, inputHash, len(values))
	}
	qpHash := perfmeasure.QPFingerprint(params.BootstrappingParameters.Q(), params.BootstrappingParameters.P())
	return residual, params, effective, values, configHash, qpHash, nil
}

func newDualCurrentBackend(opts dualCurrentOptions, params bootstrapping.Parameters, residual ckks.Parameters, configHash, qpHash string, effective perfmeasure.EffectiveParameters) (*publicBackend, dualCurrentDocument, error) {
	if err := verifyCompiledBackendSource(opts.secondaryRoot); err != nil {
		return nil, dualCurrentDocument{}, err
	}
	secondaryRoot, secondaryCommit, secondaryRef, err := cleanSecondaryState(opts.secondaryRoot, opts.backendCommit)
	if err != nil {
		return nil, dualCurrentDocument{}, fmt.Errorf("Secondary source provenance: %w", err)
	}
	if publicBackendName() != "standard" && publicBackendName() != "fast" {
		return nil, dualCurrentDocument{}, errors.New("perfprobe binary lacks a selected Standard/Fast backend tag")
	}
	if publicBackendName() == "standard" && secondaryCommit != dualCurrentStandardCommit || publicBackendName() == "fast" && secondaryCommit != dualCurrentFastCommit {
		return nil, dualCurrentDocument{}, fmt.Errorf("compiled %s backend cannot use source commit %s", publicBackendName(), secondaryCommit)
	}
	root, primaryCommit, err := cleanRepositoryState(".")
	if err != nil {
		return nil, dualCurrentDocument{}, fmt.Errorf("Primary source provenance: %w", err)
	}
	primarySource, err := primaryMeasurementSourceFingerprint(root)
	if err != nil {
		return nil, dualCurrentDocument{}, err
	}
	binaryHash, buildTag, modulePath, err := dualCurrentBinaryProvenance()
	if err != nil {
		return nil, dualCurrentDocument{}, err
	}
	backend, err := newPublicBackend(params, residual)
	if err != nil {
		return nil, dualCurrentDocument{}, err
	}
	if backend.name != publicBackendName() {
		return nil, dualCurrentDocument{}, fmt.Errorf("public evaluator dispatch selected %s in %s binary", backend.name, publicBackendName())
	}
	doc := dualCurrentDocument{
		SchemaVersion: "fast-current-dual-backend-numerical-revalidation-002/v1", Timestamp: time.Now().UTC(),
		Profile: opts.profile, Backend: backend.name, BackendCommit: secondaryCommit, BackendRef: secondaryRef,
		BackendPath: secondaryRoot, LattigoModulePath: modulePath, BackendClean: true, PrimaryPath: root, PrimaryCommit: primaryCommit,
		PrimaryClean: true, PrimarySourceSHA256: primarySource, BinarySHA256: binaryHash, BuildTag: buildTag,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
		ConfigPath: opts.config, ConfigSHA256: configHash, QPSHA256: qpHash,
		Parameters: effective, Dispatch: backend.dispatch, BootstrapBudget: 0, BootstrapCalls: 0,
	}
	return backend, doc, nil
}

func encodeDualCurrentInput(params ckks.Parameters, logSlots int, values []complex128) (*rlwe.Ciphertext, error) {
	plain, err := perfmeasure.EncodeInputPlaintext(params, logSlots, values)
	if err != nil {
		return nil, err
	}
	ct := ckks.NewCiphertext(params, 1, 0)
	*ct.MetaData = *plain.MetaData
	ct.Value[0].Copy(plain.Value)
	ct.Value[1].Zero()
	ct.IsNTT, ct.IsMontgomery = plain.IsNTT, plain.IsMontgomery
	return ct, nil
}

func validateDualCurrentInput(profile, configHash, inputHash, metadataHash string, effective perfmeasure.EffectiveParameters, ct *rlwe.Ciphertext) error {
	wantInput, wantMetadata, wantConfig := dualCurrentInput13, dualCurrentMetadata13, dualCurrentConfig13
	if profile == "logn16" {
		wantInput, wantMetadata, wantConfig = dualCurrentInput16, dualCurrentMetadata16, dualCurrentConfig16
	}
	if configHash != wantConfig || inputHash != wantInput || metadataHash != wantMetadata {
		return fmt.Errorf("%s canonical input identity mismatch: config=%s input=%s metadata=%s", profile, configHash, inputHash, metadataHash)
	}
	if ct == nil || ct.Level() != 0 || ct.Degree() != 1 || !ct.IsNTT || ct.IsMontgomery || effective.EphemeralSecretWeight != 0 {
		return errors.New("diagnostic input must be Level0 degree-one NTT non-Montgomery E=0")
	}
	if profile == "logn16" && effective.Q0Bits != 56 {
		return fmt.Errorf("LogN16 effective q0 must remain 56-bit, got %d", effective.Q0Bits)
	}
	return nil
}

func appendDualCurrentStage(backend *publicBackend, doc *dualCurrentDocument, name string, ct *rlwe.Ciphertext) error {
	decoded, method, err := decodeDualCurrentStage(backend, name, ct)
	if err != nil {
		return fmt.Errorf("decode %s: %w", name, err)
	}
	checkpoint, err := dualCurrentCheckpointOf(backend, name, ct, decoded, true)
	if err != nil {
		return err
	}
	checkpoint.DecodeMethod = method
	doc.Checkpoints = append(doc.Checkpoints, checkpoint)
	return nil
}

func dualCurrentCheckpointOf(backend *publicBackend, name string, ct *rlwe.Ciphertext, decoded []complex128, includeValues bool) (dualCurrentCheckpoint, error) {
	if ct == nil || ct.MetaData == nil || ct.Degree() != 1 || len(ct.Value) != 2 || ct.Level() < 0 || ct.Level() >= len(backend.full.Q()) {
		return dualCurrentCheckpoint{}, fmt.Errorf("%s invalid ciphertext structure/metadata", name)
	}
	rows := ct.Level() + 1
	if backend.name == "fast" {
		rows = min(rows, 4)
	}
	q := backend.full.Q()
	cp := dualCurrentCheckpoint{Name: name, State: stateOf(ct, q, rows), IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery,
		PhysicalRowLengths: make([][]int, len(ct.Value)), ActiveRowSHA256: make([][]string, len(ct.Value))}
	for row := 0; row < rows; row++ {
		cp.ActiveRowLabels = append(cp.ActiveRowLabels, fmt.Sprintf("q%d", row))
	}
	for component, poly := range ct.Value {
		cp.PhysicalRowLengths[component] = make([]int, len(poly.Coeffs))
		cp.ActiveRowSHA256[component] = make([]string, rows)
		for row, coeffs := range poly.Coeffs {
			cp.PhysicalRowLengths[component][row] = len(coeffs)
		}
		if len(poly.Coeffs) < rows {
			return dualCurrentCheckpoint{}, fmt.Errorf("%s component %d stores %d rows, needs %d authoritative rows", name, component, len(poly.Coeffs), rows)
		}
		for row := 0; row < rows; row++ {
			if len(poly.Coeffs[row]) != ct.N() {
				return dualCurrentCheckpoint{}, fmt.Errorf("%s component %d q%d row length=%d, want N=%d", name, component, row, len(poly.Coeffs[row]), ct.N())
			}
			cp.ActiveRowSHA256[component][row] = hashUint64Row(poly.Coeffs[row])
		}
		if backend.name == "fast" {
			for row := rows; row < len(poly.Coeffs); row++ {
				if len(poly.Coeffs[row]) != 0 {
					return dualCurrentCheckpoint{}, fmt.Errorf("%s Fast inactive q%d row in component %d is unexpectedly materialized", name, row, component)
				}
			}
		}
	}
	var err error
	cp.C1Nonzero, err = isCiphertextComponentNonzero(ct, 1)
	if err != nil {
		return dualCurrentCheckpoint{}, err
	}
	if backend.name == "fast" && cp.C1Nonzero {
		return dualCurrentCheckpoint{}, fmt.Errorf("%s Fast zero-c1 invariant failed", name)
	}
	if decoded != nil {
		if !allFinite(decoded) {
			return dualCurrentCheckpoint{}, fmt.Errorf("%s decode has non-finite values", name)
		}
		cp.DecodedSHA256 = perfmeasure.Fingerprint(decoded)
		if includeValues {
			cp.DecodedValues = encodeVector(decoded)
		}
	}
	return cp, nil
}

func readDualCurrentDocument(path string) (dualCurrentDocument, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return dualCurrentDocument{}, nil, err
	}
	var doc dualCurrentDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return dualCurrentDocument{}, nil, err
	}
	return doc, data, nil
}

func ensureDualCurrentOutputAbsent(path string) error {
	_, err := os.Lstat(path)
	if err == nil {
		return fmt.Errorf("refusing to overwrite existing result path %q", path)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check result path %q: %w", path, err)
	}
	return nil
}

func validateDualCurrentPairInputs(standard, fast dualCurrentDocument) error {
	if standard.Status != "PREFLIGHT_PASS" || fast.Status != "PREFLIGHT_PASS" || standard.Mode != "zero_call_public_stage_preflight" || fast.Mode != standard.Mode {
		return errors.New("both backend inputs must be completed zero-call preflights")
	}
	if standard.Backend != "standard" || fast.Backend != "fast" || standard.BackendCommit != dualCurrentStandardCommit || fast.BackendCommit != dualCurrentFastCommit {
		return errors.New("preflight artifacts do not contain the independently pinned genuine Standard and Fast sources")
	}
	if err := validateDualCurrentIndependentBackendSources(standard, fast); err != nil {
		return err
	}
	if !standard.PrimaryClean || !fast.PrimaryClean || !standard.BackendClean || !fast.BackendClean || standard.PrimaryCommit != fast.PrimaryCommit || standard.PrimarySourceSHA256 != fast.PrimarySourceSHA256 || standard.Profile != fast.Profile || standard.ConfigSHA256 != fast.ConfigSHA256 || standard.InputSHA256 != fast.InputSHA256 || standard.InputMetadataSHA256 != fast.InputMetadataSHA256 || standard.QPSHA256 != fast.QPSHA256 || !reflect.DeepEqual(standard.Parameters, fast.Parameters) {
		return errors.New("Standard/Fast preflights differ in Primary source, clean provenance, input, configuration, or effective Q/P profile")
	}
	wantStages := []string{"input", "scale_down", "mod_up", "c2s_real", "c2s_imag", "evalmod_real", "evalmod_imag", "s2c"}
	if err := validateDualCurrentDecodedCheckpoints(standard, wantStages); err != nil {
		return fmt.Errorf("Standard preflight vectors: %w", err)
	}
	if err := validateDualCurrentDecodedCheckpoints(fast, wantStages); err != nil {
		return fmt.Errorf("Fast preflight vectors: %w", err)
	}
	return nil
}

func validateDualCurrentBootstrapInputs(standard, fast dualCurrentDocument, standardPath, fastPath string) error {
	if err := ensureDistinctDualCurrentArtifacts(standardPath, fastPath); err != nil {
		return err
	}
	if standard.Status != "BOOTSTRAP_COMPLETE" || fast.Status != "BOOTSTRAP_COMPLETE" || standard.Mode != "single_cold_public_bootstrap" || fast.Mode != "single_cold_public_bootstrap" || standard.BootstrapCalls != 1 || fast.BootstrapCalls != 1 || !standard.PrimaryClean || !fast.PrimaryClean || !standard.BackendClean || !fast.BackendClean {
		return errors.New("compare-bootstrap requires exactly one completed call per backend")
	}
	if err := validateDualCurrentIndependentBackendSources(standard, fast); err != nil {
		return err
	}
	if standard.Profile != fast.Profile || standard.InputSHA256 != fast.InputSHA256 || standard.InputMetadataSHA256 != fast.InputMetadataSHA256 || standard.ConfigSHA256 != fast.ConfigSHA256 || standard.QPSHA256 != fast.QPSHA256 || standard.PrimaryCommit != fast.PrimaryCommit || standard.PrimarySourceSHA256 != fast.PrimarySourceSHA256 || !reflect.DeepEqual(standard.Parameters, fast.Parameters) || standard.PreflightPairSHA256 == "" || standard.PreflightPairSHA256 != fast.PreflightPairSHA256 {
		return errors.New("Bootstrap outputs are not from the same profile/config/input/Q/P")
	}
	if len(standard.Checkpoints) != 1 || len(fast.Checkpoints) != 1 {
		return errors.New("Bootstrap documents must each contain one final checkpoint")
	}
	if standard.OutputAgainstInput == nil || fast.OutputAgainstInput == nil || standard.OutputVsInput == nil || fast.OutputVsInput == nil {
		return errors.New("Bootstrap output-vs-input metrics are missing")
	}
	if err := validateDualCurrentDecodedCheckpoints(standard, []string{"bootstrap_output"}); err != nil {
		return fmt.Errorf("Standard Bootstrap vector: %w", err)
	}
	if err := validateDualCurrentDecodedCheckpoints(fast, []string{"bootstrap_output"}); err != nil {
		return fmt.Errorf("Fast Bootstrap vector: %w", err)
	}
	if err := validateDualCurrentBootstrapMetrics(standard); err != nil {
		return fmt.Errorf("Standard Bootstrap metrics: %w", err)
	}
	if err := validateDualCurrentBootstrapMetrics(fast); err != nil {
		return fmt.Errorf("Fast Bootstrap metrics: %w", err)
	}
	return nil
}

func validateDualCurrentIndependentBackendSources(standard, fast dualCurrentDocument) error {
	if err := validateDualCurrentBackendSource(standard, "standard"); err != nil {
		return err
	}
	if err := validateDualCurrentBackendSource(fast, "fast"); err != nil {
		return err
	}
	if filepath.Clean(standard.BackendPath) == filepath.Clean(fast.BackendPath) || filepath.Clean(standard.LattigoModulePath) == filepath.Clean(fast.LattigoModulePath) || standard.BinarySHA256 == fast.BinarySHA256 {
		return errors.New("Standard and Fast artifacts reuse the same backend source path or executable")
	}
	return nil
}

func validateDualCurrentBackendSource(doc dualCurrentDocument, wantBackend string) error {
	wantCommit, wantTag := dualCurrentStandardCommit, "perf_standard"
	if wantBackend == "fast" {
		wantCommit, wantTag = dualCurrentFastCommit, "perf_fast"
	}
	if doc.Backend != wantBackend || doc.BackendCommit != wantCommit || doc.BuildTag != wantTag {
		return fmt.Errorf("%s artifact has wrong backend source identity (backend=%q commit=%q build_tag=%q)", wantBackend, doc.Backend, doc.BackendCommit, doc.BuildTag)
	}
	if doc.BackendPath == "" || doc.LattigoModulePath == "" || filepath.Clean(doc.BackendPath) != filepath.Clean(doc.LattigoModulePath) {
		return fmt.Errorf("%s artifact backend path does not identify its compiled Lattigo module", wantBackend)
	}
	binaryHash, err := hex.DecodeString(doc.BinarySHA256)
	if err != nil || len(binaryHash) != sha256.Size {
		return fmt.Errorf("%s artifact has invalid binary SHA-256 provenance", wantBackend)
	}
	return nil
}

func ensureDistinctDualCurrentArtifacts(standardPath, fastPath string) error {
	standardInfo, err := os.Stat(standardPath)
	if err != nil {
		return fmt.Errorf("stat Standard artifact: %w", err)
	}
	fastInfo, err := os.Stat(fastPath)
	if err != nil {
		return fmt.Errorf("stat Fast artifact: %w", err)
	}
	standardAbs, err := filepath.Abs(standardPath)
	if err != nil {
		return err
	}
	fastAbs, err := filepath.Abs(fastPath)
	if err != nil {
		return err
	}
	if standardAbs == fastAbs || os.SameFile(standardInfo, fastInfo) {
		return errors.New("Standard and Fast artifact paths refer to the same file")
	}
	return nil
}

func validateDualCurrentDecodedCheckpoints(doc dualCurrentDocument, wantNames []string) error {
	wantSlots := 1 << 12
	wantLogN, wantLogSlots, wantRingN, wantQ0Bits := 13, 12, 1<<13, 55
	wantConfig, wantInput, wantQP := dualCurrentConfig13, dualCurrentInput13, dualCurrentQP13
	if doc.Profile == "logn16" {
		wantSlots = 1 << 15
		wantLogN, wantLogSlots, wantRingN, wantQ0Bits = 16, 15, 1<<16, 56
		wantConfig, wantInput, wantQP = dualCurrentConfig16, dualCurrentInput16, dualCurrentQP16
	} else if doc.Profile != "logn13" {
		return fmt.Errorf("unsupported profile %q", doc.Profile)
	}
	if doc.InputSlots != wantSlots || doc.Parameters.InputSlots != wantSlots {
		return fmt.Errorf("effective slot count mismatch: document=%d parameters=%d want=%d", doc.InputSlots, doc.Parameters.InputSlots, wantSlots)
	}
	if doc.ConfigSHA256 != wantConfig || doc.InputSHA256 != wantInput || doc.QPSHA256 != wantQP {
		return fmt.Errorf("canonical config/input/Q/P identity mismatch for %s", doc.Profile)
	}
	parameters := doc.Parameters
	if parameters.LogN != wantLogN || parameters.LogSlots != wantLogSlots || parameters.RingN != wantRingN || parameters.ResidualRingN != wantRingN || parameters.Q0Target != 55 || parameters.Q0Bits != wantQ0Bits || parameters.QPrefixRowsAtMax != 4 || len(parameters.QPrimes) != 17 || len(parameters.PPrimes) != 5 || len(parameters.QChainBits) != 17 || len(parameters.PBits) != 5 {
		return fmt.Errorf("effective CKKS parameters do not match canonical %s profile", doc.Profile)
	}
	if len(doc.Checkpoints) != len(wantNames) {
		return fmt.Errorf("checkpoint count=%d, want %d", len(doc.Checkpoints), len(wantNames))
	}
	for index, name := range wantNames {
		checkpoint := doc.Checkpoints[index]
		if checkpoint.Name != name {
			return fmt.Errorf("checkpoint %d=%q, want %q", index, checkpoint.Name, name)
		}
		if len(checkpoint.DecodedValues) != wantSlots {
			return fmt.Errorf("checkpoint %s slot count=%d, want %d", name, len(checkpoint.DecodedValues), wantSlots)
		}
		values := decodeDualValues(checkpoint.DecodedValues)
		if !allFinite(values) {
			return fmt.Errorf("checkpoint %s contains non-finite decoded values", name)
		}
		if checkpoint.DecodedSHA256 == "" || perfmeasure.Fingerprint(values) != checkpoint.DecodedSHA256 {
			return fmt.Errorf("checkpoint %s decoded values do not match decoded_sha256", name)
		}
	}
	return nil
}

func validateDualCurrentBootstrapMetrics(doc dualCurrentDocument) error {
	if doc.InputSlots <= 0 || doc.InputSHA256 == "" || doc.OutputVsInput == nil || doc.OutputAgainstInput == nil || len(doc.Checkpoints) != 1 {
		return errors.New("original input or stored output metrics are missing")
	}
	input := perfmeasure.DeterministicInput(doc.InputSlots)
	if perfmeasure.Fingerprint(input) != doc.InputSHA256 {
		return errors.New("deterministic original message does not match input_sha256")
	}
	output := decodeDualValues(doc.Checkpoints[0].DecodedValues)
	recomputedMetrics := compareVectors(input, output, 1e-7)
	recomputedSNR := numericalmetrics.Compare(input, output)
	if !reflect.DeepEqual(recomputedMetrics, *doc.OutputVsInput) || !reflect.DeepEqual(recomputedSNR, *doc.OutputAgainstInput) {
		return errors.New("stored output-vs-input metrics do not match the decoded output and canonical message")
	}
	return nil
}

func validateDualCurrentStandardCapacityAudit(standard dualCurrentDocument) error {
	audit := standard.C2SAudit
	if standard.Backend != "standard" || !audit.Applicable {
		return errors.New("genuine Standard C2S first-group capacity audit is missing")
	}
	if standard.Profile != "logn16" {
		return nil
	}
	wantLabels := func(rows int) []string {
		labels := make([]string, rows)
		for i := range labels {
			labels[i] = fmt.Sprintf("q%d", i)
		}
		return labels
	}
	checkShape := func(name string, checkpoint dualCurrentCheckpoint, level, rows int) error {
		if checkpoint.State.Level != level || checkpoint.State.QPrefixRows != rows || !reflect.DeepEqual(checkpoint.ActiveRowLabels, wantLabels(rows)) {
			return fmt.Errorf("Standard C2S audit %s state mismatch: level=%d rows=%d labels=%v", name, checkpoint.State.Level, checkpoint.State.QPrefixRows, checkpoint.ActiveRowLabels)
		}
		if len(checkpoint.PhysicalRowLengths) != 2 || len(checkpoint.ActiveRowSHA256) != 2 {
			return fmt.Errorf("Standard C2S audit %s lacks both ciphertext components", name)
		}
		for component := 0; component < 2; component++ {
			if len(checkpoint.PhysicalRowLengths[component]) < rows || len(checkpoint.ActiveRowSHA256[component]) != rows {
				return fmt.Errorf("Standard C2S audit %s component %d lacks %d active rows", name, component, rows)
			}
			for row := 0; row < rows; row++ {
				if checkpoint.PhysicalRowLengths[component][row] != standard.Parameters.RingN || checkpoint.ActiveRowSHA256[component][row] == "" {
					return fmt.Errorf("Standard C2S audit %s component %d q%d is not a full nonempty N=%d row", name, component, row, standard.Parameters.RingN)
				}
			}
		}
		return nil
	}
	if len(standard.Parameters.QPrimes) != 17 || standard.Parameters.RingN != 1<<16 || audit.MatrixGroupSize < 1 {
		return fmt.Errorf("LogN16 genuine Standard capacity audit expected 17 Q primes (Level16) and RingN=65536; got Q=%d N=%d group=%d", len(standard.Parameters.QPrimes), standard.Parameters.RingN, audit.MatrixGroupSize)
	}
	if err := checkShape("ModUp", audit.Before, 16, 17); err != nil {
		return err
	}
	if err := checkShape("group0 Rescale", audit.AfterRescale, 15, 16); err != nil {
		return err
	}
	return checkShape("next matrix", audit.AfterNextMatrix, 15, 16)
}

func compareDualCurrentCheckpoints(standard, fast []dualCurrentCheckpoint) []dualCurrentStageComparison {
	fastByName := make(map[string]dualCurrentCheckpoint, len(fast))
	for _, checkpoint := range fast {
		fastByName[checkpoint.Name] = checkpoint
	}
	comparisons := make([]dualCurrentStageComparison, 0, len(standard))
	for _, std := range standard {
		got, ok := fastByName[std.Name]
		result := dualCurrentStageComparison{Name: std.Name, Standard: std.State}
		if !ok {
			result.Reason = "Fast checkpoint missing"
			comparisons = append(comparisons, result)
			continue
		}
		result.Fast = got.State
		result.StandardIsNTT, result.FastIsNTT = std.IsNTT, got.IsNTT
		result.StandardMontgomery, result.FastMontgomery = std.IsMontgomery, got.IsMontgomery
		result.LogicalStateMatched = dualCurrentStageStateMatches(std, got)
		result.RepresentationFlagsMatched = std.IsNTT == got.IsNTT && std.IsMontgomery == got.IsMontgomery
		if !result.LogicalStateMatched {
			result.Reason = "logical Level/Scale/Degree/dimensions differ"
		} else if std.DecodedValues == nil || got.DecodedValues == nil {
			result.Reason = "decoded stage values missing"
		} else if len(std.DecodedValues) != len(got.DecodedValues) {
			result.Reason = "decoded slot count differs"
		} else {
			result.Comparable = true
			standardValues, fastValues := decodeDualValues(std.DecodedValues), decodeDualValues(got.DecodedValues)
			result.DecodedBitwiseIdentical = dualCurrentDecodedBitwiseEqual(standardValues, fastValues)
			result.FastVsStd = compareVectors(standardValues, fastValues, 1e-7)
			result.SNR = numericalmetrics.Compare(standardValues, fastValues)
			if !result.RepresentationFlagsMatched {
				result.Reason = "backend-specific NTT/Montgomery flags were decoded with their respective backend adapters"
			}
		}
		comparisons = append(comparisons, result)
	}
	return comparisons
}

func dualCurrentDecodedBitwiseEqual(standard, fast []complex128) bool {
	if len(standard) == 0 || len(standard) != len(fast) {
		return false
	}
	for index := range standard {
		if math.Float64bits(real(standard[index])) != math.Float64bits(real(fast[index])) || math.Float64bits(imag(standard[index])) != math.Float64bits(imag(fast[index])) {
			return false
		}
	}
	return true
}

func dualCurrentStageStateMatches(standard, fast dualCurrentCheckpoint) bool {
	a, b := standard.State, fast.State
	return a.Level == b.Level && a.Degree == b.Degree && a.Scale == b.Scale &&
		a.LogRows == b.LogRows && a.LogCols == b.LogCols
}

func decodeDualValues(values []complexValue) []complex128 {
	out := make([]complex128, len(values))
	for i, value := range values {
		out[i] = complex(value.Real, value.Imag)
	}
	return out
}

func reserveDualCurrentAttempt(path string, opts dualCurrentOptions, pair dualCurrentPair, inputMetadata string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fmt.Errorf("reserve single Bootstrap attempt: %w", err)
	}
	entry := map[string]any{"timestamp": time.Now().UTC(), "state": "RESERVED", "profile": opts.profile, "backend_commit": opts.backendCommit, "preflight_pair_sha256": pairHashOf(opts.preflightPair), "input_metadata_sha256": inputMetadata, "budget": 1}
	if err := json.NewEncoder(file).Encode(entry); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

func pairHashOf(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "unreadable"
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func finishDualCurrentAttempt(file *os.File, state, message string) error {
	if file == nil {
		return errors.New("nil attempt journal")
	}
	writer := bufio.NewWriter(file)
	if err := json.NewEncoder(writer).Encode(map[string]any{"timestamp": time.Now().UTC(), "state": state, "message": message}); err != nil {
		file.Close()
		return err
	}
	if err := writer.Flush(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func dualCurrentBinaryProvenance() (string, string, string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", "", "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", "", err
	}
	sum := sha256.Sum256(data)
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", "", "", errors.New("Go build info unavailable")
	}
	tag := ""
	modulePath := ""
	for _, setting := range info.Settings {
		if setting.Key == "-tags" {
			tag = setting.Value
		}
	}
	for _, dependency := range info.Deps {
		if dependency.Path == "github.com/tuneinsight/lattigo/v6" && dependency.Replace != nil {
			modulePath = dependency.Replace.Path
		}
	}
	if !strings.Contains(tag, "perf_"+publicBackendName()) {
		return "", "", "", fmt.Errorf("binary build tag %q does not identify %s", tag, publicBackendName())
	}
	if modulePath == "" {
		return "", "", "", errors.New("Go build info does not contain the local Lattigo replacement path")
	}
	return hex.EncodeToString(sum[:]), tag, modulePath, nil
}

func finiteComplex(value complex128) bool {
	return !math.IsNaN(real(value)) && !math.IsInf(real(value), 0) && !math.IsNaN(imag(value)) && !math.IsInf(imag(value), 0)
}

func dualCurrentAuditShape(backend *publicBackend, name string, ct *rlwe.Ciphertext) (dualCurrentCheckpoint, error) {
	return dualCurrentCheckpointOf(backend, name, ct, nil, false)
}

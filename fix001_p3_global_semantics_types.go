package main

import (
	"math"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

type PSGlobalMetric struct {
	MaxComponent float64 `json:"max_component_abs"`
	MaxComplex   float64 `json:"max_abs_complex"`
	MeanComplex  float64 `json:"mean_abs_complex"`
	WorstIndex   int     `json:"worst_index"`
	WorstPart    string  `json:"worst_component"`
	Pass         bool    `json:"pass_threshold"`
	Threshold    float64 `json:"threshold"`
}

type PSGlobalCheckpoint struct {
	ID                   string                       `json:"id"`
	Round                int                          `json:"round,omitempty"`
	Block                int                          `json:"block,omitempty"`
	Node                 string                       `json:"node,omitempty"`
	Operation            string                       `json:"operation"`
	Level                int                          `json:"level"`
	Degree               int                          `json:"degree"`
	Scale                string                       `json:"scale"`
	LocalConsistency     *PSGlobalMetric              `json:"local_consistency,omitempty"`
	SourceBacked         *PSGlobalMetric              `json:"source_backed"`
	ExpectedSubtree      string                       `json:"expected_subtree"`
	ExpectedSubtreeHash  string                       `json:"expected_subtree_hash"`
	Rows                 []FinalizationRowFingerprint `json:"q0_q1_rows"`
	NormalizationApplied bool                         `json:"metadata_normalization_applied,omitempty"`
	BeforeNormalization  *PSGlobalMetric              `json:"before_normalization_source_error,omitempty"`
	AfterNormalization   *PSGlobalMetric              `json:"after_normalization_source_error,omitempty"`
	ErrorDelta           float64                      `json:"source_error_delta,omitempty"`
	RatioToPrevious      float64                      `json:"source_error_ratio_to_previous,omitempty"`
	actual               []complex128                 `json:"-"`
	expected             []complex128                 `json:"-"`
	ciphertext           *rlwe.Ciphertext             `json:"-"`
}

type PSGiantStepRecord struct {
	Round       int                  `json:"round"`
	XPower      int                  `json:"x_power"`
	Checkpoints []PSGlobalCheckpoint `json:"checkpoints"`
	Final       PSGlobalCheckpoint   `json:"final_parent"`
}

type PSGlobalProbeResult struct {
	SchemaVersion           string               `json:"schema_version"`
	ScaleBits               int                  `json:"scale_bits"`
	Scale                   string               `json:"scale"`
	PowerChecks             []PSGlobalCheckpoint `json:"generated_powers"`
	BabyBlocks              []PSGlobalCheckpoint `json:"baby_blocks"`
	GiantSteps              []PSGiantStepRecord  `json:"giant_steps"`
	Root                    PSGlobalCheckpoint   `json:"pre_final_root"`
	DirectOracleAgreement   *PSGlobalMetric      `json:"ps_root_vs_direct_oracle"`
	DirectWholeOracle       *PSGlobalMetric      `json:"direct_whole_polynomial_oracle"`
	ExpectedRootDescription string               `json:"expected_root_description"`
	HashMatch               bool                 `json:"authoritative_r0_hash_match"`
	FirstFailure            string               `json:"first_source_backed_failure"`
	FirstThresholdCrossing  string               `json:"first_threshold_crossing"`
	Classification          string               `json:"classification"`
	Notes                   []string             `json:"notes,omitempty"`
}

type PSGlobalDiagnosticResult struct {
	SchemaVersion       string                 `json:"schema_version"`
	Timestamp           time.Time              `json:"timestamp"`
	Primary             RepositoryMetadata     `json:"primary_repository"`
	Lattigo             RepositoryMetadata     `json:"lattigo_repository"`
	Environment         EnvironmentMetadata    `json:"environment"`
	Config              BootstrapConfig        `json:"config"`
	Parameters          ExperimentParameters   `json:"effective_parameters"`
	Workload            CorrectnessWorkload    `json:"workload"`
	Canonical           PSGlobalProbeResult    `json:"canonical_2^91"`
	Confirmation        PSGlobalProbeResult    `json:"confirmation_2^86"`
	FirstSupportedCause string                 `json:"first_supported_cause"`
	OracleMethod        string                 `json:"oracle_method"`
	Threshold           float64                `json:"threshold"`
	Validation          map[string]interface{} `json:"validation"`
}

func psGlobalMetric(expected, actual []complex128) *PSGlobalMetric {
	m := &PSGlobalMetric{Threshold: correctnessThreshold, WorstIndex: -1}
	if len(expected) != len(actual) || len(expected) == 0 {
		return m
	}
	var sum float64
	for i := range expected {
		d := actual[i] - expected[i]
		r, im := math.Abs(real(d)), math.Abs(imag(d))
		component, part := r, "real"
		if im > component {
			component, part = im, "imag"
		}
		abs := math.Hypot(real(d), imag(d))
		if abs > m.MaxComplex {
			m.MaxComplex = abs
		}
		if component > m.MaxComponent {
			m.MaxComponent, m.WorstIndex, m.WorstPart = component, i, part
		}
		sum += abs
	}
	m.MeanComplex = sum / float64(len(expected))
	m.Pass = m.MaxComponent <= m.Threshold
	return m
}

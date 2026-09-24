package main

import (
	"fmt"
	"math/big"
	"sync"
	"time"
)

const FastMeasurementSchemaVersion = "fast-measurement.v1"

type FastMeasurementMode string

const (
	FastMeasurementOff   FastMeasurementMode = "OFF"
	FastMeasurementLight FastMeasurementMode = "LIGHT"
	FastMeasurementFull  FastMeasurementMode = "FULL"
)

type FastMeasurementPurpose string

const (
	FastMeasurementPurposeDebugVerification FastMeasurementPurpose = "debug_verification"
	FastMeasurementPurposeMLCalibration     FastMeasurementPurpose = "ml_calibration"
)

type FastInvariantID string

const (
	FastInvariantLogicalCongruence      FastInvariantID = "logical_congruence"
	FastInvariantResidueConsistency     FastInvariantID = "residue_consistency"
	FastInvariantCenteredUnique         FastInvariantID = "centered_unique"
	FastInvariantScaleTransition        FastInvariantID = "scale_transition"
	FastInvariantLogicalLevelTransition FastInvariantID = "logical_level_transition"
	FastInvariantStorageIdentity        FastInvariantID = "storage_identity"
	FastInvariantRescaleIntegerResult   FastInvariantID = "rescale_integer_result"
	FastInvariantModUpCanonicalization  FastInvariantID = "modup_canonicalization"
)

type FastMeasurementStatus string

const (
	FastMeasurementPass        FastMeasurementStatus = "pass"
	FastMeasurementFail        FastMeasurementStatus = "fail"
	FastMeasurementUnavailable FastMeasurementStatus = "unavailable"
)

type FastMeasurementProvenance struct {
	RunID           string    `json:"run_id,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	PrimaryCommit   string    `json:"primary_commit,omitempty"`
	PrimaryRef      string    `json:"primary_ref,omitempty"`
	PrimaryDirty    *bool     `json:"primary_dirty,omitempty"`
	SecondaryCommit string    `json:"secondary_commit,omitempty"`
	SecondaryRef    string    `json:"secondary_ref,omitempty"`
	SecondaryDirty  *bool     `json:"secondary_dirty,omitempty"`
	Config          string    `json:"config,omitempty"`
	GoVersion       string    `json:"go_version,omitempty"`
	OS              string    `json:"os,omitempty"`
	Architecture    string    `json:"architecture,omitempty"`
}

type FastLogicalState struct {
	Level        int     `json:"level"`
	Scale        string  `json:"scale,omitempty"`
	ScaleLog2    float64 `json:"scale_log2,omitempty"`
	Degree       int     `json:"degree,omitempty"`
	IsNTT        bool    `json:"is_ntt,omitempty"`
	IsMontgomery bool    `json:"is_montgomery,omitempty"`
}

type FastStorageState struct {
	ActiveStorageWidth        int    `json:"active_storage_width"`
	ActiveModulusBitLengths   []int  `json:"active_modulus_bit_lengths,omitempty"`
	StorageProductBitLength   int    `json:"storage_product_bit_length,omitempty"`
	MaxAbsXBitLength          int    `json:"max_abs_x_bit_length,omitempty"`
	MaxAbsXDecimal            string `json:"max_abs_x_decimal,omitempty"`
	ResidueConsistencyChecked *bool  `json:"residue_consistency_checked,omitempty"`
	SampleCoefficientIndices  []int  `json:"sample_coefficient_indices,omitempty"`
}

type FastOperationTransition struct {
	LogicalDivisor      string `json:"logical_divisor,omitempty"`
	ExpectedLevelDelta  *int   `json:"expected_level_delta,omitempty"`
	ExpectedScaleRule   string `json:"expected_scale_rule,omitempty"`
	ExpectedStorageRule string `json:"expected_storage_rule,omitempty"`
}

type FastIntegerCheckSummary struct {
	CheckedCount    int    `json:"checked_count"`
	MismatchCount   int    `json:"mismatch_count"`
	FirstMismatchAt *int   `json:"first_mismatch_at,omitempty"`
	Expected        string `json:"expected,omitempty"`
	Observed        string `json:"observed,omitempty"`
}

type FastIntegerSample struct {
	Index int    `json:"index"`
	Value string `json:"value"`
}

type FastCenteredCRTEvidence struct {
	CoefficientCount int                      `json:"coefficient_count"`
	MaxAbsX          string                   `json:"max_abs_x,omitempty"`
	NegativeCount    int                      `json:"negative_count"`
	Samples          []FastIntegerSample      `json:"samples,omitempty"`
	ResidueCheck     *FastIntegerCheckSummary `json:"residue_consistency,omitempty"`
}

type FastRescaleOracleEvidence struct {
	LogicalDivisor string                   `json:"logical_divisor"`
	InputLevel     int                      `json:"input_level"`
	OutputLevel    int                      `json:"output_level"`
	InputScale     string                   `json:"input_scale,omitempty"`
	OutputScale    string                   `json:"output_scale,omitempty"`
	IntegerResult  *FastIntegerCheckSummary `json:"integer_result,omitempty"`
}

type FastModUpOracleEvidence struct {
	LogicalModulus          string                   `json:"logical_modulus"`
	CanonicalRepresentative *FastIntegerCheckSummary `json:"canonical_representative,omitempty"`
	ExtendedResidues        *FastIntegerCheckSummary `json:"extended_residues,omitempty"`
}

type FastExactOracleEvidence struct {
	CenteredCRT         *FastCenteredCRTEvidence   `json:"centered_crt,omitempty"`
	LogicalCongruence   *FastIntegerCheckSummary   `json:"logical_congruence,omitempty"`
	StorageRoundTrip    *FastIntegerCheckSummary   `json:"storage_round_trip,omitempty"`
	ContractionIdentity *FastIntegerCheckSummary   `json:"contraction_identity,omitempty"`
	Rescale             *FastRescaleOracleEvidence `json:"rescale,omitempty"`
	ModUp               *FastModUpOracleEvidence   `json:"modup,omitempty"`
}

type FastCapacityResult struct {
	Unique                    bool     `json:"centered_unique"`
	BoundBitLength            int      `json:"bound_bit_length"`
	StorageProductBitLength   int      `json:"storage_product_bit_length"`
	CenteredCapacityBitLength int      `json:"centered_capacity_bit_length"`
	CenteredCapacityLog2      float64  `json:"centered_capacity_log2"`
	HeadroomBits              *float64 `json:"headroom_bits,omitempty"`
	UnboundedHeadroom         bool     `json:"unbounded_headroom,omitempty"`
}

type FastCenteredCRTResult struct {
	Value        *big.Int `json:"-"`
	Product      *big.Int `json:"-"`
	Negative     bool     `json:"negative"`
	AbsBitLength int      `json:"abs_bit_length"`
}

type FastContractionResult struct {
	Allowed       bool       `json:"allowed"`
	TargetProduct *big.Int   `json:"-"`
	Residues      []*big.Int `json:"-"`
	RoundTrip     *big.Int   `json:"-"`
}

type FastInvariantCheck struct {
	ID           FastInvariantID       `json:"id"`
	Status       FastMeasurementStatus `json:"status"`
	RequiredMode FastMeasurementMode   `json:"required_mode,omitempty"`
	Expected     string                `json:"expected,omitempty"`
	Observed     string                `json:"observed,omitempty"`
	Details      string                `json:"details,omitempty"`
}

type FastMeasurementAvailability struct {
	Measurement string `json:"measurement"`
	Status      string `json:"status"`
	Reason      string `json:"reason,omitempty"`
}

type FastSemanticSummary struct {
	Signal        *FastDistributionStats `json:"signal,omitempty"`
	Error         *FastDistributionStats `json:"error,omitempty"`
	SignalSamples []FastNumericSample    `json:"signal_samples,omitempty"`
	ErrorSamples  []FastNumericSample    `json:"error_samples,omitempty"`
	Threshold     *float64               `json:"threshold,omitempty"`
	ThresholdPass *bool                  `json:"threshold_pass,omitempty"`
	Metric        string                 `json:"metric,omitempty"`
	EndToEnd      bool                   `json:"end_to_end,omitempty"`
}

type FastMLCalibrationContext struct {
	ModelLayerID          string `json:"model_layer_id,omitempty"`
	ModelOperator         string `json:"model_operator,omitempty"`
	TensorRole            string `json:"tensor_role,omitempty"`
	Branch                string `json:"branch,omitempty"`
	BTSSiteID             string `json:"bts_site_id,omitempty"`
	SemanticBoundary      string `json:"semantic_boundary,omitempty"`
	PairID                string `json:"pair_id,omitempty"`
	ReferenceCheckpointID string `json:"reference_checkpoint_id,omitempty"`
}

type FastMeasurementCheckpoint struct {
	OperationID   string                        `json:"operation_id"`
	Stage         string                        `json:"stage"`
	Sequence      uint64                        `json:"sequence"`
	BeforeState   *FastLogicalState             `json:"before_state,omitempty"`
	AfterState    *FastLogicalState             `json:"after_state,omitempty"`
	BeforeStorage *FastStorageState             `json:"before_storage,omitempty"`
	AfterStorage  *FastStorageState             `json:"after_storage,omitempty"`
	Logical       *FastLogicalState             `json:"logical,omitempty"`
	Transition    *FastOperationTransition      `json:"transition,omitempty"`
	Storage       *FastStorageState             `json:"storage,omitempty"`
	Capacity      *FastCapacityResult           `json:"capacity,omitempty"`
	ExactOracles  *FastExactOracleEvidence      `json:"exact_oracles,omitempty"`
	Invariants    []FastInvariantCheck          `json:"invariants,omitempty"`
	Availability  []FastMeasurementAvailability `json:"availability,omitempty"`
	Semantic      *FastSemanticSummary          `json:"semantic,omitempty"`
	MLCalibration *FastMLCalibrationContext     `json:"ml_calibration,omitempty"`
	Notes         []string                      `json:"notes,omitempty"`
}

type FastMeasurementSummary struct {
	CheckpointCount           int      `json:"checkpoint_count"`
	FirstFailedCheckpoint     string   `json:"first_failed_checkpoint,omitempty"`
	FirstFailedInvariant      string   `json:"first_failed_invariant,omitempty"`
	MinimumHeadroomBits       *float64 `json:"minimum_headroom_bits,omitempty"`
	MinimumHeadroomCheckpoint string   `json:"minimum_headroom_checkpoint,omitempty"`
	E2EError                  *float64 `json:"e2e_error,omitempty"`
}

type FastMeasurementRun struct {
	SchemaVersion string                      `json:"schema_version"`
	Provenance    FastMeasurementProvenance   `json:"provenance"`
	Mode          FastMeasurementMode         `json:"mode"`
	Purpose       []FastMeasurementPurpose    `json:"purpose,omitempty"`
	Checkpoints   []FastMeasurementCheckpoint `json:"checkpoints,omitempty"`
	Summary       FastMeasurementSummary      `json:"summary"`
}

// FastMeasurementCollector provides a lazy, mode-aware capture boundary.
// Capture callbacks are not invoked in OFF mode, so expensive oracle work can
// be kept entirely out of that path.
type FastMeasurementCollector struct {
	mu          sync.Mutex
	mode        FastMeasurementMode
	provenance  FastMeasurementProvenance
	purposes    []FastMeasurementPurpose
	checkpoints []FastMeasurementCheckpoint
}

func NewFastMeasurementCollector(mode FastMeasurementMode, provenance FastMeasurementProvenance, purposes ...FastMeasurementPurpose) (*FastMeasurementCollector, error) {
	if mode != FastMeasurementOff && mode != FastMeasurementLight && mode != FastMeasurementFull {
		return nil, fmt.Errorf("unknown Fast measurement mode %q", mode)
	}
	if mode == FastMeasurementOff {
		return nil, nil
	}
	return &FastMeasurementCollector{mode: mode, provenance: provenance, purposes: append([]FastMeasurementPurpose(nil), purposes...)}, nil
}

func (collector *FastMeasurementCollector) Mode() FastMeasurementMode {
	if collector == nil {
		return FastMeasurementOff
	}
	return collector.mode
}

func (collector *FastMeasurementCollector) Capture(build func(FastMeasurementMode) FastMeasurementCheckpoint) {
	if collector == nil || collector.mode == FastMeasurementOff || build == nil {
		return
	}
	checkpoint := build(collector.mode)
	if collector.mode == FastMeasurementLight {
		filtered := checkpoint.Invariants[:0]
		for _, check := range checkpoint.Invariants {
			if fastInvariantRequiredMode(check) == FastMeasurementFull {
				checkpoint.Availability = append(checkpoint.Availability, FastMeasurementAvailability{
					Measurement: string(check.ID), Status: string(FastMeasurementUnavailable), Reason: "requires FULL mode",
				})
				continue
			}
			filtered = append(filtered, check)
		}
		checkpoint.Invariants = filtered
		if checkpoint.ExactOracles != nil {
			checkpoint.ExactOracles = nil
			checkpoint.Availability = append(checkpoint.Availability, FastMeasurementAvailability{
				Measurement: "exact_oracles", Status: string(FastMeasurementUnavailable), Reason: "requires FULL mode",
			})
		}
		if checkpoint.Semantic != nil {
			checkpoint.Semantic = nil
			checkpoint.Availability = append(checkpoint.Availability, FastMeasurementAvailability{
				Measurement: "semantic_statistics", Status: string(FastMeasurementUnavailable), Reason: "requires FULL mode",
			})
		}
		if checkpoint.Storage != nil {
			fastStripFullStorageDetails(checkpoint.Storage)
		}
		if checkpoint.BeforeStorage != nil {
			fastStripFullStorageDetails(checkpoint.BeforeStorage)
		}
		if checkpoint.AfterStorage != nil {
			fastStripFullStorageDetails(checkpoint.AfterStorage)
		}
	}
	collector.mu.Lock()
	if checkpoint.Sequence == 0 {
		checkpoint.Sequence = uint64(len(collector.checkpoints) + 1)
	}
	collector.checkpoints = append(collector.checkpoints, checkpoint)
	collector.mu.Unlock()
}

func fastInvariantRequiredMode(check FastInvariantCheck) FastMeasurementMode {
	if check.RequiredMode != "" {
		return check.RequiredMode
	}
	switch check.ID {
	case FastInvariantCenteredUnique, FastInvariantScaleTransition, FastInvariantLogicalLevelTransition:
		return FastMeasurementLight
	case FastInvariantLogicalCongruence, FastInvariantResidueConsistency, FastInvariantStorageIdentity, FastInvariantRescaleIntegerResult, FastInvariantModUpCanonicalization:
		return FastMeasurementFull
	default:
		return FastMeasurementFull
	}
}

func fastStripFullStorageDetails(storage *FastStorageState) {
	storage.MaxAbsXDecimal = ""
	storage.ResidueConsistencyChecked = nil
}

func (collector *FastMeasurementCollector) Run() FastMeasurementRun {
	if collector == nil {
		return FastMeasurementRun{SchemaVersion: FastMeasurementSchemaVersion, Mode: FastMeasurementOff}
	}
	collector.mu.Lock()
	checkpoints := append([]FastMeasurementCheckpoint(nil), collector.checkpoints...)
	collector.mu.Unlock()
	run := FastMeasurementRun{
		SchemaVersion: FastMeasurementSchemaVersion, Provenance: collector.provenance,
		Mode: collector.mode, Purpose: append([]FastMeasurementPurpose(nil), collector.purposes...),
		Checkpoints: checkpoints,
	}
	run.Summary = FastSummarizeMeasurement(run.Checkpoints)
	return run
}

func FastSummarizeMeasurement(checkpoints []FastMeasurementCheckpoint) FastMeasurementSummary {
	summary := FastMeasurementSummary{CheckpointCount: len(checkpoints)}
	for _, checkpoint := range checkpoints {
		if summary.FirstFailedInvariant == "" {
			for _, check := range checkpoint.Invariants {
				if check.Status == FastMeasurementFail {
					summary.FirstFailedCheckpoint = checkpoint.OperationID
					summary.FirstFailedInvariant = string(check.ID)
					break
				}
			}
		}
		if checkpoint.Capacity != nil && checkpoint.Capacity.HeadroomBits != nil {
			if summary.MinimumHeadroomBits == nil || *checkpoint.Capacity.HeadroomBits < *summary.MinimumHeadroomBits {
				value := *checkpoint.Capacity.HeadroomBits
				summary.MinimumHeadroomBits = &value
				summary.MinimumHeadroomCheckpoint = checkpoint.OperationID
			}
		}
		if checkpoint.Semantic != nil && checkpoint.Semantic.EndToEnd && checkpoint.Semantic.Error != nil {
			value := checkpoint.Semantic.Error.AbsMax
			if summary.E2EError == nil || value > *summary.E2EError {
				summary.E2EError = &value
			}
		}
	}
	return summary
}

//go:build !lattigo_standard

package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

const (
	requiredFIX001P3Q012ClosurePrimary   = "997caf2caa9657812cc2c8b9b9821ba9dff71bf4"
	requiredFIX001P3Q012ClosureSecondary = "7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797"
	q012ClosurePlanScaleExponent         = 92
	q012ClosureThreshold                 = 1e-2
)

type q012ClosureSnapshot struct {
	Name               string  `json:"name"`
	Level              int     `json:"level"`
	Scale              string  `json:"scale"`
	Degree             int     `json:"degree"`
	IsNTT              bool    `json:"is_ntt"`
	IsMontgomery       bool    `json:"is_montgomery"`
	Q01CapacityRatio   float64 `json:"q01_capacity_ratio"`
	Q012CapacityRatio  float64 `json:"q012_capacity_ratio"`
	Q01CenteredUnique  bool    `json:"q01_centered_unique"`
	Q012CenteredUnique bool    `json:"q012_centered_unique"`
}

type q012ClosureRows struct {
	RowsEqual              bool            `json:"rows_equal"`
	Q0RowsEqual            bool            `json:"q0_rows_equal"`
	Q1RowsEqual            bool            `json:"q1_rows_equal"`
	Q2RowsEqual            bool            `json:"q2_rows_equal"`
	FirstMismatchComponent int             `json:"first_mismatch_component"`
	FirstMismatchLimb      int             `json:"first_mismatch_limb"`
	FirstMismatchIndex     int             `json:"first_mismatch_index"`
	SemanticResidual       *PSGlobalMetric `json:"semantic_residual,omitempty"`
}

type q012ClosureOperationProof struct {
	Operation         string               `json:"operation"`
	SourceSnapshot    *q012ClosureSnapshot `json:"source_snapshot,omitempty"`
	OracleMetadata    string               `json:"oracle_metadata,omitempty"`
	CandidateMetadata string               `json:"candidate_metadata,omitempty"`
	Rows              q012ClosureRows      `json:"rows"`
	MetadataMatch     bool                 `json:"metadata_match"`
	Valid             bool                 `json:"valid"`
	Reason            string               `json:"reason,omitempty"`
}

type q012ClosureAlignment struct {
	Operation            string `json:"operation"`
	LeftScale            string `json:"left_scale"`
	RightScale           string `json:"right_scale"`
	ExactRatio           string `json:"exact_ratio"`
	IntegerRatio         string `json:"integer_ratio"`
	FractionalDifference string `json:"fractional_difference"`
	IntegerRatioExact    bool   `json:"integer_ratio_exact"`
	SourcePolicy         string `json:"source_policy"`
	AlignedRowsMatch     bool   `json:"aligned_rows_match"`
}

type q012ClosureG0Proof struct {
	Branch                    string                    `json:"branch"`
	Snapshots                 []q012ClosureSnapshot     `json:"source_snapshots"`
	Product                   q012ClosureOperationProof `json:"product_c0_c1_c2_proof"`
	ProductSourceQ01RowsMatch bool                      `json:"product_source_q01_rows_match"`
	Alignment                 q012ClosureAlignment      `json:"alignment"`
	Add                       q012ClosureOperationProof `json:"add_c0_c1_c2_proof"`
	AddSourceQ01RowsMatch     bool                      `json:"add_source_q01_rows_match"`
	DContractionRowsMatch     bool                      `json:"d_contraction_rows_match"`
	DContractionMetadataMatch bool                      `json:"d_contraction_metadata_match"`
	Q012CenteredUnique        bool                      `json:"q012_centered_unique"`
	Valid                     bool                      `json:"valid"`
	Reason                    string                    `json:"reason,omitempty"`
}

type q012ClosurePathProof struct {
	Branch              string                    `json:"branch"`
	DependencyPath      []string                  `json:"dependency_path"`
	G3Add               q012ClosureOperationProof `json:"g3_add_proof"`
	G3Alignment         q012ClosureAlignment      `json:"g3_alignment"`
	F0Root              q012ClosureOperationProof `json:"f0_root_proof"`
	F0Relinearized      bool                      `json:"f0_relinearized"`
	F0FinalRescale      q012ClosureOperationProof `json:"f0_final_rescale_proof"`
	PostRescaleContract q012ClosureRows           `json:"post_rescale_q01_contraction"`
	Q012CenteredUnique  bool                      `json:"q012_centered_unique_throughout"`
	Q01Contracted       bool                      `json:"q01_contracted_only_after_uniqueness"`
	Valid               bool                      `json:"valid"`
	Reason              string                    `json:"reason,omitempty"`
}

type q012ClosureBranchResult struct {
	G0       q012ClosureG0Proof   `json:"g0"`
	Path     q012ClosurePathProof `json:"path"`
	G0Output *rlwe.Ciphertext     `json:"-"`
	Final    *rlwe.Ciphertext     `json:"-"`
}

type q012ClosureResult struct {
	SchemaVersion               string                                       `json:"schema_version"`
	Timestamp                   time.Time                                    `json:"timestamp"`
	Primary                     RepositoryMetadata                           `json:"primary_repository"`
	Lattigo                     RepositoryMetadata                           `json:"lattigo_repository"`
	Environment                 EnvironmentMetadata                          `json:"environment"`
	Config                      BootstrapConfig                              `json:"config"`
	Provenance                  map[string]interface{}                       `json:"provenance"`
	Controls                    map[string]*psLocalizationDownstreamEvidence `json:"b_d_controls"`
	DCalibration                map[string]q012ClosureBranchResult           `json:"d_q056_g0_calibration"`
	BSourceFaithful             map[string]q012ClosureBranchResult           `json:"b_q055_source_faithful"`
	Downstream                  map[string]*psLocalizationDownstreamEvidence `json:"downstream"`
	ResetAttribution            []map[string]interface{}                     `json:"reset_attribution"`
	MinimalCausalResetSet       []string                                     `json:"minimal_causal_reset_set"`
	MinimalCausalResetSetProven bool                                         `json:"minimal_causal_reset_set_proven"`
	Classification              string                                       `json:"classification"`
	FirstRemainingBlocker       string                                       `json:"first_remaining_blocker"`
	ProductionReadiness         string                                       `json:"production_readiness"`
	ArchitectureRecommendation  string                                       `json:"architecture_recommendation"`
	Validation                  map[string]interface{}                       `json:"validation"`
}

func q012ClosureMetadata(ct *rlwe.Ciphertext) string {
	if ct == nil {
		return "nil"
	}
	return fmt.Sprintf("level=%d scale=%s degree=%d ntt=%t mont=%t", ct.Level(), finalizationScaleString(ct.Scale), ct.Degree(), ct.IsNTT, ct.IsMontgomery)
}

func q012ClosureSnapshotFor(params ckks.Parameters, name string, source *rlwe.Ciphertext) (q012ClosureSnapshot, error) {
	if source == nil {
		return q012ClosureSnapshot{Name: name}, fmt.Errorf("%s snapshot is nil", name)
	}
	q012, err := nativeQ012LiftQ01(params, source)
	if err != nil {
		return q012ClosureSnapshot{Name: name, Level: source.Level(), Scale: finalizationScaleString(source.Scale), Degree: source.Degree(), IsNTT: source.IsNTT, IsMontgomery: source.IsMontgomery}, err
	}
	capacity, err := nativeQ012Capacity(params, name, q012)
	if err != nil {
		return q012ClosureSnapshot{}, err
	}
	return q012ClosureSnapshot{
		Name: name, Level: source.Level(), Scale: finalizationScaleString(source.Scale), Degree: source.Degree(), IsNTT: source.IsNTT, IsMontgomery: source.IsMontgomery,
		Q01CapacityRatio: capacity.IntendedQ01.Ratio, Q012CapacityRatio: capacity.IntendedQ012.Ratio,
		Q01CenteredUnique: capacity.IntendedQ01.CenteredUnique, Q012CenteredUnique: capacity.IntendedQ012.CenteredUnique,
	}, nil
}

func q012ClosureRowProof(params ckks.Parameters, actual, oracle *rlwe.Ciphertext, expected []complex128) q012ClosureRows {
	rows, limbs := nativeQ012RowsEqual(params, actual, oracle)
	component, limb, index := nativeQ012FirstMismatch(params, actual, oracle)
	result := q012ClosureRows{RowsEqual: rows, Q0RowsEqual: limbs[0], Q1RowsEqual: limbs[1], Q2RowsEqual: limbs[2], FirstMismatchComponent: -1, FirstMismatchLimb: -1, FirstMismatchIndex: -1}
	if component != nil {
		result.FirstMismatchComponent = *component
	}
	if limb != nil {
		result.FirstMismatchLimb = *limb
	}
	if index != nil {
		result.FirstMismatchIndex = *index
	}
	if expected != nil {
		if decoded, err := psGlobalDecode(params, actual); err == nil {
			result.SemanticResidual = psGlobalMetric(expected, decoded)
		}
	}
	return result
}

func q012ClosureOracle(params ckks.Parameters, template *rlwe.Ciphertext, check *PSGlobalCheckpoint) (*rlwe.Ciphertext, error) {
	if check == nil || check.ciphertext == nil || check.expected == nil {
		return nil, fmt.Errorf("missing source-backed checkpoint oracle")
	}
	values, q012, _, err := psFirstCanonicalValues(params, *check)
	if err != nil {
		return nil, err
	}
	if !q012 {
		return nil, fmt.Errorf("checkpoint %s has no q012 oracle", check.ID)
	}
	return nativeQ012BuildFromSigned(params, template, values, template.Level(), template.Scale)
}

func q012ClosureCoefficientPoly(params ckks.Parameters, source *rlwe.Ciphertext, component, level int) ring.Poly {
	ntt := ring.NewPoly(params.N(), level)
	for limb := 0; limb <= level; limb++ {
		copy(ntt.Coeffs[limb], source.Value[component].Coeffs[limb])
		if source.IsMontgomery {
			params.RingQ().SubRings[limb].IMForm(ntt.Coeffs[limb], ntt.Coeffs[limb])
		}
	}
	coeff := ring.NewPoly(params.N(), level)
	params.RingQ().AtLevel(level).INTT(ntt, coeff)
	return coeff
}

// q012ClosureMulOracle deliberately leaves the candidate's NTT/Montgomery
// multiplication path. It performs coefficient-domain Barrett products and a
// fresh NTT, while preserving the exact degree-2 c0/c1/c2 semantics.
func q012ClosureMulOracle(params ckks.Parameters, left, right *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if left == nil || right == nil || left.Degree() != 1 || right.Degree() != 1 || left.Level() < 2 || right.Level() < 2 {
		return nil, fmt.Errorf("q012 coefficient oracle requires degree-one level >= 2 operands")
	}
	level := minInt(left.Level(), right.Level())
	left0, left1 := q012ClosureCoefficientPoly(params, left, 0, level), q012ClosureCoefficientPoly(params, left, 1, level)
	right0, right1 := q012ClosureCoefficientPoly(params, right, 0, level), q012ClosureCoefficientPoly(params, right, 1, level)
	product0, product1, product2 := ring.NewPoly(params.N(), level), ring.NewPoly(params.N(), level), ring.NewPoly(params.N(), level)
	ringQ := params.RingQ().AtLevel(level)
	polyMul := func(a, b, out ring.Poly) {
		aNTT, bNTT, productNTT := ring.NewPoly(params.N(), level), ring.NewPoly(params.N(), level), ring.NewPoly(params.N(), level)
		ringQ.NTT(a, aNTT)
		ringQ.NTT(b, bNTT)
		ringQ.MForm(aNTT, aNTT)
		ringQ.MulCoeffsMontgomery(aNTT, bNTT, productNTT)
		ringQ.INTT(productNTT, out)
	}
	polyMul(left0, right0, product0)
	polyMul(left0, right1, product1)
	polyMul(left1, right0, product2)
	ringQ.Add(product1, product2, product1)
	polyMul(left1, right1, product2)
	products := []ring.Poly{product0, product1, product2}
	out := ckks.NewCiphertext(params, 2, level)
	*out.MetaData = *left.MetaData
	out.IsNTT, out.IsMontgomery = true, true
	out.Scale = left.Scale.Mul(right.Scale)
	for component, coeff := range products {
		ntt := ring.NewPoly(params.N(), level)
		ringQ.NTT(coeff, ntt)
		for limb := 0; limb <= level; limb++ {
			params.RingQ().SubRings[limb].MForm(ntt.Coeffs[limb], ntt.Coeffs[limb])
			copy(out.Value[component].Coeffs[limb], ntt.Coeffs[limb])
		}
	}
	return out, nil
}

func q012ClosureAddOracle(params ckks.Parameters, left, right *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if left == nil || right == nil || left.Level() < 2 || right.Level() < 2 || left.Degree() != right.Degree() {
		return nil, fmt.Errorf("q012 coefficient add oracle metadata mismatch")
	}
	level := minInt(left.Level(), right.Level())
	out := ckks.NewCiphertext(params, left.Degree(), level)
	*out.MetaData = *left.MetaData
	out.IsNTT, out.IsMontgomery = true, true
	out.Scale = left.Scale
	ringQ := params.RingQ().AtLevel(level)
	for component := 0; component <= left.Degree(); component++ {
		lp, rp := q012ClosureCoefficientPoly(params, left, component, level), q012ClosureCoefficientPoly(params, right, component, level)
		sum := ring.NewPoly(params.N(), level)
		ringQ.Add(lp, rp, sum)
		ntt := ring.NewPoly(params.N(), level)
		ringQ.NTT(sum, ntt)
		for limb := 0; limb <= level; limb++ {
			params.RingQ().SubRings[limb].MForm(ntt.Coeffs[limb], ntt.Coeffs[limb])
			copy(out.Value[component].Coeffs[limb], ntt.Coeffs[limb])
		}
	}
	return out, nil
}

func q012ClosureRelinearizeOracle(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if source == nil || source.Degree() != 2 || source.Level() < 2 {
		return nil, fmt.Errorf("q012 relinearization oracle metadata mismatch")
	}
	out := ckks.NewCiphertext(params, 1, source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.Scale = source.IsNTT, source.IsMontgomery, source.Scale
	for component := 0; component <= 1; component++ {
		for limb := 0; limb <= 2; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func q012ClosureContract(params ckks.Parameters, source *rlwe.Ciphertext) (*rlwe.Ciphertext, designCapacity, error) {
	if source == nil || source.Level() < 2 {
		return nil, designCapacity{}, fmt.Errorf("q012 contraction requires level >= 2")
	}
	values, err := designQ01Values(params, source)
	if err != nil {
		return nil, designCapacity{}, err
	}
	capacity := designCapacityRow(params, "q01_contraction", "q01", values)
	if !capacity.CenteredUnique {
		return nil, capacity, fmt.Errorf("q01 contraction is not centered-unique")
	}
	out := fastckks.NewCiphertext(params, source.Degree(), source.Level())
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery = true, true
	for component := range source.Value {
		for limb := 0; limb <= 1; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, capacity, nil
}

func q012ClosurePromote(params ckks.Parameters, source *rlwe.Ciphertext, degree, level int) (*rlwe.Ciphertext, error) {
	if source == nil || degree < source.Degree() || level < 2 || level > source.Level() {
		return nil, fmt.Errorf("invalid q012 promote target degree=%d level=%d source=%s", degree, level, q012ClosureMetadata(source))
	}
	out := ckks.NewCiphertext(params, degree, level)
	*out.MetaData = *source.MetaData
	out.IsNTT, out.IsMontgomery, out.Scale = source.IsNTT, source.IsMontgomery, source.Scale
	for component := range source.Value {
		for limb := 0; limb <= 2; limb++ {
			copy(out.Value[component].Coeffs[limb], source.Value[component].Coeffs[limb])
		}
	}
	return out, nil
}

func q012ClosureAlign(params ckks.Parameters, a, b *rlwe.Ciphertext, id string) (*rlwe.Ciphertext, *rlwe.Ciphertext, q012ClosureAlignment, error) {
	if a == nil || b == nil {
		return nil, nil, q012ClosureAlignment{Operation: id}, fmt.Errorf("nil alignment operand")
	}
	level := minInt(a.Level(), b.Level())
	degree := a.Degree()
	if b.Degree() > degree {
		degree = b.Degree()
	}
	a, err := nativeQ012AtLevel(params, a, level)
	if err != nil {
		return nil, nil, q012ClosureAlignment{Operation: id}, err
	}
	b, err = nativeQ012AtLevel(params, b, level)
	if err != nil {
		return nil, nil, q012ClosureAlignment{Operation: id}, err
	}
	a, err = q012ClosurePromote(params, a, degree, level)
	if err != nil {
		return nil, nil, q012ClosureAlignment{Operation: id}, err
	}
	b, err = q012ClosurePromote(params, b, degree, level)
	if err != nil {
		return nil, nil, q012ClosureAlignment{Operation: id}, err
	}
	alignment := q012ClosureAlignment{Operation: id, LeftScale: finalizationScaleString(a.Scale), RightScale: finalizationScaleString(b.Scale), SourcePolicy: "psRescaleGuardAddAligned: ratio.BigInt() then exact target-scale metadata"}
	if a.Scale.Equal(b.Scale) {
		alignment.ExactRatio, alignment.IntegerRatio, alignment.FractionalDifference, alignment.IntegerRatioExact = "1", "1", "0", true
		alignment.AlignedRowsMatch = true
		return a, b, alignment, nil
	}
	if b.Scale.Cmp(a.Scale) > 0 {
		ratio := b.Scale.Div(a.Scale)
		integer := ratio.BigInt()
		aligned := a.CopyNew()
		if err := nativeQ012MulInteger(params, aligned, integer); err != nil {
			return nil, nil, alignment, err
		}
		aligned.Scale = b.Scale
		alignment.ExactRatio, alignment.IntegerRatio = finalizationScaleString(ratio), integer.String()
		alignment.FractionalDifference = psRescaleGuardFractionalDifference(ratio, integer)
		alignment.IntegerRatioExact = ratio.Equal(rlwe.NewScale(integer))
		alignment.AlignedRowsMatch = true
		return aligned, b, alignment, nil
	}
	ratio := a.Scale.Div(b.Scale)
	integer := ratio.BigInt()
	aligned := b.CopyNew()
	if err := nativeQ012MulInteger(params, aligned, integer); err != nil {
		return nil, nil, alignment, err
	}
	aligned.Scale = a.Scale
	alignment.ExactRatio, alignment.IntegerRatio = finalizationScaleString(ratio), integer.String()
	alignment.FractionalDifference = psRescaleGuardFractionalDifference(ratio, integer)
	alignment.IntegerRatioExact = ratio.Equal(rlwe.NewScale(integer))
	alignment.AlignedRowsMatch = true
	return a, aligned, alignment, nil
}

func q012ClosureAdd(params ckks.Parameters, a, b *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if a == nil || b == nil || a.Degree() != b.Degree() || a.Level() < 2 || b.Level() < 2 {
		return nil, fmt.Errorf("q012 add metadata mismatch")
	}
	if err := nativeQ012Add(params, a, b); err != nil {
		return nil, err
	}
	return a, nil
}

func q012ClosureG0(runtimeCase dependencyRuntimeCase, branch string) (q012ClosureG0Proof, *rlwe.Ciphertext, error) {
	var replay psRescaleGuardReplay
	var powers map[int]*rlwe.Ciphertext
	if branch == "real" {
		replay, powers = runtimeCase.RealReplay, runtimeCase.RealPowers
	} else {
		replay, powers = runtimeCase.ImagReplay, runtimeCase.ImagPowers
	}
	params := runtimeCase.Profile.BTP.BootstrappingParameters
	proof := q012ClosureG0Proof{Branch: branch, Snapshots: []q012ClosureSnapshot{}}
	if replay.G0Merge == nil || replay.G0Merge.A == nil || replay.G0Merge.BRaw == nil || replay.G0Merge.Product == nil || replay.G0Merge.Output == nil {
		proof.Reason = "actual G0 merge snapshots are incomplete"
		return proof, nil, nil
	}
	for _, item := range []struct {
		name  string
		value *rlwe.Ciphertext
	}{
		{"G0Merge.A", replay.G0Merge.A}, {"G0Merge.BRaw", replay.G0Merge.BRaw}, {"G0Merge.Product", replay.G0Merge.Product}, {"G0Merge.Output", replay.G0Merge.Output}, {"G0Merge.T2", powers[2]},
	} {
		snapshot, err := q012ClosureSnapshotFor(params, item.name, item.value)
		if err != nil {
			proof.Reason = err.Error()
			return proof, nil, nil
		}
		proof.Snapshots = append(proof.Snapshots, snapshot)
	}
	productCheck := dependencyFindCheck(replay.Checks, "G0-multiply")
	addCheck := dependencyFindCheck(replay.Checks, "G0-add")
	if productCheck == nil || addCheck == nil {
		proof.Reason = "actual G0 operation checkpoints are missing"
		return proof, nil, nil
	}
	a, err := nativeQ012LiftQ01(params, replay.G0Merge.A)
	if err != nil {
		return proof, nil, err
	}
	b, err := nativeQ012LiftQ01(params, replay.G0Merge.BRaw)
	if err != nil {
		return proof, nil, err
	}
	t2, err := nativeQ012LiftQ01(params, powers[2])
	if err != nil {
		return proof, nil, err
	}
	level := minInt(a.Level(), minInt(b.Level(), t2.Level()))
	a, err = nativeQ012AtLevel(params, a, level)
	if err != nil {
		return proof, nil, err
	}
	b, err = nativeQ012AtLevel(params, b, level)
	if err != nil {
		return proof, nil, err
	}
	t2, err = nativeQ012AtLevel(params, t2, level)
	if err != nil {
		return proof, nil, err
	}
	product, err := dependencyQ012MulDegree2(params, b, t2)
	if err != nil {
		return proof, nil, err
	}
	productOracle, err := q012ClosureMulOracle(params, b, t2)
	if err != nil {
		proof.Reason = "G0-multiply coefficient oracle: " + err.Error()
		return proof, nil, nil
	}
	proof.Product = q012ClosureOperationProof{Operation: "G0-multiply", OracleMetadata: q012ClosureMetadata(productOracle), CandidateMetadata: q012ClosureMetadata(product), Rows: q012ClosureRowProof(params, product, productOracle, productCheck.expected), MetadataMatch: nativeQ012MetadataEqual(product, productOracle)}
	proof.ProductSourceQ01RowsMatch = nativeQ01RowsEqual(params, product, replay.G0Merge.Product)
	proof.Product.Valid = proof.Product.Rows.RowsEqual && proof.Product.MetadataMatch
	if !proof.Product.Valid {
		proof.Reason = "source-faithful G0 product differs from q012 oracle"
	}
	alignedA, alignedProduct, alignment, err := q012ClosureAlign(params, a, product, "G0-add")
	if err != nil {
		proof.Reason = err.Error()
		return proof, nil, nil
	}
	proof.Alignment = alignment
	proof.Alignment.AlignedRowsMatch = alignedA != nil && alignedProduct != nil
	addOracle, err := q012ClosureAddOracle(params, alignedA, alignedProduct)
	if err != nil {
		proof.Reason = "G0-add coefficient oracle: " + err.Error()
		return proof, nil, nil
	}
	output, err := q012ClosureAdd(params, alignedA, alignedProduct)
	if err != nil {
		proof.Reason = err.Error()
		return proof, nil, nil
	}
	proof.Add = q012ClosureOperationProof{Operation: "G0-add", OracleMetadata: q012ClosureMetadata(addOracle), CandidateMetadata: q012ClosureMetadata(output), Rows: q012ClosureRowProof(params, output, addOracle, addCheck.expected), MetadataMatch: nativeQ012MetadataEqual(output, addOracle)}
	proof.AddSourceQ01RowsMatch = nativeQ01RowsEqual(params, output, replay.G0Merge.Output)
	proof.Add.Valid = proof.Add.Rows.RowsEqual && proof.Add.MetadataMatch
	if branch == "d" || runtimeCase.Control.Q0Bits == 56 {
		contracted, _, contractErr := q012ClosureContract(params, output)
		if contractErr == nil {
			proof.DContractionRowsMatch = nativeQ01RowsEqual(params, contracted, replay.G0Merge.Output)
			proof.DContractionMetadataMatch = nativeQ012MetadataEqual(contracted, replay.G0Merge.Output)
		} else if proof.Reason == "" {
			proof.Reason = "D q012→q01 contraction: " + contractErr.Error()
		}
	}
	cap, capErr := nativeQ012Capacity(params, "G0-add", output)
	if capErr != nil {
		proof.Reason = capErr.Error()
		return proof, nil, nil
	}
	proof.Q012CenteredUnique = cap.IntendedQ012.CenteredUnique
	proof.Valid = proof.Product.Valid && proof.Alignment.AlignedRowsMatch && proof.Add.Valid && proof.Q012CenteredUnique
	if branch == "d" || runtimeCase.Control.Q0Bits == 56 {
		proof.Valid = proof.Valid && proof.DContractionRowsMatch && proof.DContractionMetadataMatch
		if !proof.DContractionRowsMatch || !proof.DContractionMetadataMatch {
			proof.Reason = "D q012→q01 contraction rows or metadata differ from actual source G0-add"
		}
	}
	if !proof.Valid && proof.Reason == "" {
		proof.Reason = "G0 product/add, capacity, or D contraction proof failed"
	}
	return proof, output, nil
}

func q012ClosurePath(runtimeCase dependencyRuntimeCase, branch string, g0 *rlwe.Ciphertext) (q012ClosurePathProof, *rlwe.Ciphertext, error) {
	var replay psRescaleGuardReplay
	if branch == "real" {
		replay = runtimeCase.RealReplay
	} else {
		replay = runtimeCase.ImagReplay
	}
	params := runtimeCase.Profile.BTP.BootstrappingParameters
	proof := q012ClosurePathProof{Branch: branch, DependencyPath: []string{"G0-add", "G3-add", "F0-root", "F0-final-rescale"}}
	g3Check := dependencyFindCheck(replay.Checks, "G3-multiply")
	g3AddCheck := dependencyFindCheck(replay.Checks, "G3-add")
	rootCheck := replay.Root
	if g3Check == nil || g3AddCheck == nil || rootCheck.ciphertext == nil || replay.Final == nil {
		proof.Reason = "actual G3/F0 dependency checkpoints are incomplete"
		return proof, nil, nil
	}
	g0q, err := nativeQ012AtLevel(params, g0, minInt(g0.Level(), g3Check.ciphertext.Level()))
	if err != nil {
		return proof, nil, err
	}
	sibling, err := nativeQ012LiftQ01(params, g3Check.ciphertext)
	if err != nil {
		return proof, nil, err
	}
	alignedG0, alignedSibling, alignment, err := q012ClosureAlign(params, g0q, sibling, "G3-add")
	if err != nil {
		return proof, nil, err
	}
	proof.G3Alignment = alignment
	g3Oracle, err := q012ClosureAddOracle(params, alignedG0, alignedSibling)
	if err != nil {
		proof.Reason = "G3-add coefficient oracle: " + err.Error()
		return proof, nil, nil
	}
	g3, err := q012ClosureAdd(params, alignedG0, alignedSibling)
	if err != nil {
		return proof, nil, err
	}
	proof.G3Add = q012ClosureOperationProof{Operation: "G3-add", OracleMetadata: q012ClosureMetadata(g3Oracle), CandidateMetadata: q012ClosureMetadata(g3), Rows: q012ClosureRowProof(params, g3, g3Oracle, g3AddCheck.expected), MetadataMatch: nativeQ012MetadataEqual(g3, g3Oracle)}
	proof.G3Add.Valid = proof.G3Add.Rows.RowsEqual && proof.G3Add.MetadataMatch
	root := g3
	if rootCheck.Degree == 1 && root.Degree() == 2 {
		root, err = dependencyQ012Relinearize(params, root)
		proof.F0Relinearized = true
		if err != nil {
			return proof, nil, err
		}
	}
	rootOracle, err := root.CopyNew(), error(nil)
	if rootCheck.Degree == 1 && g3.Degree() == 2 {
		rootOracle, err = q012ClosureRelinearizeOracle(params, g3)
	}
	if err != nil {
		proof.Reason = "F0-root oracle: " + err.Error()
		return proof, nil, nil
	}
	proof.F0Root = q012ClosureOperationProof{Operation: "F0-root", OracleMetadata: q012ClosureMetadata(rootOracle), CandidateMetadata: q012ClosureMetadata(root), Rows: q012ClosureRowProof(params, root, rootOracle, rootCheck.expected), MetadataMatch: nativeQ012MetadataEqual(root, rootOracle)}
	proof.F0Root.Valid = proof.F0Root.Rows.RowsEqual && proof.F0Root.MetadataMatch
	if root.Level() < 3 {
		proof.Reason = fmt.Sprintf("F0 root level %d cannot retain q012 through final Rescale", root.Level())
		return proof, nil, nil
	}
	guard := new(big.Int).Lsh(big.NewInt(1), 3)
	if err := nativeQ012MulInteger(params, root, guard); err != nil {
		return proof, nil, err
	}
	root.Scale = root.Scale.Mul(rlwe.NewScale(guard))
	values, err := nativeQ012SignedValues(params, root)
	if err != nil {
		return proof, nil, err
	}
	divisor := new(big.Int).SetUint64(params.RingQ().SubRings[root.Level()].Modulus)
	quotients := make([][]*big.Int, len(values))
	for component := range values {
		quotients[component] = make([]*big.Int, len(values[component]))
		for i, value := range values[component] {
			quotients[component][i] = fix001P3LocalQ2Round(value, divisor)
		}
	}
	rescaled, err := nativeQ012BuildFromSigned(params, root, quotients, root.Level()-1, root.Scale.Div(rlwe.NewScale(divisor)))
	if err != nil {
		return proof, nil, err
	}
	contracted, cap, err := q012ClosureContract(params, rescaled)
	if err != nil {
		return proof, nil, err
	}
	proof.F0FinalRescale = q012ClosureOperationProof{Operation: "F0-final-rescale", OracleMetadata: q012ClosureMetadata(replay.Final), CandidateMetadata: q012ClosureMetadata(rescaled), Rows: q012ClosureRowProof(params, contracted, replay.Final, nil), MetadataMatch: nativeQ012MetadataEqual(contracted, replay.Final)}
	proof.F0FinalRescale.Valid = proof.F0FinalRescale.Rows.RowsEqual && proof.F0FinalRescale.MetadataMatch
	proof.PostRescaleContract = proof.F0FinalRescale.Rows
	proof.Q012CenteredUnique = proof.G3Add.Valid && proof.F0Root.Valid && cap.CenteredUnique
	proof.Q01Contracted = cap.CenteredUnique
	proof.Valid = proof.G3Add.Valid && proof.F0Root.Valid && proof.F0FinalRescale.Valid && proof.Q012CenteredUnique && proof.Q01Contracted
	if !proof.Valid && proof.Reason == "" {
		proof.Reason = "dependency-path q012 proof failed"
	}
	return proof, contracted, nil
}

func q012ClosureRunBranch(runtimeCase dependencyRuntimeCase, branch string) (q012ClosureBranchResult, error) {
	g0, g0ct, err := q012ClosureG0(runtimeCase, branch)
	if err != nil {
		return q012ClosureBranchResult{}, err
	}
	result := q012ClosureBranchResult{G0: g0}
	if g0ct == nil {
		return result, nil
	}
	path, final, err := q012ClosurePath(runtimeCase, branch, g0ct)
	if err != nil {
		return result, err
	}
	result.Path, result.Final = path, final
	return result, nil
}

func q012ClosureWrite(result q012ClosureResult, outPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0o644)
}

func runFIX001P3DiagLogN13Q012G0F0SourceFaithfulClosure(cfg BootstrapConfig, primaryRoot, backendRoot, outPath string) error {
	secondaryCommit := gitOutput(backendRoot, "rev-parse", "HEAD")
	secondaryClean := gitOutput(backendRoot, "status", "--porcelain") == ""
	result := q012ClosureResult{
		SchemaVersion: "fix-001-p3-diag-logn13-q012-g0-f0-source-faithful-closure.v1", Timestamp: time.Now().UTC(),
		Primary: gitMetadata(primaryRoot), Lattigo: gitMetadata(backendRoot), Environment: EnvironmentMetadata{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: cpuModel(), CPUs: runtime.NumCPU()}, Config: cfg,
		Provenance: map[string]interface{}{"required_primary_ancestor": requiredFIX001P3Q012ClosurePrimary, "required_secondary_commit": requiredFIX001P3Q012ClosureSecondary, "secondary_commit": secondaryCommit, "secondary_branch": gitOutput(backendRoot, "branch", "--show-current"), "secondary_clean": secondaryClean, "plan_scale": "2^92", "q012_oracle": "independent source-backed semantic vector materialization", "scale_alignment": "psRescaleGuardAddAligned integer ratio.BigInt policy"},
		Controls:   map[string]*psLocalizationDownstreamEvidence{}, DCalibration: map[string]q012ClosureBranchResult{}, BSourceFaithful: map[string]q012ClosureBranchResult{}, Downstream: map[string]*psLocalizationDownstreamEvidence{}, Validation: map[string]interface{}{"logn13_only": cfg.LogN == 13, "checked_in_config_unchanged": true, "secondary_exact": secondaryCommit == requiredFIX001P3Q012ClosureSecondary, "secondary_clean": secondaryClean, "no_secondary_production_changes": true, "no_serial_g0_g1_g2_g3": true, "actual_g0_snapshots": true, "generated_temporary_q012_powers": true, "no_oracle_power_map_candidate": true, "no_logn16": true, "no_benchmark": true, "no_gate_4_or_5": true, "no_exp003": true},
		MinimalCausalResetSet: []string{}, MinimalCausalResetSetProven: false, ResetAttribution: []map[string]interface{}{{"attempted": false, "reason": "no reset collapses an error; no minimal causal set claimed"}},
	}
	if cfg.LogN != 13 || len(cfg.Q0) != 1 || cfg.Q0[0] != 55 || secondaryCommit != requiredFIX001P3Q012ClosureSecondary || !secondaryClean || gitOutput(primaryRoot, "merge-base", requiredFIX001P3Q012ClosurePrimary, "HEAD") != requiredFIX001P3Q012ClosurePrimary {
		result.Classification = "logn13_q012_g0_f0_control_mismatch"
		result.FirstRemainingBlocker = "Primary/Secondary/config provenance mismatch"
		return q012ClosureWrite(result, outPath)
	}
	profile55, err := q056PrepareProfile(q056BuildConfig(cfg, 55))
	if err != nil {
		return err
	}
	profile56, err := q056PrepareProfile(q056BuildConfig(cfg, 56))
	if err != nil {
		return err
	}
	bRuntime, err := dependencyMakeRuntimeCase(profile55, 55)
	if err != nil {
		return err
	}
	dRuntime, err := dependencyMakeRuntimeCase(profile56, 56)
	if err != nil {
		return err
	}
	result.Controls["B"] = bRuntime.Control.Metrics
	result.Controls["D"] = dRuntime.Control.Metrics
	for _, item := range []struct {
		name    string
		runtime dependencyRuntimeCase
	}{{"real", dRuntime}, {"imag", dRuntime}} {
		g0, output, err := q012ClosureG0(item.runtime, item.name)
		if err != nil {
			return err
		}
		result.DCalibration[item.name] = q012ClosureBranchResult{G0: g0, G0Output: output}
	}
	dReal := result.DCalibration["real"]
	dImag := result.DCalibration["imag"]
	dCalibrationPass := dReal.G0.Valid && dImag.G0.Valid && dReal.Path.Valid && dImag.Path.Valid
	if !dCalibrationPass {
		result.Classification = "logn13_q012_g0_merge_harness_mismatch"
		result.FirstRemainingBlocker = "D q0=56 source-faithful q012 G0→F0 calibration failed"
		result.ProductionReadiness = "production_blocked_by_q012_diagnostic_harness_mismatch"
		result.ArchitectureRecommendation = "fix q012 diagnostic primitive/semantics before interpreting q0=55"
		return q012ClosureWrite(result, outPath)
	}
	for _, item := range []struct {
		name    string
		runtime dependencyRuntimeCase
	}{{"real", dRuntime}, {"imag", dRuntime}} {
		path, final, err := q012ClosurePath(item.runtime, item.name, result.DCalibration[item.name].G0Output)
		if err != nil {
			return err
		}
		entry := result.DCalibration[item.name]
		entry.Path, entry.Final = path, final
		result.DCalibration[item.name] = entry
	}
	for _, item := range []struct {
		name    string
		runtime dependencyRuntimeCase
	}{{"real", bRuntime}, {"imag", bRuntime}} {
		proof, err := q012ClosureRunBranch(item.runtime, item.name)
		if err != nil {
			return err
		}
		result.BSourceFaithful[item.name] = proof
	}
	bReal := result.BSourceFaithful["real"]
	bImag := result.BSourceFaithful["imag"]
	bG0Pass := bReal.G0.Valid && bImag.G0.Valid
	bPathPass := bG0Pass && bReal.Path.Valid && bImag.Path.Valid
	if bPathPass {
		result.Downstream["B"], err = psLocalizationRunAcceptedDownstream(bRuntime.Profile, precisionSweepScale(q012ClosurePlanScaleExponent), bReal.Final, bImag.Final)
		if err != nil {
			return err
		}
	}
	if dReal.Final != nil && dImag.Final != nil {
		result.Downstream["D"], err = psLocalizationRunAcceptedDownstream(dRuntime.Profile, precisionSweepScale(q012ClosurePlanScaleExponent), dReal.Final, dImag.Final)
		if err != nil {
			return err
		}
	}
	if !bG0Pass {
		result.Classification = "logn13_q055_plan92_g0_to_f0_q012_path_operation_mismatch"
		result.FirstRemainingBlocker = "B source-faithful G0 product/add q012 operation mismatch"
		result.ProductionReadiness = "production_blocked_by_g0_operation_mismatch"
	} else if !bPathPass {
		result.Classification = "logn13_q055_plan92_g0_to_f0_q012_path_operation_mismatch"
		result.FirstRemainingBlocker = "B source-faithful G0→G3→F0 q012 dependency path mismatch"
		result.ProductionReadiness = "production_blocked_by_dependency_path_operation_mismatch"
	} else if result.Downstream["B"] == nil || result.Downstream["B"].PublicLike == nil || !result.Downstream["B"].PublicLike.Pass {
		result.Classification = "logn13_q012_g0_merge_source_faithful_but_downstream_blocked"
		result.FirstRemainingBlocker = "source-faithful q012 G0→F0 path did not meet public-like threshold"
		result.ProductionReadiness = "production_blocked_by_downstream_numeric_result"
	} else {
		result.Classification = "logn13_q055_plan92_g0_to_f0_q012_path_system_sufficient"
		result.FirstRemainingBlocker = "none"
		result.ProductionReadiness = "diagnostic_path_system_sufficient_not_production_integrated"
	}
	result.ArchitectureRecommendation = "compare q0=56 accepted D path with source-faithful q0=55 q012 path before any production parameter change"
	result.Validation["d_calibration_passed_before_b_interpretation"] = dCalibrationPass
	result.Validation["b_g0_source_faithful"] = bG0Pass
	result.Validation["b_dependency_path_source_faithful"] = bPathPass
	result.Validation["q2_dropped_only_after_q01_uniqueness"] = bReal.Path.Q01Contracted && bImag.Path.Q01Contracted
	result.Validation["public_like_threshold"] = q012ClosureThreshold
	return q012ClosureWrite(result, outPath)
}

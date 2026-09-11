package main

import (
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func newQ01TestParameters(t *testing.T, q []int) ckks.Parameters {
	t.Helper()
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{LogN: 4, LogQ: q, LogDefaultScale: 30})
	if err != nil {
		t.Fatalf("create test parameters: %v", err)
	}
	return params
}

func fillQ01TestCiphertext(t *testing.T, params ckks.Parameters, level int, montgomery bool) *rlwe.Ciphertext {
	t.Helper()
	ct := ckks.NewCiphertext(params, 1, level)
	ct.LogDimensions = ring.Dimensions{Cols: params.LogMaxSlots()}
	ct.IsNTT = true
	ct.IsMontgomery = montgomery
	for component := range ct.Value {
		for limb := 0; limb < 2; limb++ {
			for i := range ct.Value[component].Coeffs[limb] {
				ct.Value[component].Coeffs[limb][i] = uint64(i + component + limb + 1)
			}
			if montgomery {
				params.RingQ().SubRings[limb].MForm(ct.Value[component].Coeffs[limb], ct.Value[component].Coeffs[limb])
			}
		}
	}
	return ct
}

func TestQ01ProjectionCopiesOnlyMaintainedRowsAndPreservesMetadata(t *testing.T) {
	params := newQ01TestParameters(t, []int{50, 50, 50})
	source := fillQ01TestCiphertext(t, params, 2, true)
	for component := range source.Value {
		source.Value[component].Coeffs[2] = nil
	}

	projection, evidence, err := q01Projection(params, source, true)
	if err != nil {
		t.Fatalf("q01 projection: %v", err)
	}
	if projection.Level() != 1 {
		t.Fatalf("diagnostic level = %d, want 1", projection.Level())
	}
	if !evidence.RowsCopiedExactly || !evidence.SourceUnchanged {
		t.Fatalf("projection integrity evidence: %+v", evidence)
	}
	if projection.Scale.Log2() != source.Scale.Log2() || projection.LogSlots() != source.LogSlots() || projection.IsNTT != source.IsNTT {
		t.Fatalf("projection did not preserve Scale/LogSlots/IsNTT metadata")
	}
	if projection.IsMontgomery {
		t.Fatal("Fast diagnostic projection remained Montgomery")
	}
	if evidence.DiagnosticAfterNormalize.Level != 1 || evidence.DiagnosticAfterNormalize.QModuli[0] != evidence.Source.QModuli[0] || evidence.DiagnosticAfterNormalize.QModuli[1] != evidence.Source.QModuli[1] {
		t.Fatalf("unexpected diagnostic metadata: %+v", evidence.DiagnosticAfterNormalize)
	}
}

func TestQ01ProjectionStandardDecodeMatchesFullStandardDecode(t *testing.T) {
	params := newQ01TestParameters(t, []int{50, 50})
	sk := rlwe.NewKeyGenerator(params).GenSecretKeyNew()
	pt := ckks.NewPlaintext(params, 1)
	values := make([]complex128, params.MaxSlots())
	for i := range values {
		values[i] = complex(float64(i+1)/16, -float64(i+1)/32)
	}
	if err := ckks.NewEncoder(params).Encode(values, pt); err != nil {
		t.Fatalf("encode: %v", err)
	}
	ct, err := rlwe.NewEncryptor(params, sk).EncryptNew(pt)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	full, err := decodeWithSecret(params, ct, sk)
	if err != nil {
		t.Fatalf("full Standard decode: %v", err)
	}
	projection, evidence, err := q01Projection(params, ct, false)
	if err != nil {
		t.Fatalf("Standard q01 projection: %v", err)
	}
	if !evidence.RowsCopiedExactly || !evidence.SourceUnchanged {
		t.Fatalf("projection integrity evidence: %+v", evidence)
	}
	projected, err := decodeWithSecret(params, projection, sk)
	if err != nil {
		t.Fatalf("Standard q01 decode: %v", err)
	}
	metrics, err := compareComplexVectors(full, projected, correctnessThreshold)
	if err != nil {
		t.Fatalf("compare full/projection: %v", err)
	}
	if !metrics.PassThreshold || metrics.MaxAbsComplex != 0 {
		t.Fatalf("Standard q01 projection changed decode: %+v", metrics)
	}
}

func TestQ01ProjectionFastZeroSecretDecodePath(t *testing.T) {
	params := newQ01TestParameters(t, []int{50, 50})
	zero := zeroSecret(params)
	pt := ckks.NewPlaintext(params, 1)
	values := make([]complex128, params.MaxSlots())
	values[0] = complex(0.25, -0.125)
	if err := ckks.NewEncoder(params).Encode(values, pt); err != nil {
		t.Fatalf("encode: %v", err)
	}
	ct, err := rlwe.NewEncryptor(params, zero).EncryptNew(pt)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	for component := range ct.Value {
		for limb := 0; limb < 2; limb++ {
			params.RingQ().SubRings[limb].MForm(ct.Value[component].Coeffs[limb], ct.Value[component].Coeffs[limb])
		}
	}
	ct.IsMontgomery = true
	projection, evidence, err := q01Projection(params, ct, true)
	if err != nil {
		t.Fatalf("Fast q01 projection: %v", err)
	}
	if !evidence.SourceUnchanged || projection.IsMontgomery {
		t.Fatalf("Fast projection integrity evidence: %+v", evidence)
	}
	if _, err := decodeWithSecret(params, projection, zero); err != nil {
		t.Fatalf("Fast zero-secret q01 decode path: %v", err)
	}
}

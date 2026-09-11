package main

import (
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func newFinalizationTestCiphertext(t *testing.T) (*rlwe.Ciphertext, ckks.Parameters) {
	t.Helper()
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{LogN: 4, LogQ: []int{50, 50}, LogDefaultScale: 30})
	if err != nil {
		t.Fatalf("create test parameters: %v", err)
	}
	ct := ckks.NewCiphertext(params, 1, 1)
	ct.LogDimensions = ring.Dimensions{Cols: params.LogMaxSlots()}
	ct.IsNTT = true
	ct.IsMontgomery = true
	for component := range ct.Value {
		for limb := 0; limb < 2; limb++ {
			for i := range ct.Value[component].Coeffs[limb] {
				ct.Value[component].Coeffs[limb][i] = uint64(i + component + limb + 1)
			}
			params.RingQ().SubRings[limb].MForm(ct.Value[component].Coeffs[limb], ct.Value[component].Coeffs[limb])
		}
	}
	return ct, params
}

func TestFinalizationDiagnosticCopyDoesNotAliasSource(t *testing.T) {
	source, _ := newFinalizationTestCiphertext(t)
	copy := finalizationDiagnosticCopy(source)
	copy.Value[0].Coeffs[0][0]++
	if source.Value[0].Coeffs[0][0] == copy.Value[0].Coeffs[0][0] {
		t.Fatal("diagnostic copy aliases source storage")
	}
}

func TestFinalizationMaintainedRowsDetectExactMismatch(t *testing.T) {
	source, _ := newFinalizationTestCiphertext(t)
	copy := finalizationDiagnosticCopy(source)
	if rows := compareFinalizationRows(source, copy); !finalizationRowsEqual(rows) {
		t.Fatal("identical diagnostic copies should have exact maintained rows")
	}
	copy.Value[1].Coeffs[1][3]++
	rows := compareFinalizationRows(source, copy)
	if finalizationRowsEqual(rows) {
		t.Fatal("maintained-row mismatch was not detected")
	}
	if rows[3].FirstMismatchIndex == nil || *rows[3].FirstMismatchIndex != 3 {
		t.Fatalf("unexpected mismatch location: %+v", rows[3])
	}
}

func TestFinalizationIMFormMFormRoundTrip(t *testing.T) {
	source, params := newFinalizationTestCiphertext(t)
	roundTrip := finalizationDiagnosticCopy(source)
	if err := finalizationTransformMaintained(roundTrip, params, false); err != nil {
		t.Fatalf("IMForm: %v", err)
	}
	if err := finalizationTransformMaintained(roundTrip, params, true); err != nil {
		t.Fatalf("MForm: %v", err)
	}
	if rows := compareFinalizationRows(source, roundTrip); !finalizationRowsEqual(rows) {
		t.Fatalf("IMForm->MForm round trip changed maintained rows: %+v", rows)
	}
}

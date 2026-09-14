package main

import (
	"testing"

	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

func TestFIX001P3D0ProductionPreconditionClosureClassificationSet(t *testing.T) {
	allowed := map[string]bool{
		"logn13_d0_fixed_q055_plan91_full_stack_sufficient":  true,
		"logn13_d0_fixed_q055_requires_plan92":               true,
		"logn13_d0_q0_56_width_precondition":                 true,
		"logn13_d0_q0_plan_interaction":                      true,
		"logn13_d0_production_precondition_control_mismatch": true,
		"logn13_d0_production_precondition_numeric_mismatch": true,
	}
	if len(allowed) != 6 {
		t.Fatalf("unexpected classification set size: %d", len(allowed))
	}
}

func TestFIX001P3D0ProductionPreconditionClosureFastRowsDoNotDoubleMForm(t *testing.T) {
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN:            4,
		LogQ:            []int{50, 50, 50},
		LogDefaultScale: 30,
	})
	if err != nil {
		t.Fatal(err)
	}
	a := fastckks.NewCiphertext(params, 1, 2)
	b := fastckks.NewCiphertext(params, 1, 2)
	for component := 0; component <= 1; component++ {
		for limb := 0; limb <= 1; limb++ {
			row := a.Value[component].Coeffs[limb]
			for i := range row {
				row[i] = uint64((i+1+component*3+limb)%7 + 1)
			}
			params.RingQ().SubRings[limb].MForm(row, row)
			copy(b.Value[component].Coeffs[limb], row)
		}
	}
	if !nativeQ01FastRowsEqual(params, a, b) {
		t.Fatal("identical Fast/Montgomery rows must compare equal")
	}
	double := b.CopyNew()
	for component := 0; component <= 1; component++ {
		for limb := 0; limb <= 1; limb++ {
			params.RingQ().SubRings[limb].MForm(double.Value[component].Coeffs[limb], double.Value[component].Coeffs[limb])
		}
	}
	if nativeQ01FastRowsEqual(params, a, double) {
		t.Fatal("double-MFormed rows must not compare equal")
	}
}

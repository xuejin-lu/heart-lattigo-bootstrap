package main

import "testing"

func TestPublicAddSubCompositionUsesFrozenMatchedProfile(t *testing.T) {
	result, err := runPublicComposition()
	if err != nil {
		t.Fatalf("public Add/Sub composition failed: %v", err)
	}
	if result.Status != "public_add_sub_composition_passed" {
		t.Fatalf("unexpected status %q", result.Status)
	}
	if result.BootstrapCalls != 0 {
		t.Fatalf("Bootstrap budget must stay zero, got %d", result.BootstrapCalls)
	}
	if result.Profile.LogN != 13 || result.Profile.LogSlots != 4 || result.Profile.ScaleLog2 != 45 || len(result.Q) != 4 || len(result.P) != 1 {
		t.Fatalf("unexpected profile: %+v Q=%d P=%d", result.Profile, len(result.Q), len(result.P))
	}
	if result.Profile.TestLevels[0] != 1 || result.Profile.TestLevels[1] != 3 {
		t.Fatalf("unexpected levels: %v", result.Profile.TestLevels)
	}
	if len(result.Inputs) != 4 || len(result.Checkpoints) != 6 {
		t.Fatalf("unexpected evidence cardinality: inputs=%d checkpoints=%d", len(result.Inputs), len(result.Checkpoints))
	}
	for _, input := range result.Inputs {
		if !input.Oracle.Finite || !input.State.AllActiveRowsMaterialized {
			t.Errorf("input %s failed ordinary lifecycle/oracle checks: %+v", input.Name, input)
		}
		if result.FastZeroSecret && !input.State.C1Zero {
			t.Errorf("Fast EncryptNew input %s has nonzero c1", input.Name)
		}
	}
	for _, check := range result.Checkpoints {
		if check.Status != "passed" || !check.Oracle.Finite || check.State.Level != check.Level || check.State.Degree != check.Degree || !check.State.AllActiveRowsMaterialized {
			t.Errorf("checkpoint %s failed state/oracle checks: %+v", check.ID, check)
		}
		if result.FastZeroSecret && !check.State.C1Zero {
			t.Errorf("Fast checkpoint %s has nonzero c1", check.ID)
		}
		if len(check.Decoded) != 1<<logSlots {
			t.Errorf("checkpoint %s stored %d decoded values", check.ID, len(check.Decoded))
		}
	}
}

func TestDeterministicInputHashAndPairedMetrics(t *testing.T) {
	a0, b0 := deterministicInputs(1 << logSlots)
	a1, b1 := deterministicInputs(1 << logSlots)
	if hashInputs(a0, b0) != hashInputs(a1, b1) {
		t.Fatal("deterministic workload hash changed")
	}
	want := []complex128{1 + 2i, -3 + 4i}
	got := []complex128{1 + 2i, -3 + 4i}
	if m := compare(want, got); !m.Finite || m.ComplexRMSE != 0 || m.MaxComplex != 0 {
		t.Fatalf("identical paired values did not compare exactly: %+v", m)
	}
}

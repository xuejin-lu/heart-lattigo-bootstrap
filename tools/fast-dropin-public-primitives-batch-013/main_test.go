package main

import "testing"

func TestPublicPrimitiveCoverageSuite(t *testing.T) {
	result, err := runPrimitiveSuite()
	if err != nil {
		t.Fatalf("public primitive suite failed: %v", err)
	}
	if result.Status != "all_p0_public_primitive_checkpoints_passed" {
		t.Fatalf("unexpected suite status %q: %s", result.Status, result.StopReason)
	}
	if result.BootstrapCalls != 0 {
		t.Fatalf("Bootstrap budget must remain zero, got %d", result.BootstrapCalls)
	}
	if len(result.Inputs) != 4 {
		t.Fatalf("expected two plaintexts at each of two Levels, got %d inputs", len(result.Inputs))
	}
	if len(result.Operations) != 24 {
		t.Fatalf("expected 24 operation/negative-control checkpoints, got %d", len(result.Operations))
	}
	for _, input := range result.Inputs {
		if !input.Oracle.Finite || !input.State.ActiveRowsMaterialized {
			t.Errorf("input %s failed oracle or full-active-Q validation", input.Name)
		}
		if result.FastCapability != input.State.C1Zero {
			t.Errorf("input %s c1=%t does not match Fast capability=%t", input.Name, input.State.C1Zero, result.FastCapability)
		}
	}
	for _, operation := range result.Operations {
		switch operation.Status {
		case "passed", "passed_expected_rejection", "not_applicable_to_genuine_standard":
		default:
			t.Errorf("checkpoint %s did not pass: %s (%s)", operation.ID, operation.Status, operation.Error)
		}
		if operation.Status == "passed" && (!operation.FullActiveQBacking || !operation.C1PolicyPassed || !operation.StateChecksPassed || !operation.InputUnchanged || operation.Oracle == nil || !operation.Oracle.Finite) {
			t.Errorf("checkpoint %s is missing a required correctness assertion", operation.ID)
		}
	}
	if result.FastCapability {
		if result.TotalKeyLookups.Relinearization != 0 || result.TotalKeyLookups.Galois != 0 {
			t.Fatalf("Fast zero-secret run must not perform native evaluation-key lookups: %+v", result.TotalKeyLookups)
		}
		foundC1Rejection := false
		for _, operation := range result.Operations {
			if operation.ID == "mulrelin-invalid-c1" {
				foundC1Rejection = operation.Status == "passed_expected_rejection" && operation.KeyLookupDelta.Relinearization == 0 && operation.OutputUnchanged != nil && *operation.OutputUnchanged
			}
		}
		if !foundC1Rejection {
			t.Fatal("Fast nonzero-c1 fail-closed/transactionality control did not pass")
		}
	} else {
		if result.TotalKeyLookups.Relinearization == 0 || result.TotalKeyLookups.Galois == 0 {
			t.Fatalf("genuine Standard run must show native relin and Galois key use: %+v", result.TotalKeyLookups)
		}
	}
}

func TestDeterministicInputHashStable(t *testing.T) {
	a1, b1, p1 := deterministicInputs(1 << logSlots)
	a2, b2, p2 := deterministicInputs(1 << logSlots)
	if got, want := hashInputs(a1, b1, p1), hashInputs(a2, b2, p2); got != want {
		t.Fatalf("deterministic input hash changed: %s != %s", got, want)
	}
}

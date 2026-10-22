package perfmeasure

import (
	"encoding/json"
	"math/big"
	"path/filepath"
	"testing"
)

func TestPublicDirectMulWorkloadAndCapacityPreflight(t *testing.T) {
	cfg, _, err := LoadConfig(filepath.Join("..", "..", "configs", "bootstrap_config.logN13.json"))
	if err != nil {
		t.Fatal(err)
	}
	residual, bootstrap, effective, err := ParametersFromConfigWithE(cfg, 32)
	if err != nil {
		t.Fatal(err)
	}
	if effective.LogN != 13 || effective.LogSlots != 12 || effective.InputSlots != 1<<12 || effective.EphemeralSecretWeight != 32 {
		t.Fatalf("unexpected effective direct-mul profile: %+v", effective)
	}
	q, p := bootstrap.BootstrappingParameters.Q(), bootstrap.BootstrappingParameters.P()
	if len(q) <= 5 || bootstrap.BootstrappingParameters.MaxLevel() < 5 {
		t.Fatalf("effective Q chain does not contain q0..q5: levels=%d Q=%d", bootstrap.BootstrappingParameters.MaxLevel(), len(q))
	}

	a, b, c, err := PublicDirectMulWorkload(effective.InputSlots, "direct-base")
	if err != nil {
		t.Fatal(err)
	}
	a4, b4, c4, err := PublicDirectMulWorkload(effective.InputSlots, "direct-c4")
	if err != nil {
		t.Fatal(err)
	}
	if Fingerprint(a) != Fingerprint(a4) || Fingerprint(b) != Fingerprint(b4) || Fingerprint(c) == Fingerprint(a) {
		t.Fatal("direct-c4 must preserve A/B and use a C message distinct from A")
	}
	for i := range c {
		want := complex(float64((7*i+3)%23-11)/512, float64((11*i+5)%29-14)/1024)
		if c[i] != want || c4[i] != 4*c[i] {
			t.Fatalf("direct C formula mismatch at slot %d: base=%v c4=%v", i, c[i], c4[i])
		}
	}
	if Fingerprint(c) == Fingerprint(c4) {
		t.Fatal("direct-base and direct-c4 must have distinct C inputs")
	}

	for _, workload := range []struct {
		name    string
		a, b, c []complex128
	}{{"direct-base", a, b, c}, {"direct-c4", a4, b4, c4}} {
		plan, err := BoundPublicDirectMulChain(workload.a, workload.b, workload.c, q, bootstrap.BootstrappingParameters.N())
		if err != nil {
			t.Fatalf("%s capacity: %v", workload.name, err)
		}
		if plan.Status != "STRICT_QPREFIX_FIT" || len(plan.Checks) != 8 || plan.Stage1.ActiveQPrefixRows != 4 || plan.Stage2.ActiveQPrefixRows != 4 {
			t.Fatalf("%s has incomplete strict Q-prefix evidence: %+v", workload.name, plan)
		}
		for _, check := range plan.Checks {
			if !check.StrictFit || check.QRows != 4 {
				t.Fatalf("%s capacity check failed: %+v", workload.name, check)
			}
		}
		encodedPlan, err := json.Marshal(plan)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s capacity_plan=%s", workload.name, encodedPlan)
	}

	delta := new(big.Int).Lsh(big.NewInt(1), 45)
	for _, divisor := range []*big.Int{new(big.Int).SetUint64(q[5]), new(big.Int).SetUint64(q[4])} {
		productScale := new(big.Int).Mul(new(big.Int).Set(delta), divisor)
		if got := new(big.Int).Quo(productScale, divisor); got.Cmp(delta) != 0 {
			t.Fatalf("Mul→Rescale scale does not return to Δ: got=%s want=%s", got, delta)
		}
	}
	if !residual.DefaultScale().BigInt().IsUint64() || residual.DefaultScale().BigInt().Cmp(delta) != 0 {
		t.Fatalf("configured Δ differs from 2^45: %s", residual.DefaultScale().BigInt())
	}

	inputHash, err := PublicDirectMulInputsFingerprint(a, b, c)
	if err != nil {
		t.Fatal(err)
	}
	qpHash := QPFingerprint(q, p)
	baseWorkloadHash, err := PublicDirectMulWorkloadFingerprint(a, b, c, "direct-base", delta.String(), new(big.Int).SetUint64(q[5]).String(), new(big.Int).SetUint64(q[4]).String(), qpHash)
	if err != nil {
		t.Fatal(err)
	}
	c4InputHash, err := PublicDirectMulInputsFingerprint(a4, b4, c4)
	if err != nil {
		t.Fatal(err)
	}
	c4WorkloadHash, err := PublicDirectMulWorkloadFingerprint(a4, b4, c4, "direct-c4", delta.String(), new(big.Int).SetUint64(q[5]).String(), new(big.Int).SetUint64(q[4]).String(), qpHash)
	if err != nil {
		t.Fatal(err)
	}
	if inputHash == c4InputHash || baseWorkloadHash == c4WorkloadHash {
		t.Fatal("direct workload variants must have independent input and workload identities")
	}
	t.Logf("effective_qp_sha256=%s q4=%d q5=%d", qpHash, q[4], q[5])
	t.Logf("direct-base input_sha256=%s A=%s B=%s C=%s workload_sha256=%s", inputHash, Fingerprint(a), Fingerprint(b), Fingerprint(c), baseWorkloadHash)
	t.Logf("direct-c4 input_sha256=%s A=%s B=%s C=%s workload_sha256=%s", c4InputHash, Fingerprint(a4), Fingerprint(b4), Fingerprint(c4), c4WorkloadHash)
}

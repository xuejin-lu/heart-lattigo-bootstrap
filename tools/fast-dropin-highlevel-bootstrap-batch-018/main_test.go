package main

import (
	"encoding/json"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"testing"
)

func TestBatch018IndependentDropLevelCapacityBound(t *testing.T) {
	a, b := deterministicInputs(1 << 12)
	if got := hashComplexValues(deterministicInput(1 << 12)); got != canonicalInputSHA256 {
		t.Fatalf("canonical 017 input hash=%s, want %s", got, canonicalInputSHA256)
	}
	boundA := encodedInputBound(a, math.Exp2(defaultScaleLog2))
	boundB := encodedInputBound(b, math.Exp2(defaultScaleLog2))
	boundAfterAdd := new(big.Int).Add(new(big.Int).Set(boundA), boundB)
	if boundA.String() != "2348557866436" || boundB.String() != "1804822278899" || boundAfterAdd.String() != "4153380145335" {
		t.Fatalf("unexpected conservative bounds: A=%s B=%s Add=%s", boundA, boundB, boundAfterAdd)
	}
	q0 := new(big.Int).SetUint64(36028797018652673)
	if !strictCapacity(boundAfterAdd, q0) {
		t.Fatalf("2*B=%s must fit q0=%s", new(big.Int).Lsh(boundAfterAdd, 1), q0)
	}
	q0123 := prefixProduct([]uint64{36028797018652673, 549755731969, 549756026881, 549755486209}, 4)
	if q0123.String() != "5986308565615587353347023386369277282933624144412673" {
		t.Fatalf("unexpected Level-5 q0..q3 product %s", q0123)
	}
}

func TestStrictCapacityRejectsEquality(t *testing.T) {
	if strictCapacity(big.NewInt(5), big.NewInt(10)) {
		t.Fatal("strict 2B < product must reject equality")
	}
	if !strictCapacity(big.NewInt(4), big.NewInt(10)) {
		t.Fatal("strict 2B < product should accept 8 < 10")
	}
}

func TestBatch018ProfileMatchesCanonical017(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(filepath.Join("..", "..")); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	}()
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := hashBytes(data); got != canonicalConfigSHA256 {
		t.Fatalf("config hash=%s, want %s", got, canonicalConfigSHA256)
	}
	var cfg config
	if err = json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	residual, btp, slots, err := parametersFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if residual.MaxLevel() != 1 || btp.BootstrappingParameters.MaxLevel() != 16 || slots != 12 || btp.EphemeralSecretWeight != 32 {
		t.Fatalf("wrong profile: residual=%d full=%d slots=%d E=%d", residual.MaxLevel(), btp.BootstrappingParameters.MaxLevel(), slots, btp.EphemeralSecretWeight)
	}
	qp, _ := json.Marshal(struct{ Q, P []uint64 }{btp.BootstrappingParameters.Q(), btp.BootstrappingParameters.P()})
	if got := hashBytes(qp); got != canonicalQPSHA256 {
		t.Fatalf("generated Q/P hash=%s, want %s", got, canonicalQPSHA256)
	}
}

func TestRotateLeftMatchesPublicCKKSSlotConvention(t *testing.T) {
	got := rotateLeft([]complex128{1, 2, 3, 4}, 1)
	want := []complex128{2, 3, 4, 1}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("rotation[%d]=%v, want %v", i, got[i], want[i])
		}
	}
}

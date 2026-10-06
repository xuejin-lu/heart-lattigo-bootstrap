//go:build perf_standard

package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

type standardBackend struct {
	eval     *bootstrapping.Evaluator
	residual ckks.Parameters
	secret   *rlwe.SecretKey
}

func newBackend(params bootstrapping.Parameters, residual ckks.Parameters) (backendAdapter, error) {
	keygen := rlwe.NewKeyGenerator(params.BootstrappingParameters)
	secret := keygen.GenSecretKeyNew()
	keys, _, err := params.GenEvaluationKeys(secret)
	if err != nil {
		return nil, fmt.Errorf("generate Standard evaluation keys: %w", err)
	}
	eval, err := bootstrapping.NewEvaluator(params, keys)
	if err != nil {
		return nil, fmt.Errorf("construct Standard evaluator: %w", err)
	}
	return &standardBackend{eval: eval, residual: residual, secret: secret}, nil
}

func (b *standardBackend) Name() string { return "standard" }

func (b *standardBackend) EvaluatorPath() string {
	return "fresh GenSecretKey + GenEvaluationKeys + bootstrapping.NewEvaluator"
}

func (b *standardBackend) KeyTrialEvidence() (bool, int) { return true, b.secret.LevelP() }

func (b *standardBackend) SecretKeyForTrial() *rlwe.SecretKey { return b.secret }

func (b *standardBackend) Bootstrap(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	return b.eval.Bootstrap(ct)
}

func (b *standardBackend) Decode(ct *rlwe.Ciphertext) ([]complex128, error) {
	if ct == nil || ct.Level() < 0 || ct.Level() > b.residual.MaxLevel() {
		return nil, fmt.Errorf("invalid Standard ciphertext level")
	}
	plain := rlwe.NewDecryptor(b.residual, b.secret).DecryptNew(ct)
	values := make([]complex128, b.residual.MaxSlots())
	if err := ckks.NewEncoder(b.residual).Decode(plain, values); err != nil {
		return nil, err
	}
	return values, nil
}

func (b *standardBackend) PrefixRows(ct *rlwe.Ciphertext) (int, error) {
	if ct == nil || ct.Level() < 0 {
		return 0, fmt.Errorf("invalid Standard ciphertext level")
	}
	return ct.Level() + 1, nil
}

func (b *standardBackend) Pack(cts []rlwe.Ciphertext) ([]rlwe.Ciphertext, error) {
	packed, _, _, err := b.eval.PackAndSwitchN1ToN2(cts)
	return packed, err
}

func (b *standardBackend) ScaleDown(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	out, _, err := b.eval.ScaleDown(ct)
	return out, err
}

func (b *standardBackend) ModUp(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	return b.eval.ModUp(ct)
}

func (b *standardBackend) CoeffsToSlots(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	return b.eval.CoeffsToSlots(ct)
}

func (b *standardBackend) EvalMod(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	return b.eval.EvalMod(ct)
}

func (b *standardBackend) SlotsToCoeffs(realCT, imagCT *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	return b.eval.SlotsToCoeffs(realCT, imagCT)
}

//go:build perf_standard

package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

type standardBackend struct {
	eval                *bootstrapping.Evaluator
	residual            ckks.Parameters
	bootstrappingParams ckks.Parameters
	secret              *rlwe.SecretKey
}

func newBackend(params bootstrapping.Parameters, residual ckks.Parameters) (backendAdapter, error) {
	inputBackend, err := newInputBackend(params, residual)
	if err != nil {
		return nil, err
	}
	backend := inputBackend.(*standardBackend)
	keys, _, err := params.GenEvaluationKeys(backend.secret)
	if err != nil {
		return nil, fmt.Errorf("generate Standard evaluation keys: %w", err)
	}
	eval, err := bootstrapping.NewEvaluator(params, keys)
	if err != nil {
		return nil, fmt.Errorf("construct Standard evaluator: %w", err)
	}
	backend.eval = eval
	return backend, nil
}

func newInputBackend(params bootstrapping.Parameters, residual ckks.Parameters) (backendAdapter, error) {
	keygen := rlwe.NewKeyGenerator(params.BootstrappingParameters)
	secret := keygen.GenSecretKeyNew()
	return &standardBackend{residual: residual, bootstrappingParams: params.BootstrappingParameters, secret: secret}, nil
}

func (b *standardBackend) Name() string { return "standard" }

func (b *standardBackend) InputKind() string { return standardNativeInputKind }

func (b *standardBackend) InputConstructor() string {
	return standardNativeConstructor
}

func (b *standardBackend) PrepareInput(values []complex128, logSlots int) (*rlwe.Ciphertext, error) {
	plain, err := perfmeasure.EncodeInputPlaintext(b.residual, logSlots, values)
	if err != nil {
		return nil, err
	}
	return rlwe.NewEncryptor(b.residual, b.secret).EncryptNew(plain)
}

func (b *standardBackend) EvaluatorPath() string {
	return "fresh GenSecretKey + GenEvaluationKeys + bootstrapping.NewEvaluator"
}

func (b *standardBackend) KeyTrialEvidence() (bool, int) { return true, b.secret.LevelP() }

func (b *standardBackend) SecretKeyForTrial() *rlwe.SecretKey { return b.secret }

func (b *standardBackend) Bootstrap(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	return b.eval.Bootstrap(ct)
}

func (b *standardBackend) Decode(ct *rlwe.Ciphertext) ([]complex128, error) {
	return b.decodeWithParameters(ct, b.residual)
}

func (b *standardBackend) DecodeStage(ct *rlwe.Ciphertext, name string) ([]complex128, string, error) {
	switch name {
	case "input", "scale_down", "final_public_output":
		values, err := b.decodeWithParameters(ct, b.residual)
		return values, "standard-decrypt-decode-residual-parameters", err
	default:
		values, err := b.decodeWithParameters(ct, b.bootstrappingParams)
		return values, "standard-decrypt-decode-bootstrapping-parameters", err
	}
}

func (b *standardBackend) decodeWithParameters(ct *rlwe.Ciphertext, params ckks.Parameters) ([]complex128, error) {
	if ct == nil || ct.Level() < 0 || ct.Level() > params.MaxLevel() {
		return nil, fmt.Errorf("invalid Standard ciphertext level")
	}
	plain := rlwe.NewDecryptor(params, b.secret).DecryptNew(ct)
	values := make([]complex128, params.MaxSlots())
	if err := ckks.NewEncoder(params).Decode(plain, values); err != nil {
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

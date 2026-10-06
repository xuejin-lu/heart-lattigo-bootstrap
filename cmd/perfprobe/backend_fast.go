//go:build perf_fast

package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

type fastBackend struct {
	eval                *bootstrapping.FastEvaluator
	residual            ckks.Parameters
	bootstrappingParams ckks.Parameters
}

func newBackend(params bootstrapping.Parameters, residual ckks.Parameters) (backendAdapter, error) {
	eval, err := bootstrapping.NewFastEvaluator(params)
	if err != nil {
		return nil, err
	}
	return &fastBackend{eval: eval, residual: residual, bootstrappingParams: params.BootstrappingParameters}, nil
}

func (b *fastBackend) Name() string { return "fast" }

func (b *fastBackend) EvaluatorPath() string { return "bootstrapping.NewFastEvaluator" }

func (b *fastBackend) KeyTrialEvidence() (bool, int) { return false, -1 }

func (b *fastBackend) SecretKeyForTrial() *rlwe.SecretKey { return nil }

func (b *fastBackend) Bootstrap(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	return b.eval.Bootstrap(ct)
}

func (b *fastBackend) Decode(ct *rlwe.Ciphertext) ([]complex128, error) {
	return b.decodeWithParameters(ct, b.residual)
}

func (b *fastBackend) decodeWithParameters(ct *rlwe.Ciphertext, params ckks.Parameters) ([]complex128, error) {
	if ct == nil || ct.Level() < 0 || ct.Level() > params.MaxLevel() {
		return nil, fmt.Errorf("invalid Fast ciphertext level")
	}
	plain := ckks.NewPlaintext(params, ct.Level())
	*plain.MetaData = *ct.MetaData
	plain.Value.Copy(ct.Value[0])
	plain.IsNTT, plain.IsMontgomery = ct.IsNTT, ct.IsMontgomery
	values := make([]complex128, params.MaxSlots())
	if err := ckks.NewEncoder(params).Decode(plain, values); err != nil {
		return nil, err
	}
	return values, nil
}

func (b *fastBackend) DecodeStage(ct *rlwe.Ciphertext, name string) ([]complex128, string, error) {
	switch name {
	case "input", "scale_down", "final_public_output":
		values, err := b.decodeWithParameters(ct, b.residual)
		return values, "fast-c0-decode-residual-parameters", err
	default:
		values, err := b.decodeWithParameters(ct, b.bootstrappingParams)
		return values, "fast-c0-decode-bootstrapping-parameters", err
	}
}

func (b *fastBackend) PrefixRows(ct *rlwe.Ciphertext) (int, error) {
	if ct == nil {
		return 0, fmt.Errorf("nil Fast ciphertext")
	}
	return fastckks.QPrefixWidth(ct.Level())
}

func (b *fastBackend) Pack(cts []rlwe.Ciphertext) ([]rlwe.Ciphertext, error) {
	packed, _, _, err := b.eval.PackAndSwitchN1ToN2(cts)
	return packed, err
}

func (b *fastBackend) ScaleDown(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	out, _, err := b.eval.ScaleDown(ct)
	return out, err
}

func (b *fastBackend) ModUp(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) { return b.eval.ModUp(ct) }

func (b *fastBackend) CoeffsToSlots(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	return b.eval.CoeffsToSlots(ct)
}

func (b *fastBackend) EvalMod(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	return b.eval.EvalMod(ct)
}

func (b *fastBackend) SlotsToCoeffs(realCT, imagCT *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	return b.eval.SlotsToCoeffs(realCT, imagCT)
}

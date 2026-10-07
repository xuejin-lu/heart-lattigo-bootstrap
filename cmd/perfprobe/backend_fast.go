//go:build perf_fast

package main

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

type fastBackend struct {
	eval                *bootstrapping.FastEvaluator
	residual            ckks.Parameters
	bootstrappingParams ckks.Parameters
}

func newBackend(params bootstrapping.Parameters, residual ckks.Parameters) (backendAdapter, error) {
	inputBackend, err := newInputBackend(params, residual)
	if err != nil {
		return nil, err
	}
	backend := inputBackend.(*fastBackend)
	eval, err := bootstrapping.NewFastEvaluator(params)
	if err != nil {
		return nil, err
	}
	backend.eval = eval
	return backend, nil
}

func newInputBackend(params bootstrapping.Parameters, residual ckks.Parameters) (backendAdapter, error) {
	return &fastBackend{residual: residual, bootstrappingParams: params.BootstrappingParameters}, nil
}

func (b *fastBackend) Name() string { return "fast" }

func (b *fastBackend) InputKind() string { return fastDirectInputKind }

func (b *fastBackend) InputConstructor() string {
	return fastDirectConstructor
}

func (b *fastBackend) PrepareInput(values []complex128, logSlots int) (*rlwe.Ciphertext, error) {
	plain, err := perfmeasure.EncodeInputPlaintext(b.residual, logSlots, values)
	if err != nil {
		return nil, err
	}
	ct := ckks.NewCiphertext(b.residual, 1, 0)
	*ct.MetaData = *plain.MetaData
	ct.Value[0].Copy(plain.Value)
	ct.Value[1].Zero()
	ct.IsNTT, ct.IsMontgomery = plain.IsNTT, plain.IsMontgomery
	return ct, nil
}

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
	params := b.bootstrappingParams
	switch name {
	case "input", "scale_down", "final_public_output":
		params = b.residual
	}
	projected, err := b.projectStageCiphertext(ct, params)
	if err != nil {
		return nil, "fast-c0-decode-common-q-prefix", err
	}
	values, err := b.decodeWithParameters(projected, params)
	return values, "fast-c0-decode-common-q-prefix", err
}

func (b *fastBackend) projectStageCiphertext(ct *rlwe.Ciphertext, params ckks.Parameters) (*rlwe.Ciphertext, error) {
	if ct == nil {
		return nil, fmt.Errorf("nil Fast stage ciphertext")
	}
	rows, err := fastckks.QPrefixWidth(ct.Level())
	if err != nil {
		return nil, err
	}
	level := rows - 1
	if rows < 1 || level > params.MaxLevel() {
		return nil, fmt.Errorf("Fast Q-prefix rows=%d cannot be projected into parameter level %d", rows, params.MaxLevel())
	}
	projected := ckks.NewCiphertext(params, ct.Degree(), level)
	*projected.MetaData = *ct.MetaData
	projected.IsNTT, projected.IsMontgomery = ct.IsNTT, false
	ringQ := params.RingQ()
	for component := 0; component <= ct.Degree(); component++ {
		if component >= len(ct.Value) || component >= len(projected.Value) {
			return nil, fmt.Errorf("Fast ciphertext component %d is unavailable", component)
		}
		for limb := 0; limb < rows; limb++ {
			if limb >= len(ct.Value[component].Coeffs) || limb >= len(projected.Value[component].Coeffs) {
				return nil, fmt.Errorf("Fast ciphertext component %d Q row %d is unavailable", component, limb)
			}
			copy(projected.Value[component].Coeffs[limb], ct.Value[component].Coeffs[limb])
			if ct.IsMontgomery {
				ringQ.SubRings[limb].IMForm(projected.Value[component].Coeffs[limb], projected.Value[component].Coeffs[limb])
			}
		}
	}
	return projected, nil
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

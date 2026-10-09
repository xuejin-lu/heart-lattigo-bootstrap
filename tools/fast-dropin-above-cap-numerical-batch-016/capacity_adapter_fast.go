//go:build !lattigo_standard

package main

import (
	"errors"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

type fastCapacityAudit struct {
	evaluator *fastckks.Evaluator
}

func makeCapacityAudit(params ckks.Parameters) capacityAudit {
	return &fastCapacityAudit{evaluator: fastckks.NewEvaluator(params)}
}

func (audit *fastCapacityAudit) Observe(name string, ct *rlwe.Ciphertext, rows int) (capacityObservation, error) {
	var got fastckks.QPrefixCapacitySnapshot
	seen := false
	audit.evaluator.SetQPrefixCapacityObserver(func(snapshot fastckks.QPrefixCapacitySnapshot) error {
		got, seen = snapshot, true
		return nil
	})
	if err := audit.evaluator.ObserveQPrefixCapacity(name, ct, rows); err != nil {
		return capacityObservation{}, err
	}
	if !seen {
		return capacityObservation{}, errors.New("existing observer was not invoked")
	}
	return capacityObservation{
		Rows: got.Rows, PrefixProduct: got.PrefixProduct,
		MaxAbs: got.MaxAbs, StrictFit: got.StrictFit,
	}, nil
}

//go:build !lattigo_standard

package main

import (
	"errors"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
)

type fastCapacityAudit struct{ evaluator *fastckks.Evaluator }

func isFastBackend() bool { return true }

func makeCapacityAudit(params ckks.Parameters) capacityAudit {
	return &fastCapacityAudit{evaluator: fastckks.NewEvaluator(params)}
}

func (audit *fastCapacityAudit) Observe(name string, ct *rlwe.Ciphertext, rows int) (int, string, []string, bool, error) {
	var snapshot fastckks.QPrefixCapacitySnapshot
	seen := false
	audit.evaluator.SetQPrefixCapacityObserver(func(value fastckks.QPrefixCapacitySnapshot) error {
		snapshot, seen = value, true
		return nil
	})
	if err := audit.evaluator.ObserveQPrefixCapacity(name, ct, rows); err != nil {
		return 0, "", nil, false, err
	}
	if !seen {
		return 0, "", nil, false, errors.New("existing Q-prefix capacity observer was not invoked")
	}
	return snapshot.Rows, snapshot.PrefixProduct, snapshot.MaxAbs, snapshot.StrictFit, nil
}

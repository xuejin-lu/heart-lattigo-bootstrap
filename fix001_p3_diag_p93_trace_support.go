//go:build !lattigo_standard

package main

import "github.com/tuneinsight/lattigo/v6/core/rlwe"

type fix001P3TraceComplex struct {
	Real float64 `json:"r"`
	Imag float64 `json:"i"`
}

type fix001P3TraceEvent struct {
	Kind      string                 `json:"kind"`
	Branch    string                 `json:"branch"`
	Name      string                 `json:"name"`
	Round     int                    `json:"round"`
	Level     int                    `json:"level"`
	Scale     string                 `json:"scale"`
	Degree    int                    `json:"degree"`
	Capacity  *postMod1S2CCapacity   `json:"capacity,omitempty"`
	RowHashes map[string]string      `json:"row_hashes,omitempty"`
	Values    []fix001P3TraceComplex `json:"values"`
}

var fix001P3TraceSink func(fix001P3TraceEvent)
var fix001P3TraceBranch string

func fix001P3TraceRecord(kind, name string, round int, ct *rlwe.Ciphertext, values []complex128) {
	if fix001P3TraceSink == nil || ct == nil {
		return
	}
	event := fix001P3TraceEvent{Kind: kind, Branch: fix001P3TraceBranch, Name: name, Round: round, Level: ct.Level(), Scale: finalizationScaleString(ct.Scale), Degree: ct.Degree(), RowHashes: fix001P3ActualForcedRowHashes(ct), Values: make([]fix001P3TraceComplex, len(values))}
	for i, value := range values {
		event.Values[i] = fix001P3TraceComplex{Real: real(value), Imag: imag(value)}
	}
	fix001P3TraceSink(event)
}

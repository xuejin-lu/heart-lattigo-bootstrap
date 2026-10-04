// Package numericalmetrics contains reusable decoded-domain numerical metrics.
package numericalmetrics

import (
	"math"
	"math/cmplx"
)

// SNRStatus identifies mathematically special or non-comparable SNR cases.
type SNRStatus string

const (
	Finite              SNRStatus = "FINITE"
	PositiveInfinity    SNRStatus = "POSITIVE_INFINITY"
	UndefinedZeroSignal SNRStatus = "UNDEFINED_ZERO_SIGNAL"
	NotComparable       SNRStatus = "NOT_COMPARABLE"
)

// SNR compares an observed complex vector against a reference signal. All
// powers use complex magnitude squared, |z|^2 = real(z)^2 + imag(z)^2.
// SNRDB is null for mathematically infinite, undefined, or non-comparable
// cases so the value can be serialized as strict JSON without NaN or Inf.
type SNR struct {
	SignalPower *float64  `json:"signal_power"`
	SignalRMS   *float64  `json:"signal_rms"`
	NoisePower  *float64  `json:"noise_power"`
	NoiseRMSE   *float64  `json:"noise_rmse"`
	SNRDB       *float64  `json:"snr_db"`
	Status      SNRStatus `json:"status"`
	Reason      string    `json:"reason,omitempty"`
}

// Compare computes the decoded-domain SNR for output relative to reference.
// It does not apply an epsilon floor: zero-error and zero-signal cases retain
// their explicit mathematical status.
func Compare(reference, output []complex128) SNR {
	if len(reference) == 0 || len(reference) != len(output) {
		return SNR{Status: NotComparable, Reason: "reference and output must have the same non-zero length"}
	}

	var signalSquares, noiseSquares float64
	for i, signal := range reference {
		observed := output[i]
		if !finiteComplex(signal) || !finiteComplex(observed) {
			return SNR{Status: NotComparable, Reason: "reference or output contains a non-finite complex value"}
		}
		signalMagnitude := cmplx.Abs(signal)
		noiseMagnitude := cmplx.Abs(observed - signal)
		signalSquares += signalMagnitude * signalMagnitude
		noiseSquares += noiseMagnitude * noiseMagnitude
	}

	n := float64(len(reference))
	signalPower := signalSquares / n
	noisePower := noiseSquares / n
	signalRMS := math.Sqrt(signalPower)
	noiseRMSE := math.Sqrt(noisePower)
	if !finite(signalPower) || !finite(noisePower) || !finite(signalRMS) || !finite(noiseRMSE) {
		return SNR{Status: NotComparable, Reason: "mean power overflowed the finite float64 range"}
	}

	result := SNR{
		SignalPower: ptr(signalPower), SignalRMS: ptr(signalRMS),
		NoisePower: ptr(noisePower), NoiseRMSE: ptr(noiseRMSE),
	}
	switch {
	case signalPower == 0:
		result.Status = UndefinedZeroSignal
	case noisePower == 0:
		result.Status = PositiveInfinity
	default:
		db := 10 * math.Log10(signalPower/noisePower)
		if !finite(db) {
			return SNR{Status: NotComparable, Reason: "computed SNR is outside the finite float64 range"}
		}
		result.Status = Finite
		result.SNRDB = ptr(db)
	}
	return result
}

// Unavailable constructs an explicit non-comparable metric for semantic
// checkpoints that cannot be decoded or aligned.
func Unavailable(reason string) SNR {
	return SNR{Status: NotComparable, Reason: reason}
}

func finiteComplex(value complex128) bool {
	return finite(real(value)) && finite(imag(value))
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func ptr(value float64) *float64 { return &value }

// Package numericalmetrics contains reusable decoded-domain numerical metrics.
package numericalmetrics

import (
	"math"
	"math/cmplx"
	"sort"
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

// RealVectorError contains the shared decoded-real-vector error statistics
// used by experiment reports. WorstIndex is the first slot attaining MaxAbs.
type RealVectorError struct {
	Count      int     `json:"count"`
	MaxAbs     float64 `json:"max_abs"`
	MeanAbs    float64 `json:"mean_abs"`
	RMSE       float64 `json:"rmse"`
	WorstIndex int     `json:"worst_index"`
	Status     string  `json:"status"`
	Reason     string  `json:"reason,omitempty"`
}

// RealVectorQuality contains the complete decoded-real error distribution
// and both global and per-slot precision summaries. The per-slot median is
// explicitly not interchangeable with global RMSE precision.
type RealVectorQuality struct {
	Count                        int      `json:"count"`
	MaxAbs                       float64  `json:"max_abs"`
	MeanAbs                      float64  `json:"mean_abs"`
	RMSE                         float64  `json:"rmse"`
	P99                          float64  `json:"p99"`
	P999                         float64  `json:"p99_9"`
	WorstIndex                   int      `json:"worst_index"`
	GlobalRMSEPrecisionBits      *float64 `json:"global_rmse_precision_bits,omitempty"`
	GlobalRMSEPrecisionStatus    string   `json:"global_rmse_precision_status"`
	MedianPerSlotPrecisionBits   *float64 `json:"median_per_slot_precision_bits,omitempty"`
	MedianPerSlotPrecisionStatus string   `json:"median_per_slot_precision_status"`
	ExactMatchSlots              int      `json:"exact_match_slots"`
	Status                       string   `json:"status"`
	Reason                       string   `json:"reason,omitempty"`
}

// CompareRealVectors computes absolute-error metrics for aligned real vectors.
// It rejects empty, mismatched, non-finite, or overflowed inputs explicitly.
func CompareRealVectors(reference, output []float64) RealVectorError {
	quality := CompareRealVectorQuality(reference, output)
	return RealVectorError{
		Count: quality.Count, MaxAbs: quality.MaxAbs, MeanAbs: quality.MeanAbs,
		RMSE: quality.RMSE, WorstIndex: quality.WorstIndex,
		Status: quality.Status, Reason: quality.Reason,
	}
}

// CompareRealVectorQuality computes finite real-vector error metrics, nearest-
// rank p99/p99.9 tails, global RMSE precision, and median per-slot precision.
// A zero RMSE or a zero per-slot error is represented explicitly as positive
// infinity through the corresponding status field, never serialized as Inf.
func CompareRealVectorQuality(reference, output []float64) RealVectorQuality {
	if len(reference) == 0 || len(reference) != len(output) {
		return RealVectorQuality{Status: "NOT_COMPARABLE", Reason: "vectors must have the same non-zero length"}
	}
	var absoluteSum, squaredSum float64
	var maxAbs float64
	worstIndex := 0
	exactSlots := 0
	absoluteErrors := make([]float64, len(reference))
	perSlotBits := make([]float64, len(reference))
	for index, ref := range reference {
		observed := output[index]
		if !finite(ref) || !finite(observed) {
			return RealVectorQuality{Status: "NOT_COMPARABLE", Reason: "vector contains a non-finite value"}
		}
		difference := math.Abs(observed - ref)
		if !finite(difference) {
			return RealVectorQuality{Status: "NOT_COMPARABLE", Reason: "absolute difference is not finite"}
		}
		absoluteErrors[index] = difference
		absoluteSum += difference
		squaredSum += difference * difference
		if difference == 0 {
			perSlotBits[index] = math.Inf(1)
			exactSlots++
		} else {
			perSlotBits[index] = -math.Log2(difference)
		}
		if difference > maxAbs {
			maxAbs, worstIndex = difference, index
		}
	}
	count := len(reference)
	meanAbs := absoluteSum / float64(count)
	rmse := math.Sqrt(squaredSum / float64(count))
	if !finite(absoluteSum) || !finite(squaredSum) || !finite(meanAbs) || !finite(rmse) {
		return RealVectorQuality{Status: "NOT_COMPARABLE", Reason: "error aggregation overflowed the finite float64 range"}
	}
	sort.Float64s(absoluteErrors)
	sort.Float64s(perSlotBits)
	nearestRank := func(p float64) float64 {
		rank := int(math.Ceil(p * float64(count)))
		if rank < 1 {
			rank = 1
		}
		return absoluteErrors[rank-1]
	}
	quality := RealVectorQuality{
		Count: count, MaxAbs: maxAbs, MeanAbs: meanAbs, RMSE: rmse,
		P99: nearestRank(0.99), P999: nearestRank(0.999), WorstIndex: worstIndex,
		ExactMatchSlots: exactSlots, Status: "FINITE",
	}
	if rmse == 0 {
		quality.GlobalRMSEPrecisionStatus = "POSITIVE_INFINITY_ZERO_RMSE"
	} else {
		bits := -math.Log2(rmse)
		if !finite(bits) {
			return RealVectorQuality{Status: "NOT_COMPARABLE", Reason: "global RMSE precision overflowed the finite float64 range"}
		}
		quality.GlobalRMSEPrecisionBits = &bits
		quality.GlobalRMSEPrecisionStatus = "FINITE"
	}
	median := perSlotBits[count/2]
	if count%2 == 0 {
		lower, upper := perSlotBits[count/2-1], perSlotBits[count/2]
		if math.IsInf(lower, 1) || math.IsInf(upper, 1) {
			median = math.Inf(1)
		} else {
			median = (lower + upper) / 2
		}
	}
	if math.IsInf(median, 1) {
		quality.MedianPerSlotPrecisionStatus = "POSITIVE_INFINITY_AT_LEAST_HALF_EXACT"
	} else if !finite(median) {
		return RealVectorQuality{Status: "NOT_COMPARABLE", Reason: "per-slot precision median is not finite"}
	} else {
		quality.MedianPerSlotPrecisionBits = &median
		quality.MedianPerSlotPrecisionStatus = "FINITE"
	}
	return quality
}

func finiteComplex(value complex128) bool {
	return finite(real(value)) && finite(imag(value))
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func ptr(value float64) *float64 { return &value }

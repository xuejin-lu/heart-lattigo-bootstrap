package numericalmetrics

import (
	"math"
	"math/cmplx"
)

// ComplexVectorStats summarizes a finite decoded complex vector. Variance is
// the population E[|z-mean(z)|^2], while RMS is sqrt(E[|z|^2]).
type ComplexVectorStats struct {
	Count     int     `json:"count"`
	MeanReal  float64 `json:"mean_real"`
	MeanImag  float64 `json:"mean_imag"`
	Variance  float64 `json:"variance"`
	RMS       float64 `json:"rms"`
	MaxAbs    float64 `json:"max_abs"`
	WorstSlot int     `json:"worst_slot"`
	Status    string  `json:"status"`
	Reason    string  `json:"reason,omitempty"`
}

// ComplexCovariance summarizes E[(x-mean(x))*conj(y-mean(y))] for paired
// decoded complex coordinates.
type ComplexCovariance struct {
	Count                int     `json:"count"`
	Real                 float64 `json:"real"`
	Imag                 float64 `json:"imag"`
	CorrelationMagnitude float64 `json:"correlation_magnitude"`
	Status               string  `json:"status"`
	Reason               string  `json:"reason,omitempty"`
}

// SummarizeComplexVector computes population mean, complex variance, RMS and
// maximum magnitude using the same decoded-vector validation policy as the
// shared numerical metrics package.
func SummarizeComplexVector(values []complex128) ComplexVectorStats {
	if len(values) == 0 {
		return ComplexVectorStats{Status: "NOT_COMPARABLE", Reason: "vector must be non-empty"}
	}
	var mean complex128
	var sumSquares float64
	var maxAbs float64
	worst := 0
	for i, value := range values {
		if !finiteComplex(value) {
			return ComplexVectorStats{Status: "NOT_COMPARABLE", Reason: "vector contains a non-finite value"}
		}
		mean += value
		magnitude := cmplx.Abs(value)
		sumSquares += magnitude * magnitude
		if magnitude > maxAbs {
			maxAbs, worst = magnitude, i
		}
	}
	count := float64(len(values))
	mean /= complex(count, 0)
	var centeredSquares float64
	for _, value := range values {
		difference := cmplx.Abs(value - mean)
		centeredSquares += difference * difference
	}
	variance := centeredSquares / count
	rms := math.Sqrt(sumSquares / count)
	if !finiteComplex(mean) || !finite(variance) || !finite(rms) || !finite(maxAbs) {
		return ComplexVectorStats{Status: "NOT_COMPARABLE", Reason: "complex-vector aggregation overflowed the finite float64 range"}
	}
	return ComplexVectorStats{
		Count: len(values), MeanReal: real(mean), MeanImag: imag(mean),
		Variance: variance, RMS: rms, MaxAbs: maxAbs, WorstSlot: worst, Status: "FINITE",
	}
}

// CompareComplexCovariance computes a population covariance and normalized
// magnitude for paired complex vectors. It does not claim statistical
// independence from a zero sample covariance.
func CompareComplexCovariance(x, y []complex128) ComplexCovariance {
	if len(x) == 0 || len(x) != len(y) {
		return ComplexCovariance{Status: "NOT_COMPARABLE", Reason: "vectors must have the same non-zero length"}
	}
	var meanX, meanY complex128
	for i := range x {
		if !finiteComplex(x[i]) || !finiteComplex(y[i]) {
			return ComplexCovariance{Status: "NOT_COMPARABLE", Reason: "vector contains a non-finite value"}
		}
		meanX += x[i]
		meanY += y[i]
	}
	count := float64(len(x))
	meanX /= complex(count, 0)
	meanY /= complex(count, 0)
	var covariance complex128
	var varianceX, varianceY float64
	for i := range x {
		dx, dy := x[i]-meanX, y[i]-meanY
		covariance += dx * cmplx.Conj(dy)
		varianceX += cmplx.Abs(dx) * cmplx.Abs(dx)
		varianceY += cmplx.Abs(dy) * cmplx.Abs(dy)
	}
	covariance /= complex(count, 0)
	varianceX /= count
	varianceY /= count
	correlation := 0.0
	if varianceX > 0 && varianceY > 0 {
		correlation = cmplx.Abs(covariance) / math.Sqrt(varianceX*varianceY)
	}
	if !finiteComplex(covariance) || !finite(correlation) {
		return ComplexCovariance{Status: "NOT_COMPARABLE", Reason: "covariance aggregation overflowed the finite float64 range"}
	}
	return ComplexCovariance{Count: len(x), Real: real(covariance), Imag: imag(covariance), CorrelationMagnitude: correlation, Status: "FINITE"}
}

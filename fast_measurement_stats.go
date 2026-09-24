package main

import (
	"errors"
	"math"
	"sort"
)

// FastDistributionStats intentionally stores only aggregates, never the input
// vector. Quantiles use deterministic linear interpolation on sorted values.
type FastDistributionStats struct {
	Count   int     `json:"count"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Mean    float64 `json:"mean"`
	StdDev  float64 `json:"std_dev"`
	RMS     float64 `json:"rms"`
	AbsMax  float64 `json:"abs_max"`
	P50     float64 `json:"p50"`
	P90     float64 `json:"p90"`
	P99     float64 `json:"p99"`
	P999    float64 `json:"p99_9"`
	AbsP50  float64 `json:"abs_p50"`
	AbsP90  float64 `json:"abs_p90"`
	AbsP99  float64 `json:"abs_p99"`
	AbsP999 float64 `json:"abs_p99_9"`
}

type FastNumericSample struct {
	Index int     `json:"index"`
	Value float64 `json:"value"`
}

func FastSummarizeValues(values []float64) (FastDistributionStats, error) {
	if len(values) == 0 {
		return FastDistributionStats{}, errors.New("cannot summarize an empty vector")
	}
	sorted := append([]float64(nil), values...)
	absSorted := make([]float64, len(values))
	minValue, maxValue := values[0], values[0]
	mean, sumSquares, squareMean := 0.0, 0.0, 0.0
	for i, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return FastDistributionStats{}, errors.New("statistics input contains a non-finite value")
		}
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
		absSorted[i] = math.Abs(value)
		delta := value - mean
		mean += delta / float64(i+1)
		sumSquares += delta * (value - mean)
		squareMean += value * value
	}
	sort.Float64s(sorted)
	sort.Float64s(absSorted)
	variance := sumSquares / float64(len(values))
	if variance < 0 {
		if variance > -1e-14*math.Max(1, mean*mean) {
			variance = 0
		} else {
			return FastDistributionStats{}, errors.New("statistics variance became negative")
		}
	}
	if math.IsInf(mean, 0) || math.IsNaN(mean) || math.IsInf(squareMean, 0) || math.IsInf(variance, 0) || math.IsNaN(variance) {
		return FastDistributionStats{}, errors.New("statistics aggregate exceeds float64 range")
	}
	absMax := math.Max(math.Abs(minValue), math.Abs(maxValue))
	return FastDistributionStats{
		Count: len(values), Min: minValue, Max: maxValue, Mean: mean,
		StdDev: math.Sqrt(variance),
		RMS:    math.Sqrt(squareMean / float64(len(values))), AbsMax: absMax,
		P50: fastLinearQuantile(sorted, 0.50), P90: fastLinearQuantile(sorted, 0.90),
		P99: fastLinearQuantile(sorted, 0.99), P999: fastLinearQuantile(sorted, 0.999),
		AbsP50: fastLinearQuantile(absSorted, 0.50), AbsP90: fastLinearQuantile(absSorted, 0.90),
		AbsP99: fastLinearQuantile(absSorted, 0.99), AbsP999: fastLinearQuantile(absSorted, 0.999),
	}, nil
}

func FastDeterministicSample(values []float64, limit int) []FastNumericSample {
	if len(values) == 0 || limit <= 0 {
		return nil
	}
	if limit >= len(values) {
		limit = len(values)
	}
	if limit == 1 {
		return []FastNumericSample{{Index: 0, Value: values[0]}}
	}
	samples := make([]FastNumericSample, limit)
	for i := 0; i < limit; i++ {
		index := i * (len(values) - 1) / (limit - 1)
		samples[i] = FastNumericSample{Index: index, Value: values[index]}
	}
	return samples
}

func fastLinearQuantile(sorted []float64, quantile float64) float64 {
	if len(sorted) == 1 {
		return sorted[0]
	}
	position := quantile * float64(len(sorted)-1)
	lower := int(math.Floor(position))
	upper := int(math.Ceil(position))
	if lower == upper {
		return sorted[lower]
	}
	weight := position - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

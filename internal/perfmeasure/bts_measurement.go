package perfmeasure

import (
	"math"
	"sort"
	"time"
)

const BTSNoTrace008SchemaVersion = "bts-notrace-warm-and-numerical-008-run.v1"
const BTSFastNoise009SchemaVersion = "bts-fast-noise-fidelity-009-run.v1"

// BTSRunDocument is the compact manifest emitted by the Batch008 measurement
// adapter after each completed sample. Vector payloads stay in f64le files.
type BTSRunDocument struct {
	SchemaVersion            string                     `json:"schema_version"`
	Task                     string                     `json:"task"`
	NoiseMode                string                     `json:"noise_mode,omitempty"`
	NoiseSigma               float64                    `json:"noise_sigma,omitempty"`
	NoiseRootSeed            int64                      `json:"noise_root_seed,omitempty"`
	NoiseHookSupported       *bool                      `json:"noise_hook_supported,omitempty"`
	NoiseHookEnabled         *bool                      `json:"noise_hook_enabled,omitempty"`
	ProfileSHA256            string                     `json:"profile_sha256,omitempty"`
	ArmConfigSHA256          string                     `json:"arm_config_sha256,omitempty"`
	Circuit                  string                     `json:"bootstrap_circuit,omitempty"`
	Status                   string                     `json:"status"`
	Backend                  string                     `json:"backend"`
	BackendCommit            string                     `json:"backend_commit"`
	BinarySHA256             string                     `json:"binary_sha256"`
	BTSManifestSHA256        string                     `json:"bts_source_manifest_sha256"`
	MeasurementAdapterSHA256 string                     `json:"measurement_adapter_sha256"`
	PrimaryCommit            string                     `json:"primary_commit"`
	OriginalSourceRoot       string                     `json:"original_source_root"`
	StartedAt                time.Time                  `json:"started_at"`
	UpdatedAt                time.Time                  `json:"updated_at"`
	MeasurementProcesses     int                        `json:"measurement_processes"`
	BootstrapBudget          int                        `json:"bootstrap_budget"`
	BootstrapReservations    int                        `json:"bootstrap_reservations"`
	NewBTSNS                 int64                      `json:"new_bts_elapsed_ns"`
	Environment              BTSMeasurementEnvironment  `json:"environment"`
	Parameters               BTSMeasurementParameters   `json:"effective_parameters"`
	Input                    BTSVectorArtifact          `json:"input_vector"`
	Samples                  []BTSBootstrapSample       `json:"bootstrap_samples"`
	NoiseEvidence            []BTSNoiseEvidenceArtifact `json:"noise_evidence,omitempty"`
	Failure                  string                     `json:"failure,omitempty"`
}

type BTSMeasurementEnvironment struct {
	GoVersion  string `json:"go_version"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	CPU        string `json:"cpu"`
	NumCPU     int    `json:"num_cpu"`
	GOMAXPROCS int    `json:"gomaxprocs"`
}

type BTSMeasurementParameters struct {
	LogN        int      `json:"log_n"`
	RingN       int      `json:"ring_n"`
	Slots       int      `json:"slots"`
	SlotLimbs   int      `json:"slot_limbs"`
	Mode        string   `json:"mode"`
	Seed        int64    `json:"seed"`
	Amplitude   float64  `json:"amplitude"`
	EphemeralH  int      `json:"ephemeral_h"`
	MaxLevel    int      `json:"max_level,omitempty"`
	QCount      int      `json:"q_count,omitempty"`
	PCount      int      `json:"p_count,omitempty"`
	Mod1Level   int      `json:"mod1_level,omitempty"`
	Mod1Scale   float64  `json:"mod1_scale,omitempty"`
	Mod1Degree  int      `json:"mod1_chebyshev_degree,omitempty"`
	Mod1PSDepth int      `json:"mod1_ps_depth,omitempty"`
	DoubleAngle int      `json:"mod1_double_angle,omitempty"`
	QPrimes     []uint64 `json:"q_primes"`
	PPrimes     []uint64 `json:"p_primes"`
	InputLevel  int      `json:"input_level"`
	InputScale  float64  `json:"input_scale"`
	OutputLevel int      `json:"output_level"`
	OutputScale float64  `json:"output_scale"`
}

type BTSVectorArtifact struct {
	Path        string  `json:"path"`
	Encoding    string  `json:"encoding"`
	SHA256      string  `json:"sha256"`
	Slots       int     `json:"slots"`
	FiniteSlots int     `json:"finite_slots"`
	Level       int     `json:"level"`
	Scale       float64 `json:"scale"`
}

type BTSBootstrapSample struct {
	AttemptIndex          int     `json:"attempt_index"`
	Phase                 string  `json:"phase"`
	BootstrapElapsedNS    int64   `json:"bootstrap_elapsed_ns"`
	AllocatedBytes        uint64  `json:"allocated_bytes,omitempty"`
	Allocations           uint64  `json:"allocation_count,omitempty"`
	InputLevel            int     `json:"input_level"`
	InputScale            float64 `json:"input_scale"`
	OutputCiphertextLevel int     `json:"output_ciphertext_level"`
	OutputCiphertextScale float64 `json:"output_ciphertext_scale"`
	DecodedPlaintextLevel int     `json:"decoded_plaintext_level"`
	DecodedPlaintextScale float64 `json:"decoded_plaintext_scale"`
	DecodedSlots          int     `json:"decoded_slots"`
	FiniteDecodedSlots    int     `json:"finite_decoded_slots"`
	VectorPath            string  `json:"decoded_vector_path"`
	VectorEncoding        string  `json:"decoded_vector_encoding"`
	VectorSHA256          string  `json:"decoded_vector_sha256"`
	MaxAbsErrorVsInput    float64 `json:"max_abs_error_vs_input"`
	MeanAbsErrorVsInput   float64 `json:"mean_abs_error_vs_input"`
	WorstErrorSlot        int     `json:"worst_error_slot"`
}

// BTSNoiseEvidenceArtifact points to a per-call, content-addressed Noise Hook
// event capture. Keeping the event tree out of the compact run manifest avoids
// embedding large per-multiplication arrays in summary outputs.
type BTSNoiseEvidenceArtifact struct {
	AttemptIndex        int    `json:"attempt_index"`
	Phase               string `json:"phase"`
	Path                string `json:"path"`
	SHA256              string `json:"sha256"`
	Bytes               int64  `json:"bytes"`
	OperationStart      uint64 `json:"operation_start"`
	OperationEnd        uint64 `json:"operation_end"`
	CtCtMultiplyCalls   uint64 `json:"ct_ct_multiply_calls"`
	InjectedPolynomials uint64 `json:"injected_polynomials"`
	EventCount          int    `json:"event_count"`
}

// BTSRepeatedStatistics summarizes a fixed sample set. Standard deviation
// uses the sample denominator (n-1), matching the repeated-measurement
// convention. Generic units allow identical statistics for time and memory.
type BTSRepeatedStatistics struct {
	Count                   int     `json:"count"`
	Mean                    float64 `json:"mean"`
	Median                  float64 `json:"median"`
	SampleStandardDeviation float64 `json:"sample_standard_deviation"`
	Min                     float64 `json:"min"`
	Max                     float64 `json:"max"`
}

// SummarizeBTSRepeatedSamples computes descriptive statistics for an already
// captured sample vector; it performs no measurement or warmup.
func SummarizeBTSRepeatedSamples(samples []int64) BTSRepeatedStatistics {
	if len(samples) == 0 {
		return BTSRepeatedStatistics{}
	}
	values := make([]float64, len(samples))
	for index, sample := range samples {
		values[index] = float64(sample)
	}
	return SummarizeBTSRepeatedValues(values)
}

// SummarizeBTSRepeatedValues computes the same descriptive statistics for
// captured allocation-byte and allocation-count vectors.
func SummarizeBTSRepeatedValues(samples []float64) BTSRepeatedStatistics {
	if len(samples) == 0 {
		return BTSRepeatedStatistics{}
	}
	ordered := append([]float64(nil), samples...)
	sort.Float64s(ordered)
	minimum, maximum := ordered[0], ordered[len(ordered)-1]
	median := ordered[len(ordered)/2]
	if len(ordered)%2 == 0 {
		median = (ordered[len(ordered)/2-1] + ordered[len(ordered)/2]) / 2
	}
	var sum float64
	for _, sample := range samples {
		sum += sample
	}
	mean := sum / float64(len(samples))
	var squared float64
	for _, sample := range samples {
		delta := sample - mean
		squared += delta * delta
	}
	stddev := 0.0
	if len(samples) > 1 {
		stddev = math.Sqrt(squared / float64(len(samples)-1))
	}
	return BTSRepeatedStatistics{
		Count: len(samples), Mean: mean, Median: median,
		SampleStandardDeviation: stddev, Min: minimum, Max: maximum,
	}
}

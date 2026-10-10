package main

import "time"

type Event struct {
	Scope          string   `json:"scope"`
	Name           string   `json:"name"`
	ElapsedNS      int64    `json:"elapsed_ns"`
	Sequence       uint64   `json:"sequence"`
	ParentSequence uint64   `json:"parent_sequence,omitempty"`
	LevelIn        *int     `json:"level_in,omitempty"`
	LevelOut       *int     `json:"level_out,omitempty"`
	RowsIn         *int     `json:"rows_in,omitempty"`
	RowsOut        *int     `json:"rows_out,omitempty"`
	DegreeIn       *int     `json:"degree_in,omitempty"`
	DegreeOut      *int     `json:"degree_out,omitempty"`
	ScaleLog2In    *float64 `json:"scale_log2_in,omitempty"`
	ScaleLog2Out   *float64 `json:"scale_log2_out,omitempty"`
	InPlace        *bool    `json:"in_place,omitempty"`
	Power          *int     `json:"power,omitempty"`
	SplitA         *int     `json:"split_a,omitempty"`
	SplitB         *int     `json:"split_b,omitempty"`
	Component      string   `json:"component,omitempty"`
	Count          *int     `json:"count,omitempty"`
}

type TraceRun struct {
	Index          int     `json:"index"`
	ElapsedNS      int64   `json:"elapsed_ns"`
	NumericalMatch bool    `json:"numerical_match"`
	Events         []Event `json:"events"`
}

type TracePayload struct {
	SchemaVersion string     `json:"schema_version"`
	Timestamp     time.Time  `json:"timestamp"`
	Profile       string     `json:"profile"`
	Trace         []string   `json:"trace"`
	Warmup        int        `json:"warmup"`
	Repetitions   int        `json:"repetitions"`
	InputSHA256   string     `json:"input_sha256"`
	InputLength   int        `json:"input_length"`
	LogN          int        `json:"log_n"`
	LogSlots      int        `json:"log_slots"`
	QChainBits    []int      `json:"q_chain_bits"`
	PBits         []int      `json:"p_bits"`
	Polynomial    int        `json:"polynomial_degree"`
	DoubleAngle   int        `json:"double_angle"`
	GoVersion     string     `json:"go_version"`
	OS            string     `json:"os"`
	Arch          string     `json:"arch"`
	NumCPU        int        `json:"num_cpu"`
	GOMAXPROCS    int        `json:"gomaxprocs"`
	Runs          []TraceRun `json:"runs"`
}

type RepositoryMetadata struct {
	Path   string `json:"path"`
	Commit string `json:"commit"`
	Ref    string `json:"ref,omitempty"`
	Dirty  bool   `json:"dirty"`
}

type EnvironmentMetadata struct {
	GoVersion  string `json:"go_version"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	CPU        string `json:"cpu"`
	NumCPU     int    `json:"num_cpu"`
	GOMAXPROCS int    `json:"gomaxprocs"`
}

type EventMedian struct {
	Key       string  `json:"key"`
	ParentKey string  `json:"parent_key,omitempty"`
	Scope     string  `json:"scope"`
	Name      string  `json:"name"`
	Power     *int    `json:"power,omitempty"`
	SplitA    *int    `json:"split_a,omitempty"`
	SplitB    *int    `json:"split_b,omitempty"`
	Component string  `json:"component,omitempty"`
	Count     int     `json:"count"`
	MedianNS  float64 `json:"median_ns"`
}

type TraceDocument struct {
	SchemaVersion string              `json:"schema_version"`
	Timestamp     time.Time           `json:"timestamp"`
	Profile       string              `json:"profile"`
	Trace         []string            `json:"trace"`
	Warmup        int                 `json:"warmup"`
	Repetitions   int                 `json:"repetitions"`
	Primary       RepositoryMetadata  `json:"primary_repository"`
	Secondary     RepositoryMetadata  `json:"secondary_repository"`
	Environment   EnvironmentMetadata `json:"environment"`
	Workload      TraceWorkload       `json:"workload"`
	Runs          []TraceRun          `json:"runs"`
	EventMedians  []EventMedian       `json:"event_medians"`
}

type TraceWorkload struct {
	FingerprintSHA256 string `json:"fingerprint_sha256"`
	InputLength       int    `json:"input_length"`
	LogN              int    `json:"log_n"`
	LogSlots          int    `json:"log_slots"`
	QChainBits        []int  `json:"q_chain_bits"`
	PBits             []int  `json:"p_bits"`
	PolynomialDegree  int    `json:"polynomial_degree"`
	DoubleAngle       int    `json:"double_angle"`
}

type RefTrace struct {
	RequestedRef string              `json:"requested_ref"`
	Status       string              `json:"status"`
	Repository   RepositoryMetadata  `json:"secondary_repository"`
	Environment  EnvironmentMetadata `json:"environment,omitempty"`
	Workload     TraceWorkload       `json:"workload,omitempty"`
	Runs         []TraceRun          `json:"runs,omitempty"`
	EventMedians []EventMedian       `json:"event_medians,omitempty"`
	Message      string              `json:"message,omitempty"`
}

type EventDelta struct {
	Key                  string   `json:"key"`
	Scope                string   `json:"scope"`
	Name                 string   `json:"name"`
	Power                *int     `json:"power,omitempty"`
	SplitA               *int     `json:"split_a,omitempty"`
	SplitB               *int     `json:"split_b,omitempty"`
	Component            string   `json:"component,omitempty"`
	BaselineMedianNS     float64  `json:"baseline_median_ns"`
	CandidateMedianNS    float64  `json:"candidate_median_ns"`
	Ratio                *float64 `json:"ratio,omitempty"`
	DeltaNS              float64  `json:"delta_ns"`
	ContributionToParent *float64 `json:"contribution_to_parent_delta,omitempty"`
	SampleCount          int      `json:"sample_count"`
}

type ParentClosure struct {
	ParentKey         string   `json:"parent_key"`
	ParentDeltaNS     float64  `json:"parent_delta_ns"`
	ChildrenDeltaNS   float64  `json:"children_delta_ns"`
	ClosureResidualNS float64  `json:"closure_residual_ns"`
	ResidualFraction  *float64 `json:"residual_fraction,omitempty"`
}

type RescaleShares struct {
	BaselineRescaleNS             float64  `json:"baseline_rescale_ns"`
	CandidateRescaleNS            float64  `json:"candidate_rescale_ns"`
	BaselinePreflightShare        *float64 `json:"baseline_preflight_share,omitempty"`
	CandidatePreflightShare       *float64 `json:"candidate_preflight_share,omitempty"`
	BaselineMaterializationShare  *float64 `json:"baseline_materialization_share,omitempty"`
	CandidateMaterializationShare *float64 `json:"candidate_materialization_share,omitempty"`
}

type CompareDocument struct {
	SchemaVersion string             `json:"schema_version"`
	Timestamp     time.Time          `json:"timestamp"`
	Profile       string             `json:"profile"`
	Trace         []string           `json:"trace"`
	Warmup        int                `json:"warmup"`
	Repetitions   int                `json:"repetitions"`
	Primary       RepositoryMetadata `json:"primary_repository"`
	Baseline      RefTrace           `json:"baseline"`
	Candidate     RefTrace           `json:"candidate"`
	Events        []EventDelta       `json:"event_deltas"`
	Closures      []ParentClosure    `json:"parent_closures"`
	RescaleShares *RescaleShares     `json:"rescale_shares,omitempty"`
}

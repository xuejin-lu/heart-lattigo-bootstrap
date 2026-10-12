package perfmeasure

import "time"

const BTSNoTrace008SchemaVersion = "bts-notrace-warm-and-numerical-008-run.v1"

// BTSRunDocument is the compact manifest emitted by the Batch008 measurement
// adapter after each completed sample. Vector payloads stay in f64le files.
type BTSRunDocument struct {
	SchemaVersion            string                    `json:"schema_version"`
	Task                     string                    `json:"task"`
	Status                   string                    `json:"status"`
	Backend                  string                    `json:"backend"`
	BackendCommit            string                    `json:"backend_commit"`
	BinarySHA256             string                    `json:"binary_sha256"`
	BTSManifestSHA256        string                    `json:"bts_source_manifest_sha256"`
	MeasurementAdapterSHA256 string                    `json:"measurement_adapter_sha256"`
	PrimaryCommit            string                    `json:"primary_commit"`
	OriginalSourceRoot       string                    `json:"original_source_root"`
	StartedAt                time.Time                 `json:"started_at"`
	UpdatedAt                time.Time                 `json:"updated_at"`
	MeasurementProcesses     int                       `json:"measurement_processes"`
	BootstrapBudget          int                       `json:"bootstrap_budget"`
	BootstrapReservations    int                       `json:"bootstrap_reservations"`
	NewBTSNS                 int64                     `json:"new_bts_elapsed_ns"`
	Environment              BTSMeasurementEnvironment `json:"environment"`
	Parameters               BTSMeasurementParameters  `json:"effective_parameters"`
	Input                    BTSVectorArtifact         `json:"input_vector"`
	Samples                  []BTSBootstrapSample      `json:"bootstrap_samples"`
	Failure                  string                    `json:"failure,omitempty"`
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

package perfmeasure

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BootstrapAttemptBudget provides a shared fail-closed attempt counter and
// durable per-attempt reservation journal for measurement adapters.
type BootstrapAttemptBudget struct {
	Limit       int
	Attempts    int
	JournalBase string
	Metadata    map[string]string
}

type bootstrapAttemptReservation struct {
	SchemaVersion string            `json:"schema_version"`
	Index         int               `json:"attempt_index"`
	Budget        int               `json:"bootstrap_budget"`
	Phase         string            `json:"phase"`
	ReservedAt    time.Time         `json:"reserved_at"`
	Status        string            `json:"status"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type bootstrapAttemptResult struct {
	SchemaVersion string    `json:"schema_version"`
	Index         int       `json:"attempt_index"`
	Phase         string    `json:"phase"`
	CompletedAt   time.Time `json:"completed_at"`
	ElapsedNS     int64     `json:"elapsed_ns"`
	Status        string    `json:"status"`
	Error         string    `json:"error,omitempty"`
}

// Invoke reserves the attempt before calling call. A callback error consumes
// the reservation. A process interruption leaves the exclusive reservation
// file without a result file, so a resumed process cannot silently reuse it.
func (budget *BootstrapAttemptBudget) Invoke(phase string, call func() error) (int64, error) {
	if budget == nil {
		return 0, errors.New("Bootstrap invocation has no attempt budget")
	}
	if budget.Limit < 0 {
		return 0, errors.New("Bootstrap attempt budget must be non-negative")
	}
	if budget.Attempts >= budget.Limit {
		return 0, fmt.Errorf("Bootstrap attempt budget exhausted at %d/%d", budget.Attempts, budget.Limit)
	}
	if call == nil {
		return 0, errors.New("Bootstrap attempt callback is nil")
	}
	if phase == "" {
		phase = "bootstrap"
	}

	index := budget.Attempts + 1
	if budget.JournalBase != "" {
		reservation := bootstrapAttemptReservation{
			SchemaVersion: "perfprobe-bootstrap-attempt.v1",
			Index:         index,
			Budget:        budget.Limit,
			Phase:         phase,
			ReservedAt:    time.Now().UTC(),
			Status:        "irrevocably_reserved_before_call",
			Metadata:      cloneStringMap(budget.Metadata),
		}
		path := fmt.Sprintf("%s.bootstrap-attempt-%02d.json", budget.JournalBase, index)
		if err := writeExclusiveSyncedJSON(path, reservation); err != nil {
			return 0, fmt.Errorf("reserve Bootstrap attempt %d before invocation: %w", index, err)
		}
	}

	budget.Attempts = index
	started := time.Now()
	callErr := call()
	elapsed := time.Since(started).Nanoseconds()
	if budget.JournalBase != "" {
		result := bootstrapAttemptResult{
			SchemaVersion: "perfmeasure-bootstrap-attempt-result.v1",
			Index:         index,
			Phase:         phase,
			CompletedAt:   time.Now().UTC(),
			ElapsedNS:     elapsed,
			Status:        "RETURNED",
		}
		if callErr != nil {
			result.Status = "ERROR_RETURNED"
			result.Error = callErr.Error()
		}
		path := fmt.Sprintf("%s.bootstrap-result-%02d.json", budget.JournalBase, index)
		if err := writeExclusiveSyncedJSON(path, result); err != nil {
			return elapsed, errors.Join(callErr, fmt.Errorf("record Bootstrap attempt %d result: %w", index, err))
		}
	}
	return elapsed, callErr
}

func writeExclusiveSyncedJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	clone := make(map[string]string, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

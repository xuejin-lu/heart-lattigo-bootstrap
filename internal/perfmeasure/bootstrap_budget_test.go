package perfmeasure

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestBootstrapAttemptBudgetReservesSixCallsBeforeCallbacks(t *testing.T) {
	base := filepath.Join(t.TempDir(), "attempt")
	budget := &BootstrapAttemptBudget{
		Limit:       6,
		JournalBase: base,
		Metadata:    map[string]string{"backend_commit": "pinned", "binary_sha256": "digest"},
	}
	phases := []string{"first_cold_bootstrap", "warm_bootstrap_01", "warm_bootstrap_02", "warm_bootstrap_03", "warm_bootstrap_04", "warm_bootstrap_05"}
	calls := 0
	for index, phase := range phases {
		gotPhase := phase
		elapsed, err := budget.Invoke(gotPhase, func() error {
			calls++
			reservationPath := fmt.Sprintf("%s.bootstrap-attempt-%02d.json", base, index+1)
			data, err := os.ReadFile(reservationPath)
			if err != nil {
				return fmt.Errorf("callback began without durable reservation: %w", err)
			}
			var reservation bootstrapAttemptReservation
			if err := json.Unmarshal(data, &reservation); err != nil {
				return err
			}
			if reservation.Index != index+1 || reservation.Budget != 6 || reservation.Phase != gotPhase || reservation.Status != "irrevocably_reserved_before_call" {
				return fmt.Errorf("unexpected reservation before callback: %+v", reservation)
			}
			if reservation.Metadata["backend_commit"] != "pinned" {
				return fmt.Errorf("reservation lost provenance metadata: %+v", reservation.Metadata)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("attempt %d: %v", index+1, err)
		}
		if elapsed < 0 {
			t.Fatalf("attempt %d returned negative elapsed time %d", index+1, elapsed)
		}
		resultPath := fmt.Sprintf("%s.bootstrap-result-%02d.json", base, index+1)
		if _, err := os.Stat(resultPath); err != nil {
			t.Fatalf("attempt %d result sidecar missing: %v", index+1, err)
		}
	}
	if calls != 6 || budget.Attempts != 6 {
		t.Fatalf("calls=%d attempts=%d, want 6 each", calls, budget.Attempts)
	}
	if _, err := budget.Invoke("over-budget", func() error { calls++; return nil }); err == nil {
		t.Fatal("budget permitted a seventh callback")
	}
	if calls != 6 || budget.Attempts != 6 {
		t.Fatalf("over-budget attempt changed state: calls=%d attempts=%d", calls, budget.Attempts)
	}
}

func TestBootstrapAttemptBudgetConsumesCallbackErrorAndRefusesJournalCollision(t *testing.T) {
	base := filepath.Join(t.TempDir(), "attempt")
	budget := &BootstrapAttemptBudget{Limit: 2, JournalBase: base}
	calls := 0
	wantErr := errors.New("simulated failure")
	if _, err := budget.Invoke("cold", func() error { calls++; return wantErr }); !errors.Is(err, wantErr) {
		t.Fatalf("callback error changed: got %v want %v", err, wantErr)
	}
	if budget.Attempts != 1 || calls != 1 {
		t.Fatalf("failed callback not consumed exactly once: attempts=%d calls=%d", budget.Attempts, calls)
	}
	result, err := os.ReadFile(base + ".bootstrap-result-01.json")
	if err != nil {
		t.Fatal(err)
	}
	var completed bootstrapAttemptResult
	if err := json.Unmarshal(result, &completed); err != nil {
		t.Fatal(err)
	}
	if completed.Status != "ERROR_RETURNED" || completed.Error != wantErr.Error() {
		t.Fatalf("failed callback result not recorded: %+v", completed)
	}

	restarted := &BootstrapAttemptBudget{Limit: 2, JournalBase: base}
	if _, err := restarted.Invoke("must-not-run", func() error { calls++; return nil }); err == nil {
		t.Fatal("existing reservation did not fail closed")
	}
	if calls != 1 || restarted.Attempts != 0 {
		t.Fatalf("collision invoked callback or consumed local attempt: calls=%d attempts=%d", calls, restarted.Attempts)
	}
}

func TestBootstrapAttemptBudgetZeroDoesNotWriteOrInvoke(t *testing.T) {
	base := filepath.Join(t.TempDir(), "attempt")
	budget := &BootstrapAttemptBudget{Limit: 0, JournalBase: base}
	calls := 0
	if _, err := budget.Invoke("not-authorized", func() error { calls++; return nil }); err == nil {
		t.Fatal("zero budget accepted callback")
	}
	if calls != 0 || budget.Attempts != 0 {
		t.Fatalf("zero budget changed state: calls=%d attempts=%d", calls, budget.Attempts)
	}
	if _, err := os.Stat(base + ".bootstrap-attempt-01.json"); !os.IsNotExist(err) {
		t.Fatalf("zero budget wrote reservation: stat err=%v", err)
	}
}

package perfmeasure

import "testing"

func TestSystemProfilerChipExtractsModelOnly(t *testing.T) {
	output := "Hardware Overview:\n    Model Identifier: Mac16,12\n    Chip: Apple M4\n    Serial Number: not part of the result\n"
	if got := systemProfilerChip(output); got != "Apple M4" {
		t.Fatalf("systemProfilerChip()=%q, want Apple M4", got)
	}
	if got := systemProfilerChip("Hardware Overview:\n    Model Identifier: Mac16,12\n"); got != "" {
		t.Fatalf("systemProfilerChip()=%q, want empty when no Chip line exists", got)
	}
}

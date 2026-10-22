package main

import (
	"math"
	"math/cmplx"
	"os"
	"path/filepath"
	"testing"
)

func TestFixedSigmaRoundoffOperationBoundsCoverComplex128(t *testing.T) {
	a := complex(0.0012105306038663463, -0.00017320508075688773)
	b := complex(0.03515625009549138, 0.004882812504305562)
	product := a * b
	productError := exactL1(exactSub(exactFromComplex(product), exactMul(exactFromComplex(a), exactFromComplex(b))))
	if productError.Cmp(exactComplexMulError(exactFromComplex(a), exactFromComplex(b))) > 0 {
		t.Fatalf("complex product error %s exceeds gamma_2 L1 bound", rationalUp(productError))
	}

	difference := complex(0.0012105305996361336, 0.0002718281828459045)
	computedDifference := a - difference
	exactDifference := exactSub(exactFromComplex(a), exactFromComplex(difference))
	differenceError := exactL1(exactSub(exactFromComplex(computedDifference), exactDifference))
	if differenceError.Cmp(exactAddError(exactFromComplex(a), exactFromComplex(difference))) > 0 {
		t.Fatalf("complex subtraction error %s exceeds u*(L(a)+L(b)) bound", rationalUp(differenceError))
	}
	if math.IsNaN(cmplx.Abs(product)) || math.IsInf(cmplx.Abs(product), 0) {
		t.Fatal("synthetic complex arithmetic unexpectedly became non-finite")
	}
}

func TestFixedSigmaPropagationRoundoffReviewUsesOnlyArchivedCaptures(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Errorf("restore test working directory: %v", err)
		}
	})
	root := filepath.Join(repoRoot, "results", "FIXED-SIGMA-DIRECT-MUL-004")
	report, err := buildFixedSigmaPropagationRoundoffReview(
		filepath.Join(root, "roundoff-review-manifest.json"),
		filepath.Join(root, "analysis.json"),
	)
	if err != nil {
		t.Fatalf("offline review of archived captures failed: %v", err)
	}
	if report.PriorAnalysisSHA256 != "dc45035b37a50ec35d97cbe9a4d747b75ed85012a663dba7f28022e47e027243" {
		t.Fatalf("the original fail-closed analysis was not used: SHA-256 %s", report.PriorAnalysisSHA256)
	}
	if report.Status != "DERIVED_ROUNDOFF_ENVELOPE_COVERS_ALL_COMPLEX128_CLOSURES" || len(report.Pairs) != 4 {
		t.Fatalf("unexpected offline review result: status=%s pairs=%d", report.Status, len(report.Pairs))
	}
	if !report.FailureMaskAnalysis.AllOldFailureMasksIdentical ||
		report.FailureMaskAnalysis.IntersectionFailureSlots != 4059 ||
		report.FailureMaskAnalysis.UnionFailureSlots != 4059 ||
		!report.FailureMaskAnalysis.AllPairsShareSamePassSlots {
		t.Fatalf("unexpected old gamma_64 failure position pattern: %+v", report.FailureMaskAnalysis)
	}
	for _, pair := range report.Pairs {
		if pair.OldGamma64FailureCount != 4059 || len(pair.OldGamma64PassIndices) != 37 ||
			!pair.HighPrecisionExactClosureZero || pair.HighPrecisionIdentityMismatch != 0 ||
			pair.DerivedBoundViolationCount != 0 ||
			pair.CExactZeroSlots != 0 ||
			pair.OldGamma64PassMaxProductAbs >= pair.OldGamma64FailureMinProductAbs ||
			pair.BaselinePrimaryCommit == "" || pair.BaselinePrimarySourceSHA256 == "" || pair.BaselineBackendCommit == "" || pair.BaselineBinarySHA256 == "" ||
			pair.ONOFFPrimaryCommit == "" || pair.ONOFFPrimarySourceSHA256 == "" || pair.ONOFFBackendCommit == "" || pair.ONOFFBinarySHA256 == "" {
			t.Fatalf("pair %s/%v did not satisfy the independent offline review checks: %+v", pair.Workload, pair.StageSeeds, pair)
		}
		if pair.Complex128FirstOrderResidual.Statistics.RMS <= 0 ||
			pair.Complex128FirstOrderResidual.Statistics.RMS <= pair.Complex128Closure.Statistics.RMS*1e6 {
			t.Fatalf("pair %s/%v did not separate first-order model remainder from floating closure: model=%g closure=%g",
				pair.Workload, pair.StageSeeds,
				pair.Complex128FirstOrderResidual.Statistics.RMS, pair.Complex128Closure.Statistics.RMS)
		}
	}
}

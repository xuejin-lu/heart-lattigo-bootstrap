package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func TestSummaryOfDoesNotMutateInput(t *testing.T) {
	decoded := []complexValue{{Real: 1.25, Imag: -0.5}, {Real: -2, Imag: 0.75}}
	input := runEvidence{Checkpoints: []checkpoint{{ID: "cp", Decoded: decoded}}}

	summary := summaryOf(input)

	if len(input.Checkpoints[0].Decoded) != len(decoded) || input.Checkpoints[0].Decoded[0] != decoded[0] {
		t.Fatalf("summaryOf mutated source decoded values: %+v", input.Checkpoints[0].Decoded)
	}
	if len(summary.Checkpoints) != 1 || summary.Checkpoints[0].Decoded != nil {
		t.Fatalf("summary did not strip decoded values: %+v", summary.Checkpoints)
	}
	summary.Checkpoints[0].ID = "changed"
	if input.Checkpoints[0].ID != "cp" {
		t.Fatal("summary checkpoint slice aliases the input checkpoint slice")
	}
}

func TestCompareRejectsInvalidSampleShapesAndValues(t *testing.T) {
	tests := []struct {
		name string
		want []complex128
		got  []complex128
	}{
		{name: "empty", want: nil, got: nil},
		{name: "mismatched lengths", want: []complex128{1, 2}, got: []complex128{1}},
		{name: "nonfinite expected", want: []complex128{complex(math.NaN(), 0)}, got: []complex128{0}},
		{name: "nonfinite actual", want: []complex128{0}, got: []complex128{complex(0, math.Inf(1))}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := compare(test.want, test.got)
			if got.Finite {
				t.Fatalf("compare(%v, %v) unexpectedly accepted invalid input: %+v", test.want, test.got, got)
			}
		})
	}
}

func TestCompareNonemptyValues(t *testing.T) {
	identical := compare([]complex128{1 + 2i, -3 + 0.5i}, []complex128{1 + 2i, -3 + 0.5i})
	if !identical.Finite || identical.ComplexRMSE != 0 || identical.MaxComplex != 0 {
		t.Fatalf("identical nonempty comparison should be valid zero: %+v", identical)
	}

	different := compare([]complex128{1 + 2i, -3 + 0.5i}, []complex128{1.25 + 2i, -3 + 0.5i})
	if !different.Finite || different.ComplexRMSE <= 0 || different.MaxComplex <= 0 {
		t.Fatalf("distinct nonempty comparison should be finite and positive: %+v", different)
	}
}

func TestCombinePreservesAndComparesRawDecodedSamples(t *testing.T) {
	standard := testRunEvidence(t, false)
	fast := testRunEvidence(t, true)
	standardPath, fastPath, outPath := writeRunPair(t, standard, fast)

	if err := combine(standardPath, fastPath, outPath); err != nil {
		t.Fatalf("combine valid matched runs: %v", err)
	}
	var result combinedEvidence
	if err := readJSON(outPath, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Paired) != 8 {
		t.Fatalf("paired checkpoint count = %d, want 8", len(result.Paired))
	}
	for _, pair := range result.Paired {
		if pair.FastVsStandardRMSE <= 0 || pair.FastVsStandardMax <= 0 {
			t.Errorf("%s lost the known nonzero decoded difference: RMSE=%g max=%g", pair.ID, pair.FastVsStandardRMSE, pair.FastVsStandardMax)
		}
	}
	for _, cp := range result.Standard.Checkpoints {
		if cp.Decoded != nil {
			t.Errorf("compact Standard checkpoint %s retained raw decoded samples", cp.ID)
		}
	}
}

func TestCombineRejectsMalformedEvidence(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*runEvidence, *runEvidence)
	}{
		{
			name: "empty decoded samples",
			mutate: func(standard, _ *runEvidence) {
				standard.Checkpoints[0].Decoded = nil
			},
		},
		{
			name: "mismatched decoded lengths",
			mutate: func(standard, _ *runEvidence) {
				standard.Checkpoints[0].Decoded = standard.Checkpoints[0].Decoded[:15]
			},
		},
		{
			name: "wrong checkpoint count",
			mutate: func(standard, _ *runEvidence) {
				standard.Checkpoints = standard.Checkpoints[:7]
			},
		},
		{
			name: "wrong checkpoint identity",
			mutate: func(standard, _ *runEvidence) {
				standard.Checkpoints[2].ID = "unexpected-checkpoint"
			},
		},
		{
			name: "wrong Fast backend commit",
			mutate: func(_, fast *runEvidence) {
				fast.BackendCommit = "not-the-pinned-fast-commit"
			},
		},
		{
			name: "mismatched deterministic input provenance",
			mutate: func(_, fast *runEvidence) {
				fast.InputSHA256 = "different-input"
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			standard := testRunEvidence(t, false)
			fast := testRunEvidence(t, true)
			test.mutate(&standard, &fast)
			standardPath, fastPath, outPath := writeRunPair(t, standard, fast)
			if err := combine(standardPath, fastPath, outPath); err == nil {
				t.Fatal("combine accepted malformed or mismatched evidence")
			}
		})
	}
}

func testRunEvidence(t *testing.T, fast bool) runEvidence {
	t.Helper()
	used := profile{LogN: logN, LogQ: []int{55, 39, 40, 39}, LogP: []int{60}, ScaleLog2: defaultScaleLog, LogSlots: logSlots, Levels: append([]int(nil), levels...)}
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: used.LogN, LogQ: used.LogQ, LogP: used.LogP, LogDefaultScale: used.ScaleLog2,
		Xs: ring.Ternary{H: 192},
	})
	if err != nil {
		t.Fatal(err)
	}
	profileJSON, _ := json.Marshal(used)
	qpJSON, _ := json.Marshal(struct{ Q, P []uint64 }{params.Q(), params.P()})
	a, b := deterministicInputs(1 << logSlots)
	backendCommit, backendRef := standardCommit, "pinned-standard-"+standardCommit
	if fast {
		backendCommit, backendRef = "2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac", "fast-qprefix"
	}
	result := runEvidence{
		SchemaVersion: "fast-dropin-compact-consumers-batch-015-run.v1", Status: "public_compact_add_mul_relin_rescale_rotate_passed",
		CreatedUTC: "2026-10-10T00:00:00Z", GoVersion: "go1.26.4", OS: "darwin", Architecture: "arm64",
		PrimaryCommit: "2794bcc55d67c158a81fe704a173c4a03d5bed73", PrimaryDirty: true,
		BackendCommit: backendCommit, BackendRef: backendRef, FastCapability: fast,
		Profile: used, ProfileSHA256: hashBytes(profileJSON), QPSHA256: hashBytes(qpJSON), Q: params.Q(), P: params.P(),
		InputSHA256: hashInputs(a, b), Lifecycle: "same CKKS public source and profile; NewKeyGenerator/GenSecretKeyNew; NewPlaintext/Encoder.Encode; rlwe.NewEncryptor.EncryptNew; public ckks.NewEvaluator AddNew/MulRelinNew/Rescale/RotateNew; ordinary DecryptNew/Decode",
		BootstrapCalls: 0,
	}
	if fast {
		result.KeyLookups.GaloisList = 1
	}
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	result.SourceSHA256 = hashBytes(source)

	descriptors := []struct {
		id, api string
		level   int
	}{
		{"add-l1", "AddNew", 1}, {"mul-relin-l1", "MulRelinNew", 1}, {"rescale-l1", "Rescale", 0}, {"rotate-l0", "RotateNew", 0},
		{"add-l3", "AddNew", 3}, {"mul-relin-l3", "MulRelinNew", 3}, {"rescale-l3", "Rescale", 2}, {"rotate-l2", "RotateNew", 2},
	}
	baseScale := rlwe.NewScale(math.Exp2(defaultScaleLog))
	productScale := baseScale.Mul(baseScale)
	for _, descriptor := range descriptors {
		scale := baseScale
		if descriptor.api == "MulRelinNew" {
			scale = productScale
		} else if descriptor.api == "Rescale" || descriptor.api == "RotateNew" {
			startLevel := descriptor.level + 1
			scale = productScale.Div(rlwe.NewScale(params.Q()[startLevel]))
		}
		rows := descriptor.level + 1
		state := outputState{Level: descriptor.level, Degree: 1, ScaleLog2: scale.Log2(), IsNTT: true, C1Zero: fast, RowsPerComponent: []int{rows, rows}}
		decoded := make([]complexValue, 1<<logSlots)
		for i := range decoded {
			decoded[i] = complexValue{Real: float64(i) / 64, Imag: -float64(i%3) / 128}
			if fast {
				decoded[i].Real += 0.25
			}
		}
		result.Checkpoints = append(result.Checkpoints, checkpoint{
			ID: descriptor.id, API: descriptor.api, State: state,
			Plaintext: metric{Finite: true, ComplexRMSE: 1e-12, MaxComplex: 1e-12}, Decoded: decoded, Status: "passed",
		})
	}
	return result
}

func writeRunPair(t *testing.T, standard, fast runEvidence) (standardPath, fastPath, outPath string) {
	t.Helper()
	dir := t.TempDir()
	standardPath = filepath.Join(dir, "standard.json")
	fastPath = filepath.Join(dir, "fast.json")
	outPath = filepath.Join(dir, "combined.json")
	if err := writeJSON(standardPath, standard); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(fastPath, fast); err != nil {
		t.Fatal(err)
	}
	return standardPath, fastPath, outPath
}

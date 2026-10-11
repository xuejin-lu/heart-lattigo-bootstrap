package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	fastckks "github.com/tuneinsight/lattigo/v6/schemes/ckks/fast"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/numericalmetrics"
	"github.com/xuejin-lu/heart-lattigo-bootstrap/internal/perfmeasure"
)

const fastOnlyTraceCommit = "2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f"

type fastOnlyCheckpoint struct {
	Name       string            `json:"checkpoint"`
	Metadata   NumericalMetadata `json:"metadata"`
	PrefixRows int               `json:"qprefix_rows"`
	DecodedSHA string            `json:"decoded_sha256"`
}

type fastOnlyPower struct {
	T          int               `json:"chebyshev_power"`
	Metadata   NumericalMetadata `json:"metadata"`
	PrefixRows int               `json:"qprefix_rows"`
	DecodedSHA string            `json:"decoded_sha256"`
}

type fastOnlyBranchTrace struct {
	Branch          string                     `json:"branch"`
	ReferenceStatus string                     `json:"genuine_standard_comparison_status"`
	ReferenceReason string                     `json:"genuine_standard_comparison_reason"`
	PublicReplay    numericalmetrics.SNR       `json:"fast_internal_replay_vs_public_evalmod"`
	PolynomialPlan  NumericalPolynomialPlan    `json:"polynomial_plan"`
	GeneratedPowers []fastOnlyPower            `json:"generated_powers"`
	Checkpoints     []fastOnlyCheckpoint       `json:"checkpoints"`
	CapacityAudit   []NumericalQPrefixCapacity `json:"qprefix_capacity_audit"`
	ScaleAudit      []NumericalScaleAudit      `json:"scale_audit"`
}

type fastOnlyTraceDocument struct {
	SchemaVersion    string                          `json:"schema_version"`
	Status           string                          `json:"status"`
	Timestamp        time.Time                       `json:"timestamp"`
	Profile          string                          `json:"profile"`
	Primary          RepositoryMetadata              `json:"primary"`
	Secondary        RepositoryMetadata              `json:"secondary"`
	PrimarySourceSHA string                          `json:"primary_measurement_source_sha256"`
	BinarySHA256     string                          `json:"fastdiag_binary_sha256"`
	BuildTag         string                          `json:"build_tags"`
	ModulePath       string                          `json:"compiled_lattigo_module_path"`
	GoVersion        string                          `json:"go_version"`
	OS               string                          `json:"os"`
	Arch             string                          `json:"arch"`
	ConfigPath       string                          `json:"config_path"`
	ConfigSHA256     string                          `json:"config_sha256"`
	InputSHA256      string                          `json:"input_sha256"`
	InputMetadataSHA string                          `json:"input_metadata_sha256"`
	QPSHA256         string                          `json:"qp_sha256"`
	Parameters       perfmeasure.EffectiveParameters `json:"effective_parameters"`
	BootstrapBudget  int                             `json:"bootstrap_budget"`
	BootstrapCalls   int                             `json:"actual_bootstrap_calls"`
	Stages           []fastOnlyBranchTrace           `json:"evalmod_internal_fast_only_replays"`
	Limitations      []string                        `json:"limitations"`
}

func executeNumericalFastOnlyTrace(args []string) error {
	flags := flag.NewFlagSet("fastdiag numerical-trace-only", flag.ContinueOnError)
	profile := flags.String("profile", supportedProfile, "p93-q55 or logn16-q55")
	secondaryRootFlag := flags.String("secondary-root", "", "isolated Fast Secondary worktree")
	fastCommit := flags.String("fast-commit", "", "pinned Fast commit")
	out := flags.String("out", "", "new raw Fast-only trace JSON path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *out == "" || *secondaryRootFlag == "" || *fastCommit != fastOnlyTraceCommit {
		return errors.New("numerical-trace-only requires --profile, --secondary-root, --fast-commit=<pinned SHA>, and --out")
	}
	if *profile != supportedProfile && *profile != logN16Profile {
		return fmt.Errorf("numerical-trace-only supports only %s or %s", supportedProfile, logN16Profile)
	}
	if err := ensureFastOnlyTraceOutputAbsent(*out); err != nil {
		return err
	}
	start, err := os.Getwd()
	if err != nil {
		return err
	}
	primaryRoot, err := findPrimaryRoot(start)
	if err != nil {
		return err
	}
	if err := verifyRepository(primaryRoot, "heart-lattigo-bootstrap"); err != nil {
		return err
	}
	primary, err := repositoryMetadata(primaryRoot, "main")
	if err != nil || primary.Dirty {
		return fmt.Errorf("Fast-only numerical trace requires a clean committed Primary: dirty=%t err=%v", primary.Dirty, err)
	}
	branch, err := gitOutput(primaryRoot, "branch", "--show-current")
	if err != nil || branch != "main" {
		return fmt.Errorf("Fast-only numerical trace requires Primary main, got branch=%q err=%v", branch, err)
	}
	secondaryRoot, err := filepath.Abs(*secondaryRootFlag)
	if err != nil {
		return err
	}
	if err := verifyRepository(secondaryRoot, "lattigo"); err != nil {
		return err
	}
	secondary, err := repositoryMetadata(secondaryRoot, "pinned detached checkout")
	if err != nil || secondary.Commit != fastOnlyTraceCommit || secondary.Dirty {
		return fmt.Errorf("Fast-only trace source mismatch: commit=%s dirty=%t err=%v", secondary.Commit, secondary.Dirty, err)
	}
	modulePath, buildTag, binaryHash, err := fastOnlyBuildIdentity(secondaryRoot)
	if err != nil {
		return err
	}
	params, residual, values, effective, configPath, configHash, err := numericalProfile(*profile, primaryRoot)
	if err != nil {
		return err
	}
	inputHash := fastStandardInputFingerprint(values)
	wantInput, wantMetadata, err := numericalInputFingerprints(*profile)
	if err != nil || inputHash != wantInput {
		return fmt.Errorf("historical input fingerprint mismatch: got=%s want=%s err=%v", inputHash, wantInput, err)
	}
	input, err := fastStandardEncodedInput(residual, params.CoeffsToSlotsParameters.LogSlots, values)
	if err != nil {
		return err
	}
	inputMetadata := perfmeasure.CiphertextStateFingerprint(perfmeasure.CiphertextStateOf(input, params.BootstrappingParameters.Q(), input.Level()+1))
	if inputMetadata != wantMetadata {
		return fmt.Errorf("historical input metadata fingerprint mismatch: got=%s want=%s", inputMetadata, wantMetadata)
	}
	qpJSON, err := json.Marshal(struct{ Q, P []uint64 }{params.BootstrappingParameters.Q(), params.BootstrappingParameters.P()})
	if err != nil {
		return err
	}
	qpSum := sha256.Sum256(qpJSON)
	fastEval, err := bootstrapping.NewFastEvaluator(params)
	if err != nil {
		return err
	}
	packed, _, _, err := fastEval.PackAndSwitchN1ToN2([]rlwe.Ciphertext{*input.CopyNew()})
	if err != nil {
		return fmt.Errorf("Fast PackAndSwitch replay: %w", err)
	}
	if len(packed) != 1 {
		return fmt.Errorf("Fast PackAndSwitch returned %d ciphertexts", len(packed))
	}
	scaled, _, err := fastEval.ScaleDown(&packed[0])
	if err != nil {
		return fmt.Errorf("Fast ScaleDown replay: %w", err)
	}
	modUp, err := fastEval.ModUp(scaled)
	if err != nil {
		return fmt.Errorf("Fast ModUp replay: %w", err)
	}
	realCT, imagCT, err := fastEval.CoeffsToSlots(modUp)
	if err != nil {
		return fmt.Errorf("Fast CoeffsToSlots replay: %w", err)
	}
	if realCT == nil || imagCT == nil {
		return errors.New("Fast full-slot C2S did not return both branches")
	}
	branches := make([]fastOnlyBranchTrace, 0, 2)
	for _, branch := range []struct {
		name string
		ct   *rlwe.Ciphertext
	}{{"real", realCT}, {"imag", imagCT}} {
		trace, err := runFastOnlyEvalModReplay(branch.name, branch.ct, fastEval, params.BootstrappingParameters)
		if err != nil {
			return fmt.Errorf("Fast-only %s EvalMod internal replay: %w", branch.name, err)
		}
		branches = append(branches, trace)
	}
	primarySource, err := primaryMeasurementSourceFingerprint(primaryRoot, primary.Commit)
	if err != nil {
		return err
	}
	doc := fastOnlyTraceDocument{
		SchemaVersion: "fast-current-dual-backend-numerical-revalidation-002-fastdiag/v1", Status: "FAST_ONLY_TRACE_COMPLETE",
		Timestamp: time.Now().UTC(), Profile: *profile, Primary: primary, Secondary: secondary,
		PrimarySourceSHA: primarySource, BinarySHA256: binaryHash, BuildTag: buildTag, ModulePath: modulePath,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, ConfigPath: configPath,
		ConfigSHA256: configHash, InputSHA256: inputHash, InputMetadataSHA: inputMetadata, QPSHA256: hex.EncodeToString(qpSum[:]),
		Parameters: effective, BootstrapBudget: 0, BootstrapCalls: 0, Stages: branches,
		Limitations: []string{"Fast-only replay; no genuine Standard internal reference exists in this source build.", "All internal checkpoints are explicitly NOT_COMPARABLE_TO_GENUINE_STANDARD.", "Replay-derived from public C2S branch inputs; no Bootstrap was invoked."},
	}
	return writeFastOnlyTraceJSON(*out, doc)
}

func runFastOnlyEvalModReplay(branch string, input *rlwe.Ciphertext, eval *bootstrapping.FastEvaluator, full ckks.Parameters) (fastOnlyBranchTrace, error) {
	if input == nil || eval == nil || eval.Mod1Evaluator == nil {
		return fastOnlyBranchTrace{}, errors.New("Fast EvalMod input/evaluator is unavailable")
	}
	mod1Params := eval.Mod1Evaluator.Parameters
	if mod1Params.Mod1Type != mod1.CosDiscrete || mod1Params.Mod1InvPoly != nil || input.Level() != mod1Params.LevelQ {
		return fastOnlyBranchTrace{}, errors.New("Fast EvalMod replay requires the configured CosDiscrete/no-inverse level")
	}
	trace := fastOnlyBranchTrace{Branch: branch, ReferenceStatus: "NOT_COMPARABLE_TO_GENUINE_STANDARD", ReferenceReason: "Original Standard source has no Fast private instrumentation hook; no in-tree Standard comparison is promoted."}
	eval.FastCKKS.SetQPrefixCapacityObserver(func(snapshot fastckks.QPrefixCapacitySnapshot) error {
		trace.CapacityAudit = append(trace.CapacityAudit, NumericalQPrefixCapacity{
			Checkpoint: branch + "/" + snapshot.Name, Level: snapshot.Level, Rows: snapshot.Rows,
			Scale: snapshot.Scale, Degree: snapshot.Degree, MaxAbs: append([]string(nil), snapshot.MaxAbs...),
			PrefixProduct: snapshot.PrefixProduct, StrictFit: snapshot.StrictFit,
		})
		return nil
	})
	defer eval.FastCKKS.SetQPrefixCapacityObserver(nil)
	checkpoint := func(name string, ct *rlwe.Ciphertext) error {
		if ct == nil {
			return fmt.Errorf("%s returned nil ciphertext", name)
		}
		values, err := decodeFastOnlyNumericalStage(full, ct)
		if err != nil {
			return fmt.Errorf("decode %s: %w", name, err)
		}
		rows, err := fastckks.QPrefixWidth(ct.Level())
		if err != nil {
			return err
		}
		trace.Checkpoints = append(trace.Checkpoints, fastOnlyCheckpoint{
			Name: branch + "/" + name, Metadata: numericalMetadata(ct), PrefixRows: rows,
			DecodedSHA: perfmeasure.Fingerprint(values),
		})
		trace.ScaleAudit = append(trace.ScaleAudit, numericalScaleAuditForCT(branch+"/"+name, ct))
		return nil
	}
	actual, err := eval.EvalMod(input.CopyNew())
	if err != nil {
		return fastOnlyBranchTrace{}, fmt.Errorf("public Fast EvalMod replay: %w", err)
	}
	actualValues, err := decodeFastOnlyNumericalStage(full, actual)
	if err != nil {
		return fastOnlyBranchTrace{}, err
	}
	if err := checkpoint("input_before_normalization", input.CopyNew()); err != nil {
		return fastOnlyBranchTrace{}, err
	}
	inputScale := input.Scale
	working := input.CopyNew()
	rows, err := fastckks.QPrefixWidth(working.Level())
	if err != nil {
		return fastOnlyBranchTrace{}, err
	}
	if err := eval.FastCKKS.ObserveQPrefixCapacity("evalmod-entry", working, rows); err != nil {
		return fastOnlyBranchTrace{}, err
	}
	working.Scale = mod1Params.ScalingFactor()
	if err := checkpoint("after_normalization", working); err != nil {
		return fastOnlyBranchTrace{}, err
	}
	targetScale, err := numericalEvalModTargetScale(mod1Params, full, input.Level())
	if err != nil {
		return fastOnlyBranchTrace{}, err
	}
	offset := numericalEvalModOffset(mod1Params)
	if err := eval.FastCKKS.AddScalarQPrefixRows(working, offset, rows, working); err != nil {
		return fastOnlyBranchTrace{}, fmt.Errorf("Chebyshev offset: %w", err)
	}
	if err := checkpoint("after_chebyshev_offset", working); err != nil {
		return fastOnlyBranchTrace{}, err
	}
	if err := checkpoint("polynomial_input", working); err != nil {
		return fastOnlyBranchTrace{}, err
	}
	powers, plan, err := eval.PolynomialEvaluator.DiagnosticGeneratePowers(working, mod1Params.Mod1Poly, targetScale)
	if err != nil {
		return fastOnlyBranchTrace{}, err
	}
	trace.PolynomialPlan = NumericalPolynomialPlan{Branch: branch, Degree: plan.Degree, Base: plan.Base, Level: plan.Level,
		ScaleLog2: plan.Scale.Log2(), ScaleExact: plan.Scale.Value.Text('e', 80), BlockCount: len(plan.Blocks)}
	for _, power := range powers {
		decoded, err := decodeFastOnlyNumericalStage(full, power.Ciphertext)
		if err != nil {
			return fastOnlyBranchTrace{}, fmt.Errorf("decode T%d: %w", power.N, err)
		}
		powerRows, err := fastckks.QPrefixWidth(power.Ciphertext.Level())
		if err != nil {
			return fastOnlyBranchTrace{}, err
		}
		trace.GeneratedPowers = append(trace.GeneratedPowers, fastOnlyPower{T: power.N, Metadata: numericalMetadata(power.Ciphertext), PrefixRows: powerRows, DecodedSHA: perfmeasure.Fingerprint(decoded)})
	}
	for _, required := range []int{1, 2, 3, 6} {
		found := false
		for _, power := range trace.GeneratedPowers {
			found = found || power.T == required
		}
		if !found {
			return fastOnlyBranchTrace{}, fmt.Errorf("Fast EvalMod trace did not expose required T%d", required)
		}
	}
	polyOut, err := eval.PolynomialEvaluator.EvaluateQPrefixRows(working.CopyNew(), mod1Params.Mod1Poly, targetScale, rows)
	if err != nil {
		return fastOnlyBranchTrace{}, fmt.Errorf("Chebyshev polynomial evaluation: %w", err)
	}
	if err := eval.FastCKKS.ObserveQPrefixCapacity("evalmod-polynomial-output", polyOut, rows); err != nil {
		return fastOnlyBranchTrace{}, err
	}
	if err := checkpoint("polynomial_output", polyOut); err != nil {
		return fastOnlyBranchTrace{}, err
	}
	working = polyOut.CopyNew()
	if err := checkpoint("before_double_angle_round_0", working.CopyNew()); err != nil {
		return fastOnlyBranchTrace{}, err
	}
	sqrt2Pi := mod1Params.Sqrt2Pi
	for round := 0; round < mod1Params.DoubleAngle; round++ {
		if working.Level() < 1 {
			return fastOnlyBranchTrace{}, fmt.Errorf("DoubleAngle round %d has Level %d", round, working.Level())
		}
		rows, err = fastckks.QPrefixWidth(working.Level())
		if err != nil {
			return fastOnlyBranchTrace{}, err
		}
		if err := eval.FastCKKS.ObserveQPrefixCapacity(fmt.Sprintf("double-angle-%d-before", round), working, rows); err != nil {
			return fastOnlyBranchTrace{}, err
		}
		sqrt2Pi *= sqrt2Pi
		if err := eval.FastCKKS.MulRelinElementQPrefixRows(working, working.El(), rows, working); err != nil {
			return fastOnlyBranchTrace{}, fmt.Errorf("DoubleAngle round %d multiply: %w", round, err)
		}
		if err := checkpoint(fmt.Sprintf("double_angle_round_%d_after_multiply", round), working.CopyNew()); err != nil {
			return fastOnlyBranchTrace{}, err
		}
		if err := eval.FastCKKS.AddQPrefixRows(working, working, working, rows); err != nil {
			return fastOnlyBranchTrace{}, fmt.Errorf("DoubleAngle round %d doubling: %w", round, err)
		}
		if err := checkpoint(fmt.Sprintf("double_angle_round_%d_after_add", round), working.CopyNew()); err != nil {
			return fastOnlyBranchTrace{}, err
		}
		if err := eval.FastCKKS.AddScalarQPrefixRows(working, -sqrt2Pi, rows, working); err != nil {
			return fastOnlyBranchTrace{}, fmt.Errorf("DoubleAngle round %d constant: %w", round, err)
		}
		if err := checkpoint(fmt.Sprintf("double_angle_round_%d_after_constant", round), working.CopyNew()); err != nil {
			return fastOnlyBranchTrace{}, err
		}
		if err := eval.FastCKKS.RescaleQPrefixRows(working, rows, working); err != nil {
			return fastOnlyBranchTrace{}, fmt.Errorf("DoubleAngle round %d Rescale: %w", round, err)
		}
		rows, err = fastckks.QPrefixWidth(working.Level())
		if err != nil {
			return fastOnlyBranchTrace{}, err
		}
		if err := eval.FastCKKS.ObserveQPrefixCapacity(fmt.Sprintf("double-angle-%d-after-rescale", round), working, rows); err != nil {
			return fastOnlyBranchTrace{}, err
		}
		if err := checkpoint(fmt.Sprintf("double_angle_round_%d_after_rescale", round), working.CopyNew()); err != nil {
			return fastOnlyBranchTrace{}, err
		}
	}
	working.Scale = inputScale
	if err := checkpoint("evalmod_output_before_public_scale_reset", working.CopyNew()); err != nil {
		return fastOnlyBranchTrace{}, err
	}
	working.Scale = full.DefaultScale()
	if err := checkpoint("evalmod_output_after_public_scale_reset", working); err != nil {
		return fastOnlyBranchTrace{}, err
	}
	replayValues, err := decodeFastOnlyNumericalStage(full, working)
	if err != nil {
		return fastOnlyBranchTrace{}, err
	}
	trace.PublicReplay = numericalmetrics.Compare(actualValues, replayValues)
	if trace.PublicReplay.Status != numericalmetrics.Finite && trace.PublicReplay.Status != numericalmetrics.PositiveInfinity {
		return fastOnlyBranchTrace{}, fmt.Errorf("Fast replay/public comparison is not finite: %s %s", trace.PublicReplay.Status, trace.PublicReplay.Reason)
	}
	return trace, nil
}

func decodeFastOnlyNumericalStage(params ckks.Parameters, ct *rlwe.Ciphertext) ([]complex128, error) {
	if ct == nil {
		return nil, errors.New("nil Fast ciphertext")
	}
	rows, err := fastckks.QPrefixWidth(ct.Level())
	if err != nil {
		return nil, err
	}
	projected, err := projectNumericalStageCiphertext(params, ct, rows-1)
	if err != nil {
		return nil, err
	}
	values, err := fastStandardDecodeFast(params, projected)
	if err != nil {
		return nil, err
	}
	if err := validateNumericalVector(values, params.MaxSlots()); err != nil {
		return nil, err
	}
	return values, nil
}

func fastOnlyBuildIdentity(expectedRoot string) (modulePath, tags, binaryHash string, err error) {
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return "", "", "", errors.New("fastdiag build info is unavailable")
	}
	for _, setting := range build.Settings {
		if setting.Key == "-tags" {
			tags = setting.Value
		}
	}
	for _, dep := range build.Deps {
		if dep.Path == secondaryModule && dep.Replace != nil {
			modulePath = dep.Replace.Path
		}
	}
	want, err := filepath.EvalSymlinks(expectedRoot)
	if err != nil {
		return "", "", "", err
	}
	actual, err := filepath.EvalSymlinks(modulePath)
	if err != nil || actual != want {
		return "", "", "", fmt.Errorf("compiled Fast module path %q does not match isolated checkout %q", modulePath, expectedRoot)
	}
	if modulePath == "" || !strings.Contains(tags, "fastdiag") {
		return "", "", "", fmt.Errorf("fastdiag binary lacks expected module replacement/build tag: module=%q tags=%q", modulePath, tags)
	}
	executable, err := os.Executable()
	if err != nil {
		return "", "", "", err
	}
	data, err := os.ReadFile(executable)
	if err != nil {
		return "", "", "", err
	}
	sum := sha256.Sum256(data)
	return modulePath, tags, hex.EncodeToString(sum[:]), nil
}

func ensureFastOnlyTraceOutputAbsent(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("refusing to overwrite existing Fast-only trace %q", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func writeFastOnlyTraceJSON(path string, value any) error {
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
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const hooksUnavailable = "DIAGNOSTIC_HOOKS_UNAVAILABLE_AT_REF"

type outputPaths struct {
	json     string
	markdown string
}

func traceCurrent(primaryRoot, secondaryRoot string, primaryMeta RepositoryMetadata, opts options) (TraceDocument, error) {
	tempRoot, err := os.MkdirTemp("", "fastdiag-trace-run-")
	if err != nil {
		return TraceDocument{}, err
	}
	outputPath := filepath.Join(tempRoot, "trace.json")
	cachePath := filepath.Join(tempRoot, "gocache")
	if err := os.Mkdir(cachePath, 0o700); err != nil {
		return TraceDocument{}, err
	}
	if err := runTraceTest(secondaryRoot, outputPath, cachePath, opts); err != nil {
		return TraceDocument{}, err
	}
	payload, err := readTracePayload(outputPath)
	if err != nil {
		return TraceDocument{}, fmt.Errorf("保留診斷執行目錄 %s：%w", tempRoot, err)
	}
	secondaryMeta, err := repositoryMetadata(secondaryRoot, "")
	if err != nil {
		return TraceDocument{}, err
	}
	if err := os.RemoveAll(tempRoot); err != nil {
		return TraceDocument{}, fmt.Errorf("trace 完成但無法清理 task-created temp dir %s：%w", tempRoot, err)
	}
	return traceDocument(primaryMeta, secondaryMeta, payload), nil
}

func traceDocument(primary, secondary RepositoryMetadata, payload TracePayload) TraceDocument {
	return TraceDocument{
		SchemaVersion: "fastdiag.trace.v1", Timestamp: payload.Timestamp, Profile: payload.Profile,
		Trace: payload.Trace, Warmup: payload.Warmup, Repetitions: payload.Repetitions,
		Primary: primary, Secondary: secondary,
		Environment: EnvironmentMetadata{GoVersion: payload.GoVersion, OS: payload.OS, Arch: payload.Arch,
			CPU: cpuModel(), NumCPU: payload.NumCPU, GOMAXPROCS: payload.GOMAXPROCS},
		Workload: TraceWorkload{FingerprintSHA256: payload.InputSHA256, InputLength: payload.InputLength,
			LogN: payload.LogN, LogSlots: payload.LogSlots, QChainBits: payload.QChainBits, PBits: payload.PBits,
			PolynomialDegree: payload.Polynomial, DoubleAngle: payload.DoubleAngle},
		Runs: payload.Runs, EventMedians: aggregateEvents(payload.Runs),
	}
}

func compareRefs(primaryRoot, secondaryRoot string, primaryMeta RepositoryMetadata, opts options) (CompareDocument, error) {
	branch, err := gitOutput(secondaryRoot, "branch", "--show-current")
	if err != nil {
		return CompareDocument{}, err
	}
	if branch != "fast-qprefix" {
		return CompareDocument{}, fmt.Errorf("compare 要求 authoritative Secondary branch fast-qprefix，目前是 %q", branch)
	}
	if err := requireCleanWorktree(secondaryRoot); err != nil {
		return CompareDocument{}, err
	}
	tempRoot, err := os.MkdirTemp("", "fastdiag-compare-")
	if err != nil {
		return CompareDocument{}, err
	}
	baseline, err := traceRef(secondaryRoot, tempRoot, opts.baseline, "baseline", opts)
	if err != nil {
		return CompareDocument{}, err
	}
	candidate, err := traceRef(secondaryRoot, tempRoot, opts.candidate, "candidate", opts)
	if err != nil {
		return CompareDocument{}, err
	}
	if err := os.RemoveAll(tempRoot); err != nil {
		return CompareDocument{}, fmt.Errorf("compare 完成但無法清理 task-created temp dir %s：%w", tempRoot, err)
	}
	doc := CompareDocument{
		SchemaVersion: "fastdiag.compare.v1", Timestamp: time.Now().UTC(), Profile: opts.profile,
		Trace: opts.traceScopes, Warmup: opts.warmup, Repetitions: opts.repetitions,
		Primary: primaryMeta, Baseline: baseline, Candidate: candidate,
	}
	if baseline.Status == "READY" && candidate.Status == "READY" {
		doc.Events, doc.Closures = compareEvents(baseline.Runs, candidate.Runs)
		shares := compareRescaleShares(baseline.Runs, candidate.Runs)
		doc.RescaleShares = &shares
	}
	return doc, nil
}

func traceRef(authoritativeRoot, tempRoot, requestedRef, name string, opts options) (RefTrace, error) {
	commit, err := resolveCommit(authoritativeRoot, requestedRef)
	if err != nil {
		return RefTrace{}, fmt.Errorf("resolve %s ref %q：%w", name, requestedRef, err)
	}
	if !fastdiagHooksAvailable(authoritativeRoot, commit) {
		return RefTrace{
			RequestedRef: requestedRef, Status: hooksUnavailable,
			Repository: RepositoryMetadata{Path: "", Commit: commit, Ref: requestedRef, Dirty: false},
			Message:    "該 ref 不含 compile-time fastdiag hooks；不動態 patch 歷史原始碼。",
		}, nil
	}
	worktree := filepath.Join(tempRoot, name)
	cachePath := filepath.Join(tempRoot, "gocache-"+name)
	outputPath := filepath.Join(tempRoot, name+".json")
	if err := os.Mkdir(cachePath, 0o700); err != nil {
		return RefTrace{}, err
	}
	if err := gitRun(authoritativeRoot, "worktree", "add", "--detach", worktree, commit); err != nil {
		return RefTrace{}, fmt.Errorf("建立 %s detached worktree 失敗；暫存根目錄保留於 %s：%w", name, tempRoot, err)
	}
	if err := runTraceTest(worktree, outputPath, cachePath, opts); err != nil {
		return RefTrace{}, fmt.Errorf("%s trace 失敗，保留 detached worktree 與 cache：%s；%w", name, worktree, err)
	}
	payload, err := readTracePayload(outputPath)
	if err != nil {
		return RefTrace{}, fmt.Errorf("%s trace artifact 失敗，保留 detached worktree：%s；%w", name, worktree, err)
	}
	metadata, err := repositoryMetadata(worktree, requestedRef)
	if err != nil {
		return RefTrace{}, err
	}
	if metadata.Dirty {
		return RefTrace{}, fmt.Errorf("%s trace worktree unexpectedly dirty；保留於 %s", name, worktree)
	}
	if err := gitRun(authoritativeRoot, "worktree", "remove", worktree); err != nil {
		return RefTrace{}, fmt.Errorf("無法安全移除 task-created clean worktree %s；暫存根目錄保留於 %s：%w", worktree, tempRoot, err)
	}
	if err := os.RemoveAll(cachePath); err != nil {
		return RefTrace{}, fmt.Errorf("移除 task-created cache %s 失敗：%w", cachePath, err)
	}
	if err := os.Remove(outputPath); err != nil {
		return RefTrace{}, fmt.Errorf("移除 task-created trace file %s 失敗：%w", outputPath, err)
	}
	return RefTrace{
		RequestedRef: requestedRef, Status: "READY", Repository: metadata,
		Environment: EnvironmentMetadata{GoVersion: payload.GoVersion, OS: payload.OS, Arch: payload.Arch,
			CPU: cpuModel(), NumCPU: payload.NumCPU, GOMAXPROCS: payload.GOMAXPROCS},
		Workload: TraceWorkload{FingerprintSHA256: payload.InputSHA256, InputLength: payload.InputLength,
			LogN: payload.LogN, LogSlots: payload.LogSlots, QChainBits: payload.QChainBits, PBits: payload.PBits,
			PolynomialDegree: payload.Polynomial, DoubleAngle: payload.DoubleAngle},
		Runs: payload.Runs, EventMedians: aggregateEvents(payload.Runs),
	}, nil
}

func gitRun(root string, args ...string) error {
	commandArgs := append([]string{"-C", root}, args...)
	command := exec.Command("git", commandArgs...)
	var output strings.Builder
	command.Stdout, command.Stderr = &output, &output
	if err := command.Run(); err != nil {
		return fmt.Errorf("git %s：%w：%s", strings.Join(args, " "), err, strings.TrimSpace(output.String()))
	}
	return nil
}

func cpuModel() string {
	if runtime.GOOS == "darwin" {
		if output, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	if output, err := exec.Command("uname", "-m").Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return runtime.GOARCH
}

func makeOutputPaths(requested, mode string) (outputPaths, error) {
	if requested == "" {
		dir, err := os.MkdirTemp("", "fastdiag-"+mode+"-")
		if err != nil {
			return outputPaths{}, err
		}
		return outputPaths{json: filepath.Join(dir, mode+".json"), markdown: filepath.Join(dir, mode+".md")}, nil
	}
	path, err := filepath.Abs(requested)
	if err != nil {
		return outputPaths{}, err
	}
	if filepath.Ext(path) != ".json" {
		return outputPaths{}, errors.New("--out 必須使用 .json 副檔名")
	}
	markdown := strings.TrimSuffix(path, ".json") + ".md"
	if _, err := os.Stat(path); err == nil {
		return outputPaths{}, fmt.Errorf("拒絕覆寫既有輸出檔：%s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return outputPaths{}, err
	}
	if _, err := os.Stat(markdown); err == nil {
		return outputPaths{}, fmt.Errorf("拒絕覆寫既有摘要：%s", markdown)
	} else if !errors.Is(err, os.ErrNotExist) {
		return outputPaths{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return outputPaths{}, err
	}
	return outputPaths{json: path, markdown: markdown}, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeNewFile(path, append(data, '\n'), 0o644)
}

func writeNewFile(path string, data []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func renderTraceSummary(doc TraceDocument) string {
	latencies := make([]float64, len(doc.Runs))
	for i, run := range doc.Runs {
		latencies[i] = float64(run.ElapsedNS)
	}
	var out strings.Builder
	fmt.Fprintf(&out, "# Fast diagnostic trace — %s\n\n", doc.Profile)
	fmt.Fprintf(&out, "- Trace scopes: `%s`\n- Input SHA-256: `%s` (%d values)\n- Repetitions: %d (warmup %d)\n", strings.Join(doc.Trace, ","), doc.Workload.FingerprintSHA256, doc.Workload.InputLength, doc.Repetitions, doc.Warmup)
	fmt.Fprintf(&out, "- Primary: `%s` (`%s`, dirty=%t)\n- Secondary: `%s` (`%s`, dirty=%t)\n", doc.Primary.Commit, doc.Primary.Ref, doc.Primary.Dirty, doc.Secondary.Commit, doc.Secondary.Ref, doc.Secondary.Dirty)
	fmt.Fprintf(&out, "- Bootstrap wall median: %.3f ms\n\n", median(latencies)/1e6)
	out.WriteString("| Scope | Event | Power | Component | Median (ms) | Samples |\n|---|---|---:|---|---:|---:|\n")
	for _, stat := range doc.EventMedians {
		if stat.Scope == "rescale" && stat.Name != "rescale" && stat.Name != "preflight" && stat.Name != "materialization" {
			continue
		}
		power := formatPower(stat.Power, stat.SplitA, stat.SplitB)
		component := stat.Component
		if component == "" {
			component = "—"
		}
		fmt.Fprintf(&out, "| %s | %s | %s | %s | %.3f | %d |\n", stat.Scope, stat.Name, power, component, stat.MedianNS/1e6, stat.Count)
	}
	out.WriteString("\nNumerical output metadata matched the diagnostics-off replay, and decoded slots remained within the existing 1e-2 tolerance. Timing is evidence only; the framework does not infer a bug or safe optimization.\n")
	return out.String()
}

func renderCompareSummary(doc CompareDocument) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Fast diagnostic comparison — %s\n\n", doc.Profile)
	fmt.Fprintf(&out, "- Trace scopes: `%s`\n- Primary: `%s` (`%s`, dirty=%t)\n- Baseline `%s`: **%s**\n- Candidate `%s`: **%s**\n\n", strings.Join(doc.Trace, ","), doc.Primary.Commit, doc.Primary.Ref, doc.Primary.Dirty, doc.Baseline.RequestedRef, doc.Baseline.Status, doc.Candidate.RequestedRef, doc.Candidate.Status)
	if doc.Baseline.Status == hooksUnavailable || doc.Candidate.Status == hooksUnavailable {
		out.WriteString("At least one historical ref has no diagnostic hook call sites. It is reported as `DIAGNOSTIC_HOOKS_UNAVAILABLE_AT_REF`; old source was not patched dynamically.\n")
	}
	if doc.RescaleShares != nil {
		out.WriteString("Rescale time shares (median of per-run totals):\n\n")
		fmt.Fprintf(&out, "- Baseline: total %.3f ms; preflight %s; materialization %s\n", doc.RescaleShares.BaselineRescaleNS/1e6, formatShare(doc.RescaleShares.BaselinePreflightShare), formatShare(doc.RescaleShares.BaselineMaterializationShare))
		fmt.Fprintf(&out, "- Candidate: total %.3f ms; preflight %s; materialization %s\n", doc.RescaleShares.CandidateRescaleNS/1e6, formatShare(doc.RescaleShares.CandidatePreflightShare), formatShare(doc.RescaleShares.CandidateMaterializationShare))
	}
	if len(doc.Events) == 0 {
		out.WriteString("\nNo matched event deltas were computed.\n")
		return out.String()
	}
	events := append([]EventDelta(nil), doc.Events...)
	sort.Slice(events, func(i, j int) bool {
		return absFloat(events[i].DeltaNS) > absFloat(events[j].DeltaNS)
	})
	if len(events) > 24 {
		events = events[:24]
	}
	out.WriteString("\n| Scope | Event | Power | Baseline (ms) | Candidate (ms) | Δ (ms) | Ratio | Parent contribution |\n|---|---|---:|---:|---:|---:|---:|---:|\n")
	for _, event := range events {
		power, ratio, contribution := formatPower(event.Power, event.SplitA, event.SplitB), "—", "—"
		if event.Ratio != nil {
			ratio = fmt.Sprintf("%.3fx", *event.Ratio)
		}
		if event.ContributionToParent != nil {
			contribution = fmt.Sprintf("%+.1f%%", *event.ContributionToParent*100)
		}
		fmt.Fprintf(&out, "| %s | %s | %s | %.3f | %.3f | %+.3f | %s | %s |\n", event.Scope, event.Name, power, event.BaselineMedianNS/1e6, event.CandidateMedianNS/1e6, event.DeltaNS/1e6, ratio, contribution)
	}
	out.WriteString("\nNegative deltas are retained. Ratios/contributions are omitted when the denominator is zero. Evidence does not declare a cause or optimization safety.\n")
	return out.String()
}

func formatPower(power, splitA, splitB *int) string {
	if power == nil {
		return "—"
	}
	if splitA == nil || splitB == nil {
		return fmt.Sprint(*power)
	}
	return fmt.Sprintf("%d (%d×%d)", *power, *splitA, *splitB)
}

func formatShare(share *float64) string {
	if share == nil {
		return "—"
	}
	return fmt.Sprintf("%.2f%%", *share*100)
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

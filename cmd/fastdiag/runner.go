package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const secondaryModule = "github.com/tuneinsight/lattigo/v6"

func execute(args []string) error {
	if len(args) > 0 && args[0] == "offline-e32" {
		return executePublicE32Offline(args[1:])
	}
	opts, err := parseArgs(args)
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	primaryRoot, err := findPrimaryRoot(cwd)
	if err != nil {
		return err
	}
	secondaryRoot, err := resolveSecondaryRoot(primaryRoot)
	if err != nil {
		return err
	}
	if err := verifyRepository(primaryRoot, "heart-lattigo-bootstrap"); err != nil {
		return fmt.Errorf("Primary repo 驗證失敗：%w", err)
	}
	if err := verifyRepository(secondaryRoot, "lattigo"); err != nil {
		return fmt.Errorf("Secondary repo 驗證失敗：%w", err)
	}
	paths, err := makeOutputPaths(opts.output, opts.mode)
	if err != nil {
		return err
	}
	primaryMeta, err := repositoryMetadata(primaryRoot, "")
	if err != nil {
		return err
	}
	if opts.mode == "numerical" {
		doc, err := numericalReference(secondaryRoot, primaryMeta, opts)
		if err != nil {
			return err
		}
		if field := firstNonFiniteNumber(doc); field != "" {
			return fmt.Errorf("numerical result contains non-finite JSON number at %s", field)
		}
		if err := writeJSON(paths.json, doc); err != nil {
			return err
		}
		if err := writeNewFile(paths.markdown, []byte(renderNumericalSummary(doc)), 0o644); err != nil {
			return err
		}
		fmt.Printf("Numerical JSON：%s\nNumerical 摘要：%s\n", paths.json, paths.markdown)
		return nil
	}
	if opts.mode == "trace" {
		if opts.profile == publicE32Profile {
			doc, err := tracePublicE32(primaryRoot, secondaryRoot, primaryMeta, opts, paths)
			if err != nil {
				return err
			}
			if err := writeJSON(paths.json, doc); err != nil {
				return err
			}
			if err := writeNewFile(paths.markdown, []byte(renderPublicE32Summary(doc)), 0o644); err != nil {
				return err
			}
			fmt.Printf("Public E32 Trace JSON：%s\n摘要：%s\n", paths.json, paths.markdown)
			return nil
		}
		doc, err := traceCurrent(primaryRoot, secondaryRoot, primaryMeta, opts)
		if err != nil {
			return err
		}
		if err := writeJSON(paths.json, doc); err != nil {
			return err
		}
		if err := writeNewFile(paths.markdown, []byte(renderTraceSummary(doc)), 0o644); err != nil {
			return err
		}
		fmt.Printf("Trace JSON：%s\nTrace 摘要：%s\n", paths.json, paths.markdown)
		return nil
	}

	doc, err := compareRefs(primaryRoot, secondaryRoot, primaryMeta, opts)
	if err != nil {
		return err
	}
	if err := writeJSON(paths.json, doc); err != nil {
		return err
	}
	if err := writeNewFile(paths.markdown, []byte(renderCompareSummary(doc)), 0o644); err != nil {
		return err
	}
	fmt.Printf("Compare JSON：%s\nCompare 摘要：%s\n", paths.json, paths.markdown)
	return nil
}

func findPrimaryRoot(start string) (string, error) {
	root, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		data, readErr := os.ReadFile(filepath.Join(root, "go.mod"))
		if readErr == nil && strings.Contains(string(data), "module github.com/xuejin-lu/heart-lattigo-bootstrap") {
			return root, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", errors.New("找不到 Primary heart-lattigo-bootstrap go.mod")
		}
		root = parent
	}
}

func resolveSecondaryRoot(primaryRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(primaryRoot, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == "replace" && fields[1] == secondaryModule && fields[2] == "=>" {
			path := fields[3]
			if !filepath.IsAbs(path) {
				path = filepath.Join(primaryRoot, path)
			}
			return filepath.Abs(path)
		}
	}
	return "", fmt.Errorf("Primary go.mod 未找到 %s 的 local replace", secondaryModule)
}

func verifyRepository(root, expectedName string) error {
	actual, err := gitOutput(root, "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	actual, _ = filepath.Abs(actual)
	root, _ = filepath.Abs(root)
	if actual != root {
		return fmt.Errorf("repo root 不一致：%s != %s", actual, root)
	}
	remote, _ := gitOutput(root, "remote", "get-url", "origin")
	if !strings.Contains(remote, "xuejin-lu/"+expectedName) {
		return fmt.Errorf("origin 不符合預期 repo %s", expectedName)
	}
	return nil
}

func repositoryMetadata(root, requestedRef string) (RepositoryMetadata, error) {
	commit, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		return RepositoryMetadata{}, err
	}
	ref, _ := gitOutput(root, "branch", "--show-current")
	status, err := gitOutput(root, "status", "--porcelain")
	if err != nil {
		return RepositoryMetadata{}, err
	}
	if requestedRef != "" {
		ref = requestedRef
	}
	return RepositoryMetadata{Path: root, Commit: commit, Ref: ref, Dirty: status != ""}, nil
}

func gitOutput(root string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", root}, args...)
	command := exec.Command("git", commandArgs...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("git %s：%w：%s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func fastdiagHooksAvailable(root, commit string) bool {
	for _, path := range []string{"internal/fastdiag/enabled.go", "circuits/ckks/bootstrapping/fastdiag_trace_test.go"} {
		command := exec.Command("git", "-C", root, "cat-file", "-e", commit+":"+path)
		if err := command.Run(); err != nil {
			return false
		}
	}
	return true
}

func resolveCommit(root, ref string) (string, error) {
	return gitOutput(root, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
}

func requireCleanWorktree(root string) error {
	status, err := gitOutput(root, "status", "--porcelain")
	if err != nil {
		return err
	}
	if status != "" {
		return fmt.Errorf("Secondary authoritative worktree 不乾淨，拒絕建立 compare worktree：\n%s", status)
	}
	return nil
}

func runTraceTest(secondaryRoot, outputPath, cachePath string, opts options) error {
	args := []string{"test", "-tags", "fastdiag", "./circuits/ckks/bootstrapping", "-run", "^TestFastDiagP93Q55Trace$", "-count=1"}
	command := exec.Command("go", args...)
	command.Dir = secondaryRoot
	command.Env = updateEnv(os.Environ(), map[string]string{
		"FASTDIAG_TRACE":       opts.traceCSV(),
		"FASTDIAG_WARMUP":      fmt.Sprint(opts.warmup),
		"FASTDIAG_REPETITIONS": fmt.Sprint(opts.repetitions),
		"FASTDIAG_OUTPUT":      outputPath,
		"GOCACHE":              cachePath,
		"GOMAXPROCS":           fmt.Sprint(runtime.GOMAXPROCS(0)),
	})
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	if err := command.Run(); err != nil {
		return fmt.Errorf("diagnostic Go test 失敗（保留輸出於 %s）：%w\n%s", filepath.Dir(outputPath), err, strings.TrimSpace(output.String()))
	}
	return nil
}

func updateEnv(existing []string, updates map[string]string) []string {
	values := make(map[string]string, len(existing)+len(updates))
	for _, entry := range existing {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			values[key] = value
		}
	}
	for key, value := range updates {
		values[key] = value
	}
	out := make([]string, 0, len(values))
	for key, value := range values {
		out = append(out, key+"="+value)
	}
	return out
}

func readTracePayload(path string) (TracePayload, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TracePayload{}, err
	}
	var payload TracePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return TracePayload{}, fmt.Errorf("解析 fastdiag test artifact：%w", err)
	}
	if payload.SchemaVersion != "fastdiag.trace.v1" || len(payload.Runs) == 0 {
		return TracePayload{}, errors.New("fastdiag test artifact schema 或 runs 不完整")
	}
	return payload, nil
}

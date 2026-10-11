package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const logN16E32TestAdapterPath = "circuits/ckks/bootstrapping/fastdiag_public_e32_test.go"

func applyPublicE32LogN16TestAdapter(diagnosticRoot, primaryRoot string) (string, string, error) {
	if diagnosticRoot == "" || primaryRoot == "" {
		return "", "", errors.New("LogN16 test adapter requires isolated Secondary and Primary roots")
	}
	base, err := gitOutput(diagnosticRoot, "rev-parse", "HEAD")
	if err != nil {
		return "", "", err
	}
	if base != e32LogN16ProductionFastCommit {
		return "", "", fmt.Errorf("isolated LogN16 adapter base %s differs from frozen production Fast pin %s", base, e32LogN16ProductionFastCommit)
	}
	if err := requireCleanWorktree(diagnosticRoot); err != nil {
		return "", "", err
	}
	patchEncodedPath := filepath.Join(primaryRoot, "cmd", "fastdiag", "testdata", "logn16-public-e32-test.patch.b64")
	patchEncoded, err := os.ReadFile(patchEncodedPath)
	if err != nil {
		return "", "", fmt.Errorf("read committed LogN16 test-only adapter: %w", err)
	}
	patchBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(patchEncoded)))
	if err != nil {
		return "", "", fmt.Errorf("decode committed LogN16 test-only adapter: %w", err)
	}
	patchSum := sha256.Sum256(patchBytes)
	patchSHA := hex.EncodeToString(patchSum[:])
	patchFile, err := os.CreateTemp(diagnosticRoot, ".fastdiag-logn16-adapter-*.patch")
	if err != nil {
		return "", "", fmt.Errorf("materialize isolated LogN16 adapter patch: %w", err)
	}
	patchPath := patchFile.Name()
	defer os.Remove(patchPath)
	if _, err := patchFile.Write(patchBytes); err != nil {
		_ = patchFile.Close()
		return "", "", fmt.Errorf("write isolated LogN16 adapter patch: %w", err)
	}
	if err := patchFile.Close(); err != nil {
		return "", "", fmt.Errorf("close isolated LogN16 adapter patch: %w", err)
	}
	if !strings.Contains(string(patchBytes), "diff --git a/"+logN16E32TestAdapterPath+" b/"+logN16E32TestAdapterPath) {
		return "", "", errors.New("LogN16 adapter patch does not target the single approved test file")
	}
	if err := gitRun(diagnosticRoot, "apply", "--check", patchPath); err != nil {
		return "", "", fmt.Errorf("check isolated test-only adapter patch: %w", err)
	}
	if err := gitRun(diagnosticRoot, "apply", patchPath); err != nil {
		return "", "", fmt.Errorf("apply isolated test-only adapter patch: %w", err)
	}
	if err := gitRun(diagnosticRoot, "diff", "--check"); err != nil {
		return "", "", fmt.Errorf("test-only adapter diff check: %w", err)
	}
	paths, err := gitOutput(diagnosticRoot, "diff", "--name-only", "HEAD")
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(paths) != logN16E32TestAdapterPath {
		return "", "", fmt.Errorf("isolated LogN16 adapter changed unexpected paths: %q", paths)
	}
	if err := gitRun(diagnosticRoot, "-c", "user.name=Codex Diagnostic Adapter", "-c", "user.email=codex-diagnostic-adapter@invalid", "commit", "-m", "test: adapt isolated E32 trace for LogN16"); err != nil {
		return "", "", fmt.Errorf("commit isolated test-only adapter: %w", err)
	}
	commit, err := gitOutput(diagnosticRoot, "rev-parse", "HEAD")
	if err != nil {
		return "", "", err
	}
	parent, err := gitOutput(diagnosticRoot, "rev-parse", "HEAD^")
	if err != nil || parent != e32LogN16ProductionFastCommit {
		return "", "", fmt.Errorf("isolated adapter parent %q differs from production pin %s", parent, e32LogN16ProductionFastCommit)
	}
	committedPaths, err := gitOutput(diagnosticRoot, "diff-tree", "--no-commit-id", "--name-only", "-r", commit)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(committedPaths) != logN16E32TestAdapterPath {
		return "", "", fmt.Errorf("isolated adapter commit changed paths outside the single test file: %q", committedPaths)
	}
	if _, _, paths, err := verifyCurrentFastTestOnlyClone(diagnosticRoot, commit); err != nil {
		return "", "", err
	} else if len(paths) != 1 || paths[0] != logN16E32TestAdapterPath {
		return "", "", fmt.Errorf("isolated adapter identity audit returned unexpected paths: %v", paths)
	}
	return commit, patchSHA, nil
}

func verifyCurrentFastTestOnlyClone(root, diagnosticHead string) (productionHash, diagnosticHash string, changedPaths []string, err error) {
	productionFiles, err := gitOutput(root, "ls-tree", "-r", "--name-only", e32LogN16ProductionFastCommit)
	if err != nil {
		return "", "", nil, fmt.Errorf("list current Fast production sources: %w", err)
	}
	diagnosticFiles, err := gitOutput(root, "ls-tree", "-r", "--name-only", diagnosticHead)
	if err != nil {
		return "", "", nil, fmt.Errorf("list isolated diagnostic files: %w", err)
	}
	productionSources, err := productionGoSourceBlobs(root, e32LogN16ProductionFastCommit, productionFiles)
	if err != nil {
		return "", "", nil, err
	}
	diagnosticSources, err := productionGoSourceBlobs(root, diagnosticHead, diagnosticFiles)
	if err != nil {
		return "", "", nil, err
	}
	productionHash, diagnosticHash = hashProductionGoSources(productionSources), hashProductionGoSources(diagnosticSources)
	if productionHash != diagnosticHash {
		return productionHash, diagnosticHash, nil, errors.New("isolated LogN16 adapter changed non-test Go source")
	}
	paths, err := gitOutput(root, "diff", "--name-only", e32LogN16ProductionFastCommit, diagnosticHead)
	if err != nil {
		return productionHash, diagnosticHash, nil, err
	}
	for _, path := range strings.Split(paths, "\n") {
		if path != "" {
			changedPaths = append(changedPaths, path)
		}
	}
	if len(changedPaths) != 1 || changedPaths[0] != logN16E32TestAdapterPath {
		return productionHash, diagnosticHash, changedPaths, fmt.Errorf("isolated LogN16 adapter diff is not limited to %s: %v", logN16E32TestAdapterPath, changedPaths)
	}
	for _, path := range []string{"go.mod", "go.sum"} {
		baseBlob, baseErr := gitOutput(root, "rev-parse", e32LogN16ProductionFastCommit+":"+path)
		headBlob, headErr := gitOutput(root, "rev-parse", diagnosticHead+":"+path)
		if baseErr != nil || headErr != nil || baseBlob != headBlob {
			return productionHash, diagnosticHash, changedPaths, fmt.Errorf("isolated adapter changed module manifest %s", path)
		}
	}
	return productionHash, diagnosticHash, changedPaths, nil
}

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirtyAuthoritativeCheckoutIsRefused(t *testing.T) {
	root := initTestGitRepo(t)
	require.NoError(t, requireCleanWorktree(root))
	require.NoError(t, os.WriteFile(filepath.Join(root, "dirty.txt"), []byte("user change"), 0o600))
	err := requireCleanWorktree(root)
	require.ErrorContains(t, err, "不乾淨")
	require.ErrorContains(t, err, "dirty.txt")
}

func TestTaskCreatedDetachedWorktreeCleanup(t *testing.T) {
	root := initTestGitRepo(t)
	tempRoot := t.TempDir()
	worktree := filepath.Join(tempRoot, "baseline")
	require.NoError(t, gitRun(root, "worktree", "add", "--detach", worktree, "HEAD"))
	require.NoError(t, requireCleanWorktree(worktree))
	require.NoError(t, gitRun(root, "worktree", "remove", worktree))
	_, err := os.Stat(worktree)
	require.ErrorIs(t, err, os.ErrNotExist)
	listing, err := gitOutput(root, "worktree", "list", "--porcelain")
	require.NoError(t, err)
	require.NotContains(t, listing, worktree)
}

func TestHistoricalHookAvailabilityBoundary(t *testing.T) {
	root := initTestGitRepo(t)
	oldCommit, err := gitOutput(root, "rev-parse", "HEAD")
	require.NoError(t, err)
	require.False(t, fastdiagHooksAvailable(root, oldCommit))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "internal", "fastdiag"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "circuits", "ckks", "bootstrapping"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "internal", "fastdiag", "enabled.go"), []byte("package fastdiag\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "circuits", "ckks", "bootstrapping", "fastdiag_trace_test.go"), []byte("package bootstrapping\n"), 0o600))
	require.NoError(t, gitRun(root, "add", "."))
	require.NoError(t, gitRun(root, "commit", "-m", "add diagnostic hooks"))
	newCommit, err := gitOutput(root, "rev-parse", "HEAD")
	require.NoError(t, err)
	require.True(t, fastdiagHooksAvailable(root, newCommit))
}

func initTestGitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		command := exec.Command("git", append([]string{"-C", root}, args...)...)
		output, err := command.CombinedOutput()
		require.NoError(t, err, "%s: %s", args, output)
	}
	run("init", "-q", "-b", "fast-qprefix")
	run("config", "user.name", "Fastdiag Fixture")
	run("config", "user.email", "fastdiag-fixture@example.invalid")
	require.NoError(t, os.WriteFile(filepath.Join(root, "README"), []byte("fixture\n"), 0o600))
	run("add", "README")
	run("commit", "-q", "-m", "fixture")
	return root
}

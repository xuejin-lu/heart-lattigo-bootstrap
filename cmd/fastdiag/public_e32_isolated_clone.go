package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// createPublicE32IsolatedClone creates a detached task-local Git checkout.
// Its Git directory must be a real directory owned by that checkout; the
// authoritative Secondary repository and its worktree metadata stay untouched.
func createPublicE32IsolatedClone(sourceRoot, targetRoot, pin string) error {
	if !filepath.IsAbs(sourceRoot) || !filepath.IsAbs(targetRoot) || len(pin) != 40 {
		return errors.New("absolute paths and exact forty-character pinned commit required")
	}
	if _, err := os.Lstat(targetRoot); err == nil {
		return fmt.Errorf("destination exists: %s", targetRoot)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := gitRun(sourceRoot, "clone", "--shared", "--no-checkout", "--", sourceRoot, targetRoot); err != nil {
		return fmt.Errorf("create independent local clone: %w", err)
	}
	if err := gitRun(targetRoot, "checkout", "--detach", pin); err != nil {
		return fmt.Errorf("detach pinned diagnostic clone: %w", err)
	}
	actual, err := gitOutput(targetRoot, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if actual != pin {
		return fmt.Errorf("diagnostic clone HEAD %s differs from pinned %s", actual, pin)
	}
	sourceGit, err := gitOutput(sourceRoot, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return err
	}
	cloneGit, err := gitOutput(targetRoot, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return err
	}

	// Git can resolve macOS temporary-folder aliases (/var -> /private/var),
	// so comparing these *path strings* was incorrectly rejecting a valid
	// independent clone. Verify the actual directories instead.
	expectedGit := filepath.Join(targetRoot, ".git")
	entry, err := os.Lstat(expectedGit)
	if err != nil {
		return fmt.Errorf("inspect task-local diagnostic .git: %w", err)
	}
	if !entry.IsDir() || entry.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("diagnostic .git is not an independent directory: %s", expectedGit)
	}
	expectedInfo, err := os.Stat(expectedGit)
	if err != nil {
		return err
	}
	cloneInfo, err := os.Stat(cloneGit)
	if err != nil {
		return fmt.Errorf("inspect Git-reported diagnostic directory %q: %w", cloneGit, err)
	}
	sourceInfo, err := os.Stat(sourceGit)
	if err != nil {
		return fmt.Errorf("inspect source Git directory %q: %w", sourceGit, err)
	}
	if !os.SameFile(expectedInfo, cloneInfo) || os.SameFile(sourceInfo, cloneInfo) {
		return fmt.Errorf("diagnostic Git directory is not independent (reported=%q, task-local=%q, source=%q)", cloneGit, expectedGit, sourceGit)
	}
	return nil
}

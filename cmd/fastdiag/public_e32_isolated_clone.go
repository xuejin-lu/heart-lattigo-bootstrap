package main

import (
 "errors"
 "fmt"
 "os"
 "path/filepath"
)

// createPublicE32IsolatedClone only writes Git metadata within a task-owned
// temporary checkout. It does not modify the source repository's worktrees.
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
 if err != nil { return err }
 if actual != pin { return fmt.Errorf("diagnostic clone HEAD %s differs from pinned %s",actual,pin) }
 sourceGit,err:=gitOutput(sourceRoot,"rev-parse","--absolute-git-dir")
 if err!=nil{return err}
 cloneGit,err:=gitOutput(targetRoot,"rev-parse","--absolute-git-dir")
 if err!=nil{return err}
 if sourceGit==cloneGit || filepath.Clean(cloneGit)!=filepath.Join(filepath.Clean(targetRoot),".git"){
  return errors.New("diagnostic clone lacks independent task-local Git metadata")
 }
 return nil
}

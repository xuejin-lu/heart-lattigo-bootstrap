package main

import (
 "os"
 "os/exec"
 "path/filepath"
 "strings"
 "testing"
)
func runLocalGitTest(t *testing.T,root string,args ...string) string {
 t.Helper()
 cmd:=exec.Command("git",append([]string{"-C",root},args...)...)
 buf,err:=cmd.CombinedOutput()
 if err!=nil{t.Fatalf("git %v failed: %v: %s",args,err,buf)}
 return strings.TrimSpace(string(buf))
}
func TestPublicE32IsolatedClonePreservesSourceWorktreeMetadata(t *testing.T) {
 src:=filepath.Join(t.TempDir(),"src")
 if err:=os.Mkdir(src,0700);err!=nil{t.Fatal(err)}
 runLocalGitTest(t,src,"init","-q")
 if err:=os.WriteFile(filepath.Join(src,"test.go"),[]byte("package example\n"),0600);err!=nil{t.Fatal(err)}
 runLocalGitTest(t,src,"add","test.go")
 runLocalGitTest(t,src,"-c","user.name=ResearchTest","-c","user.email=test@example.invalid","commit","-qm","pin")
 pin:=runLocalGitTest(t,src,"rev-parse","HEAD")
 stale:=filepath.Join(src,".git","worktrees","historical-prunable")
 if err:=os.MkdirAll(stale,0700);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(stale,"gitdir"),[]byte("/no-longer-existing/checkout/.git\n"),0600);err!=nil{t.Fatal(err)}
 before:=runLocalGitTest(t,src,"worktree","list","--porcelain")
 if !strings.Contains(before,"prunable"){t.Fatalf("missing expected stale fixture: %s",before)}
 target:=filepath.Join(t.TempDir(),"diagnostic")
 if err:=createPublicE32IsolatedClone(src,target,pin);err!=nil{t.Fatal(err)}
 if got:=runLocalGitTest(t,target,"rev-parse","HEAD");got!=pin{t.Fatalf("wrong pinned commit %s",got)}
 if branch:=runLocalGitTest(t,target,"branch","--show-current");branch!=""{t.Fatalf("not detached: %s",branch)}
 if status:=runLocalGitTest(t,target,"status","--porcelain");status!=""{t.Fatalf("dirty clone: %s",status)}
 if after:=runLocalGitTest(t,src,"worktree","list","--porcelain");after!=before{t.Fatalf("source metadata changed")}
 if err:=createPublicE32IsolatedClone(src,target,pin);err==nil{t.Fatal("destination overwrite allowed")}
 if err:=os.RemoveAll(target);err!=nil{t.Fatal(err)}
 if _,err:=os.Stat(filepath.Join(stale,"gitdir"));err!=nil{t.Fatalf("removed historical metadata: %v",err)}
}

func TestPublicE32IsolatedCloneAcceptsCanonicalizedTempAliases(t *testing.T) {
	// A symlinked parent is equivalent to macOS /var vs /private/var paths.
	// Git canonicalizes --absolute-git-dir; the helper must check directory
	// identity instead of requiring an exact lexical path match.
	root := t.TempDir()
	src := filepath.Join(root, "source")
	if err := os.Mkdir(src, 0o700); err != nil {
		t.Fatal(err)
	}
	runLocalGitTest(t, src, "init", "-q")
	runLocalGitTest(t, src, "-c", "user.name=ResearchTest", "-c", "user.email=test@example.invalid", "commit", "--allow-empty", "-qm", "pin")
	pin := runLocalGitTest(t, src, "rev-parse", "HEAD")

	alias := filepath.Join(root, "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	target := filepath.Join(alias, "diagnostic")
	if err := createPublicE32IsolatedClone(src, target, pin); err != nil {
		t.Fatal(err)
	}
	gitReported := runLocalGitTest(t, target, "rev-parse", "--absolute-git-dir")
	aliasNamed := filepath.Join(target, ".git")
	if gitReported == aliasNamed {
		t.Fatalf("test did not exercise alias normalization: %q", gitReported)
	}
	if got := runLocalGitTest(t, target, "rev-parse", "HEAD"); got != pin {
		t.Fatalf("alias clone HEAD %s differs from pin %s", got, pin)
	}
	if status := runLocalGitTest(t, target, "status", "--porcelain"); status != "" {
		t.Fatalf("alias clone dirty: %s", status)
	}
	if err := os.RemoveAll(target); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(src, ".git")); err != nil {
		t.Fatalf("source Git metadata was removed: %v", err)
	}
}

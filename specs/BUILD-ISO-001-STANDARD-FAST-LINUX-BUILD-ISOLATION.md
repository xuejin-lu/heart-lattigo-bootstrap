# BUILD-ISO-001 — Standard/Fast Portable Linux Build Isolation

## Status

Executable Codex task.

## Task class

I — Implementation

## Repository scope

Primary only:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary `xuejin-lu/lattigo` is read-only except for branch switching during local build validation.

## Purpose

Allow the same formal benchmark frontend in Primary to build two portable Linux amd64 executables:

- Standard: using local `../lattigo` checked out at branch `main`
- Fast: using local `../lattigo` checked out at branch `fast-ckks`

The current Fast build succeeds. The Standard build fails because some Primary diagnostic compilation units directly import:

`github.com/tuneinsight/lattigo/v6/schemes/ckks/fast`

which does not exist on Lattigo `main`.

This task is strictly build isolation. Do not change research architecture, mathematics, Fast-CKKS logic, benchmark semantics, experiment flow, or CKKS/Bootstrap parameters.

## Design

Use the smallest existing-repo pattern:

1. Add a build constraint to only those non-test Primary source files that directly require the Fast-only package or otherwise cannot compile against Standard Lattigo:
   `//go:build !lattigo_standard`

2. Add one Standard-only stub file:
   `//go:build lattigo_standard`

   The stub must define only the runner entrypoints that remain referenced by `main.go` after the Fast-only files are excluded.

3. Stub functions must preserve the exact signatures and return an explicit error such as:
   `diagnostic requires the Fast Lattigo backend`

4. Do not emulate or vendor a fake `schemes/ckks/fast` package.

5. Do not delete or rewrite diagnostic logic.

6. Do not change normal/default Fast build behavior. A normal:
   `go build .`
   against `../lattigo@fast-ckks` must continue to include the existing diagnostics.

## Required audit

Before editing, perform an exact import audit across Primary non-test `.go` files and identify every compilation unit that directly imports `schemes/ckks/fast`.

Do not tag files based only on filename patterns.

If a Fast-only tagged file exports helpers required by an untagged file, use the narrowest additional isolation/stub necessary. Do not perform package restructuring.

## Build tag semantics

`lattigo_standard` is only a source-isolation tag.

It must NOT choose Standard versus Fast CKKS semantics.

Backend selection remains the existing local replacement:

`replace github.com/tuneinsight/lattigo/v6 => ../lattigo`

and the checked-out Secondary branch determines which backend source is used.

## Build script

Add one small script, preferably:

`scripts/build-linux-amd64.sh`

It should:

1. verify Primary is in a safe/expected state;
2. verify `../lattigo` exists and is a git worktree;
3. refuse destructive branch switching if Secondary has uncommitted changes;
4. remember the current Secondary branch/commit;
5. build Standard from `../lattigo@main` with:
   - `CGO_ENABLED=0`
   - `GOOS=linux`
   - `GOARCH=amd64`
   - `GOAMD64=v1`
   - `-trimpath`
   - `-tags lattigo_standard`
6. build Fast from `../lattigo@fast-ckks` with the same portability flags but without `lattigo_standard`;
7. restore the original Secondary branch/commit on exit, including failure paths;
8. write:
   - `dist/bootstrap-standard-linux-amd64`
   - `dist/bootstrap-fast-linux-amd64`

Do not run `git reset --hard`, `git clean`, stash user work, or force-switch over dirty state.

## Acceptance

Required evidence:

1. With `../lattigo@main`:
   `go build -tags lattigo_standard .`
   succeeds.

2. With `../lattigo@fast-ckks`:
   `go build .`
   succeeds.

3. `./scripts/build-linux-amd64.sh` produces both expected binaries.

4. Both outputs are Linux x86-64 executables.

5. Default Fast diagnostic behavior remains available in the Fast build.

6. Invoking an isolated Fast-only diagnostic from the Standard build fails cleanly with the explicit stub error, rather than compile failure.

7. No Secondary source files are modified.

8. No research/math/Fast-CKKS/benchmark logic is changed.

9. `git diff --check` passes.

## Expected scope

Prefer only:
- build constraints on the exact Fast-only diagnostic files;
- one Standard stub file;
- one build script;
- optional minimal README/build note if truly necessary.

Do not change `main.go` unless compilation proves a tiny change is strictly necessary; prefer stubs.

## Completion

Codex should commit and normal-push Primary only, then report:

`READY_FOR_WEB_REVIEW`

with:
- exact isolated files;
- stubbed entrypoints;
- Standard/Fast build commands and results;
- produced binary names;
- confirmation Secondary remained source-unchanged and was restored.

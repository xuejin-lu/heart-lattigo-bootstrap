"""Read-only worktree path guard for resuming a verified task checkpoint.

This helper checks only whether Git-reported changed paths are within an
explicit task allowlist. It does not establish task ownership, validate file
contents/provenance, fetch, merge, or modify the worktree.
"""

from __future__ import annotations

import argparse
import os
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path, PurePosixPath


@dataclass(frozen=True)
class Change:
    status: str
    paths: tuple[str, ...]


@dataclass(frozen=True)
class Audit:
    branch: str
    changes: tuple[Change, ...]
    unexpected_paths: tuple[str, ...]
    unsafe_changes: tuple[Change, ...]
    expected_branch: str

    @property
    def accepted(self) -> bool:
        return (
            self.branch == self.expected_branch
            and not self.unexpected_paths
            and not self.unsafe_changes
        )


def _repo_relative_path(path: str) -> str:
    normalized = PurePosixPath(path)
    if not path or normalized.is_absolute() or ".." in normalized.parts:
        raise ValueError(f"task path must be a repository-relative path: {path!r}")
    return normalized.as_posix()


def parse_porcelain_z(output: bytes) -> tuple[Change, ...]:
    """Parse `git status --porcelain=v1 -z` without splitting filenames on spaces."""
    fields = output.split(b"\0")
    changes: list[Change] = []
    index = 0
    while index < len(fields):
        field = fields[index]
        index += 1
        if not field:
            continue
        if len(field) < 3 or field[2:3] != b" ":
            raise ValueError(f"invalid porcelain status record: {field!r}")

        status = field[:2].decode("ascii", errors="strict")
        paths = [os.fsdecode(field[3:])]
        if "R" in status or "C" in status:
            if index >= len(fields) or not fields[index]:
                raise ValueError("rename/copy status is missing its second path")
            paths.append(os.fsdecode(fields[index]))
            index += 1
        changes.append(Change(status=status, paths=tuple(paths)))
    return tuple(changes)


def classify_worktree(
    status_output: bytes,
    *,
    branch: str,
    expected_branch: str = "main",
    allowed_task_paths: set[str] | frozenset[str] = frozenset(),
) -> Audit:
    allowed = {_repo_relative_path(path) for path in allowed_task_paths}
    changes = parse_porcelain_z(status_output)
    unexpected = {
        path
        for change in changes
        for path in change.paths
        if _repo_relative_path(path) not in allowed
    }
    # Deletions, unresolved merges, renames and copies are not safe checkpoint
    # states, even if one of their paths was accidentally allowlisted.
    unsafe = tuple(
        change
        for change in changes
        if any(marker in change.status for marker in ("D", "U", "R", "C"))
    )
    return Audit(
        branch=branch,
        changes=changes,
        unexpected_paths=tuple(sorted(unexpected)),
        unsafe_changes=unsafe,
        expected_branch=expected_branch,
    )


def audit_worktree(
    repo: str | Path,
    *,
    allowed_task_paths: set[str] | frozenset[str] = frozenset(),
    expected_branch: str = "main",
) -> Audit:
    repo_path = Path(repo).resolve()

    def git(*args: str) -> bytes:
        return subprocess.run(
            ["git", "-C", str(repo_path), *args],
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        ).stdout

    branch = git("branch", "--show-current").decode("utf-8").strip()
    status = git("status", "--porcelain=v1", "-z", "--untracked-files=all")
    return classify_worktree(
        status,
        branch=branch,
        expected_branch=expected_branch,
        allowed_task_paths=allowed_task_paths,
    )


def _format_change(change: Change) -> str:
    return f"{change.status} " + " | ".join(change.paths)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", default=".", help="Git worktree to inspect")
    parser.add_argument(
        "--expected-branch", default="main", help="required current branch"
    )
    parser.add_argument(
        "--allow-task-path",
        action="append",
        default=[],
        help="exact repository-relative path already verified against the active task; repeat as needed",
    )
    args = parser.parse_args(argv)

    try:
        audit = audit_worktree(
            args.repo,
            allowed_task_paths=set(args.allow_task_path),
            expected_branch=args.expected_branch,
        )
    except (OSError, subprocess.CalledProcessError, UnicodeError, ValueError) as exc:
        print(f"STOP: worktree audit failed: {exc}", file=sys.stderr)
        return 2

    if audit.branch != audit.expected_branch:
        print(
            f"STOP: branch {audit.branch!r} does not match {audit.expected_branch!r}"
        )
    if audit.changes:
        for change in audit.changes:
            print(_format_change(change))
    else:
        print("CLEAN")

    if not audit.accepted:
        for path in audit.unexpected_paths:
            print(f"STOP: path is not an explicitly verified task path: {path}")
        for change in audit.unsafe_changes:
            print(f"STOP: unsafe worktree state: {_format_change(change)}")
        return 2

    if audit.changes:
        print("TASK_PATHS_ONLY: continue only after content/provenance review")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

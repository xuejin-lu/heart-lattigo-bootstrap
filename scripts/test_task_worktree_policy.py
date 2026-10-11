"""No-Bootstrap tests for the startup worktree checkpoint policy."""

from __future__ import annotations

import subprocess
import tempfile
import unittest
from pathlib import Path

from task_worktree_policy import audit_worktree


class TaskWorktreePolicyTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tempdir = tempfile.TemporaryDirectory(prefix="task-worktree-policy-")
        self.repo = Path(self.tempdir.name)
        subprocess.run(
            ["git", "init", "--quiet", "--initial-branch=main", str(self.repo)],
            check=True,
        )
        subprocess.run(
            ["git", "-C", str(self.repo), "config", "user.name", "Workflow Test"],
            check=True,
        )
        subprocess.run(
            [
                "git",
                "-C",
                str(self.repo),
                "config",
                "user.email",
                "workflow-test@example.invalid",
            ],
            check=True,
        )
        (self.repo / "tracked.txt").write_text("baseline\n", encoding="utf-8")
        subprocess.run(
            ["git", "-C", str(self.repo), "add", "tracked.txt"], check=True
        )
        subprocess.run(
            ["git", "-C", str(self.repo), "commit", "--quiet", "-m", "fixture"],
            check=True,
        )

    def tearDown(self) -> None:
        self.tempdir.cleanup()

    def test_expected_untracked_task_result_can_resume(self) -> None:
        result_path = "results/expected-summary.json"
        target = self.repo / result_path
        target.parent.mkdir()
        target.write_text('{"task":"fixture"}\n', encoding="utf-8")

        audit = audit_worktree(
            self.repo,
            expected_branch="main",
            allowed_task_paths={result_path},
        )

        self.assertTrue(audit.accepted)
        self.assertEqual(audit.unexpected_paths, ())
        self.assertEqual(audit.changes[0].status, "??")
        self.assertEqual(audit.changes[0].paths, (result_path,))

    def test_unknown_tracked_modification_stops_even_with_expected_result(self) -> None:
        result_path = "results/expected-summary.json"
        target = self.repo / result_path
        target.parent.mkdir()
        target.write_text('{"task":"fixture"}\n', encoding="utf-8")
        (self.repo / "tracked.txt").write_text("unreviewed edit\n", encoding="utf-8")

        audit = audit_worktree(
            self.repo,
            expected_branch="main",
            allowed_task_paths={result_path},
        )

        self.assertFalse(audit.accepted)
        self.assertEqual(audit.unexpected_paths, ("tracked.txt",))


if __name__ == "__main__":
    unittest.main()

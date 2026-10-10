# Web review: Batch024 S4 isolated clone repair

**Decision:** allow same Batch024 S4, conditional on zero-Bootstrap tests. Earlier blanket stop on unrelated prunable Git metadata is superseded.

Source diagnosis: the existing `git worktree add` writes the original Secondary repository's `.git/worktrees`, which this sandbox denied. Existing historical prunable entries are unrelated and must remain untouched. The existing `cmd/fastdiag` is changed to create a **separate task-owned local Git clone**, pinned to the exact diagnostic commit, with independent `.git` metadata. Only the task-owned checkout is removed after a certified trace. No additional measurement runner or CKKS arithmetic changes.

The Web checked the approach with a local tiny Git repository, demonstrating that an independent `git clone --shared --no-checkout` + `git checkout --detach` leaves the original `.git/worktrees` untouched. A focused repository test for preexisting prunable metadata and output collision is added. Actual user's Mac execution and Go tests **have not yet been run by Web**.

At the next Codex `開始`, first safe sync both repos. Run focused `go test ./cmd/fastdiag -count=1`, `go vet ./cmd/fastdiag`, `git diff --check` with permitted isolated `GOCACHE`; review test and clone source. Verify saved Batch023 authentic six-output vectors, original Fast and Standard pins, exact Secondary diagnostic commit, empty Batch024 attempt journals and unique output names. Do not prune/modify old `.git/worktrees` entries. The old entries' mere existence is not a blocker.

Only if all zero-call gates pass may the existing S4 procedure reserve **exactly two new Fast diagnostic attempts** (one cold, one traced warm). After reservation or process start, any failure is terminal: no retry. On success produce raw evidence before event validation, complete detailed Stage/Power/Rescale cost and S5. No math changes, Standard run or reuse of Batch023 14/14 spent calls.

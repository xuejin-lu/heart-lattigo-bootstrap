# FAST-STORAGE-006 — Salvage Private-F Scalar + Normalized Trace Foundation

## Status

Recovery/salvage task after FAST-INTEGRATION-003 performance rejection.

## Task class

`I — Implementation / controlled dirty-worktree recovery`

This task exists because FAST-INTEGRATION-003 produced correctness-valid but performance-invalid local Secondary changes and intentionally stopped before commit/push.

## Production decision

FAST-INTEGRATION-003 is **not accepted as a production integration**.

Reported same-environment LogN13 full-ModUp measurements:

- Integration-002 accepted Fast path: ~0.578–0.586 ms/op, ~314 KB/op, ~402 allocs/op.
- Integration-003 resident path: ~3.76–3.79 ms/op, ~2.70 MB/op, ~375 allocs/op.
- Standard reference: ~7.56–7.66 ms/op.

Interpretation:
- resident private-F ModUp+Trace is still faster than Standard;
- but it is about 6.5x slower than the accepted Integration-002 Fast path;
- the 20% production-regression gate therefore fails decisively;
- allocs/op did not explode (375 vs ~402), while allocated bytes increased by ~8.6x.

Do not weaken the gate. Production must remain on FAST-INTEGRATION-002 after this recovery.

## Goal

Salvage only the reusable, production-neutral private-F foundation proven by the 003 work:

1. private-F integer scalar multiplication;
2. normalized private-F Trace;
3. exact Trace bound/capacity checks;
4. correctness tests for those standalone primitives;
5. safe immutable private-F basis reuse/cache only if it is independently useful and does not alter production semantics.

Remove/revert the rejected production-resident wiring so the production ModUp path is exactly the accepted Integration-002 fused path.

## Dirty-worktree recovery authority

The Secondary worktree is expected to contain the uncommitted FAST-INTEGRATION-003 changes.

For this task only, Codex is explicitly authorized to inspect and reconcile that known dirty diff.

Do **not**:
- reset --hard;
- stash and forget;
- discard the whole diff;
- overwrite unrelated user work.

Instead:
1. identify the files/hunks belonging to FAST-INTEGRATION-003;
2. preserve only the standalone foundation listed above;
3. restore production ModUp/Bootstrap wiring to the current committed HEAD semantics (FAST-INTEGRATION-002);
4. remove rejected integration-only tests/benchmarks or rewrite them as standalone foundation tests where useful;
5. ensure no unrelated dirty files remain.

If any dirty change is unrelated to FAST-INTEGRATION-003, stop and report `NEEDS_WEB_REVIEW` before touching it.

## Frozen math

Normalized private-F Trace theorem remains accepted:

[
Y = rac{mathcal T_g(X)}{g},
qquad
mathcal T_g(X)=sum_{sigmain H_g}sigma(X).
]

Required precondition:

[
2gB_j < S_3
quadorall j.
]

The private-F implementation must:
- form the unnormalized automorphism sum first;
- only then normalize by multiplying residues by `g^{-1} mod f_i`;
- never pre-multiply the authoritative private-F state by modular `g^{-1}`.

Post-bound may conservatively remain:

[
B'_j le B_j.
]

Private-F integer scalar:

[
Y=mX,qquad B'_j=|m|B_j.
]

Reject capacity overflow transactionally.

## Standalone APIs to retain

Prefer retaining the APIs already implemented locally if clean and well-tested, conceptually:

- `FastStorageMulInteger(...)`
- `FastStorageTraceNormalized(...)`

Exact names from the local implementation may remain if reasonable.

Requirements:
- no LogicalQ/full-RNS fallback;
- no production Bootstrap dependency;
- fixed-width-3 support is sufficient for Trace production experiments;
- NTT private-F path supported;
- input semantics and bounds validated;
- no per-coefficient arbitrary-precision reconstruction in hot loops where unnecessary.

## Production state after recovery

The committed production path must be the accepted FAST-INTEGRATION-002 path:

[
	ext{Level-0 LogicalQ}
	o
	ext{FusedLevel0ModUpToCompactLogical}
	o
	ext{logical scale alignment}
	o
	ext{existing Fast logical Trace}
	o
	ext{Montgomery}.
]

Do not leave:
- private-F resident ModUp wiring;
- private-F scale-alignment wiring;
- private-F Trace wiring
in production `FastEvaluator.ModUp`.

## Required tests

Run standalone tests proving:
- signed/zero/large integer scalar semantics and exact bounds;
- Trace monomial survive/vanish theorem;
- exact divisibility by gap;
- private-F normalized Trace matches logical Trace after export for q01/q012 fixtures;
- intermediate-capacity rejection is transactional;
- gap=1 behavior;
- multiple logN/logSlots values.

Then run:
- `go test ./schemes/ckks/fast`
- `go test ./circuits/ckks/bootstrapping`
- `go test ./...`
- `git diff --check`
- gofmt check.

Production Integration-002 benchmark behavior should remain unchanged within ordinary run-to-run noise. A quick LogN13 Fast ModUp benchmark is sufficient as a guard; no new optimization gate.

## Commit / classification

Commit the salvaged foundation to Secondary `fast-ckks`.

Candidate classification:

`FAST_STORAGE_006_PRIVATE_F_TRACE_FOUNDATION_ACCEPTED_CANDIDATE`

Report:
- exact commit;
- dirty files found;
- which 003 hunks were preserved;
- which rejected production hunks were reverted;
- standalone APIs;
- tests;
- quick production ModUp benchmark guard;
- confirmation worktree is clean and production remains Integration-002.

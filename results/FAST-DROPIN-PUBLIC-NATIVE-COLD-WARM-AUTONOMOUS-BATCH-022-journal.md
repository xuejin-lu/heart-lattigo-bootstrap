# Batch 022 execution journal

## Provenance and repository state

- Primary `main` started from the synced `6bb003fefec03559b0aefaa1b6dbd27a4f6214de`; Batch022 A was committed/pushed as `cfd9c1548bc39fc697a8c264a4bfd42d00463b61`.
- Batch022 B implementation commit: `37002ff3bbd97889cf0ec8108ff2cd527ab8d3d5`; pushed normally to `origin/main`, then fetched and verified equal to local HEAD with clean worktree before measurement.
- Standard checkout: `/private/tmp/fast-standard-rebaseline-002/lattigo-standard`, clean detached `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Fast checkout: `/private/tmp/fast-standard-batch021-2d6145`, clean detached `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
- Active Secondary `fast-qprefix`: clean at `d463c336d511646994ba72f8dcfda5179431c430`; no Secondary files were modified.
- Temporary modfiles, vectors, result JSON and comparison JSON were kept under `/private/tmp/fast-dropin-batch022-egbOuI`; Primary `go.mod` and pinned Lattigo checkouts were unchanged.
- Primary measurement-source SHA-256: `41439844ad5139d07386367247c7f0537d7b752f95a3023a893f57a0e55f81d6`.

## Implementation and test sequence

1. Regression-first Batch022 A verified old comparator blind spots, then repaired vector hash binding, finite/exact-length checking, oracle recomputation, canonical coverage, state/row/capacity and provenance validation. Focused tests, both build-tag compile tests, `go vet`, and diff check passed. Committed/pushed A as `cfd9c1548bc39fc697a8c264a4bfd42d00463b61`.
2. B extended the existing `public-native` runner with an explicit `--public-bootstrap` path gated by a passing zero-call pair and exact budget 2. Fake-call tests verified independent copies of one held input, held-input immutability, exactly two reservations, first-output failure stop, warm reservation collision stop, fresh output-path checks, and metadata-bound fingerprints. No test invoked Lattigo Bootstrap.
3. B focused tests, both build-tag compile tests, `go vet`, and `git diff --check` passed. B was committed/pushed as `37002ff3bbd97889cf0ec8108ff2cd527ab8d3d5`.
4. Ordinary fetch initially hit sandbox denial writing `.git/FETCH_HEAD`; formal permission escalation succeeded. The subsequent push was a normal fast-forward. A post-push fetch confirmed `HEAD == origin/main == 37002ff3bbd97889cf0ec8108ff2cd527ab8d3d5`.

## Zero-call preflight

Using `GOWORK=off`, isolated modfiles and `--bootstrap-budget=0`, ran the same Primary source with `-tags=perf_standard` and `-tags=perf_fast`, then `compare-public`. Pair status was `PASS`, `matched_environment=true`, actual Bootstrap calls `0`, all eight canonical primitive checkpoints passed at `1e-6`, with exact pinned SHAs and the same config, Q/P, input, workload and Primary source hashes. Pair SHA-256: `e142b5d53196ddae86b7e34ba2f80b111fd70553fba6a1abae6b9e614e968ce9`.

## Bounded Bootstrap calls

After preflight passed, Standard ran exactly cold + warm with budget 2. Both outputs passed native Level1 decrypt/decode and per-output plaintext oracle validation. Its two independent input copies had fingerprint `507dfdf925636b46a1ed49fd04ce1e03342a87c4d7ed457d5fe9935b2f59a060`.

Fast then ran exactly cold + warm with budget 2. Both outputs passed the same checks and Fast zero-secret row/c1 policy. Its two independent input copies had fingerprint `e512061ab0cc1ca3ebab443ab8d03ae021286124ade946d97fc51bc216a0c95f`.

No failure path, extra warmup, retry, benchmark sweep, LogN16, or additional Bootstrap call ran. Final `compare-public` status was `PASS`, actual calls `4`; cold and warm direct Fast-vs-Standard comparisons each had RMSE `4.7262101606249815e-9` and max complex difference `4.489543437647087e-8`, below `1e-6`. See the summary and compact evidence file for phases, allocations, output hashes and checkpoint metrics.

## Source-backed initialization interpretation

The pinned Fast source documents and implements lazy circuit setup in `circuits/ckks/bootstrapping/fast_bootstrap.go:18–49`; Standard calls `buildBootstrapCircuitData` during `Evaluator.initialize` from `NewEvaluator` (`circuits/ckks/bootstrapping/evaluator.go:169–178`). This matches the measured short Fast evaluator-construction phase and larger Fast first-call phase. One cold/warm sample is descriptive only; no stable speedup claim is made.

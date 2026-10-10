# Batch 023 execution journal

## S1 — source-backed reuse map (zero Bootstrap calls)

**Status:** complete; no experiment calls spent. Primary was synchronized at
`846241ab4751115bc42c5f3dfa614a246d5439fc`; Secondary `fast-qprefix` and
`origin/fast-qprefix` both remain `d463c336d511646994ba72f8dcfda5179431c430`,
clean. Formal code pins are genuine Standard
`5dbffbdea05394de2ca3a432ed5318aa832e3f40` and original Fast
`2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.

### Reusable source map

| Work / source | Existing instrumentation and event topology | Authority / invariants | Not measured by that event |
|---|---|---|---|
| Public lifecycle — Primary `cmd/perfprobe/public_native.go` (`runPublicNative`, `newPublicBackend`, `runTwoPublicBootstrapAttempts`) and `main.go` | Shared public operation setup, phase wall time and allocation deltas; records keygen/evaluation-key generation/evaluator construction, checkpoints, held-input fingerprints, native output decode and plaintext oracle. The current public Bootstrap lane is an exact cold+warm pair, gated by a passing zero-call pair and an exclusive per-call journal. | Same canonical LogN13/E32 config, Q/P, message and public operation chain; genuine Standard keygen/EncryptNew versus pinned Fast's intended zero-secret mode; Fast compact prefix, Q0123/q0 capacity and output state checks. | No Fast in-circuit internal event tree; a public stage replay would not equal the in-circuit Bootstrap path. |
| Bootstrap root + public boundaries — Secondary `circuits/ckks/bootstrapping/fast_bootstrap.go` (`BootstrapMany`, `bootstrapCore`) | `bootstrap` root; sibling `pack_n1_to_n2`, `scale_down`, `mod_up_trace`, `coeffs_to_slots`, `evalmod_real`, optional `evalmod_imag`, `slots_to_coeffs`, `unpack_n2_to_n1`, `public_finalization`. One public ciphertext means one call at each applicable boundary; packing carries an explicit packed-count field. Stage spans are siblings under the Bootstrap root. | Entry validation enforces degree 1, NTT/non-Montgomery, positive Scale, Level, `w_Q(Level)` rows, valid backing and zero c1; finalization restores the public residual boundary. Row fields are derived from the fixed Q-prefix policy where available. | `Event` has no NTT/Montgomery fields; several stage input rows are intentionally `-1`. No keygen/evaluation-key generation or circuit data construction phase is inside these stage intervals. |
| Generated powers — Secondary `circuits/ckks/polynomial/fast.go` (`genPower` / `genPowerFromChebyshevBasis` call path) | One `generated_powers` parent and per-power spans; children classify workspace copy, multiply/relinearize, Chebyshev doubling, recurrence correction and Rescale. Actual requested powers and counts depend on the configured PS schedule. | Power event metadata contains power, split, Level/Scale/Degree and available Q-prefix row counts; no dormant row becomes authoritative. Rescale uses logical Level divisor and transactional Q-prefix implementation. | Power parent currently has `ParentSequence=0`, not a child of `evalmod_real/imag`; do not claim stage-to-power exclusive time or sum these as disjoint stage time. No domain flags in event schema. |
| Rescale internals — Secondary `schemes/ckks/internal/fastcore/rescale.go` (`ApplyToRows` workspace path) | `rescale` parent → `preflight` and `materialization`; preflight → per-component `prefix_to_coefficient` and `coefficient_loop` → `reconstruct_center_round_capacity` and `residue_materialization`; materialization → per-component `ntt_montgomery_restore`. Each event is timed by existing `internal/fastdiag` spans/records. | Explicit `sourceRows` and fixed Q-prefix rows are authoritative; centered lift/capacity is checked before commit. Source/target representation follows the existing NTT/Montgomery implementation. | The event schema has no domain flags or coefficient hashes; source is not modified by diagnostics. Counts vary with actual components and rescale calls. |

The instrumented code is selected by the existing `fastdiag` build tag. No new
stopwatch or profiler is needed. The accepted legacy test
`TestFastDiagP93Q55Trace` uses P93/E0, `NewFastEvaluator`, an encoded c0 fixture,
an unrecorded baseline Bootstrap and user-configurable warmup/repetitions; it
therefore cannot serve as E32 public evidence and would exceed this charter's
two-call trace budget.

### Minimum E32 adapter design frozen before calls

Extend existing Primary `cmd/perfprobe` zero-call public-native flow to export,
only on an explicit Fast diagnostic-fixture request, the exact held
`drop_level0` ciphertext using its binary serializer plus a compact manifest
binding ciphertext bytes, serialized `bootstrapping.Parameters`, effective
profile/Q/P hashes, input/workload hashes and Level/Scale/degree/domain/rows.
Keep ciphertext, parameter and decoded-vector files in a task-owned temporary
directory; no key material is exported. The Secondary addition is limited to a
build-tagged test in `circuits/ckks/bootstrapping/fastdiag*_test.go`: read and
validate that held ciphertext and parameter manifest, construct keys through
the public `GenEvaluationKeys` and ordinary `NewEvaluator` dispatch, run exactly
one untraced cold call followed by one traced warm call, and compare both
outputs against the manifest's decoded oracle and the uninstrumented Fast
repeatability vectors. Keep the P93 test unchanged. No Secondary non-test Go
source, API or arithmetic changes are authorized.

This avoids duplicating parameter generation: the manifest will carry the
canonical parameters produced by the shared Primary `perfmeasure` config path,
and Secondary will verify its recomputed Q/P/profile and held-cipher metadata
against that manifest before any Bootstrap. If binary round-trip, public
dispatch, vector matching, exact two-call accounting or the 1e-6 canonical
oracle cannot be proven in zero-call preflight, stop before trace calls.

### Measurement and call-budget freeze

- Profile: LogN13/E32, canonical 022 Q/P and 4096-slot workload; same Primary
  frontend and public `EncryptNew → AddNew → MulRelinNew → Rescale(q5) →
  RotateNew → DropLevelNew(4) → Bootstrap → native Level1 DecryptNew/Decode`.
- Standard: clean detached original pin; Fast timing: clean detached original
  production pin; trace: isolated test-only instrumented checkout whose every
  non-test Go source file must hash-identically to the Fast production pin.
- Timing environment: preserve the accepted 022 Apple M4/macOS arm64,
  Go 1.26.4, `GOMAXPROCS=10` conditions; record actual CPU, GC environment and
  timestamp in each artifact and reject mismatches. Setup/keygen/evaluator
  construction remain outside the Bootstrap wall interval and are reported
  separately. First call is cold; five calls on the same evaluator are warm,
  each on a verified independent copy of the same held input.
- Uninstrumented lane: maximum 6 Standard + 6 Fast calls. Independent Fast
  trace lane: maximum 2 additional calls, cold untraced then warm traced.
  Across the batch the ceiling is 14; no retries, implicit baselines,
  benchmarks or test warmups. Reserve the full two-call trace budget
  exclusively before spawning its Go test process. Any ambiguous process
  failure consumes its entire reservation.
- Historical QPREFIX-DIAG-007 / OPT-004 and Batch022 are prior art only: no
  P93 timing or single-sample 4.26x figure is transferred as E32 evidence.

### S1 inventory conclusion

Classification: **reusable capability exists; E32 adapter is missing**. The
smallest authorized gap is the held-cipher/parameter manifest bridge and a
test-only E32 consumer, plus extending the existing bounded public-native mode
from two to six Bootstrap calls without changing its legacy exact-two schema.
The actual event tree reports stage siblings and power/rescale nesting only;
stage↔power temporal attribution and NTT/Montgomery state cannot be fabricated
from current events. Later reports must label these visibility limits. S1 spent
**0** Standard and **0** Fast Bootstrap calls.

## Global expensive-call ledger

| Lane | Reserved/used | Maximum | Status |
|---|---:|---:|---|
| Standard public E32 | 6 | 6 | spent; six outputs passed oracle |
| Original Fast public E32 | 6 | 6 | spent; six outputs passed oracle |
| Fast-only E32 internal trace | 2 | 2 | spent; event-tree gate failed after both calls |
| **Total** | **14** | **14** | exhausted; do not retry |

## Checkpoint journal

| Checkpoint | Status | Commit / verification | Next |
|---|---|---|---|
| S1 source map | complete | Primary `94ba05366bd7b2b28b1e1f962eb592cfd08786f9`; pushed and clean before S2 | extend reusable tools and add zero-call tests |

## S2 — reusable six-sample and public E32 trace adapters (zero Bootstrap calls)

**Status:** implementation and zero-call tests complete; Secondary test-only
adapter and pre-reservation fixture-validation gate are committed and pushed
through `f7eb9f88d0c877331e62287508c541a1e1147bdc`. Primary S2 core is locally
committed as `8a059e54379a41c3d8aff06ba27016bc74201d1e`, with validation-order
follow-up `6e938918442c409faa6d32e159c7a3f7a1041f7a`.
No Bootstrap was invoked. The only Secondary source change is a new
`fastdiag`-tagged test file; no non-test Secondary source, API, arithmetic,
Standard code, `AGENTS.md`, or `CURRENT_TASK.md` was changed.

- Primary `cmd/perfprobe` now accepts `--public-repeatability` with the exact
  budget six and records one `first_cold_bootstrap` plus five
  `warm_bootstrap_01..05` samples on independent copies of the exact held input.
  It preserves the old exact-two mode and schemas. Fake-call regression tests
  cover the six calls, immutable held input, per-attempt reservation, output
  naming, budget exhaustion, each of six failure positions, reservation
  collision/restart behavior, and output-path collisions.
- An explicit zero-call Fast `--trace-fixture-out` option writes the exact held
  Level0 ciphertext, binary serialized public bootstrapping parameters, and a
  compact provenance/state/hash manifest. The artifacts contain no keys and
  are created exclusively with owner-only permissions.
- Existing `cmd/fastdiag trace` now has the bounded `logn13-e32-public`
  profile. It validates the fixture and six-output vectors, exact Primary and
  pinned production-Fast identities, clean Primary `main` descended from its
  synchronized `origin/main`, clean/pushed Secondary `fast-qprefix`, and equality
  of all non-test production Go source hashes plus `go.mod`/`go.sum`. It first
  runs the same Secondary adapter in validate-only mode to decode and verify
  serialized parameters/ciphertext and matched vectors without key generation,
  evaluator construction, or Bootstrap. Only after that zero-call gate passes
  does it reserve both diagnostic attempt journals and run the cold+traced-warm
  process. Its raw event tree remains in a task-owned temp
  directory outside Git; the final command emits a compact summary. The same
  already-budgeted warm traced call is wrapped by Go's built-in CPU profiler;
  a post-warm in-use heap snapshot is emitted after the call. Both profiles are
  outside Git and SHA-bound. Profiling overhead is confined to the separate
  instrumented diagnostic lane and is never folded into production timings.
- Secondary `TestFastDiagPublicE32Trace` reads the exact serialized held
  ciphertext/parameters, verifies hashes, metadata, Level0/Q-prefix row
  authority, zero-c1 semantics and decoded input identity, then constructs the
  ordinary public `GenEvaluationKeys` / `NewEvaluator` path. When invoked by
  the parent tool it runs exactly one untraced cold and one traced warm call,
  validating both against the decoded plaintext oracle and all six original
  Fast repeatability outputs. The legacy P93 test remains unchanged.
- Pure tests validate trace parent/child structure and required generated
  powers; existing `cmd/fastdiag` aggregate tests continue to cover immediate-
  child closure normalization. Reservation tests prove both tokens exist
  before the subprocess callback and that a partial collision prevents launch.

Zero-call verification:

```text
GOWORK=off go test ./cmd/perfprobe ./cmd/fastdiag ./internal/perfmeasure ./internal/numericalmetrics -count=1 — PASS
GOWORK=off go vet ./cmd/perfprobe ./cmd/fastdiag ./internal/perfmeasure ./internal/numericalmetrics — PASS
GOWORK=off go test -modfile=/private/tmp/fast-dropin-batch022-egbOuI/standard.mod -tags=perf_standard ./cmd/perfprobe ./internal/perfmeasure -run '^$' -count=1 — PASS (compile-only)
GOWORK=off go test -modfile=/private/tmp/fast-dropin-batch022-egbOuI/fast.mod -tags=perf_fast ./cmd/perfprobe ./internal/perfmeasure -run '^$' -count=1 — PASS (compile-only)
GOWORK=off go test -modfile=/private/tmp/fast-dropin-batch022-egbOuI/standard.mod -tags=perf_standard ./cmd/perfprobe ./internal/perfmeasure ./internal/numericalmetrics -count=1 — PASS
GOWORK=off go test -modfile=/private/tmp/fast-dropin-batch022-egbOuI/fast.mod -tags=perf_fast ./cmd/perfprobe ./internal/perfmeasure ./internal/numericalmetrics -count=1 — PASS
go test -tags fastdiag ./circuits/ckks/bootstrapping -run '^TestFastDiagPublicE32EventValidatorAcceptsSourceShapedTree$' -count=1 — PASS
go test -tags fastdiag ./circuits/ckks/bootstrapping -run '^$' -count=1 — PASS (compile-only)
```

Two test-harness misfires were fail-closed and spent zero Bootstrap calls:
an early Standard-tag test used Primary's default Fast dependency and correctly
failed the native nontrivial-c1 assertion; it was replaced with the exact
Standard modfile compile-only check above. A broad Secondary test regex entered
the E32 tracer without fixture variables and failed before evaluator setup or
Bootstrap; subsequent runs used only the synthetic event-tree test and
compile-only mode. Neither failure changed files or spent a call token.

S2 self-review unified the serialized-parameters payload, manifest filename,
and preflight reservation at `.parameters.bin` (with a path regression check),
strengthened fixture/vector provenance and finite-value checks, ensured public
output state is Level1/Scale-preserving/two-row, and verified production-source
identity before any test process can launch. It also requires the frozen
Go/macOS/Apple-M4/CPU-count/GOMAXPROCS environment and records/matches
GOGC/GOMEMLIMIT/GODEBUG across the trace parent and child. CPU model detection
now reuses a shared metadata helper that safely reads only the `Chip:` line
from `system_profiler` when `sysctl` is unavailable. Primary
`git diff --check` passes. The expensive-call ledger remains **0/14**.

### S3 zero-call preflight — preliminary provenance at Primary `8a059e5`

The canonical Standard and original-Fast `public-native` preflights both
completed with `bootstrap_budget=0`, `actual_bootstrap_calls=0`, LogN13/E32,
4096 slots and all eight required checkpoints. Every per-backend plaintext
oracle passed; the matched pair artifact is `PASS`, all 8/8 states match, and
the environments match. The shared capacity plan records
`rescale_bound=2271135713118062` and `q0=36028797018652673`, so `2*B < q0`;
the full Q0123 product is recorded in the raw preflight artifacts. The Fast
fixture manifest and `.parameters.bin` / `.ciphertext.bin` files were created
outside Git, with Level 0, degree 1, Scale `2^45`, one active Q-prefix row,
4096 decoded slots, and pinned Fast / Primary provenance. These artifacts are
preliminary only: the subsequent Primary validate-only ordering repair changes
the Primary commit, so S3 must regenerate and revalidate preflight artifacts
against the final committed Primary source before any expensive Bootstrap call.
No call tokens were reserved in this preliminary run. Its output is superseded
by the final provenance-bound run below.

### S3 final preflight and bounded repeatability result — Primary `6e93891`

The final zero-call preflight was regenerated after the Primary source commit.
Both backends recorded exactly the eight required checkpoints, all individual
oracles passed, and `compare-public` returned `PASS` with matched environment
and 8/8 matched states. Both had `bootstrap_budget=0` and
`actual_bootstrap_calls=0`. Provenance: Standard
`5dbffbdea05394de2ca3a432ed5318aa832e3f40`; original production Fast
`2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`; diagnostic Secondary
`f7eb9f88d0c877331e62287508c541a1e1147bdc`; Primary
`6e938918442c409faa6d32e159c7a3f7a1041f7a`. The shared LogN13/E32, 4096-slot
environment was Go 1.26.4 / darwin-arm64 / Apple M4 / NumCPU 10 /
GOMAXPROCS 10 with GOGC, GOMEMLIMIT and GODEBUG at runtime defaults. The
capacity gate has `rescale_bound=2271135713118062`, `q0=36028797018652673`,
and `2*B < q0`.

The uninstrumented formal lanes completed their exact six calls each. Every
Standard and Fast output passed its native plaintext oracle and ended at
Level 1, Scale `2^45`, two Q-prefix rows; each lane used copies of one held
input fingerprint. The six paired Bootstrap comparisons all passed the
`1e-6` max-complex gate with matched state. The worst paired max-complex
difference was `4.062428762032295e-8` (RMSE `5.0361026156396345e-9`). Warm
timing samples (ns) were:

| Backend | Cold (ns) | Warm samples (ns) | Warm median (ns) | Warm allocation median |
|---|---:|---|---:|---:|
| Standard | 356006584 | 303893791, 305669291, 306636208, 309936084, 310854375 | 306636208 | 91750008 B / 35024 allocs |
| Fast | 527153500 | 72302458, 74076542, 72580250, 72008291, 73113500 | 72580250 | 7505112 B / 8085 allocs |

The warm median ratio is `4.225x` for this pinned profile/environment only;
Fast remains the intentionally insecure zero-secret backend and this is not a
secure-FHE equivalence claim. Standard `GenEvaluationKeys` / CKKS evaluator /
Bootstrap evaluator construction were 403513666 / 1861250 / 354975917 ns;
Fast construction was 823817666 / 900958 / 2498792 ns.

### S3 trace lane — BLOCKED; reservation fully consumed

The first unsandboxed launch failed before creating a detached checkout and
before reserving tokens; a formal permission escalation then allowed the
trace runner to create an isolated, clean checkout at diagnostic commit
`f7eb9f88d0c877331e62287508c541a1e1147bdc`. Its validate-only fixture gate
passed. The runner reserved both diagnostic tokens before launching the test.
The test completed its untraced cold and traced warm Bootstrap calls, and both
outputs passed their decoded-oracle and Fast-reference checks. It then failed
the first Rescale event-tree assertion: the Rescale parent had only one direct
child, `preflight`; the validator required exactly `preflight` and
`materialization`. The failure occurred at
`circuits/ckks/bootstrapping/fastdiag_public_e32_test.go:328` / `:492`, before
the raw event JSON was written. The two reserved calls are irrevocably spent;
the global 14-call budget is exhausted and this trace must not be retried.

Preserved diagnostic checkout:
`/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-4268150718/secondary-diagnostic`
(clean, HEAD `f7eb9f88d0c877331e62287508c541a1e1147bdc`). The warm CPU profile
was written before the event validation failed at
`/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-4268150718/profiles/warm-cpu.pprof`
(3131 bytes, SHA-256
`c12b7e3bcb63fe935e5750573f5019a4cb102640328f90b5d63cbebc35755162`). No raw
event JSON or post-warm heap profile was produced. Since the accepted event
tree and Rescale closure are unavailable, no stage/power/rescale Pareto,
closure, pprof attribution, or Amdahl bound is reported. Classification:
`BATCH_BLOCKED_NEEDS_WEB_REVIEW`; no more Bootstrap calls are authorized.

The earlier S2 statement that the ledger remained 0/14 describes its state at
that checkpoint only; the final authoritative ledger above is 14/14.

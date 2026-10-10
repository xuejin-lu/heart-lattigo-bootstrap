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
| Standard public E32 | 0 | 6 | not started |
| Original Fast public E32 | 0 | 6 | not started |
| Fast-only E32 internal trace | 0 | 2 | not started |
| **Total** | **0** | **14** | within budget |

## Checkpoint journal

| Checkpoint | Status | Commit / verification | Next |
|---|---|---|---|
| S1 source map | complete | zero-call source audit; Primary compact journal commit pending | extend reusable tools and add zero-call tests |

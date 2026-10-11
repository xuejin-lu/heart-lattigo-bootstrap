# FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005

## Review decision

This task follows `FAST-LOGN16-BACKEND-PATH-AND-PERFORMANCE-004`, which ended `CURRENT_PERFORMANCE_BLOCKED / PUBLIC_NATIVE_LOGN16_PROFILE_BLOCKED`, 0/16 Bootstrap attempts spent. **Do not bypass or weaken the capacity gate. Do not alter Secondary arithmetic, Q/P, Scale or public API contracts.** The original LogN16/E32 canonical workload is archived and stays blocked under its unmodified inputs. This task authorizes a **separately named capacity-safe LogN16 public-native benchmark workload**, followed by bounded existing-platform performance measurement ONLY if all zero-call preconditions pass.

## Scientific diagnosis

The failing gate in `internal/perfmeasure/public_workload.go` is an independently calculated conservative coefficient upper bound for a specific public native Multiply–Rescale–DropLevel0 chain: `B=18169085704944493`, `q0=36028797019488257`, `2B-q0=309374390400729` (about 0.859% of q0). It is **NOT an observed coefficient overflow** and says nothing about whether Fast LogN16 kernel dispatch happened; the runner returned before construction. The condition is a valid conservative prerequisite for the existing formal workload and must be kept. The old LogN16 manually encoded c0/c1=0 diagnostic was an entirely different input protocol and cannot replace native encrypted Standard for formal timing.

A bounded, precisely specified data-only workload change is acceptable for a **new** performance profile, if separated in identity and never presented as a result of the original unchanged LogN16 canonical workload. Preserve the original failed gate and original LogN13/E32 frozen fixture.

## Contract: safe profile, no backend-specific frontend behavior

Use the existing Primary `cmd/perfprobe`, `cmd/fastdiag`, `internal/perfmeasure`, `internal/numericalmetrics`, existing build-tag public backend adapters, provenance, allocation timers, protected attempt journals. The same public frontend remains unchanged between Standard and Fast, using ordinary `ckks.NewEvaluator` and `bootstrapping.NewEvaluator`; only independently linked genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` versus current Fast `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` differs. Keep two clean, independently pinned immutable detached Secondary checkouts with separate task-local modfiles and `GOWORK=off`; do not switch a live branch or change committed `go.mod`. The ordinary Fast library may use internal `NewFastEvaluator`, but production frontend must not call it.

**New benchmark profile ID:** `logn16-e32-c15over16-v1` (or a semantically identical versioned schema name).

- N=65536, LogN=16, LogSlots=15 (32768 complex slots), original `configs/bootstrap_config.logN16.json`, original **effective Q/P**, E=32, Scale A/B=2^45, Scale C=q5, unchanged original deterministic A and B vectors.
- Generate C by first applying the existing `PublicMulRescaleWorkload(slots)` canonical C (currently C=A), then setting **C_new[i]=(15/16)*C_original[i]** for each complex slot; IEEE float64 15/16 is exact binary. No changes to A/B, ciphertext operations, matrix/DFT, CKKS frontend behavior, or effective prime generation.
- Use the **same new A/B/C values for Standard and Fast** in this profile, and update exact new `PublicMulRescaleWorkloadFingerprint`, plaintext oracle, expected vector outputs, capacity plan, and matching pair validators; old logn16 `A` input/config/Q/P SHA remain same but workload SHA MUST change. Never copy the existing LogN13 frozen workload identity into this new profile.
- Treat this as a *benchmark-specific workload variant*, NOT general improvement to cryptographic numerical precision nor a CKKS parameter change. Implement via a named explicit profile-selection helper in `internal/perfmeasure` and `cmd/perfprobe`; no hidden fallback or silent scaling when a gate fails.
- `BoundPublicMulRescaleWorkload` still must pass both `Q0123` and terminal strict `2B<q0` for the new variant. Recompute exact integer bounds with existing code; include `B`, `q0`, `2B`, positive headroom `q0-2B`, and fractions in report. Expected rough improvement from reducing C is ~6%, but **DO NOT assert actual pass until independently recomputed exact bound**.
- Regression tests MUST show (a) old LogN16 canonical C=A still fails by the original exact deficit, (b) new named C15/16 workload passes using its own exact q0/Q/P and independent deterministic hash, (c) original LogN13 frozen E32 values/hashes and 004/002/003 reports are untouched, (d) Standard/Fast consume the exact same new workload and correctly computed oracle, with fail-closed on identity mismatch.

## Phase A: ZERO Bootstrap static/preflight checks

1. Mandatory safe sync and read Primary `AGENTS.md`, `docs/BACKEND_TRANSPARENCY_CONTRACT.md`, `docs/MEASUREMENT_PLATFORM.md`, current task and `004` report, Secondary safe repository rules. Preserve local owned progress/checkpoints and historical consumed budgets.
2. Audit the original bound proof/assumptions. Do NOT reinterpret failure of an upper bound as observed overflow, and do not remove/disable the guard. Implement the small new variant in **the existing** versioned measurement platform; keep the same full E32 public native chain `EncryptNew → AddNew → MulRelinNew → Rescale(q5) → RotateNew → DropLevelNew(4) → Bootstrap → native low-level decrypt/decode`.
3. On both independent pinned binaries, run complete **zero-full-Bootstrap** preflights and eight native public checkpoints with current Fast, genuine Standard and clean matched Q/P, A/B/C and input/oracle. Require correct `c1` modes: genuine Standard native nonzero c1; Fast internally zero-secret. Require all finite values, correct Level/Scale/Degree, active Q-prefix widths, all individual plaintext-oracle checkpoints and pairwise numerical gates, full correct source provenance. Reuse `compare-public` and versioned profile labels in `perfprobe`. Any unsupported current LogN16 kernel capability or mismatch => fail closed, no performance calls.
4. **LogN16 path proof:** source-audit and demonstrate actual ordinary public Fast dispatch and exercised optimized Add/MulRelin/Rescale/Rotate with compact physical rows at LogN16. Record that special `prepareFastLogN13C2S` compression is deliberately LogN13-only; do not confuse that with entire Fast C2S/EvalMod/Bootstrap fallback. Previous LogN16 c0 diagnostic had output equality but source alone is not sufficient for in-circuit Fast runtime proof. Reuse Fast-only instrumentation / memory/physical state evidence. If runtime Fast public dispatch/primitive path cannot be established at zero calls, stop, and do NOT expend Bootstrap calls. If the full Fast Bootstrap optimized route can only be proved by actually running it, that evidence belongs to a **budgeted** Phase B call, not to a fake zero-call proof.
5. Targeted lightweight tests/vet/build/diff and fail-closed attempts; no `go test ./...` if legacy tests invoke cryptographic Bootstrap.

## Phase B: bounded performance, **transfer remaining 004 authorization — NOT an additional budget**

Exactly **zero** of the previously authorized 004 global maximum 16 full Bootstrap attempts have been reserved/used. This 005 task transfers ONLY that unused ceiling and does NOT grant another 16 attempts. Across 004+005, the ceiling remains **16 full Bootstrap calls total**, with durable journal reconciliation across Codex sessions. Do NOT run any call unless ALL Phase A gates pass.

- LogN16 new C15/16 native E32 profile: genuine Standard **1 cold + 5 warm**, Fast **1 cold + 5 warm**, = 12.
- LogN13 unchanged original frozen E32 control: genuine Standard **1 cold + 1 warm**, Fast **1 cold + 1 warm**, = 4.
- **Maximum 16 total**, no extra trace calls, no retries after failed call, hidden startup warmups, benchmark `testing.B` iterations invoking Bootstrap, implicit acceptance calls or new attempts from tests. Before each `Bootstrap`, reserve attempt under a single task-wide cumulative ledger, append reserved/finished outcome, and fail closed on missing or conflict. If a cold LogN16 call fails numerically or panics, do not proceed to its warm calls; immediately report and preserve consumed call count.
- Prefer existing `--public-bootstrap` (2 attempts) and `--public-repeatability` (6 attempts) modes or minimal reuse of existing call-budget/timer APIs, never build another stand-alone speed harness. Respect cold vs lazy initialization and measure key/eval-key/evaluator construction off the Bootstrap timer separately. Keep all actual wall-clock samples, per-call allocations and allocated bytes, comparable timing environment, `GOMAXPROCS`, Go version, GC and frequency/source build info. Report warm median and range with caveat; prior historical 4.225x LogN13 result is NOT current Fast SHA nor LogN16.
- Verify native encryption provenance and new-version workload for each timed sample. Compare output to same original **new benchmark message/oracle** and Standard-vs-Fast (complex RMSE, max error, SNR, complete slots, Level/Scale); the fixed `1e-6` public-native gate must not be loosened. Final `Bootstrap` must be paired through same source and metric helpers.
- Reuse existing Fast `fastdiag` stage/power/rescale counters/trace where possible *within* allocated calls, and separate traced overhead from uninstrumented latency. Any unmeasured in-circuit Standard stage or Fast trace is marked `UNAVAILABLE`, not approximated from unrelated replay or fabricated. Do not spend extra full Bootstrap calls for profiling beyond 16; the primary objective is correct current path and total latency.
- If LogN16 Standard/Fast outputs again bitwise equal, decide via recorded Fast dispatch/active core/profiling evidence and latency; do not equate zero RMSE to Standard fallback.

## Deliverables, safety and stop conditions

New task result artifacts, preserving originals:
- `results/FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005.md`
- `results/FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005.json`
- journal/other bounded artifacts as required by existing measurement platform. Do not commit full per-slot vectors/huge logs; include path/hash and small evidence only. Update `docs/MEASUREMENT_PLATFORM.md` only for reusable, demonstrated tool support.

Classify `CURRENT_LOGN16_FAST_SPEED_VERIFIED`, `FAST_PATH_PROVED_BUT_PERF_PARTIAL`, `CAPACITY_SAFE_PROFILE_BLOCKED`, or `PERFORMANCE_NOT_EXECUTED` with precise evidence. For unmodified original LogN16 fixture retain 004's `PUBLIC_NATIVE_LOGN16_PROFILE_BLOCKED` classification.

No Secondary modifications, no original Standard source changes, no cryptographic maths fix, no weakening gate, no q0 replacement, no secret/noise equivalence, no performance optimization in this task. Commit/push only authorized Primary source/tests/reports via normal safe fast-forward; READY_FOR_WEB_REVIEW or NEEDS_WEB_REVIEW at boundary.

**Do not ask user to run local Git commands.** The standing workflow is still `開始` (Codex) → independent Web `review`.

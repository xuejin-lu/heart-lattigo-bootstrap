# FAST-STANDARD-PERF-REBASELINE-002 — Matched LogN13 and LogN16 Speed + SNR

## Status and authority
This **supersedes** execution instructions of FAST-STANDARD-PERF-REBASELINE-001-LOGN13. Run one research campaign with **two mandatory independent profiles, LogN13 and LogN16**; neither profile is optional. The prior LogN13 spec is historical background only.

Task class: matched performance and numerical measurement, **no Secondary production changes**.

## Accepted preconditions
- Corrected Fast production SHA: `5117fc57949647182f476dc5952099c706b9f869` on `xuejin-lu/lattigo` fast-qprefix. Pin exactly for both profiles; don't silently use a different Fast build.
- Genuine Standard baseline SHA: `5dbffbdea05394de2ca3a432ed5318aa832e3f40` unless explicit incompatibility is demonstrated and Web review approves a new pinned genuine Standard SHA.
- Accepted canonical LogN13 q0=55 numerical: `FAST_STANDARD_NUMERICAL_CLOSE`, Fast-vs-Standard complex RMSE `1.44680895443e-10`, Standard/Fast Bootstrap SNR `144.713390 / 144.710750 dB`. These are **historical reference values**, not fresh results.
- Prior LogN16 speedup 6.671× belongs to **older pre-fix Fast** SHA `82601ea2517edc14784c9da250426169a1b221c7`; do not treat this as corrected Fast performance or current numerical correctness.

## Shared testing method for both profiles
Use the existing current-Primary numerical framework as the definitive **metric implementation**:
- `cmd/fastdiag/numerical.go` (numericalReference, output comparisons, precision bits, canonical input/parameter provenance, classification);
- `cmd/fastdiag/numerical_lockstep.go` and `numerical_evalmod.go` (stage alignment/capacity and EvalMod breakdown);
- `internal/numericalmetrics/snr.go` (decoded-domain SNR and nonfinite statuses).

The existing fastdiag numerical CLI currently hardcodes `p93-q55` with LogN13; **it is not already a LogN16 implementation**. Extend the **Primary-only** shared harness/CLI and parameterized helper functions for a new LogN16 profile, preserving the LogN13 reference and its existing tests. Do not duplicate SNR formulas or write a weaker ad-hoc LogN16 validator.

Use an identical, frozen, committed Primary harness source (same SHA) for the paired Fast/Standard timing runs **within each profile**; prefer the same new shared Primary harness for both profiles. The old `8186f50e...` timing harness can be consulted but is not the authoritative numerical suite. If the frozen harness cannot compile against the pinned genuine Standard without Fast imports, isolate backend-neutral measurement code shared by both in Primary and invoke the same code for each backend, without mutating the original calculation. Do not use dissimilar benchmark paths and claim matched speedup.

## Mandatory profiles (independent evidence)
### Profile A: LogN13
- LogN=13, LogSlots=12 (4096 slots), q0 bits=55, Q0123 authoritative prefix at applicable high levels.
- The numerical constructor `fastStandardP93Parameters()` and input `fastStandardP93Values()` (SHA256 `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`) are canonical **for this profile**. Do not silently substitute `configs/bootstrap_config.logN13.json`: that config may have different LogSlots/effective parameters.
- Recompute fresh numerical and timing results from corrected Fast. Retain historic values as regression references only.

### Profile B: LogN16
- LogN=16, generated full-slot LogSlots=15 (32768 slots) where valid.
- Start from `configs/bootstrap_config.logN16.json` and establish the **effective**, not just textual, parameter chain (including generated q0, Q and P), default scale, Mod1 degree=30, DoubleAngle=3, K=16, LogMessageRatio=10, ring dimensions, capacity width.
- Construct deterministic LogN16 input, compute and retain SHA256 on actual input vector and ciphertext/input metadata; feed mathematically identical ciphertext input to both backends. Report whether `c0=encoded-message, c1=0` and state this is a specialized zero-a comparator, not a normal encrypted-input security result.
- Build LogN16 stage-lockstep, real/imag EvalMod and PS/DoubleAngle numerical checkpoints **when mathematically semantically comparable**, with meaningful capacity checkpoints; if strict `2B<S_Q(Level)` fails, report first actual failing component/level and classify FAIL without weakening gates.
- Fresh LogN16 SNR is REQUIRED and may not be replaced by LogN13's SNR or old LogN16 timing metadata.

## Exactly defined SNR and accuracy
For each profile, for each backend, use existing `numericalmetrics.Compare(preDecoded, postDecoded)`:
`signal_power=mean(|pre_i|²)`,
`error_power=mean(|post_i-pre_i|²)`,
`SNR_dB=10*log10(signal_power/error_power)` if mathematically finite.
- Standard: genuine Standard generated secret and evaluation keys and **decrypt+decode** both pre- and post-Bootstrap.
- Fast: preserve current zero-a/error-retaining semantics, and decode by the existing correct Fast method. Do not equate this SNR with RLWE security noise or security strength.
- Keep finite/infinite/undefined/not-comparable statuses, signal/noise power, noise RMSE; never fabricate a floor.
- Run at least 2 Fast executions and 3 independent genuine Standard keys/trials **per profile** for numerical measurements, if existing canonical protocol is applicable; otherwise explain a concrete blocker and do not claim complete verification.
- Include Fast/Standard original-message complex RMSE and median precision bits; Fast-vs-Standard complex RMSE, max complex diff, coordinate threshold audit; final output metadata/contract; numerical classification.
- Show checkpoint stage-reference SNR and topology-aware delta SNR with genuine Standard comparisons, or explicitly flag unsupported checkpoints and explain why. Do not misrepresent partial stage coverage as full success.

## Timing and measurement boundaries
For **each profile and each backend**:
- On same machine, same Go version/config/parameter literals, input fingerprint and harness; recorded backend SHA and public constructor dispatch.
- Exclude key generation, eval-key generation, setup, warmup, encoding, decoding, SNR/RMSE calculation, reporting from the timed region. Timer measures full public Bootstrap; stage timing may be measured separately.
- Warmup >=1, repetitions >=7 per backend where feasible; preserve raw samples, median/mean/min/max ms, allocations B/op/allocs/op if available. Do not selectively report best sample.
- Time all standard stages (including real/imag EvalMod, C2S/S2C) where supported; record unavailable stages honestly. Total Bootstrap time is separately measured; not a sum of per-stage medians.
- Matching **effective** Q/P, LogSlots, degree, scale, input and environment is mandatory per Standard/Fast pair. Numerical and timed workload parameters/input SHA must also match *within profile* for a combined speed+SNR conclusion; if not, use a unified canonical profile harness to obtain matched results, or mark BLOCKED/UNMATCHED.
- Worktree hygiene: detached sibling Primary/Secondary pairs, `go.mod replace => ../lattigo`, no destructive operations on authoritative worktrees, verify each exact SHA and branch identity; commit/push Primary-only harness + result changes. Keep pinned Secondary code unchanged.
- Do not hide the distinction between genuine Standard encrypted/decrypted data and current Fast zero-a experimental semantics. A matched public operation timing comparison here is not a security-equivalent crypto-system comparison.

## Output artifacts — BOTH are required
For each profile `logN13` and `logN16`:
- `results/FAST-STANDARD-PERF-REBASELINE-002-<profile>-standard-timing.json`
- `results/FAST-STANDARD-PERF-REBASELINE-002-<profile>-fast-timing.json`
- `results/FAST-STANDARD-PERF-REBASELINE-002-<profile>-numerical.json`
- `results/FAST-STANDARD-PERF-REBASELINE-002-<profile>-report.md`

Also:
- `results/FAST-STANDARD-PERF-REBASELINE-002-comparison.md` with separate rows for LogN13 and LogN16; columns Standard/Fast full Bootstrap median ms, speedup, Standard/Fast SNR dB, Fast−Standard SNR delta, Standard/Fast original RMSE, Fast-vs-Standard RMSE, precision bits, and allocation metrics.
- Each individual report must have real/imag stage diagnostics, effective parameters, SHA provenance, threshold/capacity status, source/security caveats, and whether full aligned stage coverage is available.
- Record historical LogN16 6.671× only under a visibly separate `old Fast / not comparable` section.

## Exit criteria
- `DUAL_LOGN13_LOGN16_PERF_SNR_READY` only if **both** profiles have matched timing and fresh numerics including finite or mathematically explicit SNR statuses, comparable outputs, all specified numerical gates, full artifacts and recorded provenance; no substitutions.
- `DUAL_LOGN13_LOGN16_PERF_SNR_PARTIAL` if one profile completed but the other blocked; preserve the completed profile and clearly identify blocker; never call dual READY.
- `DUAL_LOGN13_LOGN16_PERF_SNR_NUMERICAL_FAIL` if any profile exhibits real numerical/capacity failure; report first concrete failing gate; don't quietly omit it.
- `DUAL_LOGN13_LOGN16_PERF_SNR_BLOCKED` if neither profile could be matched/run.
- Do not repair production or silently relax thresholds to force READY.

Run `go test ./cmd/fastdiag ./internal/numericalmetrics`, appropriate full tests where feasible, and `git diff --check`. Return concise Chinese table, exact commits, status, then `READY_FOR_WEB_REVIEW`.

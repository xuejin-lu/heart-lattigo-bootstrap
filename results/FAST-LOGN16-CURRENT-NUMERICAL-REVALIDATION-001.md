# FAST-LOGN16-CURRENT-NUMERICAL-REVALIDATION-001 — blocked

**Classification:** `LOGN16_CURRENT_REVALIDATION_BLOCKED`

**Generated:** 2026-10-10 21:35:47 UTC

**Primary source HEAD before this report:** `9c222531713cc46f192ae3acb43b16fc11091de5`

**Secondary:** clean `fast-qprefix` at `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`

**Requested genuine Standard pin:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`

No current LogN13 or LogN16 numerical RMSE/SNR result was produced. The LogN16 attempt panicked before output serialization, and the requested genuine Standard pin was not the code linked into that run. Historical RMSE values are not reused as current evidence.

## Provenance and matched input

Primary `go.mod` replaces `github.com/tuneinsight/lattigo/v6` with `../lattigo`, resolving to the current Fast checkout. The numerical harness records `IndependentOriginalStandard=false`. Fast and the in-tree Standard API were given the same LogN16 config, deterministic input vector, and encoded `c0`/zero-`c1` diagnostic ciphertext; this is the historical diagnostic profile, not the formal native E32 public lifecycle.

The preflight used LogN16 / LogSlots 15, config SHA-256 `16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02`, and the source-locked input fingerprint `659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf`. The generated Standard secret had `LevelP=4`; `GenEvaluationKeys` therefore took its P-capable Standard-key path rather than the Fast-compatible key-generation condition (`LevelP < 0`). The ciphertext began at Level 0 with one full 65,536-coefficient row per component, `IsNTT=true`, `IsMontgomery=false`; ModUp produced Level 16 with 17 full rows per component.

## LogN16 panic diagnosis

The previous bounded full-run attempt made two Fast Bootstrap calls and one Bootstrap call through the in-tree Standard API; that Standard attempt is **not** an independent original-Standard result. This investigation made zero additional Bootstrap calls. A no-Bootstrap preflight executed through ModUp and individual C2S matrices and reproduced the failure:

| Checkpoint | Level / Scale | Backing evidence |
|---|---|---|
| Input / Pack / ScaleDown | Level 0 / `2^45` | Both components have one full row of length 65,536; `c1` is zero-valued but fully allocated. |
| ModUp, before C2S | Level 16 / `2^50` | Both components have 17 full rows of length 65,536. |
| C2S matrix 0 | Level 16 | Completes; both components still have full rows. |
| C2S group 0 Rescale | Level 15 / `2^50` | Both components retain q0–q3 at length 65,536, but q4–q15 have length 0 while logical Level remains 15. |
| C2S matrix 1 | Level 15 | `DecomposeNTT` reaches `ring.INTT` with an empty active row and panics: `len(p1)=0`, `len(p2)=65536`, `len(roots)=65536`, `N=65536`. |

The source cause is in Secondary: `schemes/ckks/params.go` defines `FastCKKSZeroSecretSimulation()` on the general `ckks.Parameters` type. Consequently the provider check in `schemes/ckks/evaluator_fast_zero_secret.go` is true for the ordinary Standard CKKS evaluator too. Its `Rescale` takes the Fast Q-prefix `ApplyRows` path and compacts to four rows. The next Standard DFT linear transform uses `lintrans.Evaluator.EvaluateMany` → `DecomposeNTT` on `c1`, which requires populated rows across its active logical Q level; q4 is the first missing row.

The outer bootstrapping evaluator did not select its `FastEvaluator` field: the P-capable secret selected ordinary key generation. The contamination occurs lower down, in the shared CKKS evaluator's mode test and Rescale dispatch. Thus the panic is reproducibly caused by Fast Q-prefix compaction leaking into the in-tree Standard API path; it does **not** establish that the separately pinned genuine Standard implementation panics.

Fixing this requires a Secondary backend/mode-dispatch change. No Secondary code was changed because this task forbids it. The current run cannot serve as a genuine Standard comparison, so numerical validation stops here pending Web direction and an authorized backend correction or a correctly isolated Standard execution path.

## Fresh results and budgets

| Profile | Status | New Fast/Standard RMSE, EvalMod RMSE, SNR |
|---|---|---|
| LogN13 control | Not run | Not available |
| LogN16 | Blocked at in-tree Standard C2S | Not available; no result document was emitted by the failed full run |

Prior obsolete-run values in the task spec (`RMSE 0.02048549798`, EvalMod real RMSE `0.00256939892`, Secondary `5117fc57949647182f476dc5952099c706b9f869`) remain historical only and are not promoted or compared as current measurements. No numerical gate was relaxed. The `1e-7` task gates were not evaluated because no comparable genuine-Standard output exists.

## Cleanup and validation

Removed `fix001_p3_diag_p93_genuine_standard_vs_matched_reconciliation_runner_test.go`, which contained the three retired historical P93 tests, including `TestFIX001P3GenuineStandardPublicVsStagedConsistency`. No Q0123 CRT, Rescale, NTT, Chebyshev, or Bootstrap correctness tests were removed. The runner itself remains because `main.go` still exposes it through a CLI flag and the standard-build stub still provides its matching entry point; it is not currently unreferenced code.

- Passed: `go test ./cmd/fastdiag ./internal/perfmeasure ./internal/numericalmetrics -count=1`.
- Passed: `go test ./... -run '^$'` (compile-only; no test bodies or Bootstrap calls).
- Passed: `git diff --check` before final commit.
- `go vet ./...` reports the unrelated existing self-assignment `inputCheckpoint.FastQ01Safe = inputCheckpoint.FastQ01Safe` at `fix001_p3_diag_logn13_evalmod_matched_normalized_oracle_runner.go:649`; `git blame` attributes it to pre-existing commit `365b2bbb`. It is outside this task and was not changed.
- Full `go test ./...` was not run because the suite includes Bootstrap-executing tests and the LogN16 Bootstrap allowance is exhausted.
- No Secondary source, tests, configuration, or Git refs were changed.

**Disposition:** `NEEDS_WEB_REVIEW`. Do not start genuine LogN13/LogN16 numerical validation until the Standard/Fast mode-dispatch contamination is resolved under an authorized scope. No fresh RMSE/SNR claim is available.

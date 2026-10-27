# BTS-FAST-NOISE-FIDELITY-009 journal

Status: `CP_AB_IMPLEMENTED_ZERO_BOOTSTRAP_PREFLIGHT_PENDING`

## Authorization and separation

- Authorized by the user-provided 009 task specification on 2026-10-27. This is a new, independent 18-call budget.
- Prior Primary pointer was `FAST-BOOTSTRAP-NOISE-FIDELITY-003-LOGN16`; its original 18/18 Bootstrap calls remain exhausted and untouched. Its spec permits a separately authorized new task and budget.
- BTS-008 remains historical: its Standard/Fast captures and results are not 009 samples.
- No 009 Bootstrap has been reserved or executed.

## Initial synchronized source state

- Primary `main`: `3df02885e9ab99f8760f2319a7577fba5f338ad6`, clean and equal to fetched `origin/main`.
- Genuine Standard pin: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Secondary `experiment/fast-bootstrap-noise-fidelity`: `325e501709c42ab5995bf7f36fb8d19621d34131`, clean and equal to fetched origin. `fast-qprefix` is not used or modified.
- Original `external/BTS` source manifest SHA-256 verified as `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`; original BTS files are not modified.

## Frozen experiment profile

Use the exact BTS-008 effective Q/P primes and profile, LogN16 / E32, input L3 / Scale 2^32, Mod1 L26 / Scale 2^42, output L18 / Scale 2^32, `BTSModePlain`, seed 0, amplitude 1, σ `4.004280012097432`, root seed `20261027`. The authoritative full Q/P arrays are retained in the BTS-008 Standard/Fast run manifests; no regenerated profile is accepted.

## Planned gates

1. Minimal Secondary hook wiring into Fast public CKKS ciphertext×ciphertext `Mul` and `MulRelin` dispatch, default off; preserve wrappers and arithmetic semantics.
2. Targeted zero-Bootstrap tests at the BTS effective Q/P, Levels and Scales; N13–N16 regression tests.
3. Extend the existing `cmd/perfprobe` BTS adapter/report and shared `internal/perfmeasure` / `internal/numericalmetrics` owners only for evidenced capability gaps.
4. Pin final Fast source, verify provenance and shared Fast OFF/Noise binary, perform zero-call preflight, and commit/push the plan before any Bootstrap.
5. Only after all gates pass, run at most 6 calls per arm (18 total), no warmup or retry, recording each reservation and outcome.

No tests or Bootstrap calls were run during this initial checkpoint.

## Resume and implementation checkpoint (2026-10-27)

- Primary startup check: branch `main`; fetched `origin/main`; the authorized task commit `85989a005cad0e88d852f032a0d08321c22efe1f` equals `origin/main` before implementation. The earlier `3df02885e9ab99f8760f2319a7577fba5f338ad6` remains the pre-authorization source point recorded above.
- A normal `git fetch` was denied writing `.git/FETCH_HEAD`; the single formal permission-escalation retry succeeded for each repository. No working tree files were overwritten.
- Secondary experimental worktree: branch `experiment/fast-bootstrap-noise-fidelity`, clean and equal to fetched origin at `be9c42ff23662300769b77f1eead3f12fc53adef`. Genuine Standard checkout remains clean detached at `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. The separate `/Users/xuejinlu/Developer/xuejin-lu/lattigo` checkout remains on `fast-qprefix` at `1e7828293c36f9845e30723ba4e9500d84bb50d2`, unchanged.
- Primary implementation commits: `369ce3a84eb072392bacf2359d651413c159e5e8` and `42b9840a6ea42bb6d9f4ba368b7c161617d4a02d`. They reuse `cmd/perfprobe`'s BTS overlay/build/report, `internal/perfmeasure`, and `internal/numericalmetrics`; add explicit BTS-009 Noise mode, repeated-sample allocation capture, global and slot-median precision/tails, and a fail-closed three-arm comparator; and move Hook-state verification before the zero-call preflight return while persisting Hook supported/enabled state. Batch008's legacy fields and timer remain unchanged when the new fields are unset/task is 008.
- Comparator self-review found that trusting only `strict_fit=true` did not independently bind the reported prefix product to the frozen Q chain. The 009 validator now recomputes the exact first-`rows` Q-prime product and checks `2*MaxAbs < prefixProduct`; regression tests reject fabricated products and out-of-capacity bounds.

## CP-A source map and integration analysis

The historical BTS source manifest remains the frozen value `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`; the original `external/BTS/` source was not edited.

| BTS call site | Public operation | Hook expectation |
|---|---|---|
| `external/BTS/evalmod.go:868` | `MulNew` in lazy generated power | Delegates once to `Mul`; one event per actual ct×ct multiply |
| `external/BTS/evalmod.go:882` | `MulRelinNew` in Chebyshev power basis | Delegates once to `MulRelin`; one event |
| `external/BTS/evalmod.go:1063` | non-relinearized ct×ct `Mul` in PS giant-step monomial composition | One event, retaining degree-2/non-relinearized output semantics |
| `external/BTS/evalmod.go:439` | in-place `MulRelin(res,res,res)` in each double-angle round | One event per square; Secondary alias-safe tests verify source operands are copied before injection |

Other inspected `Mul` sites multiply by scalar or plaintext (including `external/BTS/myc2s.go:317` and `external/BTS/my_relu.go:160`) and are outside the ct×ct hook dispatch. `MulThenAdd` coefficient accumulation is also not a ciphertext-pair multiplication. No rotation or standalone Rescale dispatch is routed through the hook. The new Primary adapter explicitly enables the Hook only on Fast+Noise's `bts.mod1.eval`; Standard and Fast OFF remain disabled. Hook default is off. No BTS multiplication algorithm or public application operation was rewritten.

The enabled evaluator reuses the existing parent CKKS signed-Gaussian sampler, RNS/NTT/Montgomery conversion, strict Q-prefix capacity checker, seed derivation, coefficient SHA and event schema. It injects into independent copied c0 operands before the targeted multiply, rather than mutating BTS's input ciphertext. The library hook is generic over valid effective CKKS parameters; LogN16/E32 and the exact Q/P remain Primary task-profile constraints.

## CP-B coding and targeted-test evidence

No Bootstrap was called. The following targeted tests passed after the bounded self-review repair:

- Primary: `GOWORK=off GOCACHE=/private/tmp/go-cache-primary009 go test -modfile=/private/tmp/bts009-primary.mod -tags=perf_fast ./cmd/perfprobe ./internal/perfmeasure ./internal/numericalmetrics -run 'TestBTS|TestSummarizeBTSRepeatedSamples|TestCompareRealVectorQuality|TestCompareRealVectors' -count=1`.
- Secondary: `GOCACHE=/private/tmp/go-cache-bts009 go test ./schemes/ckks -run 'TestFixedSigmaBootstrap(MulNoise|NoisePublic)' -count=1`.
- Secondary: `GOCACHE=/private/tmp/go-cache-bts009 go test ./schemes/ckks/fast -run 'TestFixedSigmaBootstrapNoise|TestFast.*Noise' -count=1`.
- The Secondary selections include positive N13–N16 generic Montgomery/NTT tests, actual BTS Q/P LogN16 Level26/Scale2^42 helper coverage, public `Mul`/`MulNew`/`MulRelin`/`MulRelinNew` single-event checks, square alias safety, OFF default/bit-repeatability, non-ct×ct exclusion, strict capacity failure and exact coefficient reconstruction.
- `git diff --check` and `scripts/test_task_worktree_policy.py` passed. The Primary task-path guard found only the ten expected implementation/test files before this code commit.

## Remaining gates and budget

- Current 009 ledger remains 18 declared; reserved/entered/completed/failed = `0/0/0/0`. No formal Bootstrap or attempt reservation has occurred.
- Still required: build clean pinned Standard and Fast binaries from the refreshed clean Primary plan-checkpoint revision (which includes measurement commits `369ce3a84eb072392bacf2359d651413c159e5e8` and `42b9840a6ea42bb6d9f4ba368b7c161617d4a02d`); verify exact VCS/module provenance and SHA; execute exactly three zero-call preflights; commit/push those manifests and update the ledger while counters remain zero. Subsequent evidence-only commits may not change linked Go source.
- Only after that checkpoint passes will formal calls begin in order Fast+Noise (FN-01..06), Standard (S-01..06), Fast OFF (F-01..06), with six calls maximum per arm, one cold plus five warm, no retry/warmup. Any failure stops the sequence and permanently consumes reservations already made.
- Formal vectors and evidence will be saved in this task's result tree. No LogN16 work beyond this authorized original BTS task, no extra benchmark, no BTS/003 budget reuse, and no modification to `fast-qprefix`.

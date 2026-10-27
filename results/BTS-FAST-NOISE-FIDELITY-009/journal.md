# BTS-FAST-NOISE-FIDELITY-009 journal

Status: `PREFLIGHT_COMPLETE_AWAITING_VERSIONED_CHECKPOINT`

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
- Primary implementation commits: `369ce3a84eb072392bacf2359d651413c159e5e8`, `42b9840a6ea42bb6d9f4ba368b7c161617d4a02d`, and `dd077ac69e81192ffaa4b606d305e545d8b50202`. They reuse `cmd/perfprobe`'s BTS overlay/build/report, `internal/perfmeasure`, and `internal/numericalmetrics`; add explicit BTS-009 Noise mode, repeated-sample allocation capture, global and slot-median precision/tails, and a fail-closed three-arm comparator; and move Hook-state verification before the zero-call preflight return. Hook supported/enabled are task-only `*bool` fields so both `true` and explicit `false` are serialized while 008's zero-value JSON remains unchanged.
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

## Initial zero-call preflight set (superseded evidence; 2026-10-27)

- Built clean binaries against Primary VCS `fcedd1779cd071d59384dcfcc3f90f327f553bbc`: Standard SHA-256 `6bbba6a0d9208fe6572bb75294adfd5583e6846e3ba553e1d44e7131e8367b55`; Fast SHA-256 `643ee2bff0368504c6088593d8c7afbee8230e1bc5dabcb07ae605c047d9aca8`. Both manifests had `vcs.modified=false`, correct backend pins, the original BTS manifest, and identical Primary source/adapter hashes.
- Three `BTS_PREFLIGHT_ONLY=1` invocations completed successfully and each recorded `bootstrap_reservations=0`; all used input SHA-256 `5ca151755bd8232e0b581939fb1b2c079fa4451b454c0e1b0ad9f2b1a6198373`. The Fast+Noise preflight explicitly recorded supported/enabled/σ/root as true/true/`4.004280012097432`/`20261027`.
- Retained first build manifest SHA-256s: Standard `0d76fc06d1878e2dc8e9229135697871e6d451278165608ff0f6fb182169b909`; Fast `f4a83113fc3556cfe00bb430eb7b0f5835d36f531a2bdfe29fb0ca8a3c873354`. Retained preflight run manifest SHA-256s: Standard `34f74a2c4061de2af7a13689298274439c7af1733a8b19f061017782a3c24d1b`; Fast OFF `adc14a46a2fa67d68dec3d1d454f1714c294614d56dceb95c072b232a4e73214`; Fast+Noise `774a83907f09d585f86be09d3bc2efb68804049a2b535b62f317448b87dcd66f`. Each raw deterministic input copy is 262144 bytes with the same input SHA above; no attempt sidecar exists in any preflight directory.
- Review found Go `omitempty` removed false Hook booleans from Standard and Fast-OFF JSON. Although the code had checked those states and all three runs exited with `ZERO_BOOTSTRAP_PREFLIGHT_OK`, the artifact did not explicitly preserve false. These first preflight manifests and build artifacts are retained under `preflight/` and `builds/`; they are not accepted as the final preflight gate and will not be overwritten.
- Follow-up commit `dd077ac69e81192ffaa4b606d305e545d8b50202` switches the task-specific fields to optional pointers and the report fails closed if either boolean is absent. Targeted Primary tests passed after this fix. Rebuild and rerun exactly three zero-call preflights under new `preflight-v2/` paths before formal calls. The task budget remains `0 reserved / 0 entered / 0 completed / 0 failed`.

## Remaining gates and budget

- Current 009 ledger remains 18 declared; reserved/entered/completed/failed = `0/0/0/0`. No formal Bootstrap or attempt reservation has occurred.
- Clean pinned Standard/Fast binaries and `preflight-v2/` checks have now passed; their compact provenance, hook-state, and input evidence are recorded below. Still required before any Bootstrap: commit/push these final manifests and preflight records, set the formal run paths/ledger links, and confirm the ledger remains at `0/0/0/0`. Subsequent evidence-only commits may not change linked Go source.
- Only after that checkpoint passes will formal calls begin in order Fast+Noise (FN-01..06), Standard (S-01..06), Fast OFF (F-01..06), with six calls maximum per arm, one cold plus five warm, no retry/warmup. Any failure stops the sequence and permanently consumes reservations already made.
- Formal vectors and evidence will be saved in this task's result tree. No LogN16 work beyond this authorized original BTS task, no extra benchmark, no BTS/003 budget reuse, and no modification to `fast-qprefix`.

## Accepted source-linked build and zero-call preflight v2 (2026-10-27)

- Primary build VCS revision: `14a76be8ebb05da9f19d82d78f5d8003f3be8c3c`; both build manifests report `primary_dirty=false`, `vcs.modified=false`, Go `go1.26.4`, Darwin/arm64, and the same task adapter SHA-256 `42d7745e75663b886e7eb2e644262095c852a8b0b6f2b32f7176d068a99f01d6`.
- Genuine Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, Binary SHA-256 `6502bdc8e8b52c25c6a7afdcd9cb82bb3d09b0ba4c226f697c50e18cacaf97bc`, build-manifest SHA-256 `acf659b94e10521c5bd7d81a839e720d181e3ab4a37a9b4ec917267814846246`.
- Fast pin `be9c42ff23662300769b77f1eead3f12fc53adef`, Binary SHA-256 `df4b9c9b0dbccb370d081d239f4d0a986c769bdbd077bf83669f93a978de883e`, build-manifest SHA-256 `b8658192274ca5b36666057b7e8b3f846789a3d4984d4810a6e1583f2f4064b0`. Both Fast OFF and Fast+Noise use this exact same binary.
- Standard/OFF, Fast/OFF and Fast/ON `preflight-v2` run manifests all returned `ZERO_BOOTSTRAP_PREFLIGHT_OK`; each has zero reservations and zero samples. Their run-manifest SHA-256 values are respectively `ff8ab3dcbd95a9590d417aee526307c2395fde62fde39d8107aafd0be7e75685`, `56a1e94cf1145705943b26a70992a5b84450d16828ff99eac2c323af22c4f852`, and `f0f0ef08a72571487e48306b83b487d05f13742cffdeff414d3776c821d60250`.
- All three have exact Q/P and effective profile equal to the BTS-008 genuine Standard profile; same profile SHA-256 `229cb5a28732b755ec66a531a3b643630dee703a2e0d35398f2986a374450b47`, input SHA-256 `5ca151755bd8232e0b581939fb1b2c079fa4451b454c0e1b0ad9f2b1a6198373`, and matching environment Go/Darwin/arm64/Apple M4/NumCPU 10/GOMAXPROCS 8. Each raw input is 262144 bytes and hashes to the same input SHA.
- Explicit Hook snapshots: Standard/OFF `{supported:false, enabled:false}`, Fast/OFF `{true,false}`, Fast+Noise `{true,true, sigma:4.004280012097432, root_seed:20261027}`. All run manifests record the same input and correct source and binary pins. No attempt sidecar exists; initial Noise event counters are zero by preflight check.
- Preflight NewBTS construction time (not Bootstrap time): Standard `39.624367583s`, Fast OFF `43.008429083s`, Fast+Noise `42.768094666s`.
- At this checkpoint the durable formal ledger still has `reserved/entered/completed/failed=0/0/0/0`. The only three preflight invocations are non-Bootstrap lifecycle checks and do not consume the 18-call budget.

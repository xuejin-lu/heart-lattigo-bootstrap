# FAST-LOGN16-BACKEND-PATH-AND-PERFORMANCE-004

**Classification:** `CURRENT_PERFORMANCE_BLOCKED`
**Handoff:** `NEEDS_WEB_REVIEW`
**Specific blocker:** `PUBLIC_NATIVE_LOGN16_PROFILE_BLOCKED`

The shared `cmd/perfprobe` now accepts the pinned current Fast backend and validates both frozen E32 profile identities. The canonical LogN13/E32 zero-Bootstrap public-native comparison passed. The unchanged LogN16/E32 workload fails the existing terminal-q0 capacity gate in both backend builds, before key/evaluator construction or any CKKS operation. Therefore Phase C was not entered: this task used **0 of 16** newly authorized Bootstrap calls and produced no current LogN16 performance, Bootstrap output, allocation, or Bootstrap numerical result.

## Provenance and safety

| Item | Value |
|---|---|
| Primary measurement source | `b5fe32d4be57a507f656c6a28ceac973bcc2ba12` |
| Measurement-source SHA-256 | `0339ac39c74cbd8cce6feb62dd60089546d0d2581acea35d9add991c67c5da35` |
| Genuine Standard Lattigo | `5dbffbdea05394de2ca3a432ed5318aa832e3f40` |
| Current Fast Lattigo | `fast-qprefix@2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` |
| Secondary state | clean; no Secondary source changed |
| Build isolation | same Primary source, separate clean pinned Lattigo checkouts, isolated temporary modfiles, `GOWORK=off` |
| Environment | Go 1.26.4, darwin/arm64, Apple M4, 10 CPUs, `GOMAXPROCS=10` |
| New Bootstrap calls | `0/16`; no reservations were created |

The LogN13 raw public-native documents and decoded vectors, and the compact compare artifact, are retained outside the repository under `/private/tmp/fast-logn16-004/preflight/`. The compact paired artifact SHA-256 is `4c22b173002a5f94397142e249e0d0488b943b86737bc71941c7dc446c3abb1a`. The task-owned JSON below preserves its aggregate checkpoint evidence; multi-megabyte vectors are not committed.

## Phase A — source path and kernel attribution

| Boundary | Source-grounded behavior | LogN13 | LogN16 |
|---|---|---|---|
| Public key/evaluator selection | `Parameters.GenEvaluationKeys` marks eligible zero-secret key material Fast-compatible; ordinary `bootstrapping.NewEvaluator` installs its Fast evaluator when that marker is present. The Standard build separately verifies that the genuine Standard evaluator exists. | Standard and Fast dispatch were observed in the paired preflight. | Source path is present, but the capacity gate prevented runtime construction. |
| Bootstrap method | The ordinary public `Evaluator.Bootstrap/BootstrapMany` forwards to its `fast` field when present; this is not a frontend Fast constructor or Standard fallback. | Source-grounded. No Bootstrap called. | Source-grounded only; not executed in this preflight. |
| Fast DFT / C2S / S2C | Fast circuit initialization chooses `dft.FastEvaluator`; DFT factors call `LinearTransformQPrefixRows` and `RescaleQPrefixRows`. | Not executed: bootstrap-circuit initialization is lazy. | Same generic Fast path exists in source; not executed. |
| C2S compression | `matchesFastLogN13C2SProfile` requires `LogN()==13` and the named LogN13 schedule/Q profile. | Source predicts active for the canonical LogN13 config; runtime `C2SCompressionActive` was not observed. The existing no-Bootstrap Secondary preparation test passed on its explicit LogN13 test fixture. | Source predicts inactive because the first guard rejects `LogN()!=13`; runtime field not observed. |
| Fast EvalMod | `FastEvaluator.EvalMod` calls `EvaluateNewQPrefixRows` and restores the configured public scale. | Not executed. | Not executed. |
| Shared Fast CKKS core | Ordinary `ckks.NewEvaluator` creates shared Fast workspaces for the zero-secret mode. Public ciphertext Add/MulRelin, Rescale, and Rotate route to `fastcore.ApplyRows` with `QPrefixWidth(level)=min(level+1,4)`. | Add, MulRelin, Rescale, and Rotate were exercised through the ordinary public API; the output backing confirms the Q-prefix contract. | Source path is profile-independent, but runtime primitive execution was blocked before evaluator construction. |

This separates the source-verified LogN16 dispatch route from runtime proof: this run does **not** claim that the LogN16 Bootstrap kernels executed. It also does not classify the path as missing; the zero-call gate stopped before it could be exercised. No Secondary production code or instrumentation was added.

The current Fast zero-call LogN13 artifact records the ordinary `bootstrapping.NewEvaluator` wrapper dispatch and `c1=0` under the pinned zero-secret mode. Its native encryptions initially carry full backing; the subsequent public operations materialize only authorized rows: active width 4 at Level 5/4 and width 1 at Level 0. DFT/EvalMod/S2C spans and in-circuit timings remain unavailable without a permitted Bootstrap.

## Phase B — matched public-native preflight

### LogN13/E32: pass, zero Bootstrap

Both builds used the same 4,096-slot workload, config, Q/P, and Primary measurement source. Standard used native generated-secret encryption/decryption. Fast used the pinned zero-secret mode; it was not represented as secure-key equivalence. Both produced all eight finite canonical checkpoints and matched Level/Scale/Degree. The fixed maximum-complex gate was `1e-6`; every oracle and pair checkpoint passed.

| Checkpoint | Level | Scale (log2) | Degree | Q-prefix rows | Fast-vs-Standard complex RMSE | Max complex difference |
|---|---:|---:|---:|---:|---:|---:|
| `encrypt_a_level5` | 5 | 45 | 1 | 4 | `8.268373061552182e-12` | `2.2029381053639323e-11` |
| `encrypt_b_level5` | 5 | 45 | 1 | 4 | `8.363954287958416e-12` | `2.3785862667163578e-11` |
| `encrypt_c_q5_level5` | 5 | ≈60 | 1 | 4 | `2.5338463741002153e-16` | `6.938742113942858e-16` |
| `add_level5` | 5 | 45 | 1 | 4 | `1.170284374242679e-11` | `3.297358722995745e-11` |
| `mulrelin_level5` | 5 | ≈105 | 1 | 4 | `2.394140088465216e-13` | `8.961590950557055e-13` |
| `rescale_q5_level4` | 4 | 45 | 1 | 4 | `1.0401957370102809e-11` | `5.032792551822055e-11` |
| `rotate_level4` | 4 | 45 | 1 | 4 | `1.467903563887731e-11` | `5.930347959824538e-11` |
| `drop_level0` | 0 | 45 | 1 | 1 | `1.467903563887731e-11` | `5.930347959824538e-11` |

The worst pair difference is at `rotate_level4` and propagates unchanged to `drop_level0`; both remain well below the fixed gate. Fast output rows are compact after the initial full-RNS encryptions: Add/MulRelin use four active rows; Rescale/Rotate remain width four at Level 4; DropLevel0 contains one row. All eight self-oracles passed as well.

### LogN16/E32: blocked before backend construction

Both isolated `perf_standard` and `perf_fast` runs, on the exact canonical config/input/QP hashes and the same source, stopped with the same pre-operation error:

```text
terminal Level0 q0 capacity failure: B=18169085704944493 q0=36028797019488257 deficit=309374390400729
```

The Q0123 bound check passes; the fixed terminal centered-capacity requirement `2B < q0` then fails by `309374390400729`. The runner returns before secret/evaluation-key generation, evaluator construction, ciphertext operations, and Bootstrap. This is a real failure of the current formal public-native preflight gate, not a Standard panic or a Fast-vs-Standard numerical mismatch. The exact inputs, q0, and threshold were not altered to make the profile pass.

The resulting classification is `PUBLIC_NATIVE_LOGN16_PROFILE_BLOCKED`. Do not infer LogN16 latency, numerical parity, or runtime dispatch from source alone. Resolving whether this workload/profile needs a separately authorized mathematical or workload decision is for Web review.

## Prior evidence and interpretation

- `FAST-LOGN16-CURRENT-NUMERICAL-REVALIDATION-001` describes a previous purported Standard lane linked against the Fast checkout; it is not a genuine-original-Standard panic result and supplies no fresh RMSE/SNR for this task.
- `FAST-DUAL-LOGN16-EXACT-ZERO-EVIDENCE-AUDIT-003` confirms exact equality for the prior `diagnostic-encoded-c0-c1-zero-v1` run. It used a manually encoded `c0` and `c1=0`, not native Standard E32 encryption. Its zero terminal RMSE does not explain the equality and is not reused as current quality evidence.
- Batch023's historical `4.225x` LogN13 warm ratio used Fast `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, not current Fast `2cb8bc...`; it is not a current timing result. No old Bootstrap allowance or timing is reused.

## Validation and disposition

Passed:

- `GOCACHE=/private/tmp/fast-logn16-004-gocache go test ./cmd/perfprobe -count=1`
- `GOCACHE=/private/tmp/fast-logn16-004-gocache go vet ./cmd/perfprobe`
- The same package tests and vet with isolated `perf_standard` and `perf_fast` modfiles/build tags.
- Isolated Secondary tests `TestFastLogN13C2SCompressionPreparation` and `TestFastLogN13C2SCompressionGuardRejectsChangedProfile`; no Secondary files changed.
- `git diff --check`

**Phase C was not run** because Phase B did not pass for LogN16. No full Bootstrap was invoked; the new task budget remains `0/16`. Historical Standard/Fast output, timing, or allocation figures were not promoted.

**Disposition:** `NEEDS_WEB_REVIEW`. The bounded reusable runner extension and LogN13 zero-call evidence are complete. The canonical LogN16 terminal-q0 capacity deficit must be resolved or explicitly scoped before any Bootstrap timing/numerical campaign can proceed.

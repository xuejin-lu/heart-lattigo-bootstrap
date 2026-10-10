# FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019

**Result:** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`

**Scope:** One frozen LogN13/E32 public `Level-5 Add → MulRelin → Rescale(q5) → Rotate → DropLevel(4) → Bootstrap` chain. This is not a speed or security-equivalence claim.

## Decision and Scale closure

With both multiplicands at the frozen default Scale `2^45`, Rescale by the actual logical `q5=1152921504606830593` would yield `2^90/q5` (`log2 = 30.00000000000002`), not the required residual Scale `2^45`. The accepted Batch 017 precedent allows a per-input CKKS Scale for C without changing the config or default: A and B stay at `2^45`; C is ordinarily encoded/encrypted through the same public `EncryptNew` API at exact `Scale=q5`, identically in both backends. Consequently:

`(2^45 × q5) / q5 = 2^45`

MulRelin Scale was `40564819207302764422326571237376`; the public Rescale returned exact `2^45`. This preserves the frozen Q/P, E32, workload, frontend and default Scale. Both backends passed the native pre-Bootstrap oracle, so the per-input Scale choice was not a scale-only algebraic workaround concealing a numerical failure.

## Provenance

- Primary runner commit during both runs: `9c3a705c854a3c03d7182786c9463f0f6606722c`, clean.
- Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean and read-only.
- Fast: `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, clean and read-only.
- Config SHA-256: `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`.
- Q/P SHA-256: `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`.
- Canonical input SHA-256: `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`.
- Shared frontend SHA-256: `15700630e2fddbe1c3765cecead2e0612aab770fd23615b5ba4f8d81fa80f352`.
- 4096 slots, LogN 13, E=32; each raw run recorded identical workload/profile/Q/P hashes and `primary_dirty=false`, `backend_dirty=false`.

## Capacity and row authority

Exact bounds were derived from the deterministic binary64 input coordinates as rational values, with `B=ceil(2·max|slot|·Scale+1)`, `B_Add=B_A+B_B`, `B_MulRelin=N·B_Add·B_C` (`N=8192`, zero-c1 Fast multiplication), and `B_Rescale=floor((B_MulRelin+(q5−1)/2)/q5)`.

| Checkpoint | Conservative B | Capacity result |
|---|---:|---|
| Encrypt A, Scale `2^45` | `2348557866436` | `2B < S_Q0123` |
| Encrypt B, Scale `2^45` | `1804822278899` | `2B < S_Q0123` |
| Encrypt C, Scale `q5` | `76957544167327219` | `2B < S_Q0123` |
| Add | `4153380145335` | `2B < S_Q0123` |
| MulRelin | `2618441203534382746900245490606080` | `2B < S_Q0123` |
| Rescale / Rotate / final Level-0 projection | `2271135713118062` | `2B < q0` after DropLevel |

`S_Q0123=5986308565615587353347023386369277282933624144412673`; `q0=36028797018652673`. After Rescale the final strict Level-0 margin is `q0−2B=31486525592416549`. These component bounds describe Fast’s zero-secret encoded c0 representation, not Standard’s randomized RLWE components; Standard is validated by its native public decrypt/decode oracle.

Fast full-profile `EncryptNew` inputs physically had six rows. The public Add and MulRelin outputs at Level 5 retained q0..q3 only; Rescale consumed the logical top divisor q5 and produced Level 4 with four authoritative rows; Rotate preserved that layout; ordinary `DropLevelNew(4)` projected to Level 0 with q0 only and did not change Scale. The existing Fast Q-prefix capacity observer passed on every pre-Bootstrap checkpoint. No q4/q5 coefficient backing was read by these compact operations and no Standard/full-RNS fallback was used.

## Numerical results

The fixed maximum-error gate remained `1e-6`.

| Measurement | RMSE | Max complex error | SNR | Result |
|---|---:|---:|---:|---|
| Standard pre-Bootstrap native oracle | `1.466359017527355e-11` | `5.86605418722453e-11` | `152.5636158781992 dB` | PASS |
| Fast pre-Bootstrap native oracle | `7.442066238955308e-13` | `2.1986015127059466e-12` | `178.45455153965648 dB` | PASS |
| Fast vs Standard pre-Bootstrap | `1.4678127606723786e-11` | `5.903515384823777e-11` | `152.5550089855535 dB` | PASS |
| Standard post-Bootstrap native oracle | `4.826764480602127e-9` | `1.424471312557334e-8` | `102.21530000009507 dB` | PASS |
| Fast post-Bootstrap native oracle | `1.1795718406445398e-9` | `5.7387742390963576e-8` | `114.45393423748425 dB` | PASS |
| Fast vs Standard post-Bootstrap | `4.937529503163558e-9` | `5.038821901648196e-8` | `102.01822817774493 dB` | PASS |

Both Bootstrap outputs were Level 1 / Scale `2^45`, decoded through native `DecryptNew/Decode`, and passed the unchanged gate. Exactly one public Bootstrap ran per backend (2 total), only after the paired B gate passed; no retries were made.

Primitive key lookups were instrumented: Standard used one relinearization-key lookup for MulRelin and one Galois-key lookup for Rotate; Fast used zero direct relin/Galois lookups for those primitives. Both evaluator constructors enumerated the Galois-key list once. Bootstrap-internal key accesses were not runtime-instrumented and are not claimed to be zero.

## Validation and artifacts

- Fast targeted test and `go vet`: PASS.
- Standard targeted test (`lattigo_standard`) and `go vet`: PASS.
- `git diff --check`: PASS.
- Compact machine-readable evidence: [evidence JSON](FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019-evidence.json).
- Execution details and exact commands: [journal](FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019-journal.md).

This result establishes only the tested public chain and workload. It does not establish every CKKS overload, compact high-Level native decryption, performance, or security equivalence.

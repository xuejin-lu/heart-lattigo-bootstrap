# FAST-STANDARD-INPUT-PROVENANCE-001 — input-origin preflight

Status: `INPUT_PROVENANCE_READY` (input contract only; `FAST-STANDARD-PERF-REBASELINE-003` remains blocked).

## Pinned source and method

- Primary harness commit tested from a clean `main`: `37b73ad58983bd7689999847ed5c59c4a1e14fed`.
- Genuine unmodified Standard Lattigo: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean detached source, `perf_standard` build tag.
- Accepted Fast Q-prefix production: `5feb44917fca40c93abec6def6f26bc81a82c536`, clean detached source, `perf_fast` build tag. The safely synchronized `fast-qprefix` branch was `d1e0281305e0ae6ab883e491dc6717214d5ea6ef`; its changes after the production pin were only `AGENTS.md` and `CURRENT_TASK.md`. No Secondary file was edited or pushed.
- Go `go1.26.4`, `darwin/arm64`. Each smoke process verified that its compiled Lattigo module replacement resolved to its declared clean pinned worktree and that worktree's HEAD matched the exact SHA.
- Same `perfmeasure.DeterministicInput` and canonical config per profile; `ParametersFromConfig` produced identical full effective parameters within each Standard/Fast pair (including actual Q/P prime values, LogN, LogSlots, scale, and `CircuitOrder=ModUpThenEncode`). The original message hash matches within each pair. Ciphertext, key, decoded-input and metadata identity are **not** required to match.

## Input origin and mandatory pre-Bootstrap quality

| Profile | Config SHA-256 | Original-vector SHA-256 | Backend/input kind | Level / degree / LogCols / scale | Max complex-slot deviation (independent trials) | Gate |
|---|---|---|---|---|---|---|
| LogN13 | `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98` | `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` | Standard `standard_native_rlwe_encrypt_v1` | 0 / 1 / 12 / 2^45 | `2.852737083864921e-11`, `2.495954710855888e-11` | pass |
| LogN13 | same | same | Fast `fast_zero_secret_direct_encoded_v1` | 0 / 1 / 12 / 2^45 | `2.101955261562136e-12` | pass |
| LogN16 | `16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02` | `659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf` | Standard `standard_native_rlwe_encrypt_v1` | 0 / 1 / 15 / 2^45 | `7.461837700068314e-11`, `7.351894340649238e-11` | pass |
| LogN16 | same | same | Fast `fast_zero_secret_direct_encoded_v1` | 0 / 1 / 15 / 2^45 | `6.442117609356083e-12` | pass |

The mandatory bound is maximum absolute **complex-slot** difference `<= 1e-6` from the original message, with finite values and exact slot count (4096 for LogN13; 32768 for LogN16). All six constructed inputs passed. Their Level-0 input metadata SHA-256 was `9171dac4a79b97f7f916001d0ea5805f6137711527d364f8dfcd8e71ed2764f6` for LogN13 and `58b699119d1ca2d73cfe8812e6ae776f77d4267ff4c25e67229d7b140a6cf44f` for LogN16. Equality of these metadata hashes across backend runs happened for these profiles but is **not** a comparison prerequisite.

The source-level constructor paths are distinct:

- Standard: `rlwe.NewKeyGenerator(bootstrappingParameters).GenSecretKeyNew`, `ckks.NewEncoder.Encode` of the shared original vector, then `rlwe.NewEncryptor(residualParameters, generatedSecret).EncryptNew(plaintext)`. `rlwe.NewDecryptor(residualParameters, matchingSecret).DecryptNew(ciphertext)` plus `ckks.NewEncoder.Decode` validates every input. Each independent Standard smoke trial generated a new key and called `EncryptNew` again. Keys were compared in memory; no key bytes or fingerprints were serialized. Both `c1` rows were nonzero and their SHA-256 values differed within each profile. A synthetic Standard `c0=encoded-message,c1=0` ciphertext failed the formal input preflight in both profile tests.
- Fast: same shared-vector plaintext encoding, then the pinned Fast Bootstrap test-helper-equivalent copy to `c0` with `c1=0`. The Fast adapter alone owns this constructor. It is an **intentionally insecure zero-secret/zero-a direct-encoded simulation**, not native encryption, not secure, and not security- or input-noise-equivalent to Standard. A `KeyLayoutFast` public key was not passed to generic `rlwe.NewEncryptor`. `c0` decoding under the declared Fast semantics passed the quality gate. One untimed public Fast Bootstrap call per profile accepted the prepared input; its output was intentionally not evaluated.

Within each profile the Standard trials had distinct pre-decoded hashes and `c1` row hashes; their ciphertext content and randomness were not constrained to match Fast. Formal `perfprobe compare` now rejects old/unknown input-kind records and missing Standard native-origin evidence, checks matching original vectors and full effective profile, and no longer requires equal Standard/Fast ciphertext metadata SHA.

Standard trial `c1` q0-row SHA-256 pairs: LogN13 `e7597128fd5b41935d5b51c8f0b8991504156d469c8eb268ba0e458c7ed86d85` / `1557c7c35551c9ae23ccc8c2ba994b7e764831b161fe53a1743b2642f13a2f4e`; LogN16 `33788d674c7b9f878fe21615c1fc0b796f7e83ba5cbc14fab0b82be7c314e8a0` / `c11bfd716149512ad2e5e3373e49fe1376e7fa85d8f7a78d2630f90f4031ca65`.

## Exact modulus provenance

The following are the **actual** Q/P prime values emitted by `ParametersFromConfig`, not merely target bit sizes. Within each profile, Standard and Fast arrays matched element-for-element.

- LogN13 Q: `36028797018652673, 549755731969, 549756026881, 549755486209, 549756174337, 1152921504606830593, 1152921504606748673, 1152921504606994433, 1152921504606683137, 1152921504606601217, 1152921504606584833, 1152921504607191041, 1152921504607223809, 72057594037616641, 72057594038321153, 72057594037370881, 72057594037338113`.
- LogN13 P: `2305843009213317121, 2305843009213120513, 2305843009214414849, 2305843009212694529, 2305843009214758913`.
- LogN16 Q: `36028797019488257, 549754109953, 549753978881, 549753716737, 549753192449, 1152921504606584833, 1152921504614055937, 1152921504598720513, 1152921504615628801, 1152921504616808449, 1152921504597016577, 1152921504595968001, 1152921504618381313, 72057594038321153, 72057594036879361, 72057594035306497, 72057594040680449`.
- LogN16 P: `2305843009211596801, 2305843009210023937, 2305843009218281473, 2305843009208713217, 2305843009218936833`.

## Validation and claim boundary

- Passed targeted tests against the exact clean Standard source: `GOCACHE=/private/tmp/fast-standard-input-provenance-go-cache go test -modfile=/private/tmp/fast-standard-rebaseline-002/standard.mod -tags=perf_standard ./cmd/perfprobe ./internal/perfmeasure`.
- Passed targeted tests against the exact clean Fast source: `GOCACHE=/private/tmp/fast-standard-input-provenance-go-cache go test -modfile=/private/tmp/fast-standard-input-provenance-001-fast.mod -tags=perf_fast ./cmd/perfprobe ./internal/perfmeasure`.
- Passed untagged tests: `GOCACHE=/private/tmp/fast-standard-input-provenance-go-cache go test ./cmd/perfprobe ./internal/perfmeasure`.
- Four bounded input-smoke commands used `go run` with their respective modfiles/build tags, `-input-smoke`, profile config, exact `-backend-commit`, and clean `-secondary-root`; Fast additionally used `-fast-bootstrap-acceptance`. The compact temporary smoke JSON files were not committed. No seven-repetition campaign, timing table, speedup, Bootstrap SNR, or output-accuracy claim was made. No repository-wide suite was run for this input-only task; no unrelated failure was encountered in targeted tests.

This establishes input provenance only. A later Web-approved task must evaluate each backend's Bootstrap output against the **original message**, with cross-backend output comparison secondary and the simulator's security/noise limitations explicit.

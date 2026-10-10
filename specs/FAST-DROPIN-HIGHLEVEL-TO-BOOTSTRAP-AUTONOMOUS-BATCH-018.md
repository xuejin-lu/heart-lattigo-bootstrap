# FAST-DROPIN-HIGHLEVEL-TO-BOOTSTRAP-AUTONOMOUS-BATCH-018

**Status:** READY_FOR_CODEX
**Type:** Four bounded autonomous checkpoints, single Web scientific review.
**Prior evidence:** `results/FAST-DROPIN-PUBLIC-PRIMITIVE-BOOTSTRAP-AUTONOMOUS-BATCH-017-web-review.md`, accepted `016-web-review.md`, 015R.
**Genuine Standard:** unmodified pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Fast:** `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
**Global budget:** max ONE Standard and ONE Fast Bootstrap, only if all preflight gates pass; zero reruns/warmups, zero benchmark, zero LogN16, zero parameter sweeps. If no valid public Level/Scale/capacity path exists, STOP *without spending any Bootstrap calls*.

## End goal / scientific question

Can the **unchanged public CKKS application workflow** under the **already frozen canonical 017 E32 Bootstrap profile** perform *genuinely compact* high-Level Q-prefix arithmetic (`Level>3`, dormant q4 and above), transition through mathematically and semantically justified **ordinary public** Level/Scale operations to a valid residual Level0 ciphertext, and enter the already-proven public E32 `Bootstrap` and ordinary low-Level `DecryptNew/Decode`, using **only a library dependency swap** for genuine Standard vs zero-secret Fast?

The accepted 016 Level5 proof uses a different six-prime source profile and a special Fast measurement decoder; the accepted 017 public Bootstrap uses residual Level1→0 and two full Q rows. This charter does **not** presume their direct composition is feasible under the frozen canonical 017 E32 parameters.

## Immutable constraints

Read current Primary/Secondary `AGENTS.md`, `CURRENT_TASK.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B, `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`, Secondary `docs/FAST_QPREFIX_SPEC.md`, 016/017 source/evidence and 017 Web review before edits.

- Frontend operation signatures, native public EncryptNew/DecryptNew, `GenEvaluationKeys`, `NewEvaluator`, `Bootstrap`, real Standard baseline, E32, LogN13, effective Q/P, canonical 017 config, Scale/default and key layout must not change. No Fast-only application source imports/constructors, manual c1 patching, special measurement decoder or backend-specific business logic.
- Logical CKKS Level/Scale are authoritative. Fast maintains `w_Q(ell)=min(ell+1,4)`. Existing shared `fastcore` kernels and bounded centered CRT are source of truth; no duplicate math, full-RNS fallback on dormant rows or reading stale q4+.
- **Do not conflate independent 016 six-prime residual profile and 017 two-prime residual profile.** Changing canonical Q/P just to manufacture Level5 would invalidate matched Bootstrap provenance. No new architecture/parameter selection delegated.
- `DropLevel` merely discards levels, not a rescale or proof of numeric integrity: crossing a representation bound must satisfy independently demonstrated centered uniqueness and a correct public layout. Must prove **2B < q0** before any target Level0 centred interpretation, and preserve scale and Bootstrap input contract. Silent truncation/resize to make a test pass is prohibited.
- Existing ordinary RLWE DecryptNew may not support Level>3 compact ciphertexts. For a *preflight diagnostic only*, a previously proven measurement-boundary CRT decoder may be reused with explicit labeling, but it cannot be introduced into the **claimed final unchanged frontend**; a true end-to-end native-decrypt proof must use ordinary native DecryptNew at the supported final Level.
- Existing Bootstrap public `NewEvaluator` supports E32 and has its own residual input level/scale constraints. No forced bridge across incompatible parameter chains. Key generation and storage should remain original APIs.

## A — read-only feasibility / source-backed profile & depth audit

1. Safe-sync and verify both commits; inspect exact 017 canonical config, constructed `btpParams.ResidualParameters` Q length, Bootstrap Q prefix, plaintext/ciphertext preparation and CKKS `DropLevel/Rescale` semantics, current public Fast dispatch, actual allowed intermediate Level and q0 capacity.
2. Compare actual 016 and 017 Q/P and Level families; draw exact path from genuinely above-cap compact Level5 to canonical 017 residual Level0 using the **same allowed original source profile**. Identify whether existing canonical 017 source/residual parameter definition even contains legal high-Level inputs that can be carried into Bootstrap; map params and key domain if not.
3. Derive conservative independent bound for any proposed Level reduction (including target q0) and Scale evolution, avoiding cross-check solely with centered observer. Check backend-compatible native public decrypt and output allocation. **If no exact legitimate path exists under frozen constraints, STOP `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with source/capacity/parameter proof. Do not invent a new profile, remap keys or reinterpret Levels.**
4. Create a compact feasibility note in the new 018 journal; do not edit Secondary production and spend zero Bootstrap calls.

## B — zero-Bootstrap matched public preflight, ONLY if A passes

1. Build a new small Primary-only source-identical Standard/Fast public frontend for the approved path, including actual high-Level compact arithmetic and a justified Level/Scale reduction to canonical residual Level0.
2. Run matched Standard and Fast deterministic preflights, with independent cleartext oracles, precise Q/P and frontend/input hashes, authoritative q0..q3, dormant higher rows, no generic full-RNS fallback, native public decryption at Level0, exact canonical q0/Scale and Bootstrap keyset identity.
3. No Fast-specific constructor, measurement decoder, relin/key behavior or altered secret weight in the application frontend. Unprovable capacity, incorrect ciphertext modulus or changed config **must STOP** before Bootstrap.

## C — maximum one-pair Bootstrap, ONLY if A/B pass

1. Journal consumed budget before each actual call; run **at most one** Standard and **at most one** Fast public Bootstrap (E32) on the exact preflight chain output, no repeats even if a call fails.
2. Collect fixed cleartext oracle / Standard-vs-Fast RMSE, max, SNR, output Level/Scale/slots, ordinary DecryptNew/Decode, row authority and Bootstrap key plan. No benchmark. Never silently switch to explicit Fast API or materialize q4+ during compact evaluation.
3. If any public integration issue is discovered requiring math or library edits, STOP with exact first cause and consumed budget, do not patch Secondary during 018.

## D — evidence and safe handoff

1. Commit compact Primary-only `results/FAST-DROPIN-HIGHLEVEL-TO-BOOTSTRAP-AUTONOMOUS-BATCH-018-{summary.md,journal.md,evidence.json}` plus a bounded runner/test **only if A permits execution**. If A blocks, evidence may be read-only feasibility map (no manufactured numeric data).
2. Record exact pins, config/source hashes, key/Level/Scale/capacity and reason for any impossibility. Do not overwrite historical 016/017 artifacts. Run focused tests/vet/diff check only for new code.
3. Self-review, ordinary fast-forward push Primary and deliver `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only on full genuine composition; else `BATCH_BLOCKED_NEEDS_WEB_REVIEW` and first source-backed obstruction. Never start 019 autonomously.

**Scientific priority:** A mathematically justified infeasibility outcome is a valid result. Do not turn an inherently two-level residual profile into a fabricated Level-5 ciphertext by adding unrelated Q primes. Any new canonical profile/design should come from a **new Web-owned decision after** such a stop.

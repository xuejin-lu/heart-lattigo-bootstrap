# FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014

**Status:** READY_FOR_CODEX
**Class:** Bounded M-frozen I/E autonomous batch, 4 sequential checkpoints.
**Authority:** `AGENTS.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B, `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`, Secondary `AGENTS.md` and `docs/FAST_QPREFIX_SPEC.md`.
**Accepted predecessor:** `results/FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013-web-review.md`.
**Start state (verify after safe synchronization):** Primary `main@89b019ebacedc3f8eed307a2083ab10a26134f61`; Fast Secondary `fast-qprefix@9489925c803be09e49a840e6828b005cdeb1faca`; pinned genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. If a ref advanced, inspect changes and report material conflict rather than silently revert it.

## Single end goal

The exact same ordinary CKKS application source and public API (`ckks.NewEvaluator`, public ciphertext lifecycle, ordinary `Add/MulRelin/Rescale/Rotate`, eventual `Bootstrap`) must use Standard CKKS with the original Standard dependency, and automatically use the **already-implemented, intentionally insecure zero-secret Fast/Q-prefix engine** when the dependency is switched to the Fast fork. No Fast-specific frontend imports, constructors, flags, parameter edits, or hand-built ciphertexts.

**User's confirmed integration decision:** No parallel Normal/Standard execution backend is required inside Fast merely to compare. Use independent unmodified Standard pin as baseline. P0 public full-Q and C0 explicit compact-Q are historic transitional integration contracts, not a permanent two-path design requirement. Do not optimize or expand P0 full-Q instead of implementing real Fast dispatch.

## Frozen invariants and implementation boundaries

- **Reuse before reimplement:** Existing `schemes/ckks/fast` routines for arithmetic, rescale (including centered CRT and logical-top divisor), automorphism, compact ciphertext, and existing Fast Bootstrap are source of truth. Do **not** rewrite these algorithms in public `ckks`, duplicate kernels, or create competing numerical implementations. A pure internal relocation to break Go imports must preserve algorithm behavior and old explicit Fast tests.
- `ckks.NewEvaluator(params, evk)` must retain exact public signature and concrete return type `*ckks.Evaluator`; public operation names, signatures, operand acceptance, Level/Scale, slots, effective Q/P and E remain unchanged. An internal Fast engine is permitted, but **no import cycle**: current `fast.Evaluator` depends on `ckks.Parameters`. Prefer neutral internal layering / move shared implementation behind a private adapter rather than importing `ckks/fast` from `ckks` or making caller pick an API.
- The current Fast fork identifies intentional zero-secret simulation internally; preserve its zero-c1 semantics and fail closed on invalid Fast input. No native `GadgetProduct`, Standard KeySwitch or full-RNS fallback on dormant/stale Fast rows. Standard baseline source must not change.
- Secondary `docs/FAST_QPREFIX_SPEC.md` governs all math: actual maintained Q prefix, centered uniqueness/individual bound, materialization, logical Rescale top modulus, ModUp/Level/Scale, and dormant-row prohibition. Existing tested exact flows are reusable; a new mathematical issue only exists when a **concrete reproduced state** violates those rules.
- Do not unilaterally simplify or disable public functionality that existing application source legitimately needs. For unsupported cases: explicit error plus precise gap/provenance, not hidden fallback or a fabricated pass. Do not broaden to BFV/BGV or redesign cryptography.
- Workload/measurement source, parameters and numerical gates remain frozen where prior fixtures apply; no tuning, respecifying E, or scaling thresholds to force success. No physical worktree destruction, rebases or force pushes.

## Autonomous checkpoints

**A — Source-backed adapter map / existing-function inventory (no editing yet).**
1. Safe sync Primary and Secondary per current `AGENTS.md`; independently read current authoritative constitution, latest task, existing explicit Fast tests and 008/009/013 evidence.
2. Record exact `ckks.NewEvaluator` concrete type and embedded RLWE evaluator; public/explicit `Add/Sub/MulRelin/Rescale/Rotate` signatures, `fast.NewCiphertext`, `ckks.NewCiphertext`, `EncryptNew`, `DecryptNew`, and call-graph imports. Identify direct relocation seam to route a public call into the **same** existing Fast engine without an import cycle. Identify compatibility edge cases (nil output, aliased output, ordinary AddNew etc.) as adapter issues, not new math.
3. Review existing accepted compact-Q primitive tests and full Fast Bootstrap for proven producer row authority. Do not assume simple physical row count proves source provenance.
4. Freeze the minimal *one representative arithmetic operation* for checkpoint B, preferably Add/Sub if current test contracts allow; otherwise choose one with better existing evidence. Write a compact choice/rationale into the batch journal. **If no source-backed adapter route exists without new mathematical/representation semantics, STOP `BATCH_BLOCKED_NEEDS_WEB_REVIEW`; do not invent a new evaluator algorithm.**

**B — First actual ordinary-public-to-existing-Fast engine bridge (Secondary implementation).**
1. Make the minimum backend-only changes to dispatch the chosen ordinary `ckks.Evaluator` method through reused Fast compact-Q arithmetic; preserve the public method signature and return behavior. An import-neutral internal engine extracted from existing Fast source is allowed; old explicit Fast entrypoint must continue using the same engine.
2. The public source must create/receive a Fast-compatible zero-secret ciphertext via its **ordinary** lifecycle; public `ckks.NewCiphertext` or RLWE allocation may have surplus backing, but only the documented Q-prefix rows may be authoritative, and implementation must avoid wasting full-RNS arithmetic. No frontend conversion to `fast.NewCiphertext` or manual c1 patching.
3. Prove with source-backed instrumentation or a focused test that the intended existing Fast core executes and the native generic full-RNS counterpart does not. Preserve Fast explicit tests and ordinary Standard semantics in the separate Standard pin.
4. Use two representative Levels including a level above the four-row Q-prefix cap if reachable under the existing validated arithmetic contract. Check decoded plaintext oracle, zero-c1, output Level/Scale, unchanged logical CKKS params, and real Q rows touched. Unsupported cases fail closed without mutating outputs.
5. Run focused Secondary tests and record commit on `fast-qprefix`. If completion would demand inventing a new math rule, stop at this checkpoint with evidence.

**C — Narrow composition and migration feasibility.**
1. From the same **unchanged public frontend** source, run the new bridged operation and one or more additional public operations already supported by current Secondary, in an order covered by existing Fast mathematical evidence; use the existing small deterministic LogN13 oracle fixtures. The key goal is kernel dispatch/compact-row provenance, **not** another P0 full-Q-only success.
2. Determine whether `MulRelin`, `Rescale`, `Rotate`, and public output allocators can already consume/preserve the compact authority end-to-end. Add only the smallest additional adapter glue supported by the existing Fast kernels and same mathematical contracts; do not write a second Fast arithmetic implementation or silently materialize full Q rows on every operation.
3. If compact authority cannot safely cross a public operation boundary, record its **first exact source/operation/Level/row provenance incompatibility** and classify it as `INTEGRATION_DISPATCH`, `API_ADAPTER`, `REPRESENTATION_CONTRACT` or `GENUINE_UNSUPPORTED_CASE`. STOP for new mathematical semantics; do not conceal a partial bridge as completed E2E.
4. No Bootstrap call, no benchmark, no LogN16 run, no full repository sweep in this batch. Reuse previous 012 Bootstrap result as historical evidence only.

**D — Verification and Web handoff.**
1. Generate compact source map + dispatch/row authority and numerical oracle matrix; distinguish *real compact Fast execution*, generic public full-Q execution, explicit unsupported cases and untested operations. Include SHA of both dependencies, exact frontend/config/input hashes, Go version, tests, negative controls, any key lookups. Compare with genuine Standard using a separate unmodified checkout/workspace, never an internal Normal backend.
2. Codex self-review source diff, safety, aliases, tests, key fallback, import cycle and any dormant reads; bounded one local repair pass per checkpoint. Run affected package `go test`, `go vet` as appropriate. Commit and ordinary fast-forward push authorized Primary/Secondary changes; no history rewrite.
3. Produce `results/FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014-{summary.md,journal.md,evidence.json}` (evidence compact). Return `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only if B and C meet their own gates; else `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with first supported counterexample. Do not start 015 autonomously.

## Controls / repository scope

**Primary:** may add one new dedicated runner/test, journal, evidence, summary; do not edit prior 005/009/012/013 artifacts, change test profiles, or modify application-facing source. `CURRENT_TASK.md` is updated once by Web before Codex begins.
**Secondary:** may reorganize/bridge existing Fast implementation, modify necessary CKKS-specific integration wrappers and focused tests on `fast-qprefix`, always retaining original existing Fast algorithm identities; no changes to genuine Standard or historical `fast-ckks`.
**Cost limits:** 0 Bootstrap, 0 benchmark, 0 LogN16, no repeated expensive tests. Cheap focused tests only and one deterministic numeric runner per dependency when ready.
**Scientific stop:** fresh math/capacity decision, inability to preserve existing public signatures, proof of stale/dormant active-row read, hidden native KeySwitch fallback, differing matched original profile, first unexplained numerical oracle failure after one bounded repair, unsafe Git state.

**Important:** Passing one bridge is progress but not whole CKKS drop-in or validated Fast speedup. Do not merely report P0 numeric PASS. Preserve historical Standard-vs-Fast comparability without embedding Standard execution into the Fast fork.

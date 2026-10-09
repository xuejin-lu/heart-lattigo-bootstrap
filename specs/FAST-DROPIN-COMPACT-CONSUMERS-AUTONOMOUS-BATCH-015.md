# FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015

**Status:** READY_FOR_CODEX
**Class:** M-frozen bounded I/E batch, four sequential checkpoints.
**Predecessor:** `results/FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014-web-review.md`; Batch 014 is an accepted *partial* Add/Sub bridge and remains blocked at compact downstream composition.
**Baseline:** genuine Standard Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, independent and unmodified.
**Fast starting point:** `xuejin-lu/lattigo fast-qprefix@0c92e7ed071a1be6eb258474811db470e4463df7` (safe-sync and re-check).
**Research goal:** unchanged ordinary frontend CKKS source and parameters, library dependency change only; reuse existing intentionally insecure Fast Q-prefix kernels behind the original public API.

## Immutable mathematical and workflow authority

Read both repos' `AGENTS.md`, Secondary `docs/FAST_QPREFIX_SPEC.md`, Primary `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B, the accepted 014 review, and exact current source before edits. The Q-prefix policy is `w_Q(Level)=min(Level+1,4)`; beyond that width q rows are deliberately dormant. Standard logical Level/Scale, actual q-primes and top-level-q divisor semantics continue unchanged. Existing centered CRT and capacity contracts, producer-authoritative row counts, NTT/Montgomery domain assumptions, zero-secret c1 and materialization boundaries must be honored without making up new rules.

**Do not turn the Level-5 q4 absence into a requirement to materialize q4/q5 for routine Fast arithmetic.** Do not duplicate existing explicit `fast.Evaluator` kernels, invent alternate arithmetic, or call native full-RNS on dormant rows. Import-neutral internal extraction / thin adapter is allowed. The exact concrete `*ckks.Evaluator` public type and all public signatures/arguments must remain unchanged, including `ckks.NewEvaluator`, `MulRelin(New)`, `Rescale`, `Rotate(New)`, and existing successful Add/Sub behavior.

## Checkpoint A — precise consumer map and first math eligibility gate

1. Safely sync Primary main and Secondary fast-qprefix, never discard work; verify expected pins and re-read source/constitutions. No edits to genuine Standard or historical Fast branch.
2. Inspect `schemes/ckks/evaluator_fast_zero_secret.go` (current public MulRelin full-Q), `fast/evaluator_ntt.go` (existing `MulRelin`, `MulRelinElementQPrefixRows`, `mulElementRows`), `fast/rescale.go` and `fast/rescale_qprefix.go` (existing CRT + explicit width), `fast/evaluator.go` (explicit Rotate), and existing `fastcore` shared Add/Sub and automorphism.
3. Pin the smallest supported compact ciphertext×ciphertext `MulRelin` path: exact degree-one, metadata, NTT/Montgomery, zero-c1 and maintained-row preconditions; output Degree 1, `Scale = Scale0*Scale1`, logical Level preserved; no relin key/KeySwitch/GadgetProduct, no full-active-Q preflight. Explicitly confirm whether current Fast code already has a bounded multiplication oracle and valid capacity rule for the selected test.
4. Freeze a **separate** Level-5 high-Level structural fixture with 6 Q primes, retaining q0..q3. A pure Level-5 structural Q-prefix test is not a numerical proof; for numerical decoding, either use the accepted Level<=3 profile unchanged or a separate *explicitly labeled* safe centered-lift Level-5 fixture whose capacity is proven and decoded oracle is meaningful, **without pretending to modify the old frozen profile**.
5. If source proves the needed public adapter cannot preserve existing representation semantics, STOP `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with a concrete first cause; do not escalate a hypothetical dormant q4 need.

## Checkpoint B — public compact MulRelin via existing Fast core

1. Refactor/move only the already implemented compact Fast multiplication/relinearization row kernel and its required scratch into the internal import-neutral Fast core to avoid the `ckks` ↔ `ckks/fast` Go import cycle. The old explicit Fast calls and ordinary public Fast `MulRelin` / `MulRelinNew` **must reuse the same core**, demonstrably by source callgraph + instrumented test; no duplicated arithmetic loops.
2. Maintain zero-c1 invariant and Fast degree-two truncation/keyless semantics; no Galois/relin key retrieval. Do not apply the old full-active-Q `validateFastCKKSZeroSecretMulRows` to valid compact input. Maintain transactional fail-closed behavior for invalid input/output, including aliasing, unsupported type and malformed storage.
3. Ensure a correct compact output with logical Level (including Level 5), degree 1 and Scale product; `MulRelinNew` allocates compact backing. Do not modify the plain Standard repository or frontend to select Fast.
4. Verify with focused unit and oracle tests, including Level 1/3 paired Standard/Fast unchanged-frontend numeric checks and Level 5 row coverage, no q4 access; only claim Level-5 numerical correctness with an independently valid bounded oracle.
5. Self-review and repair only one bounded pass. If unexpected mathematical/capacity issue appears, STOP and report with first exact input/operation/bound/level rather than trying a new algorithm.

## Checkpoint C — reuse Rescale and Rotate; minimum composed public chain

1. Once B passes, inventory and bridge the *existing* Fast rescale (top logical q divisor + bounded CRT) into ordinary `ckks.Evaluator.Rescale` / relevant allocator and method entrypoints, and existing Fast automorphism shared core into public `Rotate/RotateNew` with **exact maintained-prefix row authority**. Do not reimplement the legacy Fast polynomial algorithm or a second rotation; use import-neutral extraction/adapter and existing verified implementations. **No full-active-Q row requirements** for valid compact inputs and outputs; preserve Standard public Level/Scale/metadata and Fast internal zero-secret invariant.
2. Add tests that show a compact Add→MulRelin→Rescale→Rotate chain under a supported pre-approved profile and numerical cleartext oracle per checkpoint. Cover structural Level>3 input with q4 dormant, correct output prefix contraction, explicit independent Standard baseline and no forbidden native full-RNS fallback.
3. If Rescale migration requires an unproven change to CRT bound or interpretation, or Rotate has an actual unsupported producer-domain assumption, record the concrete failure, STOP and request Web mathematical decision. Do not silently change parameters, force full Q materialization, or weaken tolerance.
4. Add a focused *inventory/control* of unsupported public scalar/plaintext/mixed operands so the Fast build cannot silently take a generic full-RNS arithmetic path on compact ciphertexts. Do not broaden this batch to complete every overload; classify unsupported cases visibly. Do not alter ordinary Standard semantics by modifying the genuine Standard pin.
5. The goal is integration dispatch evidence, not a new full Bootstrap run.

## Checkpoint D — independent-auditable handoff

1. Record committed code map, reuse identity for public/explicit Fast, domain/metadata/Level/Scale/c1 and authoritative-row checks, exact dependency SHAs, identical frontend/config/input hashes, key-lookup counters, decoded RMSE/max and first-failure negative controls. Distinguish compact arithmetic passing vs merely full-Q numerical compatibility.
2. Run focused `go test`/`go vet` on Secondary affected packages and one matched identical frontend runner in Primary with the pinned genuine Standard checkout and current Fast; do not claim local worktrees clean without checking them. Do not run Bootstrap, benchmarks or LogN16 in this batch. Reuse prior 012/013 results only as history.
3. Keep the user's agreed workflow: Codex bounded self-review (one repair pass), commit/normal fast-forward push, no destructive history or force push. Primary artifacts `results/FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015-{summary.md,journal.md,evidence.json}` and one new runner/test as needed; do not mutate historical artifacts.
4. Return `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only when B+C numerical/dispatch/representation gates genuinely pass. If stopped, `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with first causally supported failure and safely pushed evidence. Do not autonomously start another charter.

**Hard resource caps:** zero Bootstrap calls, zero benchmark runs, zero LogN16, no arbitrary parameter sweeps. Do not create a new research-level mathematical rule autonomously. Keep frontend source entirely CKKS-public, unchanged across Standard and Fast.

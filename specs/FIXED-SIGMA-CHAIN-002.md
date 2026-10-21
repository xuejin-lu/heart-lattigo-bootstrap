# FIXED-SIGMA-CHAIN-002 — Fixed-σ two-step chain and injection-cost attribution

**Status:** `AUTHORIZED_AUTONOMOUS_BATCH` — requires a new user `開始` in Codex; not executed by this GitHub documentation commit.
**Class:** M (reviewed invariants) + small E + bounded I only after gates.
**Purpose:** extend the accepted **single-operation `RMSE_ONLY`** result to a **two-step sequential ct×ct chain** at fixed σ, and identify why current diagnostic helper costs ~50–58% of one Mul+Rescale. Do NOT claim full Bootstrap support.
**Primary:** `xuejin-lu/heart-lattigo-bootstrap` / `main`; **Secondary:** `xuejin-lu/lattigo` / `fast-qprefix`.
**Starting accepted source refs:** Primary `167bd0c9c1d61a6f2ff01e37f62565eb123de5dd`; Secondary `9ef0315feaf59dbee03bf0852b72c76fdeb6b800`; genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. At runtime record synchronized exact commits and different binary hashes.
**Prior independent review:** `results/FIXED-SIGMA-MUL-NOISE-001/web-review.md`.
**Bootstrap budget:** **0 Standard, 0 Fast** (also no LogN16, teammate BTS, large benchmark). Do not borrow prior Task quota.

## Frozen scientific contract

- Same `sigma_coefficient=4.004280012097432`, from 001 CP-B. **Do not recalibrate** by Level, input, output, hidden holdout, or operation index. Sampling is fresh per logical complete operation; log distinct reproducible seeds and exact draws/hashes.
- The “new error” remains `r_op=z_decoded-x_decoded ⊙ y_decoded` at each step using the **actual decoded inputs of that step**, not the original initial cleartext product. Also report true total oracle error after each step and the error propagation of the previous injection.
- All candidate noise is defined in **signed coefficient domain after Rescale** and mapped to every currently authoritative Q-prefix row consistently with NTT/non-Montgomery representation. c1, Level/Scale, dormant limbs, Standard arithmetic remain unchanged. The existing Q-prefix centered capacity and valid source/provenance gates are non-negotiable.
- Use exact same public application CKKS source and operation order compiled with independent backend checkout via isolated `-modfile` / `GOWORK=off`, preserving `perfprobe` / `fastdiag` / `perfmeasure` / `numericalmetrics`. No per-backend frontend arithmetic branching and no new runner/stats platform.
- Existing helper `InjectFixedSigmaCoefficientNoiseDiagnostic` is opt-in, manually invoked and only supports isolated LogN13 Level4/Scale2^45. **Do not describe it as automatic injection or Bootstrap coverage.** The user eventually wants one injection per *complete* three-stage operation; no generic MulRelin API in this source includes Rescale. Mapping such boundaries requires explicit per-call source evidence; arbitrary `Rescale` hooks are forbidden.
- Keep existing `results/FIXED-SIGMA-MUL-NOISE-001/` and completed BTS/A=0 results intact. No retuning after observing held-out tests. Label all output `DIAGNOSTIC_ONLY`.

## Authorized checkpoints

### CP-A — Reuse-first source + math gate (no source edit)

1. Recheck current code/provenance and 001 source map; do not redo full audit. Identify the shortest legal two-step real ct×ct chain, e.g. first output Level5→4/Scale2^45; second Level4→3/Scale2^45 *only if* the second ciphertext input, multiplication Scale and q4 divisor can be constructed with ordinary CKKS semantics without representation violation.
2. Show algebraically that the same coefficient-domain σ at Scale2^45 is being tested at both stages; record **why 001 helper's hard-coded Level4 check must not be bypassed**. Show which rows remain authoritative, and how the existing Q-prefix capacity checker will validate the Level3 output.
3. Map all existing Primary measurement hooks and secondary tests reused; inventory components of helper overhead: PRNG/sample+hash/stats, NTT row transforms, c0 addition, clone + inverse NTT / Q-prefix CRT capacity check, allocations. Do not weaken capacity validation.
4. If input-scale/level/centered-lift or injection boundary cannot be validated using accepted mathematics, write evidence, `BATCH_BLOCKED_NEEDS_WEB_REVIEW`, stop without inventing scale hacks.

### CP-B — Zero-Bootstrap two-step *baseline* oracle and metrics

Use the same maintained `perfprobe` workload/metrics platform, adding only a focused adapter/fixture if necessary. Run at most **three matched seed pairs** for genuine Standard and unmodified Fast, with two actual ct×ct `MulRelin+Rescale` operations in series. Verify identical original input and operation schedule, exact observed Levels/Scales, full per-stage decoded and plaintext oracles, independent input fingerprints, separate SK/PK labels; the existing Fast PK unsupported condition remains explicit. Record source pin, binary SHA, timing, no fallback, 0 Bootstrap. If baseline is unsupported or inconclusive, stop; never fake second-step values.

### CP-C — Minimal opt-in Level3 extension + fixed-σ chain

Only if CP-A/B pass without new mathematics:
- Reuse existing Secondary experimental helper and existing fastcore capacity checker. **Only** allow the proven Level3/Scale2^45 output profile as a bounded extension; preserve strict errors on all other unsupported profiles. Leave zero-sigma true no-op and original Fast path unchanged.
- Wire explicit experimental injection after each **proven complete** ct×ct→Relin→Rescale in the same matched workload. Two completed operations must lead to exactly **two** injections, distinct recorded seeds, both with the *same frozen σ*. No injection for lone Rescale, ct×pt, Add, Rotate. Do not invent a global Rescale hook or claim full application-wide automatic behavior.
- Observe both stagewise incremental residual and propagated total error. Compare three columns genuine Standard, unmodified Fast, Fast+fixed-σ and include per-stage RMS, variance, complex mean, max, Level/Scale, covariance and injection count.
- If matching fails, report `FIXED_SIGMA_NOT_TRANSFERABLE` rather than tuning; maintain a clear distinction between observed RMS closeness and distribution/tail equality.

### CP-D — Cheap diagnostic overhead attribution + journal

- Using existing profiler/timing utilities, produce bounded repeated *zero-Bootstrap* observations for the existing helper, with sampling, NTT, capacity checking and addition allocated proportions where instrumentable; avoid new independent benchmark engine.
- Do not optimize or remove safety checks on the basis of noisy one-shot timing; propose minimal optimizations for subsequent Web review instead of preemptively altering validation math.
- Update `results/FIXED-SIGMA-CHAIN-002/journal.md` and `summary.md` with exact commits, test evidence, limited raw artifacts, gates/limitation and spend. Only if code/validation passes, normal safe commit+push to proper Primary/Secondary refs. At completion `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`, or `BATCH_BLOCKED_NEEDS_WEB_REVIEW` at an unresolved math/representation decision. Web performs independent acceptance.

## Safety and artifact limits

- **Zero Bootstrap calls** anywhere, including hidden in tests. No LogN16, new formal BTS workload, Standard A=0 branch, unsafe secret/security claim, `go.mod` permanent replace edit, or migration to historical `fast-ckks`.
- Preserve original Standard source, Fast stable behavior when opt-in off, qprefix dormant-row prohibition. Never reset, stash, rebase, force push or overwrite unreviewed local work. Preflight with Primary/Secondary current `AGENTS.md` and full existing safe fetch/ff-only rules; inspect size before reading large evidence files.
- Do not commit duplicate 4-MB-per-run vectors unless the mathematical validation genuinely requires them; if needed, store exact path/hash with bounded summary and avoid rereading big raw vectors for Web review.
- The next milestone does **not** authorize fully automatic injection inside Bootstrap. If the true goal requires a new general API or semantically stateful hook, stop for Web architecture review rather than implement silently.

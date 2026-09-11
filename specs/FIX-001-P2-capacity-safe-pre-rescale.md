# FIX-001-P2 — Capacity-safe pre-Rescale power multiplication

## Goal

Repair the second confirmed Fast Chebyshev power-generation capacity failure while preserving the q0/q1-only Fast architecture.

The formal coefficient-domain diagnosis proved that repaired `T3 = 2*T1*T2 - T1` fails **before** the correction term because the multiplication result already crosses the two-limb centered range:

- repaired `T1`: correct;
- repaired `T2`: correct;
- formal T3 uses strict `MulRelin`, split `(1,2,1)`;
- T3 post-Rescale error ≈ `0.031234736649481946`;
- P1=`T1_c0*T2_c0` already exceeds `q0*q1/2` for 1/8192 coefficients;
- full-Q Rescale oracle passes with max component ≈ `1.53e-5`;
- a q0/q1 wrapped simulation reproduces the actual Fast rows bit-exactly and the decoded Fast result to ~`2.22e-7`;
- therefore the supported cause is `t3_multiply_crosses_q01_capacity`.

The production repair must prevent the unsafe high-scale product from ever being formed under q0/q1.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary remote base when this spec is authored:

`e49e9e0a8cbad9cb0dfa425a28f3ca37a0d6d154`

Secondary repository: `xuejin-lu/lattigo`

Expected starting Fast commit:

`87be78ff3c591932699aba63d3be46ca306a6eea`

Standard reference remains pinned at:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Follow both repositories' `AGENTS.md` and Secondary `docs/FAST_CKKS_SPEC.md`.

If either worktree is dirty at startup, stop and report. Never reset/stash/discard unrelated work.

The parent FIX-001 remains incomplete until this repair passes formal LogN13 correctness and receives independent review.

---

# Required repair principle

Current Fast power generation effectively performs, at common logical level `L`:

```text
product = Mul/MulRelin(left, right)    // scale = Δ_left * Δ_right, level L
if Chebyshev: product *= 2
product = Rescale(product)             // divide by q_L, level L-1
apply Chebyshev correction
```

This creates the unsafe high-scale product before Rescale.

For the q0/q1 Fast backend, move the **same one required Rescale** from the product to an independently-owned operand copy before multiplication:

```text
commonLevel = min(left.Level, right.Level)
rescaledOperand = maintained-copy(chosenOperand at commonLevel)
Rescale(rescaledOperand)               // divide by q_commonLevel, level commonLevel-1
product = Mul/MulRelin(otherOperand, rescaledOperand)
                                           // level commonLevel-1
if Chebyshev: product *= 2
apply Chebyshev correction
```

There is **no post-product Rescale** for this power-generation step because the one planner-required level transition has already happened on the operand copy.

The intended final metadata is algebraically the same as the old schedule:

`outputScale = left.Scale * right.Scale / q_commonLevel`

`outputLevel = commonLevel - LevelsConsumedPerRescaling()`

For the currently supported Fast profiles `LevelsConsumedPerRescaling()` is expected to be one, but implementation should use the parameter API and should not silently hardcode assumptions if avoidable.

---

# Important structural requirements

## 1. Never mutate stored power operands merely to pre-Rescale

`pb.values[a]` and `pb.values[b]` are reusable power-basis nodes.

The operand selected for pre-Rescale must be copied into an independently-owned workspace scratch/buffer.

Do not Rescale the stored source power in place.

## 2. Align the copy to the exact common logical level before Rescale

The old post-product Rescale uses the modulus at:

`commonLevel = min(left.Level(), right.Level())`.

If the selected operand has a higher logical level, its diagnostic/working copy must first be structurally truncated to `commonLevel` while preserving the authoritative q0/q1 rows and exact Scale.

Because Fast higher q2... rows are dormant/non-authoritative, this is a structural level alignment only; it must not invoke full-Q arithmetic or read dormant rows.

Then Rescale that copy so the divisor is exactly the same `q_commonLevel` the old product Rescale would have consumed.

Add a narrow helper if needed, e.g. a maintained-copy-at-level operation, and test it.

## 3. Deterministic operand choice

Choose a deterministic operand for pre-Rescale.

Preferred rule:

- choose the operand with the larger Scale;
- if Scales are equal, use a stable tie-breaker such as the right operand.

Rationale: after division by the same `q_commonLevel`, rescaling the larger-scale operand preserves at least as much post-Rescale encoding precision.

If source inspection/testing shows another deterministic choice is materially safer while preserving planner semantics, document it and test both T2 and T3.

Do not choose based on ciphertext contents through an expensive production big.Int capacity scan.

## 4. Preserve lazy/strict multiplication semantics

The existing planner distinguishes:

- strict path: `MulRelin`, degree-one output;
- lazy path: `Mul`, degree-two output after required operand relinearization.

The repair must preserve this behavior.

Perform any existing required operand relinearization before creating/rescaling the selected working copy, or otherwise prove the resulting degree semantics are equivalent.

Do not force all power generation to strict mode.

## 5. Preserve Chebyshev correction ordering from FIX-001

After the capacity-safe product is formed at the already-rescaled level/scale:

- double for Chebyshev;
- if `c==0`, subtract scalar `1` at this safe scale;
- if `c>0`, use the existing post-Rescale correction/alignment path.

Do not reintroduce pre-Rescale Chebyshev correction.

---

# Production constraints

The repair must remain q0/q1-only.

Do not:

- materialize q2+ residues;
- call Standard polynomial evaluation or Standard Rescale as a hidden fallback;
- use full-Q/QP arithmetic in production;
- use per-coefficient `big.Int` CRT in production;
- lower the global Mod1 scale or change Bootstrap parameters;
- add an extra level transition relative to the planner;
- special-case only `T3`.

This must be a general Fast power-generation scheduling rule.

---

# Secondary regression tests

Commit production regression tests with the repair.

## A. Maintained structural copy-at-level test

If a helper is introduced, prove:

- q0/q1 rows are bit-exact;
- Scale and domain flags are preserved;
- logical Level becomes the requested lower common level;
- source is unchanged;
- no q2+ rows are read or required.

## B. Formal-scale T2 regression

Re-run/retain the FIX-001 T2 high-scale regression.

Require component error <= `1e-2`.

This guards against regression from moving the rescale point.

## C. Realistic-scale T3 regression

Add a T3 regression whose input magnitude is representative of the formal Mod1/C2S path, not the prior ~`1e-5` tiny synthetic values.

At the formal/high-scale parameter regime, evaluate real Fast Chebyshev T3 through production power generation.

Require:

- T1 pass;
- T2 pass;
- T3 pass at component threshold `1e-2`.

The test should fail on `87be78ff...` and pass after this repair.

## D. c>0 and c==0 coverage

Ensure tests cover both:

- `c==0`: T2;
- `c>0`: T3 or another non-power-of-two Chebyshev power.

## E. Lazy-mode coverage

Exercise at least one generated power through lazy mode to ensure degree-two planner behavior remains valid after pre-Rescale scheduling.

## F. Planner metadata

For representative strict and lazy powers, assert final:

- Level equals old/planned post-product-Rescale Level;
- Scale matches `left.Scale * right.Scale / q_commonLevel` within existing Scale precision rules;
- Degree matches planner lazy/strict contract;
- exactly one logical rescale depth is consumed, not two.

## G. Existing suite

Run clean Secondary:

`go test ./...`

All prior Rescale, polynomial, bootstrap, T2/T3, degree-30 and general Fast tests must remain green.

---

# Documentation update

Update `docs/FAST_CKKS_SPEC.md` narrowly:

- two-limb capacity can be violated by ciphertext-ciphertext multiplication itself, even when each operand is individually valid;
- for Fast Chebyshev power generation, the planner-required rescale is shifted to an operand copy before multiplication so the unsafe high-scale product is never formed;
- final logical Level/Scale/depth remain equivalent to the original schedule;
- stored power-basis operands remain immutable with respect to this pre-Rescale step.

Do not claim a universal theorem for arbitrary Fast CKKS multiplication outside the tested polynomial surface.

---

# Secondary implementation workflow

Starting from clean `87be78ff...`:

1. inspect synchronized `circuits/ckks/polynomial/fast.go`, Fast Rescale, and relevant tests;
2. implement the minimal general pre-Rescale scheduling repair;
3. run focused polynomial/T2/T3 tests;
4. run `go test ./...`;
5. inspect diff;
6. commit production code + regression tests + doc update;
7. push `fast-ckks` by ordinary fast-forward;
8. record exact repaired Secondary commit SHA.

No force push.

---

# Primary formal validation gates

After the new Secondary commit is pushed, use that exact clean commit for every Primary validation.

Do not change the formal LogN13 config/input.

## Gate 1 — T2/T3 primitive correctness

Re-run the corrected formal power diagnostic with `T0=1`.

Require:

- T1 pass;
- T2 pass;
- T3 pass;
- fixed threshold `1e-2`.

Also re-run the T3 capacity diagnostic at least in summarized form and confirm the **actual production pre-Rescale scheduling no longer forms the old overflowing P1 path**.

Record the new product Level/Scale and capacity evidence appropriate to the new schedule.

## Gate 2 — all generated formal powers

Re-run corrected formal power replay.

Require every generated power used by the actual degree-30 formal polynomial to pass:

`T1, T2, T3, T4, T6, T8, T16`

If the generated set differs under the synchronized implementation, validate the actual set rather than hardcoding only this historical list.

## Gate 3 — whole degree-30 polynomial

Run actual repaired Fast polynomial evaluator on formal E2.

Require source-backed plaintext polynomial oracle component error <= `1e-2`.

## Gate 4 — public EvalMod(real) and EvalMod(imag)

Re-run repaired public Mod1 stage for both branches against the already-validated Standard q0/q1 stage oracle.

Require both branches <= `1e-2`.

Do not stop after real passes; imag is mandatory.

## Gate 5 — full formal LogN13 Bootstrap correctness

Re-run authoritative full correctness:

- Standard `5dbffb...` generated-secret vs input;
- new repaired Fast zero-secret vs input;
- new repaired Fast zero-secret vs Standard final output.

All 4096 slots.

Require:

- `max_abs_real <= 1e-2`
- `max_abs_imag <= 1e-2`

Archive full vectors.

If any gate fails, stop at the first failing gate. Do not run LogN16 or performance work.

---

# Required Primary artifacts

Create new artifacts; do not overwrite historical FIX-001 diagnostics:

- `results/FIX-001-P2-POWERS-logN13-fast.json`
- `results/FIX-001-P2-EVALMOD-logN13-fast.json`
- `results/FIX-001-P2-CORRECTNESS-logN13-standard.json`
- `results/FIX-001-P2-CORRECTNESS-logN13-fast.json`
- `results/FIX-001-P2-summary.json`

Summary must record:

- old Secondary `87be78ff...`;
- new Secondary repair commit;
- exact changed Secondary files;
- T2/T3 metrics;
- all-power metrics;
- whole polynomial metrics;
- EvalMod real/imag metrics;
- full Bootstrap metrics;
- final Level/Scale equivalence evidence;
- tests run;
- Primary/Secondary clean-state evidence;
- first failing gate, or `ALL_LOGN13_GATES_PASS`.

---

# Completion criteria

Mark this P2 task COMPLETE only if:

1. production power generation shifts its one required rescale to an operand copy before multiplication;
2. no extra level is consumed;
3. no q2+/full-Q/big.Int production fallback is introduced;
4. T2 remains correct;
5. T3 becomes correct;
6. lazy-mode behavior remains correct;
7. all generated formal powers pass;
8. whole formal degree-30 polynomial passes;
9. public EvalMod real and imag pass;
10. full formal LogN13 Bootstrap passes authoritative correctness;
11. Secondary `go test ./...` passes;
12. both repositories end clean and are pushed normally.

Even if all gates pass, do **not** run LogN16, benchmarks, or EXP-003 in this task.

The parent FIX-001 should be reported as a **candidate for closure**, not silently rewritten as complete; independent review will close it.

---

# Non-goals

Do not:

- change Standard Lattigo;
- change Bootstrap parameters;
- lower Mod1LogScale;
- add active q2/q3 limbs;
- introduce dynamic per-coefficient capacity scans in production;
- fix unrelated PS/evaluator code;
- run LogN16;
- benchmark performance;
- start EXP-003.

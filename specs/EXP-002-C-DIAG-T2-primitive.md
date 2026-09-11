# EXP-002-C-DIAG-T2 — Isolate first T2 primitive failure and q0/q1 capacity

## Purpose

Previous formal diagnostics localized the first dependency-supported Fast polynomial failure to:

`T2(x) = 2*x^2 - 1`

with:

- LogN13 formal Mod1 workload
- E2 reproduction PASS
- first failing power `T2`
- T2 max component error about `0.9995117188159196`
- error almost entirely real and nearly constant across all 4096 slots
- whole polynomial failure reproduced

This task must determine whether the first failure is already present in:

1. `MulRelin(x,x) -> Rescale`,
2. only after doubling,
3. only when the `-1` constant is injected before Rescale,
4. or whether the failure is explained by loss of centered-CRT uniqueness because Fast keeps only q0/q1.

This is still diagnostic-only. Do not repair production Lattigo yet.

---

## Fixed provenance

Primary: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary remote base when this spec is authored:

`c908cf8b90eb646ef0be7ba22fbe2ca989578d21`

Fast Lattigo baseline:

`ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

Standard reference remains pinned, if needed only for supporting evidence:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Follow both repositories' `AGENTS.md`. Read `docs/FAST_CKKS_SPEC.md` before any temporary Secondary diagnostic work.

---

## Formal workload

Reproduce the exact formal LogN13 E2 input from the previous task:

- LogN = 13
- 4096 slots
- `configs/bootstrap_config.logN13.json`
- LogDefaultScale = 45
- Mod1LogScale = 60
- Mod1Degree = 30
- DoubleAngle = 3
- K = 16
- LogMessageRatio = 10
- same deterministic application input
- same `ModUpThenEncode` order
- actual Q2 real branch
- same E1 scale reinterpretation
- same E2 offset

Before T2 diagnosis, require the reproduced E2 decoded vector to match the previous EXP-002-C-DIAG-POLY E2 vector within:

- `max_abs_real <= 1e-2`
- `max_abs_imag <= 1e-2`

If not, stop with `DIAGNOSTIC_INPUT_MISMATCH`.

---

## Confirm the real T2 implementation order

At pinned Fast source, confirm from `circuits/ckks/polynomial/fast.go` that T2 is generated through the real implementation path:

1. `MulRelin(T1, T1)`
2. `Add(out, out, out)`
3. `Add(out, -1, out)`
4. `Rescale(out, out)`

Do not assume this order from this spec if synchronized source differs. If it differs, stop and report.

Also inspect the actual Fast implementations of:

- `MulRelin`
- scalar `Add`
- `Rescale`

in `schemes/ckks/fast/` before running.

---

## Diagnostic policy

Do not commit production Lattigo changes.

Temporary, uncommitted, build-tagged diagnostic helpers are allowed only if necessary to reproduce the exact E2 input or to snapshot metadata. Prefer the public Fast evaluator operations for the T2 micro-experiments.

All temporary Secondary files created by this task must be removed before completion. Secondary must finish clean at exact `ce79b861...`.

---

## Important rule about pre-Rescale decoding

Do **not** interpret a q0/q1 projection of the pre-Rescale `x^2`, `2*x^2`, or `2*x^2-1` ciphertext as a trustworthy decoded semantic value merely because Decode returns numbers.

The purpose of this task is specifically to test whether the maintained q0/q1 basis is large enough for those intermediates.

For pre-Rescale checkpoints, record:

- Degree
- Level
- Scale exact string
- `log2(scale)`
- IsNTT / IsMontgomery
- q0/q1 hashes for every component
- actual q0 and q1 moduli

but only make semantic pass/fail claims after a Rescale or other controlled operation that has an independently valid oracle.

---

## Experiment A — square then Rescale

Use an independent copy of the formal E2 ciphertext.

Execute only:

```text
A0 = E2
A1 = MulRelin(A0, A0)
A2 = Rescale(A1)
```

Decode A2 under the authoritative zero secret using the same q0/q1 projection procedure used in prior diagnostics.

Plaintext oracle for every slot:

`A2_expected[i] = E2_actual[i]^2`

Record full 4096-slot metrics.

Pass threshold:

- `max_abs_real <= 1e-2`
- `max_abs_imag <= 1e-2`

If Experiment A fails, record:

`FIRST_PRIMITIVE_REGION = square_mulrelin_or_rescale`

Do not claim MulRelin itself is wrong yet.

Still continue to the capacity accounting and low-scale control below, but do not use later baseline variants to move the first failure downstream.

---

## Experiment B — double square then Rescale

Independent formal E2 copy:

```text
B0 = E2
B1 = MulRelin(B0, B0)
B2 = Add(B1, B1)
B3 = Rescale(B2)
```

Oracle:

`B3_expected[i] = 2 * E2_actual[i]^2`

Same fixed threshold and full 4096-slot metrics.

Interpret only if Experiment A passed.

If A passes and B fails:

`FIRST_PRIMITIVE_REGION = doubling_before_rescale`

---

## Experiment C — exact baseline T2

Independent formal E2 copy:

```text
C0 = E2
C1 = MulRelin(C0, C0)
C2 = Add(C1, C1)
C3 = Add(C2, -1)
C4 = Rescale(C3)
```

Oracle:

`C4_expected[i] = 2 * E2_actual[i]^2 - 1`

This must reproduce the prior T2 failure. If it does not, stop with:

`BASELINE_T2_NOT_REPRODUCED`.

If A and B pass but C fails:

`FIRST_PRIMITIVE_REGION = pre_rescale_constant_injection`

---

## Experiment D — reorder control: Rescale before subtracting 1

This is a diagnostic control, **not** a proposed production fix.

Independent formal E2 copy:

```text
D0 = E2
D1 = MulRelin(D0, D0)
D2 = Add(D1, D1)
D3 = Rescale(D2)
D4 = Add(D3, -1)
```

Oracle:

`D4_expected[i] = 2 * E2_actual[i]^2 - 1`

Same fixed threshold.

Interpretation:

- if B passes, C fails, and D passes, this is strong evidence that scalar `-1` itself is not generally broken; rather, injecting it at the high pre-Rescale scale is the failing condition.
- do not change production operation order in this task.

---

## q0/q1 centered-CRT capacity accounting

For the **actual formal E2/T2 path**, record exactly:

- q0
- q1
- `Q01 = q0*q1`
- `Q01_half = floor(Q01/2)`
- `log2(Q01)`
- `log2(Q01_half)`
- E2 input scale `Delta`
- `log2(Delta)`
- post-MulRelin scale `Delta2`
- `log2(Delta2)`
- scale after Rescale
- actual divisor modulus used by the first T2 Rescale

Then evaluate the representability of the `-1` scalar added at C3.

Fast scalar Add encodes `-1` at the current ciphertext scale. Therefore compute:

`encoded_minus_one = -round(Delta2)`

and record:

`abs(encoded_minus_one) < Q01/2 ?`

Also record:

`capacity_ratio = abs(encoded_minus_one) / (Q01/2)`

This comparison must use integer / high-precision arithmetic, not float64 for the decision.

If `abs(encoded_minus_one) >= Q01/2`, record:

`SCALAR_MINUS_ONE_CENTERED_CRT_UNIQUE = false`

This is a mathematical capacity failure for reconstructing that scalar from q0/q1 alone; it is not merely an empirical numerical mismatch.

Also record the q0 and q1 residues of `encoded_minus_one` and its centered-CRT representative under Q01, but do not confuse that representative with the original integer when uniqueness is false.

---

## Low-scale control — same semantic values, same level, safe nominal capacity

Construct a **diagnostic-only** ciphertext representing the same decoded E2 slot values at the same logical input Level as formal E2, but with:

`diagnostic input scale = 2^45`

Use ordinary CKKS Encode to encode the E2 actual values, then construct a degree-1 zero-secret diagnostic ciphertext with:

- c0 = encoded plaintext
- c1 = 0
- same Level as formal E2
- same LogSlots / batching metadata
- q0/q1 maintained as required by Fast

Run exact T2 order:

```text
L1 = MulRelin(L0, L0)
L2 = Add(L1, L1)
L3 = Add(L2, -1)
L4 = Rescale(L3)
```

Oracle:

`L4_expected[i] = 2 * E2_actual[i]^2 - 1`

Record the same metrics and capacity accounting.

This is only a causal control. It does not replace the formal Mod1 profile.

Expected interpretation rule:

- if formal scale path fails while the 2^45 control passes, and formal `encoded_minus_one` violates Q01/2 while low-scale `encoded_minus_one` does not, record strong evidence for a q0/q1 capacity boundary.
- if low-scale also fails, do not attribute the formal failure solely to scale capacity.

Do not sweep multiple scales in this task.

---

## Classification

Use the earliest supported classification below.

### Case 1 — Experiment A fails

`FIRST_SUPPORTED_CAUSE = square_mulrelin_or_rescale`

If formal capacity accounting also shows the pre-Rescale scale or required represented values exceed q0/q1 uniqueness bounds, additionally record:

`Q01_CAPACITY_INVOLVED = true`

but do not yet choose between MulRelin and Rescale implementation.

### Case 2 — A passes, B fails

`FIRST_SUPPORTED_CAUSE = doubling_before_rescale`

### Case 3 — A/B pass, C fails, D passes

`FIRST_SUPPORTED_CAUSE = pre_rescale_constant_q01_capacity`

Require the formal `-1` encoded magnitude to violate Q01/2 for this exact classification.

### Case 4 — A/B/C all pass

`FIRST_SUPPORTED_CAUSE = diagnostic_inconsistency`

Stop and reconcile with the prior T2 evidence.

### Case 5 — C and D both fail while A/B pass

`FIRST_SUPPORTED_CAUSE = scalar_add_or_downstream_rescale_unresolved`

Do not guess further.

The low-scale control is supporting causal evidence and must be reported alongside the selected case.

---

## Required artifacts

Write temporary output to `/tmp` while any Secondary diagnostic file exists.

After Secondary is restored clean, copy only final result artifacts into Primary:

- `results/EXP-002C-DIAG-T2-logN13-fast.json`
- `results/EXP-002C-DIAG-T2-logN13-summary.json`

Summary must include:

- exact provenance
- E2 reproduction metrics
- A/B/C/D full comparison metrics
- low-scale control metrics
- metadata before/after every primitive
- q0/q1/Q01 capacity accounting
- exact formal `encoded_minus_one` bound decision
- selected `FIRST_SUPPORTED_CAUSE`
- whether q0/q1 capacity is mathematically implicated
- confirmation that prior T2 failure was reproduced
- final Primary/Secondary clean-state evidence

---

## Validation

Before completion:

- exact formal E2 reproduction passes
- prior T2 failure is reproduced
- all required experiments executed unless an execution error prevents them
- no threshold is loosened
- no production Lattigo source is committed
- temporary Secondary files removed
- Secondary exact HEAD remains `ce79b861...`
- Secondary worktree clean
- Secondary clean `go test ./...` passes
- Primary `go test ./...` passes
- Primary results committed and pushed fast-forward
- LogN16 not run
- EXP-003 not started

---

## Non-goals

Do not:

- fix Rescale
- fix MulRelin
- change active-limb architecture
- change Mod1LogScale
- change q moduli
- reorder production T2
- introduce q2 as a production fix
- benchmark performance
- run LogN16
- start EXP-003

This task ends when the first T2 primitive region and the q0/q1 centered-CRT capacity relationship are supported by evidence.
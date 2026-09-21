# FIX-001-P3-DIAG-P93-DOUBLEANGLE-PLAINTEXT-CAUSAL-DECOMPOSITION

## Purpose

The preceding final-restore causal proof at Primary commit
`58fd5ecc2f4ac4063da23736895fe3ed7ddc7c42`
rejected the hypothesis that the final maintained multiplier itself is the main bug.

Accepted findings:

- DA2 post-Rescale Fast-vs-Genuine-Standard canonical error:
  - real `4.084980061155408e-6`
  - imag `4.161680802892462e-6`
- final internal error:
  - real `0.004183019582623534`
  - imag `0.004261561142162278`
- ratio is essentially exactly:
  [
  1024=2^{10}.
  ]
- runtime scale transition is approximately:
  [
  2^{60}ightarrow2^{50}.
  ]
- Genuine Standard itself performs the final restore by assigning the caller input Scale after DoubleAngle.

Therefore the ~1024 amplification is part of the common final Scale contract and is **not by itself evidence of a Fast bug**.

The `2^17` counterfactual restore factor was experimentally rejected:
- internal error only improved to about `0.00376/0.00377`;
- public error remained about `0.120`;
- exact E2E remained about `0.1875`.

Do not pursue restore-factor repair.

The stable semantically trustworthy checkpoints from the existing diagnostics are the polynomial output and the **post-Rescale** DoubleAngle states:

Real context:
- polynomial output: `4.986373945969902e-7`
- DA0 post-Rescale: `1.5490788308758496e-6`
- DA1 post-Rescale: `3.6151065426759388e-6`
- DA2 post-Rescale: `4.084980061155408e-6`

Imag context:
- polynomial output: `5.07761187318323e-7`
- DA0 post-Rescale: `1.5784306544031068e-6`
- DA1 post-Rescale: `3.6833666752222882e-6`
- DA2 post-Rescale: `4.161680802892462e-6`

This task must determine whether the growth across DoubleAngle is:

1. the mathematically expected propagation of the Fast polynomial-output semantic error through the correct DoubleAngle recurrence; or
2. additional error introduced by the Fast DoubleAngle implementation.

No production repair in this task.

---

## Repository safety

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- committed HEAD:
  `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- expected dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact branch/HEAD/fingerprint match.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the dirty Secondary worktree.

No Secondary source modification, commit, or push is authorized.

---

## Fixed controls

Authoritative:

- Genuine Standard exact E2E:
  `5.830057349387463e-8`
- C2S Fast-vs-Standard:
  ~`1e-13`
- public target:
  `1e-2`
- internal final pre-public target:
  `3.125e-4`
- final restore scale amplification:
  `1024`
- therefore required DA2 post-Rescale error budget:
  [
  3.125e-4 / 1024 = 3.0517578125e-7.
  ]

No q/planScale tuning.

---

# R0 — retire invalid pre-Rescale causal metrics

Explicitly mark as non-causal:

- pre-DA raw/coherent-scale comparisons;
- DA square pre-Rescale canonical metrics;
- DA after-constant pre-Rescale metrics

from the preceding logical trace, because their virtual-coordinate interpretation was not validated across maintained-scalar/Rescale boundaries.

Use only:
- polynomial output;
- DA0 post-Rescale;
- DA1 post-Rescale;
- DA2 post-Rescale;
- final internal/public endpoints

for this task's causal classification.

---

# R1 — establish actual semantic vectors

For real and imag, obtain decoded semantic vectors for:

## Genuine Standard
- polynomial output
- DA0 post-Rescale
- DA1 post-Rescale
- DA2 post-Rescale

## Fast
- polynomial output
- DA0 post-Rescale, canonicalized with the validated post-Rescale virtual exponent
- DA1 post-Rescale, canonicalized
- DA2 post-Rescale, canonicalized

Prove the endpoints reproduce the accepted aggregate metrics.

Do not commit full vectors; keep them in memory or `/tmp`.

---

# R2 — build the plaintext DoubleAngle oracle

Use the exact runtime Mod1 parameters.

For each round define the Standard mathematical recurrence from the source:

[
M_r(x)=2x^2-c_r,
]

where the round constant (c_r) is derived exactly from the runtime `Sqrt2Pi` update used by Genuine Standard.

Do not hardcode constants if they can be read from parameters.

Apply the recurrence elementwise in complex plaintext space.

Validate the oracle:

1. apply (M_0) to Genuine Standard polynomial-output decoded values;
2. compare with Genuine Standard DA0 post-Rescale decoded values;
3. apply (M_1), compare with Standard DA1 post-Rescale;
4. apply (M_2), compare with Standard DA2 post-Rescale.

Required oracle fidelity:
- target <= `1e-10` max component;
- if decode/CKKS noise prevents this, report the measured floor and require it to be negligible relative to `3.0517578125e-7`.

If the plaintext oracle does not reproduce Standard semantic evolution sufficiently, stop:

`P93_DOUBLEANGLE_PLAINTEXT_ORACLE_MISMATCH`.

---

# R3 — pure propagation of Fast polynomial error

Starting only from the **Fast polynomial-output decoded semantic vector**, apply the validated plaintext recurrence:

[
P_0=M_0(F_{poly})
]
[
P_1=M_1(P_0)
]
[
P_2=M_2(P_1).
]

At each round compare:

## Upstream-propagation error

[
P_r - S_r
]

where (S_r) is Genuine Standard actual post-Rescale semantic output.

Record:
- max component
- max complex
- mean complex
- worst slot/component
- amplification from previous stage.

This represents what would happen if the DoubleAngle implementation were mathematically perfect and only the Fast polynomial-output error were present.

---

# R4 — Fast DoubleAngle implementation effect

At each round compare:

[
F_r-P_r
]

where:
- (F_r) is actual canonical Fast post-Rescale semantic output;
- (P_r) is the plaintext-oracle result from R3.

This isolates Fast DoubleAngle implementation error from propagated polynomial error.

Also verify vector closure:

[
(F_r-P_r)+(P_r-S_r)=F_r-S_r.
]

Target closure residual:
<= `1e-10`, or a justified numerical floor.

For every round report:
- propagation contribution;
- implementation contribution;
- observed total;
- implementation fraction of observed total;
- closure residual.

Do not compare only norms; use vector differences before reducing to metrics.

---

# R5 — causal attribution at DA2

At DA2 post-Rescale, classify contribution dominance.

Use these rules:

## Polynomial-error dominated

If:
- plaintext-oracle Standard validation passes;
- closure passes;
- (|F_2-P_2|) is <=10% of (|F_2-S_2|);
- propagated polynomial error accounts for >=90% of total DA2 error.

## Fast-DA implementation dominated

If:
- implementation effect >50% of total DA2 error.

## Mixed

Otherwise.

Record whether the total DA2 error:
[
le 3.0517578125e-7
]
or still fails the pre-final-restore budget.

---

# R6 — polynomial-output budget implication

The current polynomial-output Fast-vs-Standard error is approximately `5e-7`.

Using the validated plaintext recurrence, numerically determine the maximum polynomial-output perturbation scale that would keep DA2 post-Rescale error within:

[
3.0517578125e-7.
]

Do this by local linearization / measured directional scaling around the actual error vector, not an open-ended parameter sweep.

Required:
- scale the actual Fast-minus-Standard polynomial error vector by a small bounded set of factors sufficient to estimate the local gain;
- no ciphertext reruns;
- plaintext-only diagnostic.

Report:
- empirical DA gain from polynomial-output error to DA2;
- implied polynomial-output error budget.

This tells the next task how far PS/polynomial accuracy must improve if DoubleAngle implementation is not the blocker.

---

# R7 — source audit

Read-only source audit:

## Standard
`circuits/ckks/mod1/mod1_evaluator.go`

Confirm:
- recurrence is `MulRelin -> Add(res,res) -> Add(-sqrt2pi) -> Rescale`;
- final restore is Scale assignment only.

## Fast
current dirty:
`circuits/ckks/mod1/fast.go`

Confirm:
- maintained-scalar operations are intended to implement the same plaintext recurrence;
- identify any differences relevant to post-Rescale semantics.

Do not infer a repair unless R4 proves implementation effect.

---

# Decision classification

Choose exactly one:

## A — `P93_DOUBLEANGLE_PLAINTEXT_ORACLE_MISMATCH`

Plaintext recurrence does not reproduce Genuine Standard DoubleAngle closely enough.

## B — `P93_DOUBLEANGLE_ERROR_PROPAGATION_DOMINATED`

Fast DoubleAngle actual post-Rescale semantics closely follow the plaintext oracle; DA2 error is >=90% explained by propagation of polynomial-output error.

Next task should return to PS/polynomial accuracy.

## C — `P93_DOUBLEANGLE_IMPLEMENTATION_DIVERGENCE`

Fast DA implementation effect is >50% of DA2 total error.

Next task should localize the first DA round with significant implementation effect.

## D — `P93_DOUBLEANGLE_MIXED_BLOCKER`

Both polynomial propagation and Fast DA implementation contribute materially.

## E — `P93_DOUBLEANGLE_ALREADY_WITHIN_PRE_RESTORE_BUDGET`

DA2 total error is <= `3.0517578125e-7`, contradicting current endpoint accounting.

## F — `P93_DOUBLEANGLE_DECOMPOSITION_ALIGNMENT_INVALID`

Vector semantics or recurrence alignment cannot support the decomposition.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-DOUBLEANGLE-PLAINTEXT-CAUSAL-DECOMPOSITION-summary.json`

Keep <=300 pretty-printed JSON lines.

Include:
- provenance;
- R0 retired-metric statement;
- Standard plaintext-oracle validation;
- per-round propagation/implementation/total decomposition;
- closure residuals;
- DA2 attribution;
- pre-restore budget result;
- implied polynomial-output budget;
- source audit;
- one classification A-F.

No full vectors, no coefficient arrays, no repeated large row tables.

---

## Validation

Run:
- focused plaintext-oracle test;
- vector-closure test;
- per-round decomposition test;
- directly affected Primary tests;
- `go test ./...`;
- `git diff --check`.

On success, commit and push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

---

## Prohibitions

- no Secondary source modification
- no Secondary commit/push
- no destructive operation on dirty Secondary
- no production repair
- no q/planScale tuning
- no public Scale-contract modification
- no P92/P94 sweep
- no C2S/S2C/finalizer work
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

The deliverable is a causal decomposition of DoubleAngle error growth, not a fix.

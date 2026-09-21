# FIX-001-P3-DIAG-P93-FINAL-MAINTAINED-RESTORE-FACTOR-CAUSAL-PROOF

## Purpose

The preceding logical-semantics trace at Primary commit
`469c24171b0b480830eac52dc7ec4a5814f1059d`
fixed one earlier checkpoint-pairing defect, but its classification
`D = P93_LOGICAL_DOUBLEANGLE_ROUND0_DIVERGENCE`
is not accepted.

The reason is that several pre-Rescale logical metrics remain nonphysical:

- polynomial output error ≈ `5e-7`;
- DA0 square/after-constant raw/canonical diagnostics appear near `0.6`;
- immediately after Rescale the error falls back to ≈ `1.5e-6`.

A semantics-preserving Rescale should not erase a true ~0.6 plaintext divergence by six orders of magnitude. Therefore those intermediate pre-Rescale virtual-coordinate comparisons remain unsuitable for causal classification.

However, the same trace exposed a much stronger endpoint invariant:

### Real branch

- DA2 post-Rescale canonical error:
  `4.084980061155408e-6`
- internal final-restore error:
  `0.004183019582623534`

Ratio:

[
0.004183019582623534 / 4.084980061155408e-6
= 1024.000000000097
]

### Imag branch

- DA2 post-Rescale canonical error:
  `4.161680802892462e-6`
- internal final-restore error:
  `0.004261561142162278`

Ratio:

[
0.004261561142162278 / 4.161680802892462e-6
= 1024.000000000096
]

Thus the final internal error is the DA2 post-Rescale error amplified almost exactly by:

[
1024 = 2^{10}.
]

The current Fast restore sequence is source-described as:

1. virtual exponent `e = 27`;
2. current raw Scale ≈ `2^60`;
3. `MulIntegerMaintained(2^27)`;
4. metadata Scale reset to caller input Scale `2^50`.

This sequence is suspicious because preserving the virtual logical value while changing Scale from approximately `2^60` to `2^50` requires a physical factor satisfying:

[
F^* = 2^e cdot rac{S_{out}}{S_{in}}.
]

With:

- (e=27),
- (S_{in}approx 2^{60}),
- (S_{out}=2^{50}),

the required factor is approximately:

[
F^* approx 2^{17},
]

not (2^{27}).

The current (2^{27}) factor is larger by approximately:

[
2^{10}=1024,
]

matching the observed error amplification exactly.

This task must prove or reject this final-restore factor hypothesis with a diagnostic-only counterfactual.

No Secondary production source modification is allowed.

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

Treat these as authoritative:

- Genuine Standard exact E2E:
  `5.830057349387463e-8`
- Fast-vs-Genuine C2S:
  ~`1e-13`
- public threshold:
  `1e-2`
- internal pre-public budget:
  `3.125e-4`
- current final internal Fast-vs-Standard:
  - real `0.004183019582623534`
  - imag `0.004261561142162278`
- public/internal ratio:
  exactly `32`
- public DefaultScale reset is part of the Genuine Standard contract and must remain unchanged.

No q/planScale tuning.

---

# R0 — explicitly reject the prior DA0 causal classification

Record compactly:

- prior classification `D`;
- why it is not accepted;
- the pattern:
  - polynomial output ≈ `5e-7`
  - pre-Rescale DA metrics ≈ `0.6`
  - DA0 post-Rescale ≈ `1.5e-6`;
- statement that pre-Rescale virtual-coordinate metrics are retained only as diagnostic context.

Do not use those ~0.6 values as the causal basis for this task.

---

# R1 — reproduce the final-restore boundary exactly

Run the current Fast normalized LogN13 path source-faithfully through:

1. DA2 post-Rescale;
2. current final maintained-scalar materialization;
3. caller-input Scale restoration.

For real and imag record at DA2 post-Rescale:

- current virtual exponent (e);
- raw Scale (S_{in});
- caller/input Scale (S_{out});
- Level/Degree;
- raw decoded metric vs Genuine Standard corresponding DA2 post-Rescale;
- canonical logical metric using the validated post-Rescale interpretation;
- Fast q01/q012 capacity.

Then record after current final restore:

- current physical maintained multiplier;
- resulting Scale;
- Fast-vs-Genuine Standard internal metric.

Require the observed ratio:

[
E_{after}/E_{before} approx 1024
]

for both real and imag.

Tolerance:
relative error <= `1e-9`.

If not reproduced, classify:

`P93_FINAL_RESTORE_1024_REPLAY_CONFLICT`.

---

# R2 — derive the mathematically required materialization factor

Do not hardcode `2^17` first.

Compute:

[
F^* = 2^e cdot rac{S_{out}}{S_{in}}
]

using the actual runtime Scale values.

Record:

- exact (e);
- (S_{in});
- (S_{out});
- (F^*);
- nearest power-of-two exponent:
  [
  k^* = operatorname{round}(log_2 F^*)
  ]
- relative error between (F^*) and (2^{k^*}).

Expected:
[
k^* approx 17.
]

Also record the current implementation exponent:
[
k_{current}=27.
]

Then derive expected over-amplification:

[
G = rac{2^{k_{current}}}{F^*}.
]

Expected:
[
G approx 1024.
]

Compare this predicted (G) to the observed R1 error ratio.

---

# R3 — diagnostic-only corrected final restore

Starting from the exact same DA2 post-Rescale Fast ciphertext clone:

## Control

Apply the current source-faithful restore:
- current maintained multiplier;
- current Scale assignment.

## Counterfactual

Do not modify Secondary.

In Primary diagnostic code only:

1. apply the nearest mathematically required power-of-two maintained multiplier (2^{k^*});
2. set Scale to the original caller/input Scale;
3. do not change any other metadata;
4. do not alter any other coefficient operation.

If the exact non-power-of-two (F^*) cannot be applied by the maintained integer primitive, use (2^{k^*}) and report the approximation floor.

Compare counterfactual Fast internal final vs Genuine Standard internal final:

- real max component;
- imag max component;
- max complex;
- mean complex;
- worst slot/component;
- Level/Scale/Degree.

Primary internal acceptance:
[
E_{internal} le 3.125e-4.
]

Also report whether the corrected result is consistent with the DA2 post-Rescale error floor instead of being amplified by 1024.

---

# R4 — coefficient/metadata causal proof

Prove what changes between control and counterfactual:

- same DA2 post-Rescale input ciphertext;
- same Level/Degree;
- same virtual exponent before restore;
- only the final maintained multiplier differs;
- both end at the same caller input Scale;
- no unrelated coefficient manipulation.

Record row hashes:

- before restore;
- after current restore;
- after counterfactual restore.

Do not claim metadata-only correction: this counterfactual intentionally changes the physical maintained multiplier.

The causal claim must be:

> the current final maintained multiplier is inconsistent with the simultaneous virtual exponent and Scale transition.

---

# R5 — public EvalMod consequence under the real Standard contract

After each internal final result, apply the **unchanged** public DefaultScale reset exactly as Genuine Standard does.

Record:

## Current control public error

Expected context:
- real ≈ `0.1338566266`
- imag ≈ `0.1363699565`

## Counterfactual public error

Compare vs Genuine Standard public EvalMod.

Acceptance:
[
E_{public}le1e-2.
]

Verify public/internal ratio remains the common contract factor (~32), rather than changing the public API semantics.

---

# R6 — full LogN13 exact-E2E counterfactual

If R3 and R5 pass, continue the counterfactual through the unchanged production path:

1. public DefaultScale reset;
2. Fast SlotsToCoeffs;
3. UnpackAndSwitchN2ToN1;
4. standard-compatible finalization/public boundary;
5. decode against original deterministic message.

Do **not** preserve (2^{50}) through public finalization. Use the Genuine Standard public contract.

Record exact E2E:

[
E_{max}=
max_i
{
|Re(hat m_i-m_i)|,;
|Im(hat m_i-m_i)|
}.
]

Compare:

- current production E2E;
- counterfactual E2E;
- Genuine Standard E2E.

Primary system milestone:
[
E_{max}le1e-2.
]

If the counterfactual reaches <= `1e-2`, this is sufficient evidence to authorize a later minimal Secondary repair task.

---

# R7 — source audit

Read current dirty Secondary source, read-only.

Record exact file/function/lines for:

- initialization/update of virtual exponent;
- DA2 post-Rescale current Scale;
- final `MulIntegerMaintained`;
- final caller-input Scale assignment.

Explain whether source currently computes the restore multiplier as:

- only (2^e); or
- (2^e) adjusted by Scale ratio.

State the smallest candidate production repair formula, but do not implement it.

Expected conceptual repair candidate:

[
F_{restore}approx 2^e rac{S_{out}}{S_{in}}.
]

Do not assume integer/power-of-two implementation details until source audit confirms them.

---

# Decision classification

Choose exactly one:

## A — `P93_FINAL_RESTORE_1024_REPLAY_CONFLICT`

The 1024 amplification does not reproduce.

## B — `P93_FINAL_RESTORE_FACTOR_HYPOTHESIS_REJECTED`

Derived factor does not explain the observed amplification, or counterfactual fails to improve internal semantics.

## C — `P93_FINAL_RESTORE_FACTOR_CAUSAL_INTERNAL_PASS`

Corrected restore factor reduces internal Fast-vs-Genuine Standard EvalMod error to <= `3.125e-4`, but public or full E2E still fails `1e-2`.

Record next failing boundary.

## D — `P93_FINAL_RESTORE_FACTOR_CAUSAL_PUBLIC_PASS_E2E_FAIL`

Internal and public EvalMod pass, but exact full E2E remains > `1e-2`.

## E — `P93_FINAL_RESTORE_FACTOR_CAUSAL_AND_1E2_SUFFICIENT`

All hold:

1. observed current amplification is ~1024;
2. derived restore factor predicts that amplification;
3. counterfactual internal EvalMod <= `3.125e-4`;
4. counterfactual public EvalMod <= `1e-2`;
5. exact LogN13 E2E <= `1e-2`.

If E is selected, the next task may implement the smallest Secondary production repair plus regression validation.

## F — `P93_FINAL_RESTORE_FACTOR_ALIGNMENT_INVALID`

The DA2 post-Rescale logical state or restore boundary cannot be validated reliably enough for causal use.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-FINAL-MAINTAINED-RESTORE-FACTOR-CAUSAL-PROOF-summary.json`

Keep it compact and pretty-printed:
- <= 300 lines;
- no minified one-line JSON;
- no full vectors;
- no coefficient arrays;
- no repeated row tables.

Include:

- provenance;
- R0 rejection of old classification;
- R1 observed 1024 amplification;
- R2 factor derivation;
- R3 counterfactual internal metrics;
- R4 causal-change proof;
- R5 public metrics;
- R6 exact E2E;
- R7 source audit;
- one classification A-F.

---

## Validation

Run:

- focused final-restore factor causal-proof test;
- current-vs-counterfactual restore test;
- public 32x contract test;
- exact E2E counterfactual test;
- directly affected Primary tests;
- `go test ./...`;
- `git diff --check`.

On success, commit and push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

---

## Prohibitions

- no Secondary source modification;
- no Secondary commit/push;
- no destructive operation on dirty Secondary;
- no q/planScale tuning;
- no public Scale-contract modification;
- no C2S/S2C/finalizer repair;
- no P92/P94 sweep;
- no threshold relaxation;
- no LogN16;
- no Gate4/5;
- no EXP-003;
- no benchmark campaign.

The deliverable is a causal proof or rejection of the final maintained-restore factor hypothesis, not a production fix.

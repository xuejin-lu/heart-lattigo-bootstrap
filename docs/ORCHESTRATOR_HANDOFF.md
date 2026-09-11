# Fast-CKKS Orchestrator Handoff

這份文件是給下一個 ChatGPT／指揮手 AI 的 durable handoff。目的不是保存每一段對話，而是讓新的 orchestrator 能快速恢復：研究目標、固定規範、重要里程碑、已被推翻的舊結論、目前卡點，以及下一步決策方式。

---

## 1. Project identity

Primary repository:

`xuejin-lu/heart-lattigo-bootstrap`

Secondary repository:

`xuejin-lu/lattigo`

Secondary active branch:

`fast-ckks`

Primary 是固定 experiment harness；Secondary 是 CKKS implementation。核心實驗契約：

> 固定 workload、parameters、measurement、frontend；只換 Lattigo implementation/commit。

不要把 Fast/Standard runtime selector、q0/q1 residue policy、zero-secret implementation details 搬進 Primary。

---

## 2. Research goal

目前 Fast-CKKS 是 intentionally insecure、speed-first CKKS backend，用來研究：

- 大型重複 workload（尤其 encrypted CNN / Bootstrap）如何用最小 active residue subset 加速；
- 保留 logical CKKS structure / Level / Scale / API，但只 materialize 真正需要的 physical state；
- Fast 與 Standard 並存並可由同一 frontend 比較。

長期架構原則：

> Preserve logical structure; allocate / compute physical state only when required.

目前 Fast 通常只維護 q0/q1；Level metadata 與 q2...qL logical chain 仍存在，但 dormant rows 不可被 Standard full-RNS path 當成有效資料使用。

目前 zero-secret 是 implementation mode，不是永久 scientific invariant。

---

## 3. Durable Fast semantics relevant to current investigation

- q0/q1 centered CRT uniqueness requires：

  `|x| < q0*q1/2`

  when a subsequent operation reconstructs / interprets the centered integer from only q0/q1.

- current formal q0/q1 roughly：

  - q0 ≈ 2^55
  - q1 ≈ 2^39
  - q0*q1 ≈ 2^94
  - centered capacity ≈ 2^93~2^94 depending exact primes.

- Fast scalar `Add` encodes scalar directly at the ciphertext's current Scale into q0/q1.
- Fast `Rescale` reconstructs a centered q0/q1 representative and divides by the logical q_level divisor. Therefore an intended pre-Rescale integer outside q0/q1 centered uniqueness aliases before division.
- Fast `MulRelin` under current zero-secret semantics drops/relinearizes c2 according to Fast mode; this is not the current known failure source after later diagnostics.

---

## 4. Formal LogN13 Bootstrap profile

The authoritative formal profile used throughout correctness work:

- LogN = 13
- LogDefaultScale = 45
- SecretHamming = 192
- q0 bit-size 55
- SlotsToCoeffs primes 39,39,39
- EvalMod / Mod1 primes around 60-bit × 8
- CoeffsToSlots primes around 56-bit × 4
- P primes around 61-bit × 5
- LogSlots = -1 => 4096 slots
- Mod1LogScale = 60
- degree = 30
- DoubleAngle = 3
- K = 16
- LogMessageRatio = 10
- InvDegree = 0
- circuit order = `ModUpThenEncode`

Deterministic workload:

`real=((i%7)-3)/16; imag=((i%5)-2)/32`

All formal correctness gates use all 4096 slots and threshold `1e-2` unless a spec explicitly states otherwise.

---

## 5. Performance baseline before correctness investigation

After earlier Fast arithmetic / DFT / ScaleDown / ModUp / EvalMod / Packing / Bootstrap work, speed results showed large gains.

Important clean experiment milestones:

- EXP-001-P-FIX:
  - LogN13 ~12.55× Fast vs Standard
  - LogN16 ~20.58×
- EXP-002 stage breakdown:
  - LogN13 Fast full ~24.3ms vs Std ~303.6ms
  - LogN16 Fast ~256.9ms vs Std ~5719ms

These are **execution/performance results only** until formal numerical correctness is restored.

Do not present them as correctness-preserving speedups yet.

---

## 6. Why correctness investigation started

Initial formal dual-decryption check showed:

- Standard generated-secret output matched input with tiny error.
- Fast zero-secret output failed badly; error roughly comparable to original input magnitude.

This proved prior small-unit tests were insufficient for formal Bootstrap correctness.

The investigation then localized the first supported divergence progressively instead of guessing.

---

## 7. Correctness localization history

### 7.1 DIAG-Q01

First direct supported stage failure moved to EvalMod(real):

`Q3_eval_mod_real`

Q1 ModUp passed; Q2 C2S real/imag passed.

### 7.2 DIAG-EVALMOD

Within EvalMod, first failure localized to polynomial evaluation.

### 7.3 Historical polynomial-power diagnostic mistake

An early Primary plaintext Chebyshev oracle incorrectly treated `T0` as zero. It falsely classified T2 / power generation as failing.

This conclusion is **superseded**.

Never use that old replay as evidence.

The independent T2 primitive diagnostic remained valid.

### 7.4 Independent T2 primitive diagnostic

For formal T2 = `2*T1*T1 - 1`:

- `MulRelin -> Rescale` passed.
- `MulRelin -> double -> Rescale` passed.
- `MulRelin -> double -> Add(-1) -> Rescale` failed by ~0.9995.
- Reordered `MulRelin -> double -> Rescale -> Add(-1)` passed.
- Low-scale control passed.

Root cause:

`pre_rescale_constant_q01_capacity`

The encoded `-1` at the high pre-Rescale scale exceeded q0/q1 centered capacity by ~1.34e8.

Conclusion: Standard Chebyshev operation order is not automatically valid in the two-limb Fast representation.

### 7.5 FIX-001 first repair

Production changed Chebyshev correction to occur after Rescale.

This repaired T2, but formal polynomial still failed.

Corrected oracle then showed:

- T2 passed ~4.88e-4
- T3 failed ~0.03123

### 7.6 T3 localization

T3 split `(1,2,1)`.

Failure occurred before correction, after multiply/double/rescale.

Capacity diagnostic proved T1*T2 multiplication itself produced a coefficient outside q0/q1 centered uniqueness before Rescale, while a full-Q oracle survived the same operation.

Root cause:

`t3_multiply_crosses_q01_capacity`

### 7.7 P2 one-sided pre-Rescale idea failed conceptually

Attempted one-sided pre-Rescale:

`Rescale(one operand) -> multiply`

Codex inserted a heuristic gate because unconditional one-sided pre-Rescale would collapse an operand scale from ~2^60 to ~1.

The formal T3 therefore bypassed the repair and reproduced the old error.

Important lesson:

> Do not merely remove the heuristic. The original one-sided design itself is wrong for formal scale.

### 7.8 Balanced two-sided pre-Rescale feasibility

Diagnostic tested:

`left' = Rescale(m1*left)`

`right' = Rescale(m2*right)`

then multiply, with `m1*m2 ≈ q_L`.

Formal validated factors:

`m1 = m2 = 2^30`

Results:

- operand preservation errors around 1e-7 / 1e-9
- no q0/q1 product capacity violation
- product error ~1.84e-7
- T3 correction error ~1.84e-7

Classification:

`balanced_pre_rescale_validated`

### 7.9 P3 production balanced scheduling

Secondary production commit:

`61607bb4bb82591009ce768d9a3773bed1497565`

This replaced the incomplete one-sided P2 logic with balanced two-sided scheduling for eligible high-scale generated powers, while retaining low-scale compatibility path.

Formal Gate 1–2:

- T2 ~5.75e-9
- T3 ~1.84e-7
- T1/T2/T3/T4/T6/T8/T16 all passed.

Therefore **power generation is no longer the known formal correctness blocker**.

Gate 3 whole degree-30 polynomial still failed:

~`0.7909176408709179`

So the first remaining cause moved into Paterson–Stockmeyer (PS) polynomial accumulation.

---

## 8. PS localization history

### 8.1 PS first-divergence diagnostic

Primary result commit:

`8222fd9f6addc479bf87626ee4c24efdc9287fd9`

Actual PS plan:

- degree 30
- base 8
- reversal order `[4,3,2,1,0]`

Blocks 0 and 1 passed.

First failure:

`baby_block=2 constant_add coefficient=0`

Local error:

`0.03665425827255296`

The block-2 constant is about:

`0.03665425157863994`

Accumulator Scale is ~`2^120`.

Nominal encoded magnitude:

~`4.87e34`

while `Q01/2` is ~`9.90e27`, giving:

`nominal_over_half ≈ 4.919651e6`.

Actual centered q0/q1 residue is of course wrapped back inside the interval, so simply checking `outside_half_count` on residues is insufficient. The intended integer itself is non-unique.

Classification:

`baby_constant_add`

Important interpretation:

This is not safely fixed by special-casing only block-2 constant because the same ~2^120 baby accumulator scale also controls subsequent non-integer coefficient terms.

### 8.2 Compressed internal PS scale diagnostic

Primary result commit:

`605d864d02b87793530e6bf53e3d67cd800c7503`

Tested internal scales:

`2^91, 2^90, 2^89, 2^88, 2^87, 2^86`

All prechecks passed.

Minimum coefficient encoding precision remained ~2^31 to 2^26, all above fixed 2^20 diagnostic floor.

All five baby blocks passed for every candidate; max baby error ~1.5e-8.

This proves:

> Lower q0/q1-safe internal PS scales can fix the baby-step scalar-capacity problem without destroying coefficient precision.

However all candidates stopped at the first giant step (`T8`) because after:

`Rescale(b) -> Mul(b,T8)`

both local arithmetic oracles passed but the actual resulting Scale did not satisfy production's very strict `Scale.InDelta(a.Scale, ScalePrecision-12)` gate.

Example at candidate 2^91:

- a.Scale = `2.475880078570760549798248448e27`
- b.Scale = `2.47588007856713654229323309002828985927e27`

Relative mismatch is ~1.46e-12, around 2^-39.3.

Production gate effectively expects around 2^-116 relative closeness.

Classification:

`compressed_giant_steps_fail`

Important: giant Rescale oracle passed and Mul(T8) oracle passed. The current blocker is scale bookkeeping / alignment, not yet a demonstrated arithmetic-value failure.

---

## 9. Current task at handoff

Always confirm current remote `CURRENT_TASK.md` because this section can become stale.

At the time this handoff was written, Primary `CURRENT_TASK.md` points to:

`specs/FIX-001-P3-DIAG-PS-scale-normalization.md`

The task is diagnostic only.

Hypothesis:

After a semantically-correct compressed giant-step product, change **metadata Scale only** from the actual b.Scale to sibling a.Scale when drift is tiny and explicitly bounded, then continue PS.

The diagnostic must test complete:

`baby blocks -> all normalized giant steps -> final Rescale -> target-scale restoration -> whole degree-30 polynomial`

No production tolerance relaxation yet.

No LogN16 / benchmark / EXP-003.

---

## 10. Current authoritative Secondary state

At handoff, authoritative Fast production commit is:

`61607bb4bb82591009ce768d9a3773bed1497565`

Do not assume a newer commit until GitHub remote is checked.

This commit productionized balanced pre-Rescale power scheduling.

All later compressed-scale / normalization work so far is diagnostic-only in Primary; Secondary must remain exact and clean unless a later production repair spec explicitly changes it.

---

## 11. Review protocol for Codex completion reports

When user pastes a Codex report:

1. Do not approve from text alone.
2. Fetch the claimed Primary/Secondary commit.
3. Verify remote HEAD / branch / files changed.
4. Inspect production diff if any.
5. Inspect result JSON/summary directly.
6. Check methodology matches the spec, especially:
   - corrected `T0=1` oracle;
   - exact formal LogN13 profile;
   - actual P3 generated powers / real production path;
   - first-failing-gate stop rule;
   - no hidden q2+/full-Q fallback;
   - no parameter/threshold cheating.
7. Distinguish:
   - GitHub independently verified facts;
   - Codex-reported local test/worktree facts not independently observable.
8. If evidence passes, do not invent extra fix.
9. If a real gap exists, write the next narrow diagnostic/repair spec in GitHub and update `CURRENT_TASK.md`.

---

## 12. Orchestrator workflow

Preferred cycle:

1. User pastes Codex result.
2. ChatGPT independently reviews GitHub.
3. ChatGPT identifies one supported conclusion.
4. If repair is not yet causally justified, create smallest diagnostic spec.
5. If repair is justified, create bounded production repair spec with explicit gates.
6. Update `CURRENT_TASK.md`.
7. Tell user only: `開始`.
8. Codex syncs safely, executes, tests, commits/pushes.
9. Repeat.

Avoid long ad-hoc instructions in chat when they belong in repository specs.

---

## 13. Things that are explicitly superseded

Do not resurrect these conclusions without new evidence:

- “T2 power generation fundamentally fails” from the old `T0=0` oracle — invalid.
- “unpack/switch is the root cause” — only first directly comparable checkpoint at the time, later superseded by earlier EvalMod localization.
- “finalization / Scale restore causes the full Bootstrap failure” — ruled out by earlier DIAG-P.
- “one-sided pre-Rescale is the correct P2 repair if heuristic is removed” — false; it collapses precision.
- “current whole polynomial failure is still caused by generated powers” — false under P3; all formal powers now pass.
- “block-2 constant can be fixed safely with a one-off special case” — unsupported and likely incomplete.
- “compressed internal PS scales are infeasible” — false; all baby blocks passed. The current question is giant-step scale normalization / alignment.

---

## 14. New-chat startup recommendation

A new ChatGPT should first read:

1. `NEW_CHAT_START.md`
2. `AGENTS.md`
3. this file
4. `docs/DEVELOPMENT_LESSONS.md`
5. `CURRENT_TASK.md`
6. current task spec
7. Secondary `AGENTS.md`
8. Secondary `docs/FAST_CKKS_SPEC.md`

Then independently verify remote HEADs before giving advice.

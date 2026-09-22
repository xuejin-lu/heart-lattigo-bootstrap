# LogN13 / P93 Scale Semantics — Static & Mathematical Audit

Secondary under audit:

`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Frozen Fast exact E2E:

[
E_{max}=0.009705381393898434.
]

This document records the orchestrator-led **static + mathematical** audit before any runtime measurement task. It does not authorize a Secondary change.

## Scale taxonomy

| Type | Meaning |
|---|---|
| M | Metadata-only reinterpretation |
| P | Physical scalar with matching metadata scalar |
| R | CKKS Rescale |
| V | Virtual/deferred scalar |
| L | Planner/logical scale |
| A | Approximate alignment/rounding |
| C | Public Standard-compatible scale contract |

## Static ledger

| Boundary | Source operation | Type | Static conclusion | Status |
|---|---|---:|---|---|
| Pack / Unpack | Fast packing helpers | — | No explicit Scale mutation found; metadata is propagated through packing/ring-degree operations. | PASS_STATIC |
| ScaleDown level drop | cheap level resize | — | Same logic as Standard; no coefficient/Scale reinterpretation beyond dropping unused levels. | PASS_STANDARD_CONTRACT |
| ScaleDown scale-up | `scaleUp.BigInt(); MulIntegerMaintained; Scale *= integer` | P/A | `Scale.BigInt()` rounds to nearest. Fast code matches Standard. Semantic preservation is exact only for the rounded integer, not the pre-rounded real ratio. Runtime mismatch must be measured. | WARNING_RUNTIME_ROUNDING |
| ScaleDown RescaleTo | coefficient division by dropped Q + Scale division | R | Fast follows Standard RescaleTo contract. | PASS_STANDARD_CONTRACT |
| ModUp basis extension | centered q0 lift into maintained limbs | — | No Scale change. | PASS_STATIC |
| ModUp message lift | physical `scalar=round(scale)`; metadata `Scale *= scale` | P/A | **Standard uses the same code pattern.** Not Fast-specific. Physical/metadata ratio is `round(scale)/scale`; runtime mismatch must be quantified. | WARNING_STANDARD_ROUNDING |
| C2S compressed matrix | matrix Scale divided by `2^k`, k={4,2} | L | Same mathematical matrix encoded at lower scale for bounded-capacity Fast path. This changes quantization, not plaintext transform semantics. Existing C2S Fast-vs-Standard error is ~1e-13. | PASS_NUMERICAL_FLOOR |
| C2S restore plan | `MulIntegerMaintained(2^k); Scale *= 2^k` | P | Exact power-of-two physical + metadata scaling, hence decoded semantics invariant. | PASS_EXACT |
| DFT group Rescale | `Rescale` once per factor group | R | Same level/scale progression as Standard DFT. | PASS_STANDARD_CONTRACT |
| EvalMod normalization | `res.Scale = ScalingFactor()` | M | Coefficients unchanged. **Standard performs the identical reinterpretation.** This is the x mod 1 normalization contract. | PASS_STANDARD_CONTRACT |
| EvalMod targetScale | Q-index recurrence with square roots | L | Fast computes the same target-scale schedule as Standard source. | PASS_STANDARD_CONTRACT |
| P93 planScale | `planScale=2^93`; override PS plan node scales | L | Not a public ciphertext reset. It changes planned intermediate scales and coefficient encoding through PS evaluation. Historical/final P93 polynomial evidence validates the resulting arithmetic, but exact alignment ratios still need runtime audit. | PASS_WITH_RUNTIME_ALIGNMENT_CHECK |
| Generated powers | post-product recurrence then one Rescale | R/L | T2/T3/T4/T6/T8/T16 end near 2^60 and match exact Chebyshev oracle at ~1e-15. | PASS_NUMERICAL_FLOOR |
| PS add/sub alignment | `Scale.Div(...).BigInt()`; integer physical multiply; metadata set to target | A/P | `BigInt()` rounds to nearest. Semantic ratio is `round(r)/r`. Need actual executed ratios; static code alone cannot prove they are integral. | WARNING_RUNTIME_ROUNDING |
| PS giant-step snap | after `InDelta(...,32)`, assign `b.Scale=a.Scale` | M/A | Coefficients unchanged; relative metadata snap is statically bounded by about 2^-32, but actual executed deltas must be measured. | WARNING_RUNTIME_SNAP |
| Polynomial output | final Rescale | R | Final repaired P93 polynomial Scale ~2^33; Fast-vs-Genuine Standard error real 1.857e-8 / imag 1.622e-8. | PASS_NUMERICAL_FLOOR |
| Mod1 coherent scale | `res.Scale = res.Scale * 2^k`, k=27 | M/V | Metadata-only reinterpretation paired with explicit virtual exponent. Static algebra is consistent: logical value = raw value × 2^k at this coordinate. | PASS_MATHEMATICAL |
| DA virtual recurrence | `a = 1 + 2e - e_next`; physical `2^a`; constant divided by `2^e_next`; Rescale | V/P/R | Algebraically implements `y_next = 2 y^2 - c` in deferred coordinate: raw result represents logical result / 2^e_next. | PASS_MATHEMATICAL |
| DA exponent selection | nearest power-of-two exponent; source asserts k=27 and multiplier exponent=28 | V/A | Exact chosen exponent is guarded; closeness of coherent/target and next/working scales is checked by `InDelta(...,32)`. Runtime deltas should be recorded. | PASS_WITH_RUNTIME_DELTA_CHECK |
| Mod1 final materialization | physical `2^27`, then `res.Scale=inputScale` | V/C | Materializes deferred scalar, then performs the same Standard x-mod-1 “multiply-back” metadata contract. This transition is **not semantic-invariant by design**. Stable endpoint matches Genuine Standard within the accepted internal error. | PASS_STANDARD_CONTRACT |
| EvalMod public reset | `Scale=BootstrappingParameters.DefaultScale()` | C | Fast and Standard source are identical here. The earlier “2^50→2^45 Fast bug” interpretation is permanently rejected. | PASS_STANDARD_CONTRACT |
| S2C DFT | group LT + Rescale | R | Same structural progression as Standard; no Fast restore plan on S2C. | PASS_STANDARD_CONTRACT |
| Final public bootstrap reset | `Scale=ResidualParameters.DefaultScale()` | C | Fast and Standard both explicitly set the same residual DefaultScale. Final Fast and Standard metadata are both 3.5184372088832e13 (~2^45). | PASS_STANDARD_CONTRACT |

## Important mathematical clarifications

### 1. `Scale.BigInt()` does not truncate

Lattigo implements:

[
operatorname{BigInt}(Delta)=lfloor Delta+0.5floor.
]

Therefore the relevant alignment error is:

[
epsilon=rac{operatorname{round}(r)}{r}-1,
]

not a floor/truncation error.

### 2. ModUp rounding is Standard behavior

Both Fast and Genuine Standard compute a real ratio through `Float64()`, physically multiply by:

[
s=operatorname{round}(r),
]

but update metadata with (r).

Thus:

[
m_{after}=rac{s}{r}m_{before}.
]

This is a real numerical contract, but not a Fast-specific divergence.

### 3. P93 planScale is not the ciphertext public scale

[
S_{plan}=2^{93}
]

is a PS planning/intermediate scale. It is distinct from:

- Mod1 `ScalingFactor`;
- Mod1 `targetScale`;
- generated-power Scale (~2^60);
- polynomial output Scale (~2^33);
- public DefaultScale (~2^45).

### 4. Virtual DoubleAngle recurrence is algebraically closed

If raw decoded value is:

[
x_{raw}=rac{x}{2^e},
]

Fast chooses:

[
a=1+2e-e'
]

and computes the raw recurrence with a constant divided by (2^{e'}), so before/after Rescale:

[
x'_{raw}=
rac{2x^2-c}{2^{e'}}.
]

The deferred factor is therefore explicit, not an accidental hidden scale.

## Static findings requiring runtime measurement

Only the following remain unresolved from source inspection:

1. **ScaleDown rounded ratio**
   - real `scaleUp`
   - rounded `BigInt()`
   - relative mismatch.

2. **ModUp Float64/round contract**
   - exact 128-bit inputs to the ratio;
   - Float64 ratio;
   - rounded integer scalar;
   - `round(r)/r-1`;
   - measured semantic distortion.

3. **Executed PS add/sub alignment ratios**
   - all `Scale.Div(...)` values reaching `BigInt()`;
   - integer multiplier;
   - relative mismatch;
   - maximum observed semantic residual.

4. **Executed PS giant-step metadata snaps**
   - pre-snap `a.Scale/b.Scale`;
   - `Log2Delta`;
   - implied semantic interpretation shift.

5. **Normalized Mod1 exact delta ledger**
   - polynomial output Scale;
   - targetScale;
   - coherentScale/targetScale ratio;
   - each DA nextScale/workingScale ratio;
   - exact exponent chosen.

6. **C2S restore runtime confirmation**
   - matrix compression exponents {4,2};
   - physical restore factors exactly match metadata restore factors.

## Current static classification

`STATIC_SCALE_AUDIT_NO_CAUSAL_DEFECT_FOUND_RUNTIME_ROUNDING_CHECKS_REQUIRED`

No Secondary modification is justified from static analysis.

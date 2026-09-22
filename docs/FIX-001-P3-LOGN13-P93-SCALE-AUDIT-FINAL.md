# LogN13 / P93 Scale Semantics — Final Audit

Secondary:

`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Frozen Fast exact E2E:

[
0.009705381393898434
]

Genuine Standard exact E2E:

[
5.830057349387463e-8
]

## Final classification

`SCALE_AUDIT_ALL_CONTRACTS_VERIFIED`

No causal scaling-factor defect was found in the finalized LogN13/P93/Q012 candidate.

The observed scale deviations are either:

- exact semantic invariants;
- numerical-floor effects far below the current and next precision goals; or
- explicit contracts shared with Genuine Standard.

No Secondary scaling change is authorized by this audit.

## Runtime findings

| Area | Runtime evidence | Judgment |
|---|---:|---|
| ScaleDown target ratio | ideal ratio `1.9999999999913598...`; rounded integer `2`; relative target mismatch `4.3200859556522646e-12`; semantic residual `0` | PASS_STANDARD_CONTRACT |
| ModUp scale lift | exact ratio `16`; Float64 ratio `16`; scalar `16`; semantic residual `0` | PASS_EXACT |
| PS add/sub alignment | 8 executed alignments; every ratio exactly `1`; max rounding mismatch `0`; semantic residual `0` | PASS_EXACT |
| PS metadata snaps | worst relative Scale difference `2.927449949114433e-12`; worst semantic residual `2.6065816172149425e-12` | PASS_NUMERICAL_FLOOR |
| C2S restore | restore exponents `4,2,0,0`; physical and metadata factors identical; residual `0` | PASS_EXACT |
| Final Fast / Standard public Scale | both `3.5184372088832e13` | PASS_STANDARD_CONTRACT |

## Normalized Mod1 ledger

For both real and imag:

- EvalMod input Scale:
  [
  2^{50}=1.125899906842624e15
  ]
- `ScalingFactor()`:
  approximately (2^{60})
- `planScale`:
  [
  2^{93}=9.903520314283042...e27
  ]
- polynomial output Scale:
  [
  8.5899345920012207e9
  ]
  which is (2^{33}) with relative deviation about (1.42e-13)
- `kIn=27`
- coherent Scale:
  [
  1.1529215046070108e18
  ]
- coherent/target ratio:
  [
  1+1.0125071354e-13
  ]

Thus:

[
S_{poly}cdot2^{27}approx S_{target}approx2^{60}.
]

### DoubleAngle

All three rounds choose:

[
e=27,qquad a=28,qquad 2^a=268435456.
]

The post-Rescale Scale remains close to (2^{60}).

Relative drift from exact (2^{60}):

- coherent entry: about `1.42e-13`
- DA0 post-Rescale: about `1.56e-13`
- DA1 post-Rescale: about `3.98e-13`
- DA2 post-Rescale: about `8.10e-13`

These deviations are many orders below the active precision target and do not place the nearest-power-of-two exponent selection near a boundary.

### Virtual-scalar semantics

The Fast normalized recurrence is algebraically consistent.

If raw decoded state represents:

[
x_{raw}=x/2^e,
]

then using:

[
a=1+2e-e'
]

with the constant encoded as (c/2^{e'}) gives:

[
x'_{raw}=rac{2x^2-c}{2^{e'}}.
]

The measured runtime values use (e=e'=27), hence (a=28), exactly matching source.

### Final Mod1 / public contracts

Fast materializes (2^{27}), then resets metadata to the caller EvalMod Scale (2^{50}).

Genuine Standard performs the equivalent internal-to-caller Scale contract at the same logical boundary.

The subsequent public EvalMod reset to residual DefaultScale (2^{45}) is also shared with Standard.

Therefore the earlier interpretation that (2^{50}ightarrow2^{45}) was a Fast-specific scale bug is definitively rejected.

## What does *not* explain the current 1e-2 error

The audit specifically rules out the following as material causes of the current Fast error:

- ScaleDown integer rounding;
- ModUp Float64/round scalar selection;
- PS `BigInt()` scale alignment;
- PS metadata snapping;
- C2S restore factors;
- Mod1 virtual-exponent selection;
- public DefaultScale resets.

The current (1e-2) error must therefore be attributed to approximation/arithmetic error elsewhere, not an unresolved Scale contract.

## Test-suite maintenance note

Full Primary `go test ./...` currently fails at:

`TestFIX001P3GenuineStandardPublicVsStagedConsistency`

with:

`P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT`.

This is **not a Standard numerical replay failure**.

That historical runner checks, before running the Standard comparison, that Secondary is the old pre-finalization dirty handoff state. The finalized Secondary is now committed at `40532b4d...` and clean, so the runner exits through its stale provenance guard.

Focused scale measurements, compile-only full package tests, and finalized exact-E2E replay pass.

This stale historical test should be treated as Primary diagnostic maintenance debt, separate from scale correctness.

## Final decision

- Scaling-factor audit: complete.
- Secondary scale repair: not justified.
- Precision-tightening phase: not started.
- Frozen (1e-2) milestone remains valid.

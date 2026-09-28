# QPREFIX-IMPL-009 — Performance Review Result

## Status

`QPREFIX_V2_PERFORMANCE_REVIEW`

Production candidate:
`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

Historical comparison point:
`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

## Release gates

Accepted from the QPREFIX-IMPL-009 execution report:

- focused regression: PASS;
- full `go test ./... -count=1`: PASS;
- generated-secret Standard comparator: PASS;
- zero-secret ordinary decrypt/decode: PASS;
- full-slot real/imag: PASS;
- sparse/repacked: PASS;
- odd-count BootstrapMany: PASS;
- N1=N2 and factor-two boundary: PASS;
- C2S strict capacity: PASS;
- EvalMod/PS/DoubleAngle strict capacity: PASS;
- no Standard/full-RNS/private-F production fallback found;
- public output contract: PASS.

Final production Level/authority trace:

| Stage | Logical Level | Authoritative Q prefix |
|---|---:|---|
| Public input | 0 | q0 |
| ModUp | 16 | q0123 |
| C2S group outputs | 15,14,13,12 | q0123 |
| EvalMod output | 4 | q0123 |
| S2C group | 3 | q0123 |
| S2C group | 2 | q012 |
| S2C group | 1 | q01 |
| Public output | 1 | q01 |

This satisfies the constitution:

[
w_Q(ell)=min(ell+1,4).
]

## Performance summary

Environment reported:
- Apple M4;
- Go 1.26.4;
- darwin/arm64;
- GOMAXPROCS=10.

### Historical q0=55 matched public-API harness

Count 1:
- baseline: 30.670 ms;
- candidate: 116.390 ms;
- candidate / baseline: 3.79x.

Count 3:
- baseline: 89.891 ms;
- candidate: 348.166 ms;
- candidate / baseline: 3.87x.

### Current q0=56 production P93

Fast count 1:
- 114.846 ms;
- 7,373,272 B/op;
- 12,901 allocs/op.

Fast count 3:
- 322.244 ms;
- 22,119,570 B/op;
- 38,695 allocs/op.

Standard count 1:
- 303.035 ms;
- 91,892,148 B/op;
- 34,635 allocs/op.

Standard / Fast count-1 latency ratio:

[
303.035/114.846 approx 2.64.
]

Therefore Fast remains materially faster than Standard, but the historical matched-baseline regression is substantial and above measurement noise.

## Critical interpretation of the historical baseline

The q0=55 historical benchmark must **not** be described as a simple q012 -> q0123 implementation comparison.

At `40532b4...`, the legacy row selector is:

`MaintainedLimbCount = 3` only when `bits.Len64(q0) == 56`.

For the matched q0=55 harness, ordinary high-Level legacy Fast arithmetic therefore uses:

[
q01
]

rather than the Q-prefix-v2 production policy:

[
q0123.
]

Thus the approximately 3.8x result measures the cost of migrating from the historical narrow/profile-dependent Fast policy to the fixed Q-prefix-v2 constitution, plus any implementation overhead.

It does **not** by itself prove that the four-row implementation is inefficient.

The current release classification remains:

`QPREFIX_V2_PERFORMANCE_REVIEW`

because:
- all hard semantic/structural/capacity gates pass;
- Fast is faster than Standard;
- matched historical regression is <10x but clearly substantial.

## Next action

Run a performance-attribution task before any production optimization.

Separate:

1. expected width cost (legacy q01/q012 -> q0123);
2. non-linear fixed-width CRT/Rescale cost;
3. polynomial/PS workspace and copy cost;
4. scalar-guard cost;
5. allocation/scratch effects;
6. public packing/ring-degree cost.

No architecture reopening and no production optimization should occur until attribution identifies the dominant terms.

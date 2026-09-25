# FAST-INTEGRATION-002 — Fused Level-0 ModUp Bridge Optimization

## Status

Executable Codex task.

## Task class

`I — Implementation`

This task changes no CKKS mathematics. It optimizes the already accepted FAST-INTEGRATION-001 production seam.

Authoritative architecture:
- Secondary `docs/FAST_CKKS_SPEC.md`
- especially Sections 4.9, 4.10, 4.10.1, 4.11, 4.14.

If implementation appears to require changing canonicalization semantics, Logical-Q metadata, maintained-row policy, Scale behavior, or downstream Bootstrap contracts, stop with `NEEDS_WEB_REVIEW`.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- orchestration/spec only.

Secondary:
- `xuejin-lu/lattigo@fast-ckks`
- implementation target.

Accepted prerequisite:
- FAST-INTEGRATION-001 at `31efadc693559217b48d3e76a2e3655b9e6cd14d`
- fusion architecture clarification at `dc698e2d99a488f0de2cf4f3207b09ef94521303`

## Motivation

The accepted production bridge is correct but slow.

Reported LogN13 basis-raise measurements:

- Fast bridge: ~4.50 ms/op
- Standard: ~2.94 ms/op
- Fast: ~87,263 allocs/op
- Standard: 46 allocs/op

The current path physically performs:

[
q_0	ext{-NTT}
	o
F_3
	o
F_3	ext{ canonical ModUp}
	o
compact LogicalQ
]

even though no arithmetic operation observes the transient private-F state.

The accepted semantics prove that at this specific seam:

[
C=operatorname{Center}_{q_0}(r)
]

is already determined directly from the Level-0 logical residue and the private-F ModUp step is idempotent on that canonical lift.

Therefore the production bridge may be fused to compute maintained logical rows directly from the same canonical lift while retaining the unfused private-F path as a semantic oracle.

---

# 1. Target production kernel

Replace only the internal implementation of the production basis-raise seam with a fused path:

[
q_0	ext{-NTT}
	o
q_0	ext{-coeff}
	o
C=operatorname{Center}_{q_0}(r)
	o
(q_0,q_1[,q_2])	ext{-coeff}
	o
(q_0,q_1[,q_2])	ext{-NTT}.
]

Then preserve the existing scale-alignment logic exactly.

Do **not** physically materialize transient `FastCiphertext/F_3` state in the production hot path.

The semantic reference remains:

[
ImportLevel0(width=3)
	o
FastStorageModUpLevel0
	o
ExportToCompactLogical.
]

The fused output must match that reference exactly on all maintained logical rows and metadata.

---

# 2. API and placement

Prefer a small CKKS-Fast helper in `schemes/ckks/fast` that can be called by Bootstrap without exposing private implementation details.

Conceptually:

`FusedLevel0ModUpToCompactLogical(params, source, targetLevel, targetDomain)`

or equivalent.

Exact naming is a coding choice.

Required behavior:
- input ordinary Level-0 LogicalQ ciphertext;
- accepted canonical centered q0 convention;
- explicit target logical Level;
- output compact logical-Q ciphertext with only maintained rows materialized;
- preserve metadata and degree;
- non-Montgomery input/output;
- support NTT input and NTT output required by production;
- coefficient-domain support is optional if not needed by production, but may be added if it simplifies testing.

Do not delete or weaken:
- `ImportLevel0`
- `FastStorageModUpLevel0`
- `ExportToCompactLogical`

They remain the semantic oracle and standalone foundation.

---

# 3. Canonicalization rule

For odd `q0=2m+1` and canonical residue `r in [0,q0)`:

[
C=
egin{cases}
r,&0le rle m,\
r-q_0,&m+1le r<q_0.
end{cases}
]

Therefore:

[
r=q_0>>1=m
]

maps to the positive representative `+m`.

This rule is mandatory.

Do not restore historical `>= q0>>1` behavior.

---

# 4. Maintained logical rows only

At target logical Level `L`:

[
m = MaintainedLimbCount(params,L).
]

Allocate and compute only:

[
q_0,dots,q_{m-1}.
]

All logical rows:

[
q_m,dots,q_L
]

must remain structurally present but nil/unmaterialized according to the compact Fast ciphertext policy.

No full-RNS output.

---

# 5. Fixed-width-3 architectural relation

The initial production storage policy remains private-F width 3.

This fused seam is permitted because the transient F state is unobservable here.

Do not reinterpret this optimization as:
- removing private-F architecture;
- reducing the permanent private storage width;
- permitting logical q rows to contain f residues;
- bypassing private-F in later stages that actually execute F arithmetic.

Later production stages that perform private-F arithmetic must still use physically separate width-3 F storage.

---

# 6. Allocation/performance design

The main acceptance concern is removal of the pathological allocation behavior.

Required implementation direction:

- do not construct a new `fastStorageBasis` in the fused hot path;
- do not allocate three private-F polynomials;
- do not NTT/INTT any private-F rows;
- use one logical q0 coefficient buffer per reusable workspace/component strategy;
- allocate only the compact maintained logical output and minimal scratch;
- no per-coefficient `math/big`;
- use fixed-width/uint64 centered reduction directly from q0 residue;
- use logical subring NTT only for maintained output rows.

A reusable evaluator-local scratch/workspace is allowed if ownership and concurrency assumptions are explicit and compatible with current evaluator usage.

Do not introduce unsafe shared mutable global scratch.

---

# 7. Production integration

Update `FastEvaluator.modUpBasis` to call the fused kernel instead of:

- `ImportLevel0`
- `FastStorageModUpLevel0`
- `ExportToCompactLogical`

in the hot path.

Preserve existing:
- validation;
- scale-alignment behavior;
- target logical Level;
- NTT output;
- ordinary non-Montgomery output before `FastEvaluator.ModUp`;
- downstream Trace and Montgomery conversion.

No downstream migration.

---

# 8. Exact semantic oracle

For every focused correctness fixture, compute both:

## Reference

[
ImportLevel0(width=3)
	o
FastStorageModUpLevel0(targetLevel)
	o
ExportToCompactLogical(targetDomain).
]

## Fused

new fused helper.

Verify exact equality for:
- all maintained logical rows;
- logical Level;
- Scale;
- degree;
- NTT/domain metadata;
- plaintext metadata.

The reference path should remain in tests only, not production hot path.

---

# 9. Required tests

## T1 — fused vs unfused q01

For q01 maintained profile, test:
- positive residues;
- negative centered residues;
- zero;
- `q0>>1 - 1`;
- `q0>>1`;
- `q0>>1 + 1`.

Require exact maintained-row equality with the unfused private-F reference.

## T2 — fused vs unfused q012

Repeat T1 for q012 maintained profile.

## T3 — metadata equivalence

Verify exact equality of:
- target logical Level;
- Scale;
- degree;
- IsNTT;
- IsMontgomery;
- IsBatched;
- IsBitReversed;
- LogDimensions;
- parameter identity/value.

## T4 — no full-RNS materialization

At high MaxLevel:
- maintained rows have N backing;
- all dormant rows nil.

## T5 — input immutability

The fused helper must not mutate the Level-0 input.

Snapshot q0 row and metadata before call.

## T6 — production modUpBasis equivalence

Compare production `modUpBasis` after optimization against the unfused private-F reference plus existing scale alignment.

Cover q01 and q012.

## T7 — full ModUp boundary

Run existing `FastEvaluator.ModUp` tests:
- Trace succeeds;
- output NTT;
- output Montgomery;
- maintained rows equivalent.

## T8 — Bootstrap regressions

Run supported Fast Bootstrap tests with unchanged correctness thresholds.

## T9 — standalone private-F regressions

All storage/import/ModUp/compact-export tests remain passing.

## T10 — full regressions

Run:
- `go test ./schemes/ckks/fast`
- `go test ./circuits/ckks/bootstrapping`
- `go test ./...`
- `git diff --check`
- gofmt check.

---

# 10. Benchmark acceptance

Re-run the same LogN13 basis-raise benchmark pair used for FAST-INTEGRATION-001.

Record at least three runs each:

- Fast fused
- Standard

Report:
- ns/op
- B/op
- allocs/op

Also compare Fast fused against the accepted Integration-001 baseline:

Baseline Fast:
- ~4.50 ms/op
- ~3.35 MB/op
- ~87,263 allocs/op

## Required acceptance gates

Correctness gates above are mandatory.

Performance gates:

1. `allocs/op` must decrease by at least **100x** versus Integration-001 baseline:
   [
   allocs_{new} le 873.
   ]

2. `B/op` must be materially reduced versus Integration-001 baseline.

3. `ns/op` must improve versus the ~4.50 ms Integration-001 baseline.

No requirement yet to beat Standard 2.94 ms/op in this task.

If correctness passes but any performance gate fails:
- do not silently accept;
- report `NEEDS_WEB_REVIEW` with profile evidence.

If the fused path becomes faster than Standard, report that clearly but do not widen scope.

---

# 11. Profiling if gate fails

If allocation or time gates fail after one bounded repair pass, collect a lightweight profile or allocation source breakdown sufficient to identify the dominant remaining cause.

Do not begin a broad optimization campaign.

Stop with `NEEDS_WEB_REVIEW`.

---

# 12. Prohibitions

Do not:
- change CKKS mathematics;
- modify accepted canonical midpoint semantics;
- delete private-F standalone APIs;
- migrate DFT/EvalMod/S2C;
- modify production Rescale;
- implement KeySwitch/Relinearize/Rotate;
- add contraction/adaptive width;
- add frontend flags;
- materialize full logical RNS;
- add Standard/full-RNS fallback;
- weaken correctness tests;
- broaden into general memory-pool refactoring.

---

# 13. Success classification

Candidate classification:

`FAST_INTEGRATION_002_FUSED_MODUP_BRIDGE_ACCEPTED_CANDIDATE`

Successful handoff:

`READY_FOR_WEB_REVIEW`

## Completion report

Report:
- exact Secondary commit;
- changed files;
- fused helper/API;
- proof-by-test of exact equivalence to unfused private-F reference;
- production `modUpBasis` change;
- regression results;
- three-run Fast/Standard benchmark table;
- change versus Integration-001 baseline in time, B/op, allocs/op;
- whether all performance gates passed;
- confirmation that downstream Bootstrap stages and standalone private-F semantics were not changed.

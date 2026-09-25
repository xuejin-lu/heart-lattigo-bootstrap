# FAST-INTEGRATION-003 — Private-F Resident ModUp + Scale Alignment + Normalized Trace

## Status

Executable Codex task.

## Task class

`I — Implementation`

This task is the first production stage in which a private-F ciphertext remains physically resident across real arithmetic after ModUp.

Authoritative architecture:
- Secondary `docs/FAST_CKKS_SPEC.md`
- especially Sections 4.9, 4.10.2, 4.11–4.15.

If implementation requires changing Trace semantics, CKKS Scale semantics, canonicalization, fixed-width-3 policy, or downstream C2S contracts, stop with `NEEDS_WEB_REVIEW`.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- orchestration/spec only.

Secondary:
- `xuejin-lu/lattigo@fast-ckks`
- implementation target.

Accepted prerequisites:
- FAST-INTEGRATION-002 at `57ffb88744c82778c0a9392ecab394e19f712a3d`
- private-F normalized Trace theorem at `8f823fdb9464c2738f71c30d156ce574098d8605`

---

# 1. Purpose

Replace the current production sequence

[
	ext{fused ModUp to compact LogicalQ}
	o
	ext{logical scale alignment}
	o
	ext{logical Trace}
]

with a private-F resident sequence:

[
	ext{Level-0 LogicalQ}
	o
	ext{ImportLevel0}(F_3)
	o
	ext{logical ModUp metadata/canonical boundary}
	o
	ext{integer scale alignment in }F_3
	o
	ext{normalized Trace in }F_3
	o
	ext{compact LogicalQ export}
	o
	ext{existing Montgomery conversion}.
]

The private-F state must remain authoritative from import through Trace.

Do not migrate C2S/EvalMod/S2C in this task.

---

# 2. Input/output production contract

Input to the resident segment:
- ordinary LogicalQ ciphertext;
- logical Level 0;
- degree 1;
- NTT;
- non-Montgomery;
- positive Scale;
- Standard ring.

Private production state:
- fixed storage width 3;
- NTT;
- non-Montgomery;
- logical target Level = Bootstrap MaxLevel.

Output after resident Trace:
- ordinary compact LogicalQ ciphertext;
- logical Level = Bootstrap MaxLevel;
- degree 1;
- NTT;
- non-Montgomery;
- Scale exactly matching current production ModUp semantics;
- only maintained logical q rows materialized.

The existing `FastEvaluator.ModUp` may then perform its current Montgomery conversion exactly as before.

---

# 3. Integer scale alignment over private-F

Current production ModUp computes:

[
s =
rac{	ext{Mod1 ScalingFactor}/	ext{MessageRatio}}
     {	ext{ciphertext Scale}}
]

and when `s > 1`, uses:

[
m=operatorname{round}(s)
]

then multiplies the ciphertext by integer `m` and updates:

[
Delta'=Deltacdot s
]

using the same existing metadata behavior.

Implement a private-F integer-scalar primitive, conceptually:

`FastStorageMulInteger(op *FastCiphertext, scalar *big.Int) (*FastCiphertext, error)`

or a minimal equivalent API.

Required semantics:

[
Y=mX.
]

Bound:

[
B'_j=|m|B_j.
]

Requirements:
- fixed width 3 in the production path;
- transactional / non-mutating preferred;
- NTT and coefficient domains may be supported, but production requires NTT;
- residue computation must use `m mod f_i`;
- no per-coefficient reconstruction;
- no logical-Q materialization;
- reject if the proven output bound violates:
  [
  2B'_j<S_3.
  ]

The production metadata Scale update must remain exactly where current ModUp semantics place it. Do not silently change Scale to `Delta*m` if current code records `Delta*s`; preserve existing behavior.

---

# 4. Normalized private-F Trace

Implement standalone private-F normalized Trace, conceptually:

`FastStorageTraceNormalized(op *FastCiphertext, logN int) (*FastCiphertext, error)`

Production requirements:
- Standard ring;
- degree 1;
- width 3;
- NTT;
- non-Montgomery;
- logical Level >= 1.

Let `g` be the exact current Lattigo Trace normalization gap:

[
g=2^{operatorname{LogN}-operatorname{logN}-1},
]

with the existing extra factor two when `logN=0`.

Let:

[
mathcal T_g(X)
]

be the unnormalized automorphism sum using the exact same Galois-element schedule as current Fast/Standard Trace.

Compute:

[
Y=rac{mathcal T_g(X)}{g}.
]

## 4.1 Intermediate bound gate

For every component:

[
B^{sum}_j=gB_j.
]

Before executing the Trace sum require:

[
oxed{2gB_j<S_3}.
]

This is mandatory because the unnormalized sum must remain uniquely reconstructible throughout the private-F accumulation.

If the condition fails:
- return an explicit capacity error;
- do not fall back to compact LogicalQ;
- production integration must propagate the failure;
- Codex must report whether the supported Bootstrap profile ever hits this condition.

## 4.2 Automorphisms

Automorphism in NTT private-F rows is a coefficient-index permutation using the same:

`ring.AutomorphismNTTIndex(N, NthRoot, galEl)`

schedule.

For each active private row independently:
- permute NTT coefficients;
- add into the private-F accumulator modulo that row's `f_i`.

No evaluation key is used, matching current Fast Trace semantics.

The implementation may cache automorphism indices/workspace in a private-F evaluator/workspace if scoped safely to a single execution stream, or use bounded local scratch.

Do not use ordinary logical-Q ring operations on F rows.

## 4.3 Final exact normalization

Do not pre-multiply by `g^{-1}`.

After the complete unnormalized Trace sum, normalize each private row with:

[
g^{-1}mod f_i.
]

Because:
- exact integer divisibility by `g` is proven;
- each `f_i` is odd;
- the unnormalized integer lift is unique under the bound gate,

this residue-wise multiplication must represent exact integer division.

Post-bound may conservatively remain:

[
B'_j=B_j.
]

A tighter exact observed bound is not required.

Preserve:
- logical Level;
- Scale;
- degree;
- NTT domain;
- width 3;
- plaintext metadata.

---

# 5. Production integration

Refactor only the `FastEvaluator.ModUp` boundary enough to make private-F resident across:
- canonical ModUp;
- scale alignment;
- normalized Trace.

Preferred high-level production flow:

1. validate existing Level-0 ModUp input;
2. `ImportLevel0(params, ct, 3, NTT)`;
3. `FastStorageModUpLevel0(private, MaxLevel)`;
4. if scale alignment scalar is required:
   - private-F integer multiply;
   - preserve existing Scale metadata update;
5. `FastStorageTraceNormalized(private, logSlots)`;
6. `ExportToCompactLogical(NTT)`;
7. run existing logical-Q Montgomery conversion;
8. return existing public `*rlwe.Ciphertext`.

It is acceptable to introduce a small resident helper that composes steps 2–6.

The already accepted fused ModUp helper from Integration-002 must remain in the repository as:
- a performance/reference primitive;
- a useful isolated boundary;
- regression evidence.

Do not delete it merely because production ModUp now uses a longer F-resident path.

---

# 6. Semantic oracle

Production resident result before Montgomery must match the accepted Integration-002 path:

[
	ext{FusedLevel0ModUpToCompactLogical}
	o
	ext{existing scale alignment}
	o
	ext{existing Fast Trace}.
]

Compare exact maintained logical rows whenever the contracts are exact.

This oracle is especially important because normalized private-F Trace changes the order from:

[
g^{-1}pmod Q quad	ext{then sum}
]

to:

[
	ext{integer sum}quad	ext{then exact divide by }g.
]

The theorem says these are equivalent modulo the logical Q basis under the bounded lift conditions; tests must verify it.

---

# 7. Required tests

## T1 — private-F integer scalar

For positive, negative, zero, and representative large integer scalars:
- exact output lifts;
- exact bound multiplication;
- Scale unchanged inside the primitive itself;
- width remains 3;
- NTT preserved;
- capacity failure is transactional.

Production may restrict actual scale-alignment scalar to positive values, but standalone arithmetic should have a clear signed policy.

## T2 — normalized Trace monomial theorem

Construct coefficient-domain integer polynomials containing monomials:
- that survive Trace;
- that vanish under Trace.

Independently compute the automorphism sum and verify:
- every summed coefficient is divisible by `g`;
- normalized private-F Trace equals the exact integer normalized polynomial.

Use small LogN for transparent fixtures.

## T3 — private-F Trace vs existing logical Trace

For q01 and q012 logical parameter profiles:
1. create a bounded Level-0 logical input;
2. import/canonical ModUp into F;
3. run private-F normalized Trace;
4. export compact logical;
5. compare maintained q rows against existing `FastCKKS.Trace` applied after accepted fused ModUp.

Cover multiple `logN/logSlots`, including:
- gap = 1;
- gap > 1;
- `logN = 0` if valid for fixture.

## T4 — intermediate capacity rejection

Construct a private-F state with valid input bound:

[
2B<S_3
]

but invalid Trace intermediate:

[
2gBge S_3.
]

Require explicit failure before mutation.

## T5 — production scale-alignment + Trace equivalence

Compare:
- old accepted Integration-002 production semantics;
- new resident private-F semantics.

Require exact:
- maintained logical rows before Montgomery;
- Level;
- Scale;
- metadata.

Cover both cases:
- scale alignment scalar applied;
- no scale alignment.

## T6 — complete ModUp boundary

Run full `FastEvaluator.ModUp`:
- output Level unchanged from accepted path;
- NTT true;
- Montgomery true;
- maintained rows equivalent to old accepted oracle;
- dormant rows remain nil.

## T7 — Bootstrap regression

Existing supported Fast Bootstrap tests must pass with unchanged thresholds.

## T8 — standalone foundation regression

All storage arithmetic, Rescale, ModUp, compact export, fused ModUp tests remain passing.

## T9 — full regression

Run:
- `go test ./schemes/ckks/fast`
- `go test ./circuits/ckks/bootstrapping`
- `go test ./...`
- `git diff --check`
- gofmt check.

---

# 8. Performance measurement

Measure the full ModUp boundary, not only basis raise, because Trace is now resident.

Run at least three LogN13 runs each of the existing/full equivalents:

- Fast resident `ModUp`
- accepted pre-003 Fast ModUp reference if benchmarkable
- Standard/reference ModUp

Record:
- ns/op;
- B/op;
- allocs/op.

Also separately benchmark private-F normalized Trace if practical.

There is no mandatory Standard speed ratio in this task.

Required performance acceptance:
- new resident Fast ModUp must not regress by more than 20% versus the accepted Integration-002 Fast full-ModUp baseline measured in the same environment;
- if it is faster, report the gain;
- allocations must not show a new pathological explosion (no >10x increase without explicit explanation).

If the same-environment Integration-002 full-ModUp baseline is not already recorded, benchmark the parent commit or reproduce the old path in a benchmark/reference helper.

If the gate fails after one bounded repair pass, report `NEEDS_WEB_REVIEW`.

---

# 9. Prohibitions

Do not:
- migrate C2S, DFT, EvalMod, S2C;
- modify production Rescale;
- implement KeySwitch/Relinearize;
- change public CKKS parameters;
- add frontend flags;
- contract storage width;
- pre-multiply private-F Trace by modular `g^{-1}`;
- hide a capacity failure with LogicalQ/full-RNS fallback;
- reinterpret F rows as q rows;
- delete Integration-002 fused helper;
- weaken existing Bootstrap thresholds.

---

# 10. Success classification

Candidate classification:

`FAST_INTEGRATION_003_PRIVATE_F_RESIDENT_MODUP_TRACE_ACCEPTED_CANDIDATE`

Successful handoff:

`READY_FOR_WEB_REVIEW`

Use `NEEDS_WEB_REVIEW` for:
- architecture/math conflicts;
- supported-profile Trace capacity failure;
- performance gate failure after the allowed repair pass.

## Completion report

Report:
- Secondary commit;
- changed files;
- private-F integer-scalar API;
- private-F normalized Trace API;
- exact gap/divisibility implementation;
- bound/capacity checks;
- production ModUp resident-flow change;
- exact oracle equivalence results;
- supported-profile capacity evidence;
- regressions;
- benchmark table;
- performance change versus Integration-002 full ModUp;
- confirmation downstream C2S/EvalMod/S2C/Rescale were untouched.

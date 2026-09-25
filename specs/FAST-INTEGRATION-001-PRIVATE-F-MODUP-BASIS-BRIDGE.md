# FAST-INTEGRATION-001 — Private-F ModUp Basis Production Bridge

## Status

Executable Codex task.

## Task class

`I — Implementation`

The mathematics and architecture are already frozen by Secondary `docs/FAST_CKKS_SPEC.md`.

This task is the first production integration of the accepted private-F foundation. It must change only one bounded production seam: the **basis-raise portion of Fast Bootstrap ModUp**.

If implementation requires changing CKKS mathematics, the fixed-width-3 policy, Trace semantics, ScaleDown semantics, or downstream DFT/EvalMod contracts, stop with `NEEDS_WEB_REVIEW`.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- orchestration/spec only.

Secondary:
- `xuejin-lu/lattigo@fast-ckks`
- implementation target.

Accepted prerequisites:
- FAST-STORAGE-005: `531aca50b5b38741e4e71cc98ea4b626bf88cb84`
- fixed initial production width 3: `d9919f9c080e0dfa731746f5c447f93633ae2f36`

## Purpose

Replace only the historical logical-q0/q1/q2 basis-raise implementation inside:

`circuits/ckks/bootstrapping/fast_modup.go: modUpBasis`

with the accepted private-F path:

[
	ext{Level-0 LogicalQ}
	o
	ext{ImportLevel0}(F_3)
	o
	ext{FastStorageModUpLevel0}
	o
	ext{compact maintained LogicalQ}
]

while leaving the rest of the existing production Bootstrap path unchanged.

The downstream contract remains the historical compact Fast `*rlwe.Ciphertext` representation used by Trace, DFT, EvalMod, packing, and finalization.

This task does **not** migrate DFT/EvalMod arithmetic to `FastCiphertext`.

---

# 1. Production seam

Current:

[
	ext{ScaleDown}
	o
	ext{historical }modUpBasis(q_0	o q_0/q_1[/q_2])
	o
	ext{scale alignment}
	o
	ext{Trace}
	o
	ext{Montgomery}
	o
	ext{C2S/EvalMod/S2C}.
]

Target for this task:

[
	ext{ScaleDown}
	o
	ext{ImportLevel0 width 3}
	o
	ext{FastStorageModUpLevel0(target=MaxLevel)}
	o
	ext{compact logical-Q bridge}
	o
	ext{existing scale alignment}
	o
	ext{existing Trace}
	o
	ext{existing Montgomery}
	o
	ext{unchanged downstream circuit}.
]

Only the basis raise is replaced.

---

# 2. Fixed private storage width

Production private-F state must use:

[
oxed{StorageWidth=3}.
]

Do not invoke adaptive contraction or choose width 1/2.

The historical compact logical-Q output may maintain only the existing required q rows (normally q0/q1, or q0/q1/q2 for the bounded q012 profile). This is distinct from private-F storage width.

Do not confuse:

- private-F width = 3;
- number of maintained logical-Q bridge rows = existing historical `MaintainedLimbCount` policy.

---

# 3. Compact private-F -> logical-Q bridge

The existing generic `FastCiphertext.ExportToLogical` materializes all logical rows through `LogicalLevel`; do **not** use that in the production hot path.

Add a narrow Fast-private conversion helper conceptually equivalent to:

`ExportToCompactLogical(target FastCiphertextDomain) (*rlwe.Ciphertext, error)`

or another minimal name.

Required semantics:

1. input is a valid private-F `FastCiphertext`;
2. output logical Level equals `FastCiphertext.LogicalLevel()`;
3. output Scale/degree/plaintext metadata are preserved;
4. output is ordinary non-Montgomery;
5. output contains N-sized backing only for the historical actively maintained logical rows selected by existing Fast policy;
6. dormant logical rows remain nil/unmaterialized;
7. for each maintained logical row:
   [
   r_i=Xmod q_i;
   ]
8. coefficient/NTT conversion follows the explicit cross-basis rule;
9. no private `f_i` residue may be reinterpreted as logical `q_i`;
10. no full-RNS Standard fallback.

For an NTT target:

[
F	ext{-NTT}
	o
F	ext{-coeff}
	o
X
	o
Xmod q_i
	o
q_i	ext{-NTT}.
]

The helper should use the existing compact Fast ciphertext allocator/policy where practical.

---

# 4. modUpBasis integration

Refactor only enough of `modUpBasis` to use the private-F foundation.

Required sequence:

1. preserve existing input validation;
2. input remains Level 0, degree 1, NTT, ordinary non-Montgomery;
3. call:
   [
   ImportLevel0(..., width=3, target=NTT)
   ]
   so the private lift is exactly the canonical centered q0 representative;
4. call:
   [
   FastStorageModUpLevel0(..., MaxLevel)
   ]
   to keep the canonicalization boundary explicit;
5. export to compact logical-Q NTT form using the new bridge;
6. preserve the current Mod1 scale-alignment behavior exactly;
7. return a compact ordinary NTT `*rlwe.Ciphertext` suitable for the existing Trace boundary.

Do not call generic full `ExportToLogical` here.

Do not alter Trace or Montgomery conversion in `FastEvaluator.ModUp`.

## Input ownership

Prefer non-destructive construction until the private-F conversion and compact export have succeeded.

Whether `modUpBasis` ultimately returns a newly owned compact ciphertext or commits the maintained rows back into the input is a coding choice, provided:
- public semantics are unchanged;
- dormant rows are never read as authoritative;
- no full logical basis is materialized;
- existing caller behavior remains safe.

Do not preserve stale dormant data at the cost of reading or synchronizing it.

---

# 5. Equivalence contract

The authoritative rule is the accepted canonical centered-q0 semantics, not byte-for-byte preservation of every historical boundary convention.

For a canonical Level-0 residue `r in [0,q0)`, let

[
C=\operatorname{Center}_{q_0}(r).
]

For odd `q0 = 2m+1`:

[
C=
\begin{cases}
r,&0\le r\le m,\\
r-q_0,&m+1\le r<q_0.
\end{cases}
]

Therefore `r = q0>>1 = m` maps to the positive representative `+m`.

Historical Fast/Standard ModUp source contains an off-by-one convention in some no-key component paths using `r >= q0>>1`, which maps this single residue to `-(m+1)`. That historical choice is congruent modulo q0 but is not the accepted canonical representative and is not an authority for this integration.

For every maintained logical row the integrated bridge must satisfy:

[
row_i=C\bmod q_i.
]

Equivalence requirements:
- for all non-boundary residues where the historical path and canonical rule agree, maintained rows must match the historical basis-raise result;
- for the single residue `r=q0>>1`, the new path must follow canonical `+m` semantics even if the historical row differs by q0;
- after scale alignment, preserve all unchanged scale/metadata behavior.

The target is semantic correctness on maintained authoritative rows, not preservation of dormant-row garbage, backing-pointer identity, or the historical midpoint off-by-one.

---

# 6. No downstream migration

Do not modify:
- `FastEvaluator.ModUp` Trace logic except any minimal call plumbing required by the new `modUpBasis` result;
- Trace implementation;
- Montgomery conversion semantics;
- DFT evaluator;
- polynomial evaluator;
- EvalMod;
- packing/unpacking;
- Rescale production implementation;
- KeySwitch/Relinearize/Rotate;
- public Bootstrap frontend.

After the compact bridge, downstream execution remains on the existing historical compact logical-Q Fast path.

---

# 7. Required tests

## T1 — compact export structure

Construct private-F ciphertexts at a high logical Level and export compact logical-Q.

Verify:
- logical Level exact;
- degree/Scale/metadata exact;
- NTT/coefficient target behavior correct;
- only `MaintainedLimbCount(params, level)` logical rows have N-sized backing;
- all higher dormant rows are nil/unmaterialized.

## T2 — compact export exactness

For known private lifts, independently compute:

[
Xmod q_i
]

for every maintained row and verify exact coefficient residues.

Repeat for coefficient and NTT private-F inputs.

## T3 — no full-RNS materialization

At a parameter set with MaxLevel much larger than maintained count:
- verify no dormant q row receives N-sized backing;
- verify the bridge does not call/produce generic full logical materialization.

A structural test is required.

## T4 — basis-raise equivalence

For Level-0 inputs with:
- positive values;
- negative centered values;
- q0-half boundary-near values;
- zero;

compare the new `modUpBasis` maintained rows against an independent canonical centered-q0 oracle.

Also compare against the historical basis-raise path for non-boundary residues where both definitions agree.

The fixture must include `q0>>1` and prove that this boundary follows the accepted positive canonical representative rather than the historical `>= q0>>1` off-by-one convention.

Cover q01 and q012 maintained profiles where practical.

## T5 — metadata and scale alignment

Verify the new integrated `modUpBasis` preserves the existing:
- target logical Level;
- NTT domain;
- ordinary representation before `ModUp` Montgomery step;
- degree;
- plaintext metadata;
- scale-alignment result.

## T6 — downstream ModUp contract

Run full existing `FastEvaluator.ModUp` and verify:
- Trace still succeeds;
- output remains NTT;
- output is Montgomery after the existing boundary;
- maintained rows match the existing semantic reference.

## T7 — poisoned/dormant rows

Provide Level-0 inputs whose backing capacity or historical higher rows contain poison/stale values where constructible.

Verify:
- those values are not read as input semantics;
- maintained output depends only on q0 input semantics;
- no requirement is imposed to preserve poison pointer identity.

## T8 — production Bootstrap regression

Run the existing supported Fast Bootstrap targeted tests and ensure the public Bootstrap result remains within its current accepted correctness thresholds.

No threshold may be weakened.

## T9 — standalone storage regressions

Run all private-F storage tests from FAST-STORAGE-001 through 005.

## T10 — full regressions

Run:
- `go test ./schemes/ckks/fast`
- `go test ./circuits/ckks/bootstrapping`
- `go test ./...`
- `git diff --check`

Any new unexplained failure is blocking.

---

# 8. Measurement requirement

This is the first production integration task, so record performance rather than postponing all measurement.

Run at least the existing ModUp-basis benchmarks for the supported large profile, preferably:

- `BenchmarkFastModUpBasisLogN13`
- `BenchmarkStandardModUpBasisLogN13`

and record:
- ns/op;
- B/op;
- allocs/op.

If practical, also run LogN16.

There is **no speed threshold in this task**. The result is diagnostic evidence.

If the new private-F path is materially slower than the previous Fast basis raise, do not revert the architecture or bypass canonicalization. Report the cost so GPT Web can decide whether the next task should fuse:

[
ImportLevel0 + ModUp canonicalization + compact export
]

to remove redundant domain conversions.

---

# 9. Performance constraints

Do not:
- use generic full logical export in the production path;
- materialize dormant q rows;
- add per-coefficient arbitrary-precision work when current fixed-width helpers suffice;
- synchronize stale dormant logical residues;
- add Standard/full-RNS fallback.

One explicit F/logical basis conversion at this integration seam is allowed.

---

# 10. Prohibitions

Do not:
- migrate C2S/EvalMod/S2C to `FastCiphertext`;
- modify production Rescale semantics;
- implement KeySwitch/Relinearize/Rotate;
- implement storage contraction/adaptive width;
- change public CKKS parameters;
- add frontend Fast flags;
- alter Bootstrap correctness thresholds;
- delete the accepted standalone private-F APIs;
- expand into a repository-wide representation refactor.

---

# 11. Success classification

Candidate classification:

`FAST_INTEGRATION_001_PRIVATE_F_MODUP_BRIDGE_ACCEPTED_CANDIDATE`

Successful handoff:

`READY_FOR_WEB_REVIEW`

## Completion report

Report:
- Secondary commit;
- changed files;
- exact compact bridge API;
- evidence no full logical basis is materialized;
- maintained-row equivalence evidence;
- ModUp/Bootstrap regression results;
- storage regression results;
- benchmark numbers;
- whether Import -> ModUp -> compact export shows a measurable optimization opportunity;
- confirmation downstream DFT/EvalMod/packing and production Rescale were not migrated or changed.

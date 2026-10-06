# FAST-QPREFIX-REMOVE-BALANCED-PRESCALE-003 — Delete obsolete operand-side pre-scale

## Decision and priority
Task class: I — **direct Secondary code deletion**, not another experiment or redesign. The user explicitly directs removal of the **entire** legacy balanced/pre-scale mechanism. This task **supersedes and suspends** `FAST-STANDARD-PERF-REBASELINE-003` until cleanup is tested and numerically accepted. Do not time obsolete Fast code or conduct a preliminary A/B test.

## Architecture
In active Secondary `xuejin-lu/lattigo`, branch `fast-qprefix`, baseline `75ef5dbe7bbf7d3947fb2b9fb232c4a56f05c948`, Q-prefix storage remains `w_Q(l)=min(l+1,4)`. The math is independent of LogN-specific paths. The accepted **Chebyshev** power schedule is source-level multiplication -> full recurrence correction -> single logical Rescale on corrected product, with strict pre-Rescale capacity checks. No operand-side integer scaling, independent operand Rescales, or replacement numerical workaround.

## Exact removal inventory, confirmed against baseline source
Only `circuits/ckks/polynomial/fast.go` and the corresponding relevant tests/documentation may be edited. Remove:
- `balancedMinScaleBits`, `balancedScaleMarginBits`, `balancedFactorCandidates`, `balancedScaleToleranceBits`;
- `balancedFactors`, `balancedSchedule`, `integerSqrt`, `balancedFactorPair`, `balancedScheduleFor`;
- `fastPolynomialWorkspace.balancedLeft`, `.balancedRight`, `.balancedCopy`;
- the entire `balanced` decision/branch in `genPowerInternal`, including factor scaling, pre-Rescale of both operands, scale metadata adjustment, and balanced-copy/old-target `Scale.InDelta` logic;
- `TestBalancedFactorPairDeterministic`, `TestBalancedScheduleBranchSelection` and tests referring specifically to deleted scratch fields or the old schedule. Retain and **rename/retarget** any valuable source-immutability or dormant-residue tests so they check the **remaining** generic source-level path without requiring removed scratch fields.
- remove imports that become unused (e.g. `math/big` if used only by the removed code); retain `math/bits`, `errors` or helpers if independently used. Do not delete the generic `copyQPrefixAtLevel` helper unless all remaining call sites, including tests, are accounted for.

## Replacement production behavior, no algorithmic options
At `genPowerInternal`, take one direct Q-prefix source-level `MulElementQPrefixRows` / `MulRelinElementQPrefixRows` path for generated powers regardless of basis, LogN or q0 bit size. Preserve existing degree-2 `lazy` relinearization preconditions, source/operand ownership and correct output buffer. For Chebyshev, apply recurrence correction at product scale, assert exact pre-Rescale strict Q-prefix capacity, then perform exactly one `RescaleQPrefixRows`. For the **internally supported non-Chebyshev/Monomial** code, preserve its mathematical contract and level/scale with a single post-product Rescale, not a legacy factor split. Do not change unrelated Paterson–Stockmeyer polynomial planner behavior. Public unsupported Monomial entry points remain unsupported unless independently authorized.

Keep logical divisor `q_l`, `Level'=l-1`, `Scale'=Scale/q_l`, and `w_Q(l)` including Level 2 -> 3 rows and Level 3 -> 4 rows. Do not touch Standard arithmetic, other CKKS schemes, key semantics, residual parameters, or numerical thresholds. If some intermediate violates `2B<S_Q`, fail loudly with exact checkpoint, coefficient component, B, prefix S_Q and deficit; **do not revive pre-scale**.

## Tests / acceptance
- Secondary: `gofmt`, focused polynomial tests covering T2/T3/T5 (including lazy powers), distinct LogN13/LogN16, Level 2/3 Q012/Q0123, source operand immutability, stale dormant residues, and surviving Monomial internal path; `go test ./circuits/ckks/polynomial ./circuits/ckks/mod1 ./circuits/ckks/bootstrapping ./schemes/ckks/fast -count=1`; `git diff --check`. Run against a clean safely synchronized branch.
- Primary existing `cmd/fastdiag numerical` canonical LogN13/LogN16 using the **new pinned Secondary SHA**, with 2 Fast and 3 Standard trials/profile; report Fast-vs-Standard final RMSE, decoded-domain Standard/Fast SNR, 0.01 coordinate gate, 80 capacity checkpoints including pre-Rescale if applicable. Primary changes limited to necessary SHA pin and compact result summary `results/FAST-QPREFIX-REMOVE-BALANCED-PRESCALE-003-summary.md`.
- Repository-wide source audit within Secondary scoped to **active production references**: no `balanced/pre-scale` symbol in `circuits/ckks/polynomial/fast.go` or tests that depend on the deleted mechanism. Historical commit history and archived background need not be rewritten. Any separate pre-scale code outside this generated-power mechanism must be identified rather than silently deleted.
- Preserve existing `FAST-QPREFIX-CHEBYSHEV-ORDER-002` results as history. No full 7x timing campaign yet.
- The known historical `TestFIX001P3GenuineStandardPublicVsStagedConsistency` conflict must be tracked separately and not used to excuse a newly failing test.

## Completion and sequencing
Only after all focused tests and both numerical profiles pass: normal fast-forward push Secondary `fast-qprefix`, then Primary pin/report commit and normal push. Do not force push, rewrite history or overwrite unrelated user changes. Report `BALANCED_PRESCALE_REMOVED` on successful implementation plus numerical regression; `BALANCED_PRESCALE_REMOVAL_BLOCKED` for a concrete implementation/capacity/test blocker with evidence. End `READY_FOR_WEB_REVIEW`.

**After Web accepts this cleanup**, reactivate existing `FAST-STANDARD-PERF-REBASELINE-003` against the newly pinned Secondary SHA. Do not perform that performance task before acceptance.

# FAST-QPREFIX-CHEBYSHEV-ORDER-002 — Remove pre-scale in Q0123 generated powers

## Decision — implement, not experiment
Task class: **I — authorized Secondary numerical-correctness repair**. Directly remove the legacy Q01-era pre-scale/balanced generated-power schedule from the **supported Fast Q-prefix Chebyshev/EvalMod production path**, for BOTH LogN13 and LogN16. **Do not first run an A/B hypothesis study**, build a new trace harness, or iterate on performance optimizations.

Mathematical and architectural authority: `xuejin-lu/lattigo/docs/FAST_QPREFIX_SPEC.md` §4, §8, §15 and Primary `specs/FAST-STANDARD-NUMERICAL-FIX-001-REMOVE-NORMALIZED-EVALMOD.md`. The maintained-row rule `w_Q(l)=min(l+1,4)` controls Q storage/capacity, **not the Standard CKKS recurrence/order**. At supported levels, a Fast Chebyshev generated power must execute the Standard-equivalent sequence:
1. multiply original operands at their source logical level/scale;
2. apply Chebyshev doubling and recurrence correction at the product scale;
3. execute one Rescale of the corrected result using the logical top `q_l`, `Level'=Level-1`, `Scale'=Scale/q_l`.

Never Rescale the two operands independently before multiplication; never Rescale the product before the Chebyshev recurrence correction; do not just set metadata scale to an unproven target. This order is a **semantic rule**, not a `LogN==13` or `LogN==16` special case.

## Confirmed source defect and frozen starting point
- Secondary active `fast-qprefix` known-correct starting commit `5117fc57949647182f476dc5952099c706b9f869`; safely sync current branch before work.
- `circuits/ckks/polynomial/fast.go` `postProductLogN13Schedule()` hardcodes `pb.params.LogN()==13`; this excludes LogN16 even though its real Q prefix is `[56,39,39,39]` bits and the existing Q0123 row policy covers its high logical levels. Non-selected LogN16 goes through `balancedScheduleFor()`, the pre-operand scaling/Rescale alternative or low-scale early product Rescale.
- FIX-001 repaired LogN13 by using **multiply -> recurrence correction -> Rescale**; LogN13 is currently PASS (Fast-vs-Standard RMSE `1.44680895e-10`, SNR Standard/Fast `144.713390/144.710750 dB`).
- Frozen LogN16 baseline reports `results/FAST-STANDARD-PERF-REBASELINE-002-logN16-report.md`: 5.6701x with INVALID numerical result, Fast/Standard RMSE `0.020485498`, SNR Standard/Fast `130.947666/-0.000462 dB`. First internal nonzero Chebyshev difference T2; public first material divergence `evalmod_real`.

## Bounded implementation
1. Modify only the necessary **Secondary** Fast polynomial source/tests. Delete or retire the `postProductLogN13Schedule` LogN-specific path selection and **remove pre-scale/balanced operand Rescale from the supported Q-prefix Chebyshev power generation**. Use one shared, Standard-equivalent algorithmic ordering for BOTH LogN13 and LogN16; do not replace `LogN==13` with `LogN==13 || LogN==16` and keep the old fallback. Preserve Standard full-RNS code unchanged and keep genuine Standard CKKS multiplication and Scale/Level semantics.
2. Keep existing minimal/capped real Q-prefix rows `w_Q(l)=min(l+1,4)`; **Q0123 only when Level >= 3**, with natural contraction at lower levels. Don't impose four rows on levels <3; don't use dormant q rows, private-F, full-Q fallback, or key-switching workaround.
3. Keep capacity observers/assertions for post-multiply/pre-Rescale intermediate and final states (add only the smallest missing guard needed). Require strict centered `2B < S_Q(l)` at every lifted-integer boundary where it is needed. Do not assert that 56 historical post-operation checks automatically prove a newly exposed intermediate fits. On a real capacity failure, **stop with exact coefficient component, logical Level, actual prefix, observed/proven B and S_Q/2 deficit**; do not restore pre-scale or invent another workaround.
4. Preserve `c0=encoded-message,c1=0` Fast zero-a semantics and sampled/error-retaining invariants; do not touch Normal Lattigo, application/frontend, or Primary `cmd/perfprobe`.
5. If the old `balancedSchedule`, factorization helpers and tests become dead code after removing the production branch, delete precisely the unused code and superseded tests; add narrow tests asserting both LogN13 and LogN16 now follow the **same ordering** independently of LogN, source Level/Scale, Q-prefix coverage and Chebyshev correction before Rescale. Avoid repository-wide cleanup.

## Focused validation — regression checks, not hypothesis exploration
- Secondary focused polynomial generated-power tests, relevant Fast Mod1 and Bootstrap tests, `go test` for changed packages, plus `git diff --check`.
- Fresh final numerical run using **existing** Primary harness: LogN13 regression and LogN16 original canonical profile, each against genuine Standard on unchanged effective Q/P/input. Record final output Fast–Standard RMSE, coordinate threshold violations, Standard/Fast decoded SNR and first failing checkpoint if any. **No seven-repeat timing campaign**, new harness, parameter adjustment or threshold relaxation.
- LogN13 must remain numerically close, and LogN16 must clear existing 0.01 real/imag threshold with dramatically improved RMSE/SNR, ideally match Standard precision/SNR. If not, report exact failure and stop rather than treating the commit as numerically accepted. Do not assert all numerical parity solely from a unit test.
- Preserve existing `FAST-STANDARD-PERF-REBASELINE-002` historical reports untouched; a production change makes its old 5.6701x speedup stale for the new Secondary SHA.
- Keep documented pre-existing unrelated `TestFIX001P3GenuineStandardPublicVsStagedConsistency` failure separate; do not use it to hide a new failure.
- Provide compact new Primary `results/FAST-QPREFIX-CHEBYSHEV-ORDER-002-summary.md` with old/new Secondary SHA, exact diff/scope, two profile numerical outcomes, needed capacity evidence, tests and limitations. If authorized Primary report writing is blocked, return the complete measured facts and the push SHA for Web review instead of expanding task scope.

## Git and execution authority
- Read Primary and Secondary `AGENTS.md` and Secondary `docs/FAST_QPREFIX_SPEC.md`, safely synchronize Primary `main` then Secondary `fast-qprefix`. Secondary CURRENT_TASK is a local pointer; this Primary spec is authoritative for this execution.
- **Secondary production edits and normal fast-forward push to `origin/fast-qprefix` are explicitly authorized by the user.** Primary test/report commits may be pushed normally once source is pinned; avoid ref conflicts and never amend/force/reset.
- One implementation pass, one bounded coding self-review/repair, no new redesign cycle or repeated exploratory trials. The constitutional invariant is already decided.
- Return `QPREFIX_CHEBYSHEV_ORDER_RESTORED` when both profiles pass; `QPREFIX_CHEBYSHEV_CAPACITY_BLOCKED` for demonstrated strict capacity failure; `QPREFIX_CHEBYSHEV_NUMERICAL_UNCLOSED` if exact ordering is installed but numerical output still fails; `QPREFIX_CHEBYSHEV_IMPLEMENTATION_BLOCKED` for workflow/build blockers. Finish `READY_FOR_WEB_REVIEW` with SHAs, tests and before/after results. Never label a failed numerical profile as restored.

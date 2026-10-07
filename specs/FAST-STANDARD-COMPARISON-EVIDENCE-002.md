# FAST-STANDARD-COMPARISON-EVIDENCE-002 — Separate formal paired results from legacy synthetic fastdiag evidence

## Task class and purpose

Class **I** — bounded Primary research-harness report/validation repair. This task follows accepted `FAST-STANDARD-INPUT-PROVENANCE-001` and addresses one remaining formal-comparison blocker.

**Do not run a formal benchmark campaign in this task.** The prior `FAST-STANDARD-PERF-REBASELINE-003` remains blocked; do not execute or amend its old specification.

## Source-backed defect

`cmd/perfprobe/compare.go:runCompare` still requires `--fastdiag` and reads `cmd/fastdiag numerical` results before it can generate a paired report. That legacy diagnostic constructs a single artificial `(c0=encoded-message,c1=0)` input even for its in-tree Standard evaluator. It is a useful specialized stage diagnostic, **not an independently encrypted Standard reference**. The current paired document copies the diagnostic classification/capacity and labels it alongside the new genuine Standard/Fast paired results, which risks promoting non-equivalent experiment evidence to formal quality conclusions. The current report also hardcodes a REBASELINE-002 raw artifact filename. This is an evidence-boundary defect, not a defect in the accepted Standard `EncryptNew` input path.

## Frozen invariants

1. Keep the **genuine native Standard input** from `37b73ad58983bd7689999847ed5c59c4a1e14fed` and the clearly labeled Fast direct-encoded simulator input unchanged. Pin Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and Fast production `5feb44917fca40c93abec6def6f26bc81a82c536`. No cryptographic source, production arithmetic, parameter, or key-semantic edits.
2. Only normal decrypted Standard outputs from the independent pinned Standard build, and declared Fast simulator outputs from the pinned Fast build, may support **formal paired output comparisons**. Both inputs must pass the previously accepted v2 input-provenance checks. Preserve common-original-message hash, actual effective CKKS parameters, and environmental/provenance gates. Do not require equal ciphertext metadata or coefficients.
3. The archived `fastdiag numerical` stage/PS/DoubleAngle/capacity analysis must never be interpreted as evidence generated from a native Standard-encrypted input. It may remain available as a **separate, explicit diagnostic-only artifact** with its own input kind, source provenance, and limitations, but it must not be mandatory or contribute to formal paired success/failure classification.
4. Numerical quality and latency are distinct measurements. Existing decoded-domain `numericalmetrics` helpers can be reused. Do not replace output-versus-original gates with merely small Fast-versus-Standard differences. Do not promise numerical equality or security equivalence.
5. Historical result files are immutable in this task. Do not delete, relabel, or overwrite them. Secondary source and `cmd/fastdiag numerical` implementation are out of scope.

## Authorized implementation

- Repair `cmd/perfprobe/compare.go` and only the minimal adjacent Primary adapter/schema/tests needed so **formal compare runs solely on the independently pinned Standard/Fast v2 probe records and trial vectors**. Remove the mandatory `--fastdiag` prerequisite. Make legacy synthetic-diagnostic inputs invalid as formal inputs.
- If optional diagnostic attachment is retained, it must be clearly separated under a `diagnostic_only` or equivalent namespace with its own provenance and a machine-checkable `not_formal_reference` status. No diagnostic classification/capacity pass may silently determine a formal comparison acceptance. Simpler omission of legacy diagnostic attachment is preferred.
- Formal classification should be derived only from validated fresh paired source/provenance, preflight, and per-backend output-against-original metrics and declared thresholds. If the task cannot define a mathematically sound quality acceptance threshold from the already-accepted durable contract, use explicit `UNASSESSED` / `REQUIRES_WEB_REVIEW` instead of inventing or relaxing a gate. Keep pairwise Fast-versus-Standard RMSE as secondary descriptive information.
- Repair report text that wrongly calls in-tree legacy stages 'genuine' independently encrypted Standard evidence, and remove the hardcoded `FAST-STANDARD-PERF-REBASELINE-002` artifact link in new formal reports. Keep honest Standard native-encryption versus insecure Fast simulation caveats visible.
- Add unit tests (no costly Bootstrap execution): valid v2 paired records with independent per-trial Standard input evidence; failure for legacy schema/kind; report construction **without** fastdiag; rejection of absent/contradictory provenance; separation or omission of legacy synthetic diagnostic classification; and no hardcoded legacy file reference in new output.
- Run focused `go test` (including tagged variants as needed), `git diff --check`, bounded self-review, then Primary-only normal commit/push. No Secondary changes, no benchmark run, no old result deletions, no broad root runner rewrite.

## Acceptance

- `FORMAL_COMPARISON_EVIDENCE_BOUNDARY_READY`: formal comparison can consume only native-Standard/Fast-simulation v2 evidence, without requiring synthetic Standard fastdiag; tests prove the separation and report labels are accurate. No new timing/numerical research claim.
- `FORMAL_COMPARISON_EVIDENCE_BOUNDARY_BLOCKED`: precise code/schema/contract blocker with evidence; return `NEEDS_WEB_REVIEW`, do not substitute old diagnostic.
- On success return `READY_FOR_WEB_REVIEW`. After independent Web acceptance, a **separate** experiment task may measure LogN13/LogN16 formal outputs and runtime using the repaired pipeline.

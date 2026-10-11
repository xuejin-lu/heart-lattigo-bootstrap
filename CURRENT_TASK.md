# CURRENT TASK — Integrated LogN16 Bootstrap stage/power/rescale measurement

Task: FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006
Status: READY_FOR_CODEX
Authoritative spec: `specs/FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006.md`

Task005 already completed the formal matched original-input LogN16/E32 Standard-vs-current-Fast native public lifecycle, numerical-oracle checks, latency/allocations and 16/16 strictly budgeted Bootstrap calls. **DO NOT repeat its Standard/Fast benchmark or reuse/reset its exhausted budget.** Fixed hardware Apple M4; Standard 5dbffbdea05394de2ca3a432ed5318aa832e3f40, Fast 2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f. Existing results: LogN16 warm median 6425.276 ms vs 663.878 ms (9.6784x); terminal complex RMSE 4.09656e-8, max 4.34568e-7, gate 1e-6; 8/8 public checkpoints passed.

The user's next priority is the **EXISTING complete measurement platform**, not another custom RMSE-only runner: reuse `cmd/perfprobe` for accepted Standard/Fast full-bootstrap timing/numerics, `cmd/fastdiag` and Secondary internal fastdiag hooks for **actual Fast in-circuit** C2S, EvalMod real/imag, S2C, generated powers, DoubleAngle/Rescale timing, alongside `internal/perfmeasure` and `internal/numericalmetrics`. Batch025 offline trace validator and event-root analysis are reusable, but current public-E32 trace integration still pins historical LogN13 Fast and needs **minimal versioned LogN16/current-Fast support**. Preserve backend-transparent app-facing `NewEvaluator`/ordinary public methods. Do not create another measurement framework or modify current Fast production math.

Mandatory Phase A: safely sync and read AGENTS/measurement/backend transparency rules; validate Task005 provenance/ledger; repair only necessary **Primary platform adapters, versioned fixtures, source-integrity tests and isolated test-only diagnostic path** for new LogN16 trace. Run all no-full-Bootstrap schema/fixture/recursive event-tree/fail-closed tests first. If an instrumented production-source change is needed, stop for Web review; no unapproved Secondary edits.

Only after every zero-call gate passes, **new Task006 exclusive maximum of TWO Fast-only full Bootstrap calls** (one cold, one traced warm) for the same original native LogN16/E32 Level0 ciphertext, separately journalled and fully counted. No additional Standard runs, no hidden attempts, no repeat after failed call. Validate output oracle, decode and event trace integrity, report in-circuit stage/power/rescale time and separately labelled 005 uninstrumented speed/error. Traced overhead cannot be treated as matched Standard speedup; non-comparable private Standard stages must be labelled unavailable.

Output `results/FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006.md` / `.json`, task-specific attempt ledger, compact event Pareto/provenance, and any safely validated platform improvements. Commit/push Primary only, READY_FOR_WEB_REVIEW or NEEDS_WEB_REVIEW. User says `開始` in Codex.

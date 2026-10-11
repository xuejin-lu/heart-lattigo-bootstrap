# CURRENT TASK — Measure LogN16 Rescale instrumentation perturbation

Task: FAST-LOGN16-RESCALE-INSTRUMENTATION-BIAS-004
Status: READY_FOR_CODEX
Authoritative spec: `specs/FAST-LOGN16-RESCALE-INSTRUMENTATION-BIAS-004.md`

Previous Task003 source-backed LogN16 EvalMod audit at `6271aae2fee2cdfb286f112f309e0a8e0a9e90ec` passed as a hypothesis report, **not** causal proof. Task006 trace shows EvalMod real+imag 71.4485% of one traced Bootstrap. However `fastcore/rescale.go` currently uses fine-grained `StartTimer/ElapsedNS` calls inside the coefficient loop when tracing is enabled, creating a possibly large observer effect. The existing BenchmarkFastRescaleLogN16 also uses a different Q chain from historical Task006.

On `開始`, use AGENTS.md mandatory safe Git sync, then autonomously execute bounded Task004: audit instrumentation source, reuse current package benchmarks/oracles to compare matched LogN16 Rescale with tracing absent, disabled and enabled, and determine whether CRT coefficient-loop is a credible uninstrumented optimization target. NO new full Bootstrap, no new measurement platform, no production math changes, no Secondary committed branch changes. Check file sizes first; preserve Task005/006 historical traces. Source and matched-measurement evidence must be separated from inference. Finish with compact report, safe Primary fast-forward push, READY_FOR_WEB_REVIEW or NEEDS_WEB_REVIEW for actual blockers.

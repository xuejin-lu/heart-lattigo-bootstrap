# CURRENT TASK — Batch026 E32 Rescale Coefficient Hotspot

Task: FAST-E32-RESCALE-COEFFICIENT-OPT-AUTONOMOUS-BATCH-026
Status: READY_FOR_CODEX
Mode: ZERO-BOOTSTRAP REPRESENTATIVE E32 RESCALE MICROBENCH + ONE BOUNDED CANDIDATE

**Authoritative spec:** `specs/FAST-E32-RESCALE-COEFFICIENT-OPT-AUTONOMOUS-BATCH-026.md`
**Independent scientific review:** `results/FAST-DROPIN-E32-TRACE-OFFLINE-SALVAGE-AUTONOMOUS-BATCH-025-web-review.md`
**Permanent rules:** Primary and Secondary AGENTS, docs/MEASUREMENT_PLATFORM.md, workflow §4A/4B.

Batch025 `OFFLINE_REVALIDATED_TRACE` accepted as **post-hoc**, original raw remains `TRACE_UNVERIFIED`. 539 events / zero new crypto calls. One instrumented warm E32 Fast Bootstrap 122.193458ms; EvalMod real+imag 91.857958ms (75.17% of Bootstrap traced stage root). Thirty-five independent Rescale roots sum 108.326291ms, which is NOT additive to Bootstrap root or Power roots. Inside those roots, coefficient loop own exclusive plus CRT/reconstruct+round/capacity sum **72.933231ms / ~67.33% of separately-rooted Rescale time**, more than NTT restore 9.567417ms (~8.83%). Single traced sample has measurement overhead; original uninstrumented 72.580ms Fast median / 306.636ms Standard median historical 4.225x is distinct and intentionally insecure zero-secret.

R1: clean startup sync, source map and authentic repeated E32 real-Rescale **regular-build** microbench baseline, no Bootstrap.
R2: only if justified, one scoped source-backed safe fixed-width/coefficient-loop optimization hypothesis; no CKKS math/representation changes without Web.
R3: extensive cheap real Rescale oracle equality (rows 1–4, alias, metadata, capacity/rounding), matched baseline/candidate repeated microbench; no benchmark overclaim; revert only own candidate safely if worse.
R4: compact result + explicit 0 new Bootstrap, safe commits/push and one Web review. No Batch027 authority.

**Exact expensive call cap: ZERO Bootstrap (Standard and Fast), ZERO full CNN/CKKS benchmark.** Never reuse Batch023 14/14 or Batch024 2/2 spent attempts. Avoid touching old Git worktree metadata or changing frozen genuine Standard and formal original Fast pins.

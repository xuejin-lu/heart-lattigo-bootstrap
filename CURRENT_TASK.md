# CURRENT TASK — Q0123 Rescale realism and bounded optimization pilot

Task: FAST-Q0123-RESCALE-REALISM-OPT-PILOT-005
Status: READY_FOR_CODEX
Authoritative spec: `specs/FAST-Q0123-RESCALE-REALISM-OPT-PILOT-005.md`

Previous Task004 at `d8ea95d2d005db96c7a54750fa243fef62e8412b` is accepted as a matching synthetic instrumentation-perturbation result: uninstrumented LogN16 Q0123 Rescale median 21.340 ms, trace-on median 31.852 ms (+48.82% vs tagged mask-off). This is NOT a real Bootstrap improvement measurement.

When user says `開始`, follow AGENTS.md mandatory safe Primary/Secondary Git sync, then autonomously run one bounded Task005 pilot. First audit actual EvalMod c1 state; compare matched zero/nonzero c1 on exact LogN16 Q/P with existing Task004 test-only benchmark and BigInt oracle, without tracing. Then attempt at most one justified bit-exact coefficient-loop optimization in isolated Secondary prototype, if appropriate, with existing regression and finite A/B timing. No full Bootstrap, no new tools/framework, no Secondary production commit/push, no stale data overwrite or compromised math/security. Preserve prior experimental evidence. Publish compact Primary summary and candidate patch only if correct and repeatably beneficial; safe Primary fast-forward push. Finish READY_FOR_WEB_REVIEW with PROMOTE/REJECT/INCONCLUSIVE or NEEDS_WEB_REVIEW on genuine blockers.

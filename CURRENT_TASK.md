# Current Task

Task: QPREFIX-ARCH-PLAN-001
Status: READY_FOR_CODEX

Task class:
`M — Architecture planning only`

Purpose:
Produce a top-level Q-prefix v2 production migration plan from the current pre-F branch architecture.

Authoritative repositories:
- Primary: `xuejin-lu/heart-lattigo-bootstrap@main`
- Secondary: `xuejin-lu/lattigo@fast-qprefix`

Authoritative architecture:
- `xuejin-lu/lattigo@fast-qprefix:docs/FAST_QPREFIX_SPEC.md`

Important decision:
- private F is not the production direction;
- Q-prefix v2 is the production direction;
- do not resume `QPREFIX-AUDIT-003` unless a later Primary task explicitly re-authorizes it.

The user will provide the detailed architecture-planner prompt directly in Codex chat. Follow that prompt as the authoritative task instructions.

Required repository action:
- create only `docs/QPREFIX-V2-PRODUCTION-MIGRATION-PLAN.md` in Primary;
- do not modify production code, tests, CURRENT_TASK, specs, parameters, or Secondary source;
- commit and push Primary normally;
- final chat response should be concise and report `READY_FOR_WEB_REVIEW`, commit SHA, report path, milestone count, top risks, and any `NEEDS_LOCAL_IMPLEMENTATION_INVESTIGATION` items.

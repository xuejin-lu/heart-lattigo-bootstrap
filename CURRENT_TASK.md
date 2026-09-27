# Current Task

Task: QPREFIX-IMPL-004
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-IMPL-004-RESCALE-LEVEL-TRANSITIONS.md`

Revision:
Primary spec commit `87125e13570a4e88dc26343d88810c23966768f7`.

Reason for revision:
The first 004 attempt correctly discovered that current C2S LinearTransform still produces only legacy q012 authority, so production Rescale cannot globally consume q0123 yet.

Updated rule:
- q0123 Rescale capability must be implemented and tested;
- explicit-width Rescale/DropLevel kernels may consume width 4;
- existing production Rescale/RescaleTo wrappers remain legacy-authority activated until their upstream producers are migrated;
- current C2S/DFT must continue to pass using q012 authority;
- q3 backing alone never implies q3 authority.

Controlled dirty-state continuation is authorized:
- Secondary local uncommitted work from the stopped first attempt is expected and must be preserved;
- do not reset, clean, or stash it away indiscriminately;
- first sync task/spec metadata as safely as possible without discarding implementation work;
- adapt the existing local changes to the revised spec.

Accepted prerequisites remain QPREFIX-IMPL-001..003.

Do not modify LinearTransform/DFT production routing in this task.
Do not activate q0123 globally.
Do not touch ModUp/Trace, EvalMod/PS/DA, Bootstrap orchestration, `fast-ckks`, or add F/full-RNS fallback.

Required end state:
- q0123 CRT + explicit width-4 Rescale kernels tested;
- production C2S/DFT regression passes on legacy q012 authority;
- full required tests pass;
- commit/push Secondary `fast-qprefix`;
- report `READY_FOR_WEB_REVIEW`.

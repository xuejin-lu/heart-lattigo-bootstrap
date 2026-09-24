# Current Task

Task: FAST-STORAGE-002
Status: READY_FOR_WEB_REVIEW

Specification:
`specs/FAST-STORAGE-002-CONTAINER-CONVERSION-BOUNDARIES.md`

Task class:
`I — Implementation`

Purpose:
Implement a physically separate Fast storage ciphertext container and exact Level-0 LogicalQ <-> FastStorage conversion boundaries.

Frozen architecture:
- Fast storage rows are modulo private `f_i`, never logical `q_i`.
- logical Level is explicit metadata independent of storage width.
- Bootstrap production entry is planned after ScaleDown reaches logical Level 0.
- no production Bootstrap wiring in this task.

Secondary architecture authority:
`docs/FAST_CKKS_SPEC.md`
including commits:
- `3c3fe59f24fd9e80ab03ca566a39336c53b6c121`
- `aa99f85499899af53af87b569cec48d8ce4236c6`

Accepted storage foundation:
`cc5028c872a89aff05ca43aa7e6f8c4269fcf8b5`

Codex should execute the bounded implementation -> self-review -> one repair pass -> final validation workflow, then return `READY_FOR_WEB_REVIEW`.

Implementation candidate:
- Secondary `xuejin-lu/lattigo@fast-ckks`: `3b57a52b311397e0e1cf8298027782eec20d4ffc`

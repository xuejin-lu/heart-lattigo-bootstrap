# Current Task

Task: FAST-STORAGE-007
Status: READY_FOR_CODEX

Specification:
`specs/FAST-STORAGE-007-PRIVATE-F-PLAINTEXT-LINEAR-TRANSFORM.md`

Task class:
`I — Implementation / feasibility foundation`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: orchestration/spec only.
- Secondary `xuejin-lu/lattigo@fast-ckks`: implementation target.

Accepted prerequisites:
- FAST-STORAGE-006 at `532319346d8235fb42c72bfd22b57a6468675c82`.
- Private-F plaintext mirror architecture at `22b9f07969af38705573686fe96a873e4bd001a3`.

Implement only the private-F plaintext mirror + standalone linear-transform feasibility foundation:
- exact one-time LogicalQ plaintext CRT reconstruction into authoritative integer coefficients;
- mirror those coefficients into fixed width-3 private F;
- private-F plaintext multiplication;
- private-F automorphism;
- standalone private-F LinearTransform;
- reproducible LogN13 real-C2S factor-by-factor capacity audit using the bound `B_out <= B_in * sum_d ||P_d||_1`.

Do not modify production Bootstrap/C2S, DFT matrices, restore plan, EvalMod, S2C, production Rescale, storage width, or public CKKS parameters.

If real LogN13 C2S capacity fails at any factor, stop with `NEEDS_WEB_REVIEW` and report the first failing factor and exact numbers; do not attempt production integration.

If feasible, commit/push Secondary and report `READY_FOR_WEB_REVIEW`.

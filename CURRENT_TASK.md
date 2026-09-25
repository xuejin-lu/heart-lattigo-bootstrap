# Current Task

Task: FAST-INTEGRATION-001
Status: READY_FOR_CODEX_AFTER_WEB_RESOLUTION

Specification:
`specs/FAST-INTEGRATION-001-PRIVATE-F-MODUP-BASIS-BRIDGE.md`

Task class:
`I — Implementation`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: orchestration/spec only.
- Secondary `xuejin-lu/lattigo@fast-ckks`: implementation target.

Accepted prerequisites:
- FAST-STORAGE-005 at `531aca50b5b38741e4e71cc98ea4b626bf88cb84`.
- Fixed-width-3 initial production policy at `d9919f9c080e0dfa731746f5c447f93633ae2f36`.

Implement only the first bounded production integration seam:
- replace the historical basis-raise portion of `FastEvaluator.modUpBasis` with
  Level-0 LogicalQ -> ImportLevel0(width=3) -> FastStorageModUpLevel0(MaxLevel) -> compact maintained LogicalQ;
- add a compact private-F -> logical-Q bridge that materializes only the existing maintained q rows;
- preserve existing scale alignment, Trace, Montgomery conversion, downstream DFT/EvalMod/S2C, packing, and production Rescale semantics;
- record ModUp-basis benchmark evidence.

Do not migrate downstream arithmetic to `FastCiphertext`, do not add KeySwitch/Relinearize/Rotate, contraction/adaptive width, frontend flags, or full-RNS fallback.

Codex must follow the normal startup sync and bounded implementation -> self-review -> one repair pass -> validation workflow, commit/push Secondary `fast-ckks`, then report `READY_FOR_WEB_REVIEW`.


Web resolution:
- Accepted canonical centered-q0 semantics are authoritative.
- For odd q0=2m+1, residue r=m=q0>>1 maps to +m, not -(m+1).
- Historical >= q0>>1 behavior at that single residue is treated as an off-by-one convention and is not an integration acceptance oracle.
- Codex may continue FAST-INTEGRATION-001 using the updated spec commit 242b279c351fe4b13d399e9aed6d0891f8673ac4 and Secondary durable-spec clarification a98c00aa5e9c3fd4118c745159f5e9dc8e89db4b.

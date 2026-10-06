# Current Task

Task: FAST-QPREFIX-CHEBYSHEV-ORDER-002
Status: READY_FOR_CODEX
Task class: I — Direct Secondary correctness repair

Authoritative executable spec:
`specs/FAST-QPREFIX-CHEBYSHEV-ORDER-002.md`

**Direct user decision**: LogN13 and LogN16 Q-prefix Chebyshev/EvalMod generated powers must share **multiply → recurrence correction → Rescale**, not historical balanced/pre-scale. The LogN-specific selector is unconstitutional: Q-prefix width affects representation/capacity, not mathematical CKKS schedule. **Implement directly, NO preliminary A/B study.**

Supersedes `specs/FAST-LOGN16-EVALMOD-ROOTCAUSE-001.md` (unexecuted diagnostic proposal; historical only).

Secondary `xuejin-lu/lattigo` `fast-qprefix` production edits and ordinary push are authorized. Retain q0123 capacity safeguards and prove failure if any, not a new pre-scale fallback. No rerun of 7x timing benchmarking. Run only necessary LogN13 regression + LogN16 final numerical correctness against original genuine Standard.

Prior committed evidence:
- LogN13 PASS: Fast-Standard RMSE 1.44681e-10.
- LogN16 FAIL at Secondary `5117fc57949647182f476dc5952099c706b9f869`: Fast-Standard RMSE 0.0204855, decoded SNR Standard/Fast 130.947666/-0.000462 dB.

Codex: safe-sync both repositories; read both AGENTS and the new spec; edit Secondary narrowly, validate both profiles, commit/push per standing safety rules, report classification + SHAs then `READY_FOR_WEB_REVIEW`.

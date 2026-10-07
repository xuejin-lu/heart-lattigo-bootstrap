# Current Task

Task: FAST-STANDARD-PERF-REBASELINE-003
Status: BLOCKED_BY_BASELINE_CONTRACT
Task class: E — clean-source independent-Standard paired performance + SNR

**Authoritative executable spec:** `specs/FAST-STANDARD-PERF-REBASELINE-003-POST-CHEBYSHEV.md`

## Accepted predecessors
- `FAST-QPREFIX-CHEBYSHEV-ORDER-002`: removed LogN-specific Chebyshev ordering.
- `FAST-QPREFIX-REMOVE-BALANCED-PRESCALE-003`: the entire historical balanced/pre-scale generated-power mechanism deleted from production, regression verified. Secondary final pinned **`5feb44917fca40c93abec6def6f26bc81a82c536`**, Primary result `5a7e17e861df7fd7ae68cfffc38c4444446c00cf`.

Existing post-cleanup zero-a numerical gates:
- LogN13 Fast-vs-Standard RMSE `1.446809e-10`, Standard/Fast SNR `144.713390 / 144.710750 dB`.
- LogN16 in-tree Standard API comparison RMSE `0`, Standard/Fast SNR `130.947666 / 130.947666 dB`.
- 80/80 capacity strict fit per profile, no `1e-2` coordinate violations.
- These are **not** independent historical-Standard-SHA result confirmations and **not** updated performance results.

## Blocked — do not execute

The referenced REBASELINE-003 specification and current `cmd/perfprobe` / numerical harness reuse a plaintext-like Standard input with manually constructed ciphertext components (`c0=encoded-message, c1=0`). That path conflicts with the current Formal Standard baseline contract in `AGENTS.md`, which requires native Standard key generation, encryption, measured operation, and decryption/decoding for any formal Standard numerical-quality or performance baseline.

Do **not** execute REBASELINE-003, do not run its benchmark campaign, and do not treat its existing zero-a results as formal Standard baseline evidence.

This task remains blocked until Web review provides a replacement or amended experiment specification that satisfies the durable baseline contract. Do not improvise a substitute input path or silently weaken that contract.

The Secondary `CURRENT_TASK.md` is also known stale legacy state; do not use it to resume historical FIX-001 work.

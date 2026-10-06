# Current Task

Task: FAST-STANDARD-PERF-REBASELINE-003
Status: READY_FOR_CODEX
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

## Execute now
Use unchanged existing Primary `cmd/perfprobe`, `internal/perfmeasure`, `cmd/fastdiag` and `internal/numericalmetrics`. Compare clean detached Fast `5feb44917...` against clean detached genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, same frozen committed Primary source/config/input per profile, both LogN13 and LogN16. Seven full-Bootstrap timing samples each backend/profile, allocation metrics, fresh independent-build SNR/RMSE/precision, stage/capacity evidence and concise joint report. No production edits, harness rewrites, speculative fixes, or reuse of obsolete speedups.

**Secondary's `CURRENT_TASK.md` is a known stale legacy pointer** to FIX-001; follow the newly synced authoritative Primary spec and do not touch Secondary files/HEAD during this campaign. Report genuine conflict if found.

Return one classification defined in the spec and `READY_FOR_WEB_REVIEW`. Primary-only reporting commit and normal push once validated.

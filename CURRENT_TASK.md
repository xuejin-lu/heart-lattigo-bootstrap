# CURRENT TASK — BTS LogN16 no-trace Standard/Fast baseline

Task: BTS-LOGN16-NOTRACE-BASELINE-001
Status: NEEDS_WEB_REVIEW
Authoritative spec: `specs/BTS-LOGN16-NOTRACE-BASELINE-001.md`

`FAST-Q0123-RESCALE-EXACT-DIVISION-PILOT-006` is suspended by explicit user instruction. Do not run its Rescale pilot as part of this task.

This task is limited to the unchanged `external/BTS` module and exactly one original-default Bootstrap attempt per linked backend: genuine Standard Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, and Fast `fast-qprefix` `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`. Standard completed; Fast panicked in `MyC2S_DHManual → ckksEval.DecomposeNTT → ring.INTT` because the first polynomial operand had length zero. Both attempt budgets are exhausted; do not rerun. See `results/BTS-LOGN16-NOTRACE-BASELINE-001-summary.md` and its journal.

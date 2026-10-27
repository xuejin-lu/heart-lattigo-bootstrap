# BTS-FAST-NOISE-FIDELITY-009

Status: `PREFLIGHT_COMPLETE_AWAITING_VERSIONED_CHECKPOINT`

- Independent budget: 18 Bootstrap calls (Standard 6, Fast OFF 6, Fast+Noise 6).
- Reserved / entered / completed / failed: 0 / 0 / 0 / 0.
- Original BTS manifest: `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`.
- Frozen parameters: BTS-008 effective LogN16 / E32 / Q30 / P5 / input L3, Scale 2^32 / output L18, Scale 2^32.
- σ=`4.004280012097432`, root seed=`20261027`.
- Primary measurement implementation commits: `369ce3a84eb072392bacf2359d651413c159e5e8`, `42b9840a6ea42bb6d9f4ba368b7c161617d4a02d`, `dd077ac69e81192ffaa4b606d305e545d8b50202`.
- Genuine Standard source: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Task-specific Fast source: `experiment/fast-bootstrap-noise-fidelity@be9c42ff23662300769b77f1eead3f12fc53adef`.
- Targeted Primary perfprobe/perfmeasure/numericalmetrics tests and Secondary public Noise Hook/fast tests passed; these include N13–N16 generic hook and BTS LogN16 helper coverage, but no complete Bootstrap.
- The first three zero-call preflights and binaries are preserved but superseded. Clean binaries and all three explicit-state `preflight-v2/` checks passed; each has zero reservations, the same original BTS input/profile, and explicit Standard/OFF/ON Hook states. Their hashes and provenance are in the attempt ledger and journal. The v2 build/preflight checkpoint still needs to be committed and pushed before any formal call.

See `journal.md` for source map, synchronization, tests and gates; `experiment-plan.md` specifies the frozen preflight and arm order; `attempt-ledger.json` contains the 18 predeclared, unreserved attempt slots.

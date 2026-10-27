# BTS-FAST-NOISE-FIDELITY-009

Status: `CP_AB_IMPLEMENTED_ZERO_BOOTSTRAP_PREFLIGHT_PENDING`

- Independent budget: 18 Bootstrap calls (Standard 6, Fast OFF 6, Fast+Noise 6).
- Reserved / entered / completed / failed: 0 / 0 / 0 / 0.
- Original BTS manifest: `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`.
- Frozen parameters: BTS-008 effective LogN16 / E32 / Q30 / P5 / input L3, Scale 2^32 / output L18, Scale 2^32.
- σ=`4.004280012097432`, root seed=`20261027`.
- Primary measurement implementation commits: `369ce3a84eb072392bacf2359d651413c159e5e8`, `42b9840a6ea42bb6d9f4ba368b7c161617d4a02d`, `dd077ac69e81192ffaa4b606d305e545d8b50202`.
- Genuine Standard source: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Task-specific Fast source: `experiment/fast-bootstrap-noise-fidelity@be9c42ff23662300769b77f1eead3f12fc53adef`.
- Targeted Primary perfprobe/perfmeasure/numericalmetrics tests and Secondary public Noise Hook/fast tests passed; these include N13–N16 generic hook and BTS LogN16 helper coverage, but no complete Bootstrap.
- Three initial zero-call preflights completed with 0 reservations and the same input SHA, but their false Hook booleans were omitted by `omitempty`; those original artifacts remain preserved and are superseded. Clean binaries plus exactly three explicit-state preflights under `preflight-v2/` are pending. Do not start a Bootstrap until the new artifacts are committed and pushed.

See `journal.md` for source map, synchronization, tests and gates; `experiment-plan.md` specifies the frozen preflight and arm order; `attempt-ledger.json` contains the 18 predeclared, unreserved attempt slots.

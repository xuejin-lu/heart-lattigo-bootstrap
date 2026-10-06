# Current Task

Task: FAST-LOGN16-EVALMOD-ROOTCAUSE-001
Status: READY_FOR_CODEX
Task class: E/M — bounded LogN16 numerical root-cause diagnosis

**Authoritative executable spec**:
`specs/FAST-LOGN16-EVALMOD-ROOTCAUSE-001.md`

## Prior result accepted as measurement, NOT numerical correctness
`FAST-STANDARD-PERF-REBASELINE-002` completed, Primary report commit `d24396a76ce7773c8b70200ee0d7633de48f8126`; outcome `DUAL_LOGN13_LOGN16_PERF_SNR_NUMERICAL_FAIL`.
- LogN13 PASS: 4.3755x, Fast-vs-Standard RMSE 1.44681e-10.
- LogN16 FAIL: measured 5.6701x **on numerically invalid output**; Standard/Fast SNR 130.947666/-0.000462 dB, RMSE 0.0204855. First public material checkpoint `evalmod_real`, but internal Chebyshev T2 has earlier nonzero difference.
- Full provenance and SNR: `results/FAST-STANDARD-PERF-REBASELINE-002-comparison.md`, profile reports, raw fastdiag and paired numerical JSON.

**Scope**: trace earliest LogN16 generated-power T2 multiplication/recurrence/Rescale discrepancy using existing diagnostics; verify representation and exact-scale comparability, including genuine Standard vs fast-tree Standard API. No Secondary production changes, no speculative repair, no full 7x timing reruns, no threshold/parameter relaxation. Preserve canonical LogN13 passing result. Web retains mathematical repair authority.

Codex must safe-sync Primary/Secondary per both `AGENTS.md`, read the spec, execute bounded self-review, push only authorized changes and report one spec-defined outcome then `READY_FOR_WEB_REVIEW`.

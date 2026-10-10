# CURRENT TASK — Batch027 E32 Rescale causal hotspot

Task: FAST-E32-RESCALE-CAUSAL-HOTSPOT-AUTONOMOUS-BATCH-027
Status: READY_FOR_CODEX
Mode: ZERO-BOOTSTRAP REUSABLE REAL-RESCALE BENCHMARK + CAUSAL CHECK + AT MOST ONE NEW CANDIDATE

**Authoritative spec:** `specs/FAST-E32-RESCALE-CAUSAL-HOTSPOT-AUTONOMOUS-BATCH-027.md`
**Web review of negative Batch026:** `results/FAST-E32-RESCALE-COEFFICIENT-OPT-AUTONOMOUS-BATCH-026-web-review.md`
**Permanent rules:** both AGENTS.md; Primary docs/MEASUREMENT_PLATFORM.md, workflow §4A/4B; Secondary docs/FAST_QPREFIX_SPEC.md.

Batch026 is ACCEPTED `NO_BENEFICIAL_CANDIDATE`: one safe temporary `mod192By64` Hi==0 candidate tested, five 100ms samples; baseline median 2.692458ms, candidate 2.709171ms (+0.62%, sample ranges overlap), unchanged 20 allocs/op. Candidate was removed; Secondary production source clean/unchanged at `4f2557062cb5c1ffb9a671bc3df67401fd7092b4`. This proves no robust win, not a statistically significant regression or CRT irrelevance. Captured Bootstrap trace stays post-hoc only, raw unchanged.

P1: safe startup sync; reconstruct identical documented Batch026 fixed seed E32 synthetic Level9→8 4→4 in-place public Rescale fixture and verify input SHA. Preserve reusable test-only benchmark in Secondary when correct, even if no production candidate wins.
P2: longer, source-grounded regular-build baseline and CPU pprof if already supported; fallback bounded helper timing if not. Do not invent kernel attribution or inflate synthetic fixture to representative full Bootstrap.
P3: optionally ONE non-math-changing candidate only when P2 proves worthwhile, strict exact correctness and paired/interleaved A/B timing, preserve candidate patch hash even if rejected.
P4: one compact scientific report/tests/vet/diff and safe authorized pushes, then Web review. If no clear hypothesis, complete an honest hotspot/benchmark study rather than guessing a production rewrite.

**Expensive call budget: 0 Fast + 0 Standard Bootstrap; 0 P93/Standard performance.** Never reuse historical Batch023 14/14 or Batch024 2/2. Preserve original formal Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and production Fast `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` pins, unchanged arithmetic/security. No Batch028 permission.

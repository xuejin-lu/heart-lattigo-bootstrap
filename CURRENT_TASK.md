# Current Task

Task: FAST-DROPIN-PUBLIC-NATIVE-REPEATABILITY-AUTONOMOUS-BATCH-023
Status: READY_FOR_CODEX
Mode: REUSE-FIRST FULL E32 BOOTSTRAP REPEATABILITY + DEEP STAGE/POWER/RESCALE ATTRIBUTION (S1–S5)
Web handoff: ONE Web review at S5 or immediate mathematical/safety STOP

**Authoritative revised spec:** `specs/FAST-DROPIN-PUBLIC-NATIVE-REPEATABILITY-AUTONOMOUS-BATCH-023.md`
**Accepted previous review:** `results/FAST-DROPIN-PUBLIC-NATIVE-COLD-WARM-AUTONOMOUS-BATCH-022-web-review.md`
**Permanent rules:** both `AGENTS.md`, Primary `docs/MEASUREMENT_PLATFORM.md`, workflow §4A/4B, Secondary `docs/FAST_QPREFIX_SPEC.md`.

## Scientific correction to previous Batch023

The prior charter focused almost exclusively on collecting five warm total times, with Fast E32 internal stage tracing treated as optional/unavailable. This is insufficient for source-backed diagnosis. Current **Fast production already includes `fastdiag` Stage/Power/Rescale hooks** at Bootstrap root, packing, ScaleDown, ModUp, C2S, EvalMod real/imag, S2C, unpack, finalization, generated powers and per-Rescale phases. The old `cmd/fastdiag` runner invokes a P93/E0 hardcoded test; the missing capability is a **thin E32 public-native diagnostic bridge**, not another tracing system.

S1: inventory/map actual existing E32-capable stage/power/rescale hooks, nested phase hierarchy and profiler tools; freeze fair conditions and a precise reuse-based adapter design, zero Bootstrap.
S2: extend existing `cmd/perfprobe` to 1 cold+5 warm original Standard/Fast without duplicating timers; and adapt existing `cmd/fastdiag` E32 profile with an isolated **test-only Secondary fastdiag diagnostic fixture** (no production code/math changes), based on exact 022 public E32 input rather than old P93/direct-c0. Tests use fake calls only; safe progress commits.
S3: frozen production-timing lane MAX **6 genuine Standard+6 original Fast Bootstrap**, plus distinct diagnostic E32 tracing lane MAX **2 extra Fast**, 14 total real calls including failures/warmups. Trace lane must **not** contaminate production timings; same held Level0 input and valid oracle; conservative irrevocable pre-reservations before spawning possible two-call test processes.
S4: deep ranked E32 stage/power/rescale exclusive/inclusive time, Level/Scale/Q-rows, call counts, event closure, CPU profile only within call budget, source mapping and honest Amdahl bottleneck upper bounds. Reuse P93 DIAG007/OPT004 as history only, not as E32 metrics. No autonomous CKKS algorithm rewriting.
S5: compact results and permanent platform docs, tests/vet, safe Primary and **test-only diagnostic** Secondary commits if needed, Web scientific acceptance. No next Batch authorization.

**Backend pins:** unchanged pristine true Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, original Fast production `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` for all formal public timing. Optional isolated diagnostic Secondary test-only commit must be recorded separately and must prove production Go source identity (and no change to math). Active Secondary branch HEAD may differ from pinned production by prior AGENTS-only history.

**Frozen experiment:** canonical LogN13 E32 full Q/P, 4096 slots, A/B Scale2^45 C actual q5, Public Add→MulRelin→Rescale(q5)→Rotate→DropLevel(4)→Bootstrap, original 1e-6 numeric and q0 capacity, identical Primary frontend source, genuine Standard native vs intentionally insecure Fast zero-secret. NO LogN16, parameter sweeps, new measurement runner, Standard internal hook fabrication, or new production algorithm.

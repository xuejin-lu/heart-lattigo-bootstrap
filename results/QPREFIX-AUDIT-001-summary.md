# QPREFIX-AUDIT-001 summary

**Classification:** `QPREFIX_V2_AUDIT_EVIDENCE_INCOMPLETE`

The fixed LogN13/P93, q0=56 public Bootstrap control passed: generated-secret and zero-secret max component error were both 0.009705381393898434, below 0.01. Secondary `fast-qprefix` remained clean at `636953688f33f00f3404924e3d3f519752583c6a`; focused fast/bootstrapping/DFT/Mod1/polynomial tests passed.

The inherited PS checkpoints (38 names × real/imag) have maximum q012 ratio 1.6321616569370316e-12 and maximum q01 ratio 0.8972907077452122; inherited DoubleAngle local-q2 checkpoints have maximum q012 ratio 8.841439275724772e-12. Recorded S2C post-rescale q01 ratios peak at 0.000005002817953068453. These measurements show substantial capacity in the saved P93 evidence, but their recorded Secondary provenance is older `fast-ckks`, not the task's `fast-qprefix` branch.

The first unresolved required state is **C2S group 0 LinearTransform output before its explicit Rescale at Level 16**. The available C2S trace came from `fast-ckks@7d05f1f` and records the transform at Level 15 / Scale 2^50, identical to the following Rescale checkpoint. It cannot establish the target branch's separate raw transform bound. The per-Rescale source→divisor→predicted/observed output recurrence and target-branch contraction checkpoints are also not fully retained.

Therefore this audit does **not** classify the capped q-prefix architecture as failed, but it also cannot authorize `QPREFIX_V2_FULL_PROFILE_CAPACITY_PROVEN`. The missing current-branch C2S raw output and complete Rescale recurrence are listed in the JSON artifact; no production behavior was changed.

Primary `go test .` has the documented pre-existing failure `TestFIX001P3GenuineStandardPublicVsStagedConsistency` / `P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT`; the task-specific public control and Secondary focused tests passed. The historical test was not changed or bypassed.

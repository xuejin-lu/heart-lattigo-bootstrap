# Web Review — FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A

**Decision: ACCEPT the bounded safety repair.** Do not infer full CKKS drop-in, cryptographic security or runtime speedup.

Primary result `23ee62d01116d84a2575981d126db1cecca07068`; Secondary repair `93ab7ecc5a864fd9bbacea55f0af730941552691` (parent `6ce15cf8b949c8a89f610bcca5ab506dd3f560f8`). Both heads matched remote during review.

Source-level evidence: in Fast-marked `ckks.Evaluator.MulRelin`, ciphertext/ciphertext calls now directly return `mulRelinFastCKKSZeroSecret` result. Nonzero c1 or unsupported degree/storage/representation yields a descriptive error and never returns `handled=false` into native `GadgetProduct`. `MulRelinNew` checks nil operands before allocating a result. Tests use an actual Standard-layout relin key and instrument `GetRelinearizationKey` lookups: rejected pairs leave serialized input/output unchanged with zero lookups. Valid keyless pairs succeed with nil or normal keys. Pinned genuine Standard remains untouched.

The previous same-source LogN13/16-slot Fast primitive regression retained c1=0 at eight checkpoints and reported MulRelin RMSE `5.320587702687035e-15`; no Standard rerun, Bootstrap or benchmark. **Provenance qualification:** the full eight-checkpoint standalone runner used the worktree before the final nil-preallocation guard was added; the final Secondary commit passed the targeted package test suite, including keyless positive and safety-negative fixtures. This supports acceptance of the safety repair but not a claim that the final SHA independently reran the entire original numerical experiment.

This validates **only** full-active-Q degree-one NTT non-Montgomery ciphertext/ciphertext zero-secret MulRelin. It does not establish Q-prefix compact-row semantics, optimized Rotate/Rescale, public-key encryption, general CKKS API replacement or speedup. `FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.

**Next scientific priority:** keyless public Rotate/RotateNew under identical zero-secret contract with full active Q authority. A normal public Rotate currently reaches native `rlwe.Evaluator.Automorphism` and uses a Galois key, although zero-secret rotation can act on c0 alone. Use a separately authorized scope and test no-key operation without promoting the existing separate Q-prefix kernels as general full-Q semantics.

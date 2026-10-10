# Independent Web Scientific Review — FAST-DROPIN-HIGHLEVEL-TO-BOOTSTRAP-AUTONOMOUS-BATCH-018

**Date:** 2026-10-10 (Asia/Taipei).
**Verdict: ACCEPT BOUNDED HIGH-LEVEL COMPACT ADD/ROTATE → JUSTIFIED DROP → E32 PUBLIC BOOTSTRAP COMPOSITION.**
**No full unrestricted Fast drop-in, high-level MulRelin/Rescale-to-Bootstrap, universal DropLevel safety, speedup or cryptographic security claim.**

## GitHub provenance

- Verified Primary remote `main@94309a11b68ee7f3f99115b1b992f5019284d916`; independently inspected three pushed commits: `be70b758a20aa0198a62d700e0b03322214cf884` (initial runner and journal), `86af97092ba666a4a16ab2e7c4dde861ee669a96` (runner validation repair), and `94309a11b68ee7f3f99115b1b992f5019284d916` (results).
- Both final backend runs report clean Primary runner commit `86af97092ba666a4a16ab2e7c4dde861ee669a96`. Verified Secondary remote still `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, not modified by this batch. Genuine unmodified Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40` is recorded, local clean status developer-reported.
- Independently read the committed `018-summary.md`, `018-journal.md`, compact `018-evidence.json`, shared `tools/fast-dropin-highlevel-bootstrap-batch-018/main.go` plus companion adapters/tests, and earlier 016–017 evidence/spec. Web did **not** independently run Go tests/Bootstrap or read local raw 4096-slot vector files. Numerical outcomes are accepted against source-backed committed provenance with this explicit limitation.

## Validated scope and why it is meaningful

The prior conjecture that canonical E32 *residual* `MaxLevel=1` made a full Level5 public source impossible was too strong: the SAME frozen 017 `btpParams.BootstrappingParameters` supply a full-Q `MaxLevel=16` arithmetic key domain, while the original public `GenEvaluationKeys(sk)` makes an extended compatible secret. The shared public source can legitimately encrypt A/B under that full-domain secret at Level5; all source/config/input/QP hash gates and visible E32/LogN13/default 2^45 Scale are matched.

Actual public workflow run:
`EncryptNew(A,B) [full-profile Level5] → AddNew → RotateNew(1) [Level5 Fast Q-prefix compact] → DropLevelNew(5) [Level0, q0 only, Scale unchanged] → native DecryptNew/Decode preflight → public E32 Bootstrap → native DecryptNew/Decode [Level1]`.
There are **no high-level MulRelin or Rescale operations in this 018 workflow**, unlike the earlier separate lower-level 017 chain. `DropLevel` is a projection/discard of higher moduli, not a division/rescaling operation and not a general-purpose way to fix incompatible Scale.

- Initial Fast `EncryptNew` A/B still allocate all six logical active rows; authoritative prefix of 4 used for measurements. The public Fast Add and Rotate outputs at Level5 physically store q0..q3 only; q4/q5 row backing is absent (zero lengths), maintaining c1=0. At `DropLevelNew(5)`, only q0 backing remains. No evidence of stale q4+ reads or silent full-RNS fallback in these paths.
- Conservative workload-specific Fast zero-secret bound: `B_A=2348557866436`, `B_B=1804822278899`, `B_add/rotate=4153380145335`. At the target residual q0 `36028797018652673`, strict `2B=8306760290670 < q0` holds, so this PARTICULAR centered numeric ciphertext safely survives plain DropLevel without relying on q4+ or a special decoder. Existing capacity observer was used at inputs, Add, Rotate and DropLevel; independent bounds were propagated instead of interpreting observer centered maxima as sufficient on their own.
- Genuine Standard retained randomized nonzero c1 and native evaluation; Fast used zero-secret public shared Add/Rotate and public Fast Bootstrap dispatch. Same public API method names, no Fast-only frontend constructors; diagnostic Fast-only capacity observer lives in distinct build adapter, not in the application operation calls.
- Pre-Bootstrap at native residual Level0: direct Fast-vs-Standard RMSE `1.5656982925854002e-11`, max `7.13774172628691e-11`. Post-Bootstrap: RMSE `4.963927268857064e-9`, max `3.93386413412319e-8`. Standard oracle RMSE/max `4.880485998172208e-9/1.8458745857066216e-8`; Fast `1.1829277433501467e-9/5.764906238946737e-8`. All under original `1e-6` maximum complex error gate. Bootstrap returned Degree1/Level1/Scale2^45, native public DecryptNew/Decode, 4096 slots.
- Exactly one Standard and one Fast Bootstrap called in final held sessions. First preflight pair aborted *before* the Bootstrap gate due to an overstrict **test validator** incorrectly requiring Fast public EncryptNew itself to be compact; its fix was committed as `86af970...`. Retried only zero-Bootstrap preflight; expensive call ceiling respected. Focused tests/vet/diff reported passed under both pinned workspaces.

## Precise acceptance boundaries

This is a **real** accepted high-Level compact public Add/Rotate→Level0→Bootstrap composition. Do NOT claim Batch018 ran high-Level `MulRelin` or `Rescale`, demonstrated numerical correctness of q1..q3 decoded components independently at Level5 in this experiment, or proved general high-Level-native `DecryptNew`. The final Level0 cleartext oracle sees q0 only. The strict capacity bound is specific to these input magnitudes, one addition, one rotation and the fixed frozen profile. A different workload may violate `2B<q0` and make such a direct DropLevel unsound.

**Remaining end-goal gaps:** full high-Level compact arithmetic depth involving existing MulRelin and logical-top Rescale before canonical q0→Public E32 Bootstrap, native `DecryptNew` on compact high-Level ciphertexts, public operand overload coverage, measured acceleration. Prior 016 independently covered high-Level MulRelin/Rescale under a different six-Q test profile and measurement-only decoder; the two are not yet proven in one frozen E32 Bootstrap Q chain.

**Next authorized experiment:** bounded `FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019`: mathematically preflight Level5 `MulRelin → Rescale(q5) → Rotate → justified DropLevel` with exact E32 original Q/P, and only if the independent `2B<q0` and fixed Bootstrap Scale gates pass, expend max one Standard+one Fast Bootstrap. No backend kernel rewriting or performance claims.

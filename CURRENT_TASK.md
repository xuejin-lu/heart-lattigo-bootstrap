# Current Task

Task: FAST-STORAGE-004
Status: COMPLETE

Accepted Secondary implementation:
`1b9ecd7973505cac1c4a1673a9578260a761950f`

Classification:
`FAST_STORAGE_004_LOGICAL_RESCALE_ACCEPTED`

Accepted result:
- standalone one-step private-storage Rescale implemented as `FastStorageRescale`;
- coefficient rounded division and Scale division both use the frontend logical modulus `q_ell = params.Q()[LogicalLevel]`;
- private storage moduli `f_i` are never used as CKKS Rescale divisors;
- coefficient-domain and NTT-domain private-F inputs are supported;
- NTT inputs follow the explicit F-INTT -> integer rounded division -> F-NTT path;
- logical Level decrements by one;
- Scale divides by the same logical `q_ell`;
- degree, storage width, representation domain, public parameters, and compatible plaintext metadata are preserved;
- per-component bounds use the frozen exact transition
  `floor((B_j + (q_ell-1)/2)/q_ell)`;
- storage contraction is not performed;
- fixed-width signed 192-bit rounded division is used in the coefficient hot loop without per-coefficient `math/big`;
- Standard CKKS logical Rescale oracle passes, including noncanonical lifts `X = c + k Q_ell`;
- chained `FastStorageMul -> FastStorageRescale` passes while preserving raw multiplication degree.

Independent review:
- commit changes only the new `storage_rescale.go` and its focused test file;
- historical production `Evaluator.Rescale` / `RescaleTo` are unchanged;
- `rlwe.Scale.Div` is non-mutating, so the implementation preserves the input transactional contract;
- fixed-width long division uses `bits.Div64` with valid remainder preconditions and exact odd-divisor nearest rounding;
- no hidden Standard/full-RNS fallback is present.

Validation reported:
- targeted storage Rescale tests passed;
- `go test ./schemes/ckks/fast ./circuits/ckks/bootstrapping` passed;
- `go test ./...` passed;
- `git diff --check` passed.

No benchmark gate applied for this correctness-foundation task.

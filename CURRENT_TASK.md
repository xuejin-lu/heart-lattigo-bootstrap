# Current Task

Task: FAST-STORAGE-005
Status: COMPLETE

Accepted Secondary implementation:
`531aca50b5b38741e4e71cc98ea4b626bf88cb84`

Classification:
`FAST_STORAGE_005_LEVEL0_MODUP_CANONICALIZATION_ACCEPTED`

Accepted result:
- standalone `FastStorageModUpLevel0` canonicalizes a valid Level-0 width-3 private-F ciphertext before logical basis growth;
- canonicalization is exactly `Center_q0(X mod q0)`;
- target logical Level is explicit and bounded by public CKKS parameters;
- Scale, degree, domain, plaintext metadata, parameter identity, and width 3 are preserved;
- output bounds are reset to the exact observed maximum magnitude after canonicalization;
- coefficient and NTT private-F inputs are supported;
- private storage remains exactly three F rows regardless of logical target Level;
- no adaptive width or contraction is performed;
- noncanonical lifts `c + k q0` collapse to the same canonical output;
- `Rescale -> ModUp` boundary behavior is covered by focused tests;
- historical production `FastEvaluator.modUpBasis`/`ModUp` and Bootstrap are unchanged.

Independent review:
- commit changes only `storage_modup.go` and `storage_modup_test.go`;
- implementation reconstructs the authoritative private-F lift before canonicalization rather than merely changing Level metadata;
- signed reduction for negative lifts is consistent with the centered odd-q0 convention;
- NTT inputs explicitly follow F-INTT -> integer canonicalization -> F-NTT;
- component bounds are recomputed from canonicalized coefficients;
- transactionality is preserved because outputs are newly allocated and input state is not modified.

Validation evidence available from repository review:
- focused tests cover centered q0 boundaries, coefficient/NTT exactness, target levels, bound reset, noncanonical-lift equivalence, transactional failures, logical export, and Rescale -> ModUp chaining;
- repository has no configured remote commit status/check runs for this commit, so acceptance does not claim independent CI execution.

The initial production storage policy remains fixed width 3.

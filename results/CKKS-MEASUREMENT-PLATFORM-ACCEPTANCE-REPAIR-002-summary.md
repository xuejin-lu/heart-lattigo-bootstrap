# CKKS measurement platform acceptance repair — 002

## Outcome

Status: `READY_FOR_WEB_REVIEW`. This closeout completed the existing platform acceptance gaps without creating a second measurement framework, rerunning a Bootstrap, changing Secondary production code, or altering Task005/Task006 source evidence. The first implementation commit is `e8289d484daff610dbb517f0d1ff4f10861a1bfc`.

## Parameter-driven shared entrypoints

- `cmd/perfprobe` now distinguishes immutable `frozen-historical-fixture-v1` profiles from `effective-parameterized-v1` profiles. Frozen LogN13/LogN16 artifact and hash contracts remain intact; a supplied config is resolved from constructed `perfmeasure.EffectiveParameters`, and an optional profile label must agree with its effective dimensions.
- `cmd/fastdiag numerical --config ... --ephemeral-secret-weight ...` uses the same shared parameter builder and deterministic input. Effective LogN, LogSlots, slots, Scale, Q/P and E are recorded/derived from the constructed parameters. Its generic c0/zero-c1 input remains diagnostic-only, not the formal native Standard lifecycle.
- Structural construction tests cover LogN13/14/15/16. They are not numerical acceptance claims. The formal public-native workload remains E32; unsupported backend/E/topology conditions fail explicitly.
- The noncanonical generic LogN16/E32 strict-capacity preflight is intentionally rejected before backend construction or CKKS operations: `B=18169085704944493`, `q0=36028797019488257`, deficit `309374390400729`. No workload or capacity threshold was changed. This does not relabel the separate pinned historical LogN16 observed-oracle fixture.

## Task006 offline trace repair

The existing 539-event capture was replayed using Batch025's `analyzePublicE32Events`; no live trace capture or Bootstrap was run. The raw artifact remains `TRACE_UNVERIFIED`, with its original `TRACE_COLLECTED_UNVERIFIED` classification unchanged. The new derived result is `OFFLINE_RECURSIVE_TRACE_REVALIDATED`; new Bootstrap calls: **0**. Existing Task006 ledger and two reserved attempts are preserved.

The Bootstrap root is 1,060,517,791 ns. The compact direct-child table contains nine nonoverlapping Bootstrap stages, including CoeffsToSlots (152.657 ms), EvalMod real (375.138 ms), EvalMod imag (382.586 ms), SlotsToCoeffs (145.066 ms), and public finalization (0.073 ms). Generated-power branches and all 35 Rescale roots are reported as independent instrumentation trees, not added to Bootstrap wall time. Full counts, ns/ms values, shares, root closures, and recorded numerical metadata are in the linked report; no decrypt or vector reconstruction was repeated.

The retained LogN16/E32 live trace entrypoint now refuses the retired adapter path; any remaining event validation delegates to the shared recursive validator. Positive and negative synthetic ancestry tests cover the former direct-child-only defect. The captured diagnostic adapter commit is recorded as `649da85757a97dc8c60b5f9cc031634f39ad4a40` and its patch digest matches the captured metadata, but that historical commit object is absent from the current Secondary clone; attribution is therefore an offline captured-trace analysis, not a fresh rebuild of that adapter.

Artifacts:

- [Task006 compact stage report](CKKS-MEASUREMENT-PLATFORM-ACCEPTANCE-REPAIR-002-task006-summary.md)
- [Task006 replay journal](CKKS-MEASUREMENT-PLATFORM-ACCEPTANCE-REPAIR-002-task006-journal.md)
- [Task006 compact evidence](CKKS-MEASUREMENT-PLATFORM-ACCEPTANCE-REPAIR-002-task006-evidence.json)

Raw capture SHA-256: `9bd5ba81abce75b8fd7c06ac544410246fb385bafb1585a646d7d71e2f8e872d`; sidecar SHA-256: `f598b481b776448dcc78ff80083610e9376cee12451716ddb88fd414abfc756b`.

## Retired-test coverage audit

The earlier accepted retirement removed 188 Go files, including 75 `_test.go` files. Representative high-risk invariants were checked against maintained Secondary tests; no unique production invariant was found that required restoring the deleted diagnostic runners or adding another adapter.

| Retired test concern | Maintained coverage |
|---|---|
| Ciphertext copy/degree growth and output ownership around packing/finalization | `TestFastCiphertextCopyAndDegreeGrowthPreservePrefixShape`; `TestFastPackingIgnoresDormantResiduesAndOwnsOutputs` |
| Finalization row authority and malformed-row transactionality | `TestFastBootstrapPublicBoundaryRejectsRowsBeyondContract`; `TestFastBootstrapFinalizationRejectsMalformedRowsTransactionally` |
| IMForm/MForm evaluator-domain transitions and scalar paths | `TestFastEvaluatorMulAndMulRelinNTT`; `TestFastEvaluatorScalarMontgomery`; `TestFastEvaluatorScalarMulThenAdd` |
| Aliasing/squaring and dormant-row isolation | `TestFastEvaluatorMulAliasingSquaringAndPoison`; `TestFastLinearTransformIgnoresDormantAndSupportsAlias`; packing ownership test above |
| Malformed public/keyless MulRelin input rejection | `TestPublicCKKSEvaluatorKeylessZeroSecretMulRelinRejectsUnsafeInputs`; `TestPublicCKKSEvaluatorFastMulRelinFailsClosedWithStandardRelinKey`; finalization transactionality test above |

The historical diagnostic-copy helpers were not restored. The table is a representative invariant map, not a claim that every deleted test was migrated one-for-one; historical source and exact deleted paths remain recoverable from Git history and the prior retirement evidence.

## Validation and repository scope

Passed:

- `go test ./cmd/fastdiag -count=1`
- `go test ./cmd/perfprobe -run 'Test(ParseArgsAndScopeValidation|PublicBootstrapGateRequiresExactPassingZeroCallPair|PublicBootstrapGateBindsTheCurrentFastLogN13Pin|PublicBootstrapGateRequiresVersionedLogN16ObservedOracleEvidence|PublicNativeProfilesResolveFromEffectiveLogN|PublicNativeProfileConstructionUsesEffectiveParametersAcrossLogN13To16|PublicNativeFrozenProfilesRequireTheirCanonicalConfigHash|ParameterizedPublicPairValidatesAgainstItsEffectiveMetadata|LogN16PublicNativeProfileUsesOnlyObservedOracleCapacityPolicy|PublicPairRejectsDecodedVectorIntegrityFailures|PublicPairRejectsForgedStateRowsAndCapacity)' -count=1`
- Captured-trace replay: `FASTDIAG_TASK006_RAW=... FASTDIAG_TASK006_SIDECAR=... go test ./cmd/fastdiag -run '^TestTask006OfflineReplayCapturedTraceWithoutBootstrap$' -count=1`
- `go test ./... -run '^$'` (all Primary packages compile; no cryptographic tests executed)
- `go vet ./cmd/fastdiag ./cmd/perfprobe ./internal/perfmeasure ./internal/numericalmetrics`
- `git diff --check`

No full Bootstrap test or benchmark campaign was run. Secondary `fast-qprefix` was not modified; final HEAD and clean status are recorded in the completion message after verification. The focused tests and structural construction checks do not establish new LogN14/15/16 Fast numerical correctness.

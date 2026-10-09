# Independent Web Milestone Review — FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012

**Verdict: ACCEPT BOUNDED MILESTONE**, not general CKKS drop-in completion, security, performance, or compact-Q public interoperability.

## Repositories and scope inspected

- Primary remote `main@a00ceb7b4dbd50cbd4b5f442d3880760ac531127` at review; Secondary Fast remote `fast-qprefix@00ac70ba136d190fa31bbb26c2f51d003a221634`, unchanged. Genuine Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Inspected committed `012-summary.md`, `012-journal.md`, `012-coverage-audit.md`, and compact paired preflight and Bootstrap JSON artifacts. This Web review checks the committed evidence and its internal consistency; it **did not independently run Go tests or the Bootstrap experiments**, and machine-local full 4096-slot arrays are not committed, so per-slot error claims have not been independently recomputed by Web.
- 012 used an authorized 4-checkpoint autonomous charter, with no intermediate Web handoff. Its key-plan fix only modified a Primary shared test runner; Lattigo Standard/Fast production code and frozen historical 004/010 sources were untouched.

## What the evidence supports

**Checkpoint A, key-plan correction:** Previously, native Standard public Rotate panicked because a bootstrap-generated Galois Key with five P primes was paired with a P-free residual evaluator. The new shared runner constructs the pre-Rotate `ckks.NewEvaluator(btpParams.BootstrappingParameters, keys.MemEvaluationKeySet)`, checks same N and Standard ring, actual ordered residual-Q prefix, rotation key presence, P-level matching, and secret-q0 extension. No param tuning or library repair was necessary.

**Checkpoint B, public Rotate preflight:** Both pinned builds reported `preflight_passed`, with the same fixed config/input and the same measured frontend SHA **for this B checkpoint** (`14435c478ee9af0ba2ae0e3548850ed3c121b69141c4e0918d8a05bee4650e7f`). Genuine Standard had nonzero c1 and one `GetGaloisKey` lookup; Fast c1 stayed zero, with zero lookup. Their Rotate oracles passed; respective RMSE `1.342087606921074e-11` and `7.44997342660252e-13`. A comparison-schema defect required derived nested Galois element fields using preexisting captured top-level values; the repair is transparently documented without rerunning B.

**Checkpoint C, same-source Rotate → Bootstrap:** Both pinned builds reported `bootstrap_completed`, **exactly 1 Bootstrap call each**. Crucially, both C records used identical measured frontend source SHA `a19841d94311d95a6f0cbab9a3301e49d5820fb5660885f1037a088e1a9204ad`, config SHA `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`, and input SHA `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`, with matching actual Q and P hash fields. Separate Primary HEADs during C (Standard `30a312996b5916de97b1e14b4af4438d95660ebe`, Fast `a682e8616e441f0c9fba15b3bf888187fba276d2`) reflect journal bookkeeping; common frontend/config/input hashes match. Same residual input Level 0, log2Scale 45, 4096 slots, E32; Bootstrap outputs both Degree 1, Level 1, log2Scale 45, two full active Q rows. Standard nonzero c1, Fast zero c1.

- Bootstrap-vs-rotated-cleartext oracle RMSE: Standard `4.885854924910455e-9`; Fast `1.1814120448423705e-9`.
- Direct decoded Fast-vs-Standard Bootstrap RMSE `4.954605783029611e-9`. Max complex output error to oracle: Standard `1.5351218791077742e-8`, Fast `5.771241651805313e-8`; do not claim Fast is better on every metric or across inputs.
- Real Fast public Bootstrap dispatch is evidenced by `fast_bootstrap_selected=true`, paired key layout, and the source-backed coverage report. No runtime speedup or crypto-equivalent noise/security is implied.

**Checkpoint D:** The 012 coverage audit correctly distinguishes directly executed `EncryptNew`, `Rotate`, `Bootstrap`, `DecryptNew/Decode` from standalone `Add/Sub`, `MulRelin` variants and `Rescale`, which are not individually demonstrated by *this batch*. Their other historical tests remain valid within their original narrower scopes; do not describe them as entirely absent implementations. It also correctly preserves the P0 full-active-Q vs explicit C0 compact-prefix distinction and no implicit stale-row reads.

## Provenance caveats

The preflight B and Bootstrap C executions used **different frontend source hashes** because of the documented reporting-schema correction. Comparison must therefore be per phase, not mislabel B and C as an identical byte-for-byte runner across both phases. Within each phase Standard and Fast used identical source, input and config. Later modifications were reporting/test focused; focused Go tests and vet under both pinned workspaces passed according to Codex's report, but Web has not independently rerun them. The committed compact evidence is reviewable; raw full-output vectors stay machine-local. A larger performance or accuracy claim would require its own source- and run-verifiable artifacts.

## Decision and next architecture gate

**Accept 012 as a bounded transparent public Rotate→Bootstrap numerical composition milestone.** No Secondary repair is authorized by this result. Preserve immutable Standard, intentional Zero-Secret insecurity, no automatic P0↔C0 conversion, no speedup claim, and retain `FAST-STANDARD-PERF-REBASELINE-003` as blocked.

Authorize a new **bounded autonomous E/I coverage batch** with no modifications to either Lattigo library. Reuse already-existing 005 and 006A/009 source evidence. Test same-source public Add/Sub, ciphertext×ciphertext and supported plaintext/scalar MulRelin branches, standalone Rescale and a representative public-operation chain on P0 full-active-Q. For each distinguish (1) cleartext oracle, (2) correct branch actually used, (3) Level/Scale/active Q authority and forbidden native fallback, and (4) true performance evidence (none authorized). Do not implement a second Fast arithmetic kernel. Unexpected mathematical/representation gaps escalate to Web instead of becoming autonomous backend refactors.

Next task charter: `specs/FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013.md`.

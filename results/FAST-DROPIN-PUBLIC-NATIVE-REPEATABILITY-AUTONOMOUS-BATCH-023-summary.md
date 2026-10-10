# FAST-DROPIN-PUBLIC-NATIVE-REPEATABILITY-AUTONOMOUS-BATCH-023

**Outcome: `BATCH_BLOCKED_NEEDS_WEB_REVIEW`**

The frozen LogN13/E32 public-native preflight and complete uninstrumented
Standard/Fast repeatability lanes passed. The separate Fast trace lane used
its two reserved calls; both outputs passed their numerical oracles, but the
first Rescale event-tree assertion failed. The emitted parent had a
`preflight` child only, while the validator requires both `preflight` and
`materialization`. Raw event JSON was not emitted. All 14 authorized
Bootstrap calls are spent. Available evidence cannot distinguish a missing
span from an overly strict validator; no retry or further experiment is
permitted pending Web review.

## Provenance and fixed conditions

- Primary measurement source: `6e938918442c409faa6d32e159c7a3f7a1041f7a`,
  SHA-256 `b63fef5bbbdfe1e679b2e50ab66e0a5c43afefc526bd2228b16b7541ae01ea63`.
- Genuine Standard Lattigo: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Formal original Fast Lattigo: `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
- Diagnostic Secondary HEAD: `f7eb9f88d0c877331e62287508c541a1e1147bdc`.
  Non-test Go production-tree SHA-256 is
  `a1ddabc7bfcdefcb8e0e0e2028966184f24d3fc4c68530f8e2919f0507a3aa00`,
  identical to the formal Fast pin. The only Go-file delta is the test-only
  `circuits/ckks/bootstrapping/fastdiag_public_e32_test.go`, SHA-256
  `43bf6e467c9e6f4f86ad54c0fa29f708747bc76c7aa78c470671d8323b88c0bb`.
- Workload: LogN13/E32, 4096 slots. Config/workload/Q-prefix SHA-256:
  `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`,
  `00b70a2e77887c7d6a41db1859246f6513f2c984915de648734e5dd8c584cf74`,
  `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`.
- Go 1.26.4, darwin/arm64, Apple M4, 10 CPUs, `GOMAXPROCS=10`; GC settings
  were runtime defaults. Formal build tags: `perf_standard` / `perf_fast`;
  trace used the `fastdiag` test-only lane. Trace overhead is unavailable
  because no complete event tree/root duration was serialized; traced timing
  is not mixed with formal samples.

## Bootstrap ledger

| Lane | Calls | Limit | Result |
|---|---:|---:|---|
| Genuine Standard public-native | 6 | 6 | All six decoded-output oracles passed |
| Original Fast public-native | 6 | 6 | All six decoded-output oracles passed |
| Fast diagnostic trace | 2 | 2 | Both call oracles passed; event-tree gate failed afterward |
| **Total** | **14** | **14** | **Exhausted; do not retry** |

Preflight and test processes spent zero calls. Both trace tokens were reserved
before process launch and are irrevocably spent.

## Numerical and capacity gates

Both preflight lanes passed all eight required checkpoints, plaintext oracles,
and matched-environment comparison (8/8 matched states). Capacity was
`B=2271135713118062`, `q0=36028797018652673`, with `2B < q0`. Formal results
matched at Level 1, Scale `2^45`, and two Q-prefix rows. All six paired
Bootstrap comparisons passed the `1e-6` max-complex threshold. Each had
complex RMSE `5.0361026156396345e-9`, max complex difference
`4.062428762032295e-8`, worst slot 0. Fast remains intentionally zero-secret;
this is not a security/noise-equivalence claim.

| Backend | Cold ns | Five warm samples ns | Warm median ns | Warm median allocations |
|---|---:|---|---:|---:|
| Standard | 356006584 | 303893791, 305669291, 306636208, 309936084, 310854375 | 306636208 | 91750008 B / 35024 allocs |
| Fast | 527153500 | 72302458, 74076542, 72580250, 72008291, 73113500 | 72580250 | 7505112 B / 8085 allocs |

Warm median ratio is `4.225x` for this profile/machine only. Off-timer
GenEvaluationKeys / CKKS evaluator / Bootstrap evaluator construction (ns):
Standard `403513666 / 1861250 / 354975917`; Fast
`823817666 / 900958 / 2498792`.

## Trace gate and attribution limits

The failure was at
`circuits/ckks/bootstrapping/fastdiag_public_e32_test.go:328` (assertion
`:492`): first Rescale direct children were `{preflight}`, expected
`{preflight, materialization}`. It failed before raw event JSON serialization.
The diagnostic cold and traced warm outputs both passed decoded-oracle and
original-Fast-reference checks.

The warm CPU profile is preserved outside Git at
`/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-4268150718/profiles/warm-cpu.pprof`
(3131 bytes, SHA-256
`c12b7e3bcb63fe935e5750573f5019a4cb102640328f90b5d63cbebc35755162`). No
heap profile or raw event JSON was produced. Stage/power/rescale Pareto,
event closure, trace overhead, pprof attribution, and Amdahl bounds are
therefore **unavailable**, not zero or inferred. The clean diagnostic checkout
is preserved outside Git at
`/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-4268150718/secondary-diagnostic`.

Primary focused tests/vet, pinned Standard- and Fast-tagged tests, and the
Secondary synthetic event-shape and compile-only tests passed; exact commands
are in the journal. `git diff --check` and JSON syntax validation passed for
these report artifacts. `docs/MEASUREMENT_PLATFORM.md` was not updated because
the failed event-tree gate does not establish a valid reusable E32 trace.
Recommended next candidate for Web approval: inspect the first Rescale event
producer and validator expectation, then add/adjust zero-call shape tests
before proposing any new trace budget. Any future Bootstrap call needs a new
explicit task/budget authorization.

**Handoff:** `BATCH_BLOCKED_NEEDS_WEB_REVIEW`.

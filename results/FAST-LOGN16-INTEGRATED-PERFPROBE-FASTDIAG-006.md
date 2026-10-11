# FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006

**Classification:** `TRACE_COLLECTED_UNVERIFIED`
**Trace status:** `TRACE_UNVERIFIED`
**Bootstrap accounting:** Task005 remains exhausted at 16/16 and was not reused; Task006 reserved and used exactly 2/2 Fast-only calls. No retry occurred.

## Provenance and numerical outputs

The run used Primary `aff322665f6d178f159408d66eae39a2823815f5`, clean production Secondary Fast `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`, and an isolated test-only adapter commit `649da85757a97dc8c60b5f9cc031634f39ad4a40`. The fixture was the LogN16/E32 native public input with explicit zero Bootstrap budget/calls and eight passing zero-call pre-Bootstrap checkpoints. The isolated adapter's `FASTDIAG_VALIDATE_ONLY=1` gate passed against the same fixture and the committed Task005 Fast result/vectors before the two calls.

| Evidence source | Timing | Numerical result |
|---|---:|---|
| Task005 Standard, uninstrumented warm median | 6425.275750 ms | — |
| Task005 Fast, uninstrumented warm median | 663.878208 ms | 9.6784× vs Standard |
| Task005 terminal Fast-vs-Standard | — | RMSE `4.09656347139061e-8`; max `4.345682976228596e-7` |
| Task006 Fast cold call | 12279.909583 ms | Oracle RMSE `5.82126707846636e-9`; max `4.6694655714754984e-7` |
| Task006 Fast traced warm call | 1060.782625 ms | Oracle RMSE `5.82126707846636e-9`; max `4.6694655714754984e-7` |

Both Task006 outputs were Level 1, degree 1, Scale `2^45`, Q-prefix rows 2. Their stored comparison to all six compatible Task005 Fast reference outputs was exact (`RMSE=0`, max difference `0`). The two Task006 output-oracle gates passed the `1e-6` max-complex threshold. Task006 timings are instrumented, Fast-only observations; they are not a new Standard/Fast speed comparison and are not directly comparable with Task005 timings.

## Why the trace is unverified

The warm run captured 539 events, but the isolated test-only event validator rejected the recursive generated-power tree with `power event is not nested under generated_powers`. The first rejected event is sequence 61 (`T8`), parent sequence 60 (`T16`); sequence 60 is itself a child of `generated_powers` sequence 59. Thus the observed chain is `generated_powers#59 → T16#60 → T8#61 → T4#62 → T2#63`. The validator required every power to be a direct child of the root, rather than accepting the task-required recursive recurrence ancestry.

Because the runtime structural gate failed, this report does **not** certify stage, power, or Rescale attribution; it does not publish hotspot shares, Pareto/Amdahl results, or inferred stage timings. The raw trace and failure record are preserved at the exact paths and SHA-256 values in the JSON artifact. The attempt ledger and both irrevocable reservation journals are committed alongside this report. Task006's 2-call budget is exhausted; no further Bootstrap is authorized or attempted.

See [`FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006.json`](FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006.json) for compact provenance, metrics, evidence hashes, and call accounting.

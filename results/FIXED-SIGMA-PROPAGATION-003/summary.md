# FIXED-SIGMA-PROPAGATION-003 result summary

**BATCH_COMPLETE_READY_FOR_WEB_REVIEW** — CP-A was already passed and was not repeated. Web classified the one incidental LogN16 parameter-construction test as `NON_BLOCKING_INCIDENTAL_TEST_SCOPE`; it did not run a LogN16 operation, workload, Bootstrap, or measurement. This one-time decision does not authorize future LogN16 activity or increase budgets. Prior completed 002 evidence is unchanged. Freeze σ exactly `4.004280012097432`.

| Question | Result |
|---|---|
| OFF/OFF Stage-2 total oracle RMS | canonical 7.3940e-13; stronger C×4 7.5777e-13 |
| ON/OFF transmitted Stage-1 error P1 RMS | canonical 6.6046–6.6229e-13; stronger C×4 1.3261–1.3270e-12 |
| OFF/ON Stage-2 total oracle RMS | canonical 1.0407–1.0427e-11; stronger C×4 1.0283–1.0366e-11 |
| ON/ON Stage-2 total oracle RMS | canonical 1.0413–1.0429e-11; stronger C×4 1.0298–1.0398e-11 |
| First-order δ1×C RMS; residual RMS | canonical 2.0977–2.1384e-13; 6.2146–6.2554e-13. C×4 8.5013–8.5416e-13; 1.0197–1.0224e-12. Residual left unattributed. |
| ON/ON−OFF/ON consistency with P1 | canonical difference RMS 7.1354–7.2013e-21; C×4 1.2229–1.2477e-19 |
| Genuine Standard Stage-2 total oracle RMS | canonical 1.0378e-11; stronger C×4 1.0415e-11; one native reference per workload |
| Inputs and capacity | canonical plus predeclared C×4; both strict centered-Q-prefix runtime gates passed; frozen `σ=4.004280012097432` |
| CP-C checker | NO_CHANGE; suggested scratch reuse needs adversarial exact parity and comparable post-change profiling |
| CP-D Bootstrap map | PASS_READ_ONLY; no single safe public injection hook; zero Bootstrap calls |
| Chain executions | genuine Standard 2/2; Fast 16/16; Bootstrap 0/0 |
| LogN16 / BTS | 0 LogN16 workload or chain runs; BTS 0. The one incidental parameter-construction test was accepted by Web with no future authorization. |
| Targeted test evidence | Fast and genuine Standard `TestFixedSigmaAblation|TestFixedSigmaChain`; `TestSummarizeComplexMagnitudeTails`; bounded JSON reader and scale fingerprint regression; all passed. Avoid package-wide Go tests. |
| Compact artifact | `cp-b-analysis.json`, 95145 bytes, SHA-256 `5d59e622ce281e8e59e9d7db58dfdf7250d7eb63b9a672f3dca0d729b93c0ed4`; includes per-run result/vector SHA-256 and aggregate/tail metrics. |
| Source / binary provenance | Captures from Primary `2d9ffd4643b590d38da3c892b670a998a39b92c1`; Fast `1e7828293c36f9845e30723ba4e9500d84bb50d2`; genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Binary hashes and analyzer provenance are in journal and artifact. |

Propagation is observed and the ON/ON context agrees with P1 to numerical precision. The elementary `δ1×C` model does not explain the full P1; operation/encoding/rounding contributions remain unattributed. Two seed pairs across two workloads do not establish transferability of the frozen sigma. No Gaussianity, security equivalence, or Bootstrap noise-fidelity claim is made. The Web incident decision and prevention steps are recorded in the journal.

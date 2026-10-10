# Batch025 — Fast E32 trace offline salvage

- Classification: **`OFFLINE_REVALIDATED_TRACE`**
- Source raw status: `TRACE_UNVERIFIED` (relabelled=false)
- New Bootstrap calls: **0**
- Raw: `/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-2533882906/raw-trace-unverified.json` (177696 bytes, SHA-256 `f922587fa944da7ef07dcb94a9497505018f2fee2bc67e7f87af6d35bc3d1c74`)
- Sidecar: `/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-2533882906/trace-failure.json` (SHA-256 `59a872b2df18ca32859f4753650c3f1e14633dd79d1358b283204c4445b3f907`)
- Fixture manifest: `/private/tmp/fast-dropin-batch023-final.3VxTOp/fast-e32-trace.manifest.json` (SHA-256 `5bc19b1eb6e67997ca3d19f39e6fd781bf93b27ad305a1be733f0c883611b847`)
- Six-output Fast vectors: `/private/tmp/fast-dropin-batch023-final.3VxTOp/fast-repeatability-vectors.json` (SHA-256 `63ad89ce9ccc70a0b737397810d4f4a15d8a1ed65e82ccd3af0435d1df4d6881`)

## Provenance

- Captured Primary commit: `4d77c511b27ea303fc5a65870dff1404011d1807`; fixture Primary commit: `6e938918442c409faa6d32e159c7a3f7a1041f7a`; current offline analyzer: `75a8563d7371970141bb6705f0e7132d8aab610f` (`main`, dirty=false).
- Captured measurement source SHA-256: `b63fef5bbbdfe1e679b2e50ab66e0a5c43afefc526bd2228b16b7541ae01ea63`; config `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`; Q/P `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`; input `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`; workload `00b70a2e77887c7d6a41db1859246f6513f2c984915de648734e5dd8c584cf74`.
- Production Fast pin: `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`; diagnostic Secondary: `4f2557062cb5c1ffb9a671bc3df67401fd7092b4` (current `4f2557062cb5c1ffb9a671bc3df67401fd7092b4`, `fast-qprefix`, dirty=false).
- Diagnostic-vs-production source delta: `true`, paths `[schemes/ckks/internal/fastcore/rescale.go]`.
- Reservation journals: 2/2; captured actual calls 2/2. Reservation entries prove spent tokens; the raw holds the two output records.

## Preserved output metadata

These are checks of already-recorded metadata and vector hashes, not a new native decrypt or numerical experiment.

| Phase | Elapsed (ms) | Level | Scale | Degree | Q-prefix rows | Oracle RMSE | Oracle max | Fast-reference max | Vector SHA match | Gate metadata |
|---|---:|---:|---|---:|---:|---:|---:|---:|---|---|
| first_cold_bootstrap | 568.862 | 1 | 3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13 | 1 | 2 | 1.18e-09 | 5.739e-08 | 0 | true | true |
| warm_bootstrap_01 | 122.202 | 1 | 3.51843720888320000000000000000000000000000000000000000000000000000000000000000000e+13 | 1 | 2 | 1.18e-09 | 5.739e-08 | 0 | true | true |

Captured numerical gate: `1e-06`. The Fast native output's reported decoded hashes match the corresponding held repeatability vectors; raw records report oracle RMSE 1.18e-09 and max complex difference 5.739e-08. No numerical metric was recomputed from a new decrypt.

## Event structure and timing

The full per-event inclusive/exclusive closures, root-partitioned Pareto, and conditional Amdahl scenarios are in the sibling `-evidence.json` artifact under `root_partitioned_event_analysis`. Negative signed residuals indicate observed child timing exceeds its parent; they are retained rather than normalized. Independent Power and Rescale trees are never combined with Bootstrap-root percentages.

- Events: 539; independent roots: 38; events with negative exclusive residual: 0.
- Stage counts: `coeffs_to_slots=1, evalmod_imag=1, evalmod_real=1, mod_up_trace=1, pack_n1_to_n2=1, public_finalization=1, scale_down=1, slots_to_coeffs=1, unpack_n2_to_n1=1`.

### Descriptive event-type medians (no shares)

| Scope | Event | Component | Parent type | Count | Median (ms) |
|---|---|---|---|---:|---:|
| power | chebyshev_doubling | — | power/power/T16 | 2 | 0.016 |
| power | chebyshev_doubling | — | power/power/T2 | 2 | 0.016 |
| power | chebyshev_doubling | — | power/power/T3 | 2 | 0.016 |
| power | chebyshev_doubling | — | power/power/T4 | 2 | 0.016 |
| power | chebyshev_doubling | — | power/power/T6 | 2 | 0.016 |
| power | chebyshev_doubling | — | power/power/T8 | 2 | 0.016 |
| power | copy_workspace | — | power/power/T16 | 2 | 0.001 |
| power | copy_workspace | — | power/power/T2 | 2 | 0.000 |
| power | copy_workspace | — | power/power/T3 | 2 | 0.000 |
| power | copy_workspace | — | power/power/T4 | 2 | 0.000 |
| power | copy_workspace | — | power/power/T6 | 2 | 0.000 |
| power | copy_workspace | — | power/power/T8 | 2 | 0.000 |
| power | generated_powers | — | — | 2 | 19.528 |
| power | mul_relin | — | power/power/T16 | 2 | 0.082 |
| power | mul_relin | — | power/power/T2 | 2 | 0.079 |
| power | mul_relin | — | power/power/T3 | 2 | 0.092 |
| power | mul_relin | — | power/power/T4 | 2 | 0.078 |
| power | mul_relin | — | power/power/T6 | 2 | 0.079 |
| power | mul_relin | — | power/power/T8 | 2 | 0.078 |
| power | power | — | power/generated_powers | 2 | 12.981 |
| power | power | — | power/power/T4 | 2 | 3.241 |
| power | power | — | power/power/T6 | 2 | 3.280 |
| power | power | — | power/power/T8 | 2 | 6.468 |
| power | power | — | power/generated_powers | 2 | 6.546 |
| power | power | — | power/power/T16 | 2 | 9.726 |
| power | recurrence_correction | — | power/power/T16 | 2 | 0.009 |
| power | recurrence_correction | — | power/power/T2 | 2 | 0.008 |
| power | recurrence_correction | — | power/power/T3 | 2 | 0.049 |
| power | recurrence_correction | — | power/power/T4 | 2 | 0.008 |
| power | recurrence_correction | — | power/power/T6 | 2 | 0.008 |
| power | recurrence_correction | — | power/power/T8 | 2 | 0.008 |
| power | rescale | — | power/power/T16 | 2 | 3.146 |
| power | rescale | — | power/power/T2 | 2 | 3.135 |
| power | rescale | — | power/power/T3 | 2 | 3.122 |
| power | rescale | — | power/power/T4 | 2 | 3.122 |
| power | rescale | — | power/power/T6 | 2 | 3.161 |
| power | rescale | — | power/power/T8 | 2 | 3.153 |
| rescale | coefficient_loop | c0 | rescale/preflight | 6 | 1.458 |
| rescale | coefficient_loop | c0 | rescale/preflight | 6 | 1.438 |
| rescale | coefficient_loop | c0 | rescale/preflight | 2 | 1.461 |
| rescale | coefficient_loop | c0 | rescale/preflight | 1 | 1.440 |
| rescale | coefficient_loop | c0 | rescale/preflight | 1 | 1.488 |
| rescale | coefficient_loop | c0 | rescale/preflight | 1 | 1.453 |
| rescale | coefficient_loop | c0 | rescale/preflight | 1 | 1.513 |
| rescale | coefficient_loop | c0 | rescale/preflight | 1 | 1.135 |
| rescale | coefficient_loop | c0 | rescale/preflight | 1 | 1.421 |
| rescale | coefficient_loop | c0 | rescale/preflight | 1 | 1.434 |
| rescale | coefficient_loop | c0 | rescale/preflight | 2 | 1.460 |
| rescale | coefficient_loop | c0 | rescale/preflight | 2 | 1.463 |
| rescale | coefficient_loop | c0 | rescale/preflight | 2 | 1.466 |
| rescale | coefficient_loop | c0 | rescale/preflight | 2 | 1.450 |
| rescale | coefficient_loop | c0 | rescale/preflight | 6 | 1.450 |
| rescale | coefficient_loop | c1 | rescale/preflight | 6 | 1.088 |
| rescale | coefficient_loop | c1 | rescale/preflight | 6 | 1.083 |
| rescale | coefficient_loop | c1 | rescale/preflight | 2 | 1.078 |
| rescale | coefficient_loop | c1 | rescale/preflight | 1 | 1.094 |
| rescale | coefficient_loop | c1 | rescale/preflight | 1 | 1.077 |
| rescale | coefficient_loop | c1 | rescale/preflight | 1 | 1.078 |
| rescale | coefficient_loop | c1 | rescale/preflight | 1 | 1.077 |
| rescale | coefficient_loop | c1 | rescale/preflight | 1 | 0.941 |
| rescale | coefficient_loop | c1 | rescale/preflight | 1 | 1.081 |
| rescale | coefficient_loop | c1 | rescale/preflight | 1 | 1.090 |
| rescale | coefficient_loop | c1 | rescale/preflight | 2 | 1.078 |
| rescale | coefficient_loop | c1 | rescale/preflight | 2 | 1.097 |
| rescale | coefficient_loop | c1 | rescale/preflight | 2 | 1.100 |
| rescale | coefficient_loop | c1 | rescale/preflight | 2 | 1.074 |
| rescale | coefficient_loop | c1 | rescale/preflight | 6 | 1.078 |
| rescale | materialization | — | rescale/rescale | 6 | 0.278 |
| rescale | materialization | — | rescale/rescale | 6 | 0.277 |
| rescale | materialization | — | rescale/rescale | 2 | 0.278 |
| rescale | materialization | — | rescale/rescale | 1 | 0.281 |
| rescale | materialization | — | rescale/rescale | 1 | 0.278 |
| rescale | materialization | — | rescale/rescale | 1 | 0.276 |
| rescale | materialization | — | rescale/rescale | 1 | 0.278 |
| rescale | materialization | — | rescale/rescale | 1 | 0.139 |
| rescale | materialization | — | rescale/rescale | 1 | 0.211 |
| rescale | materialization | — | rescale/rescale | 1 | 0.280 |
| rescale | materialization | — | rescale/rescale | 2 | 0.279 |
| rescale | materialization | — | rescale/rescale | 2 | 0.282 |
| rescale | materialization | — | rescale/rescale | 2 | 0.291 |
| rescale | materialization | — | rescale/rescale | 2 | 0.276 |
| rescale | materialization | — | rescale/rescale | 6 | 0.277 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 6 | 0.138 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 6 | 0.138 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 2 | 0.139 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 1 | 0.142 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 1 | 0.138 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 1 | 0.138 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 1 | 0.138 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 1 | 0.068 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 1 | 0.107 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 1 | 0.140 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 2 | 0.137 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 2 | 0.141 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 2 | 0.151 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 2 | 0.138 |
| rescale | ntt_montgomery_restore | c0 | rescale/materialization | 6 | 0.138 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 6 | 0.138 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 6 | 0.138 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 2 | 0.136 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 1 | 0.137 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 1 | 0.137 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 1 | 0.137 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 1 | 0.138 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 1 | 0.068 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 1 | 0.103 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 1 | 0.139 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 2 | 0.141 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 2 | 0.139 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 2 | 0.139 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 2 | 0.137 |
| rescale | ntt_montgomery_restore | c1 | rescale/materialization | 6 | 0.137 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 6 | 0.145 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 6 | 0.146 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 2 | 0.147 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 1 | 0.146 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 1 | 0.161 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 1 | 0.150 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 1 | 0.145 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 1 | 0.114 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 1 | 0.150 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 1 | 0.146 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 2 | 0.151 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 2 | 0.145 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 2 | 0.146 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 2 | 0.146 |
| rescale | prefix_to_coefficient | c0 | rescale/preflight | 6 | 0.146 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 6 | 0.145 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 6 | 0.144 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 2 | 0.151 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 1 | 0.145 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 1 | 0.145 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 1 | 0.147 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 1 | 0.147 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 1 | 0.108 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 1 | 0.145 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 1 | 0.144 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 2 | 0.145 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 2 | 0.144 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 2 | 0.144 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 2 | 0.144 |
| rescale | prefix_to_coefficient | c1 | rescale/preflight | 6 | 0.144 |
| rescale | preflight | — | rescale/rescale | 6 | 2.841 |
| rescale | preflight | — | rescale/rescale | 6 | 2.809 |
| rescale | preflight | — | rescale/rescale | 2 | 2.840 |
| rescale | preflight | — | rescale/rescale | 1 | 2.825 |
| rescale | preflight | — | rescale/rescale | 1 | 2.875 |
| rescale | preflight | — | rescale/rescale | 1 | 2.829 |
| rescale | preflight | — | rescale/rescale | 1 | 2.884 |
| rescale | preflight | — | rescale/rescale | 1 | 2.299 |
| rescale | preflight | — | rescale/rescale | 1 | 2.798 |
| rescale | preflight | — | rescale/rescale | 1 | 2.814 |
| rescale | preflight | — | rescale/rescale | 2 | 2.835 |
| rescale | preflight | — | rescale/rescale | 2 | 2.850 |
| rescale | preflight | — | rescale/rescale | 2 | 2.857 |
| rescale | preflight | — | rescale/rescale | 2 | 2.815 |
| rescale | preflight | — | rescale/rescale | 6 | 2.822 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 6 | 0.699 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 6 | 0.690 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 2 | 0.701 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 1 | 0.690 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 1 | 0.683 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 1 | 0.699 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 1 | 0.732 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 1 | 0.416 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 1 | 0.667 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 1 | 0.687 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 2 | 0.704 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 2 | 0.712 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 2 | 0.713 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 2 | 0.701 |
| rescale | reconstruct_center_round_capacity | c0 | rescale/coefficient_loop/c0 | 6 | 0.702 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 6 | 0.312 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 6 | 0.304 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 2 | 0.298 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 1 | 0.320 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 1 | 0.299 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 1 | 0.302 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 1 | 0.300 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 1 | 0.225 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 1 | 0.316 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 1 | 0.303 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 2 | 0.304 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 2 | 0.309 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 2 | 0.317 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 2 | 0.305 |
| rescale | reconstruct_center_round_capacity | c1 | rescale/coefficient_loop/c1 | 6 | 0.305 |
| rescale | rescale | — | — | 6 | 3.133 |
| rescale | rescale | — | — | 6 | 3.101 |
| rescale | rescale | — | — | 2 | 3.135 |
| rescale | rescale | — | — | 1 | 3.123 |
| rescale | rescale | — | — | 1 | 3.173 |
| rescale | rescale | — | — | 1 | 3.123 |
| rescale | rescale | — | — | 1 | 3.179 |
| rescale | rescale | — | — | 1 | 2.451 |
| rescale | rescale | — | — | 1 | 3.027 |
| rescale | rescale | — | — | 1 | 3.112 |
| rescale | rescale | — | — | 2 | 3.132 |
| rescale | rescale | — | — | 2 | 3.150 |
| rescale | rescale | — | — | 2 | 3.166 |
| rescale | rescale | — | — | 2 | 3.108 |
| rescale | rescale | — | — | 6 | 3.122 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 6 | 0.196 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 6 | 0.200 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 2 | 0.194 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 1 | 0.189 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 1 | 0.197 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 1 | 0.192 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 1 | 0.209 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 1 | 0.122 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 1 | 0.169 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 1 | 0.191 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 2 | 0.191 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 2 | 0.191 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 2 | 0.191 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 2 | 0.193 |
| rescale | residue_materialization | c0 | rescale/coefficient_loop/c0 | 6 | 0.192 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 6 | 0.236 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 6 | 0.235 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 2 | 0.238 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 1 | 0.226 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 1 | 0.238 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 1 | 0.238 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 1 | 0.238 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 1 | 0.166 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 1 | 0.175 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 1 | 0.240 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 2 | 0.243 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 2 | 0.236 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 2 | 0.234 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 2 | 0.237 |
| rescale | residue_materialization | c1 | rescale/coefficient_loop/c1 | 6 | 0.234 |
| stage | bootstrap | — | — | 1 | 122.193 |
| stage | coeffs_to_slots | real+imag | stage/bootstrap | 1 | 16.469 |
| stage | evalmod_imag | imag | stage/bootstrap | 1 | 45.890 |
| stage | evalmod_real | real | stage/bootstrap | 1 | 45.968 |
| stage | mod_up_trace | — | stage/bootstrap | 1 | 0.451 |
| stage | pack_n1_to_n2 | — | stage/bootstrap | 1 | 0.006 |
| stage | public_finalization | — | stage/bootstrap | 1 | 0.010 |
| stage | scale_down | — | stage/bootstrap | 1 | 0.016 |
| stage | slots_to_coeffs | real+imag | stage/bootstrap | 1 | 13.363 |
| stage | unpack_n2_to_n1 | — | stage/bootstrap | 1 | 0.010 |

These medians are descriptive event-type summaries only; they do not sum nested durations or form cross-root percentages. Consult the root-partitioned evidence for exact per-root Pareto and closure.

Bootstrap root closure: 122.193 ms inclusive, 122.183 ms direct children, signed residual +0.011 ms. Amdahl scenarios are conditional mathematical bounds on measured inclusive stage fractions, not speedup predictions.

## Profile evidence

- CPU: `9e551ee7bd6ac6bf8c32a56900eba7eafed57ef04cf08cd0f71840dc28e60aeb`, file available=true, raw-bound hash=false, disposition `file_present_but_original_raw_has_no_digest; excluded_from_attribution`.
- Heap: ``, file available=false, raw-bound hash=false, disposition `profile_bytes_unavailable`.

## Validation anomalies

None. The separate offline analysis passes the strict frozen-source, output-metadata, full event-tree and root-partition gates. The original raw artifact remains `TRACE_UNVERIFIED`.

## Interpretation boundary

This is a Fast-only offline analysis of the already captured 539-event trace. It does not establish a Standard timing comparison, a speedup, secure-CKKS equivalence, or a new Bootstrap result. Batch023's historical 4.225× uninstrumented observation remains separate and is not inferred from this trace.

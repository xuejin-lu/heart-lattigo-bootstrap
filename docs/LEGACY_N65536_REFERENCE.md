# Historical N=65536 Reference

The historical `legacy-hardware-model` test `TestSlimBootstrapperHWN65536` used:

```go
cfg := DefaultBootstrapConfig()
cfg.LogN = 16
```

It otherwise inherited these old default values:

| Field | Historical value |
| --- | --- |
| `LogDefaultScale` | `45` |
| `SecretHamming` | `192` |
| `Q0` | `[55]` |
| `QSlotsToCoeffs` | `[39,39,39]` |
| `QCircuitSlots` | `[45]` |
| `QEvalMod` | `[60,60,60,60,60,60,60,60]` |
| `QCoeffsToSlots` | `[56,56,56,56]` |
| `P` | `[61,61,61,61,61]` |
| `SlotsToCoeffsDFT` | `[1,1,1]` |
| `CoeffsToSlotsDFT` | `[1,1,1,1]` |
| `LogBSGSRatio` | `1` |
| `LogSlots` | `-1` |
| `Mod1LogScale` | `60` |
| `Mod1Degree` | `30` |
| `Mod1DoubleAngle` | `3` |
| `Mod1K` | `16` |
| `LogMessageRatio` | `10` |
| `Mod1InvDegree` | `0` |

The modern experiment profile intentionally preserves many high-level numerical
settings, but it is not construction-identical to the legacy builder. The
legacy path used a full residual/full Q chain and `DecodeThenModUp`; the current
Standard/Fast compatibility path uses a two-prime residual and
`ModUpThenEncode`, with the bootstrapping Q chain derived by the current public
Lattigo parameter builder.

The current `log_n=16`, `log_slots=-1` profile therefore exercises Standard
CKKS `N=65536` with `32768` logical slots, but it must not be described as an
exact reproduction of the old hardware model.

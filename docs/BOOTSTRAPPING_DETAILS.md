# Hardware CKKS Bootstrapping Details

This document describes the bootstrapping method implemented by the current
hardware-modeled code path. It is based on the code in `bootstrapping_hw.go`,
`bootstrap_config.go`, `bootstrap_usage_report.go`, `hardware_evaluator.go`,
`basic_operation.go`, and `functional_unit.go`.

## Important Scope Note

The local paper `docs/slim_bootstrapping.pdf` describes "slim mode" bootstrapping
for FV/BGV using homomorphic lower digit removal. This repository's active
bootstrapping code is a CKKS bootstrapping flow implemented through Lattigo's
CKKS bootstrapping parameters and a hardware-modeled evaluator. The useful
high-level analogy is that both methods move between packed slots and a
coefficient-like representation, perform a modular/digit cleanup, and move back.
The actual current code does not implement the FV/BGV lower-digit-removal
algorithm. It implements CKKS `EvalMod` with a cosine modular-reduction
polynomial.

## Current Bootstrapping Method

The current hardware bootstrapper uses this order:

```text
SlotsToCoeffs
-> ScaleDown
-> ModUp
-> CoeffsToSlots
-> EvalMod(real branch)
-> EvalMod(imag branch, if present)
-> Recombine
```

This is selected by `CircuitOrder: bootstrapping.DecodeThenModUp` in
`NewBootstrapParametersFromConfig`.

Current default parameter summary:

| Item | Current value |
| --- | --- |
| `LogN` | 13 |
| Ring degree `N` | 8192 |
| Max slots | 4096 |
| Secret hamming weight | 192 |
| Default CKKS scale | `2^45` |
| Q bit lengths | `[55, 39, 39, 39, 45, 60, 60, 60, 60, 60, 60, 60, 60, 56, 56, 56, 56]` |
| Total Q bit length | 921 bits |
| P bit lengths | `[61, 61, 61, 61, 61]` |
| Total P bit length | 305 bits |
| S2C DFT levels | `[1, 1, 1]` |
| C2S DFT levels | `[1, 1, 1, 1]` |
| `LogSlots` | `-1`, resolved to `LogMaxSlots = 12` |
| EvalMod type | `mod1.CosDiscrete` |
| EvalMod degree | 30 |
| EvalMod double-angle rounds | 3 |
| EvalMod `K` | 16 |
| EvalMod log message ratio | 10 |
| Sparse bootstrap keys | disabled, `EphemeralSecretWeight = 0` |

Although `BootstrapConfig.LogBSGSRatio` is set to 1 by default,
`ConfigureNaiveDFTMatrices` forces both S2C and C2S to `LogBSGSRatio = -1`.
Therefore the current hardware DFT path uses the direct diagonal method, not
the baby-step/giant-step linear transform.

## Data and Object Types

The main encrypted object is a degree-1 RLWE/CKKS ciphertext:

```text
ct = (c0, c1)
```

Each `c_i` is an RNS polynomial over the current Q level. The hardware model
often converts it into the local type:

```go
type Ciphertext struct {
    Value []ring.Poly
}
```

The bootstrapping code also uses these plaintext-like objects:

- DFT diagonal plaintexts: `ringqp.Poly` values stored in
  `dft.Matrix.Matrices[i].Vec[k]`.
- Scalar plaintext constants: complex constants such as `i`, `-i`, EvalMod
  offsets, and polynomial coefficients.
- Evaluation keys: rotation/Galois keys and the relinearization key.
- Modulus conversion constants: `BasisConvConstants`, precomputed from the
  active Q and P bases.

## Linear Transform Used by S2C and C2S

S2C and C2S are implemented as diagonal linear transforms. For a linear matrix
`M`, Lattigo decomposes it into plaintext diagonals:

```text
M(x) = sum_k diag_k(M) * Rot_k(x)
```

In the code:

- `matrix.Vec[k]` is the plaintext diagonal `diag_k(M)`.
- `Rot_k(x)` is implemented by Galois automorphism plus key switching.
- `k = 0` means no rotation, so the ciphertext is directly multiplied by the
  plaintext diagonal.
- `k != 0` means the ciphertext must be rotated first, then multiplied by that
  diagonal.

The hardware path for one `LinearTransform` is:

1. Sort all diagonal indexes `k` in `matrix.Vec`.
2. For each non-zero rotation:
   - Get the Galois key for `rot = k mod slots`.
   - Key-switch `c1` into QP accumulators with `KeySwitchQP`.
   - Add `P * c0` into the Q side of the first key-switch output.
   - Apply `AutomorphismGaloisUnit` to Q and P parts.
   - Multiply by the plaintext diagonal `matrix.Vec[k]`.
   - Accumulate the products in Q and P.
3. Mod-down the QP accumulation back to Q with `ModDownQPtoQNTTUnit`.
4. If `matrix.Vec[0]` exists, multiply the unrotated ciphertext by the zero
   diagonal and add it into the result.
5. Set the output scale to `input_scale * matrix.Scale`.

The plaintext matrix needed by the ciphertext is therefore not a dense matrix at
runtime. It is a set of encoded diagonal plaintext polynomials:

```text
S2C: btp.S2C.Matrices[j].Vec[k]
C2S: btp.C2S.Matrices[j].Vec[k]
```

Each plaintext diagonal is already encoded in the polynomial/RNS form expected
by Lattigo's linear-transform generator. The hardware model consumes those
plaintexts and performs the ciphertext operations with hardware units.

## Step 1: SlotsToCoeffs

Source function: `SlimBootstrapperHW.SlotsToCoeffs`.

Purpose:

- Convert the encrypted CKKS slot vector into the coefficient representation
  needed before modular reduction.
- This is the homomorphic CKKS decoding transform.
- In Lattigo terms, the matrix literal has `Type: dft.HomomorphicDecode`.

Input ciphertext:

- In the default test/report path, the input is one ciphertext `ctReal`.
- The optional `ctImag` argument is `nil` in the current full bootstrap test.
- If both real and imaginary ciphertexts are supplied, the code first computes
  `ctReal + i * ctImag`, then applies S2C.

Plaintexts and keys needed:

- S2C diagonal plaintexts from `btp.S2C.Matrices`.
- Rotation keys for all non-zero S2C diagonals.
- P-basis mod-down constants for key switching.

Operation process:

1. Select the input ciphertext. With only `ctReal`, use it directly.
2. Call `dft(ct, btp.S2C, "SlotsToCoeffs")`.
3. The S2C matrix has 3 levels in the current config: `[1, 1, 1]`.
4. For each level group:
   - Apply one diagonal `LinearTransform`.
   - Rescale after the group.
5. The result is a ciphertext whose encrypted message is represented in the
   coefficient domain used by the following scale-down and mod-up stages.

Hardware operations used:

- `LinearTransform`
- `KeySwitchQP`
- `AutomorphismGaloisUnit`
- `EvalMulPlaintext`
- `ModDownQPtoQNTTUnit`
- `RescaleCiphertext`

Current measured functional-unit usage:

```text
NTT forward=12, inverse=87, total=99
BaseConversion=6
Automorphism=300
EWU total=17326080
DoubleRNS add=0, mul=0
```

## Step 2: ScaleDown

Source function: `SlimBootstrapperHW.ScaleDown`.

Purpose:

- Adjust the coefficient-domain ciphertext scale/message ratio so that EvalMod
  receives values in the expected interval.
- Consume remaining levels until the ciphertext is at level 0.

Input ciphertext:

- Output of S2C.
- In the current config, S2C already consumes the level down to level 0, so
  this stage mainly applies an integer scalar multiply.

Plaintexts and constants needed:

- No matrix plaintext is needed.
- The stage computes an integer scalar `scaleUp` from:
  - current message ratio `Q_level / ct.Scale`
  - target message ratio `Mod1Parameters.MessageRatio()`

Operation process:

1. Compute:

   ```text
   currentMessageRatio = Q_current_level / ct.Scale
   targetMessageRatio = Mod1Parameters.MessageRatio()
   scaleUp = currentMessageRatio / targetMessageRatio
   ```

2. Convert `scaleUp` to an integer. The current hardware path requires this
   integer to fit in 63 bits.
3. Multiply every ciphertext polynomial coefficient by `scaleUp` using the
   element-wise MAD datapath.
4. If the ciphertext level is still above 0, repeatedly rescale until level 0.
5. Return the scaled ciphertext and the scale error.

Hardware operations used:

- `mulScalarPolyUnit`
- `ElementWiseUnit(OpMAD)`
- `Rescale`, if levels remain

Current measured functional-unit usage:

```text
NTT total=0
BaseConversion=0
Automorphism=0
EWU total=16384
DoubleRNS add=0, mul=0
```

## Step 3: ModUp

Source function: `SlimBootstrapperHW.ModUp`.

Purpose:

- Lift the ciphertext from the single base modulus `q0` back to the full Q
  chain.
- This gives the later C2S and EvalMod stages enough levels to run.

Input ciphertext:

- Level-0 ciphertext after S2C and ScaleDown.
- Degree-1 ciphertext `(c0, c1)`.
- The code only supports the non-sparse full-packing path:
  - `EphemeralSecretWeight = 0`
  - `CoeffsToSlotsParameters.LogSlots == params.LogMaxSlots()`

Plaintexts and constants needed:

- No plaintext matrix is needed.
- Uses the full Q modulus chain and centered reduction from `q0` to all other
  Q limbs.

Operation process:

1. Convert each ciphertext polynomial from NTT to coefficient domain at `q0`
   with `NTTUnit(..., isNTT=0)`.
2. Resize the ciphertext to the maximum Q level.
3. For every coefficient:
   - Read its residue modulo `q0`.
   - Interpret it as a centered signed representative:

     ```text
     coeff >= q0/2 means negative
     ```

   - Broadcast that signed integer into every higher Q limb by modular
     reduction.
4. Convert every polynomial back to NTT domain over the full Q basis with
   `NTTUnit(..., isNTT=1)`.
5. If the scaling factor computed from EvalMod parameters is greater than 1,
   multiply the ciphertext by that scalar.

Hardware operations used:

- `NTTUnit` inverse and forward
- coefficient-wise centered broadcast
- `mulScalarPolyUnit`, only if the final scalar correction is needed

Current measured functional-unit usage:

```text
NTT forward=2, inverse=2, total=4
BaseConversion=0
Automorphism=0
EWU total=278528
DoubleRNS add=0, mul=0
```

## Step 4: CoeffsToSlots

Source function: `SlimBootstrapperHW.CoeffsToSlots`.

Purpose:

- Convert the coefficient representation back into CKKS slots.
- This is the homomorphic CKKS encoding transform.
- In Lattigo terms, the matrix literal has:

  ```text
  Type   = dft.HomomorphicEncode
  Format = dft.RepackImagAsReal
  ```

Input ciphertext:

- Output of ModUp, now back at the maximum Q level.

Plaintexts and keys needed:

- C2S diagonal plaintexts from `btp.C2S.Matrices`.
- Rotation keys for all non-zero C2S diagonals.
- Complex conjugation Galois key.
- Scalar plaintext `-i` for the imaginary branch.

Operation process:

1. Apply `dft(ctIn, btp.C2S, "CoeffsToSlots")`.
2. The current C2S matrix has 4 level groups: `[1, 1, 1, 1]`.
3. Each group applies one diagonal `LinearTransform` followed by rescale.
4. The result is named `zV`.
5. Because the format is `RepackImagAsReal`, the code splits real and
   imaginary data:

   ```text
   conj = Conjugate(zV)
   ctImag = zV - conj
   ctImag = (-i) * ctImag
   ctReal = zV + conj
   ```

   The normalization factors are part of the generated Lattigo C2S matrix
   scaling and the exact RepackImagAsReal convention.
6. Return `ctReal` and `ctImag`.

What matrix is multiplied:

- Each C2S submatrix is stored as plaintext diagonals:

  ```text
  btp.C2S.Matrices[0].Vec[k]
  btp.C2S.Matrices[1].Vec[k]
  btp.C2S.Matrices[2].Vec[k]
  btp.C2S.Matrices[3].Vec[k]
  ```

- The ciphertext is multiplied by these diagonal plaintexts after the required
  rotations. There is no dense matrix multiply in the hardware path.

Hardware operations used:

- Same diagonal `LinearTransform` path as S2C
- `Conjugate`, implemented by Galois automorphism plus key switching
- `EvalSub` and `EvalAdd`
- scalar multiply by `-i` through double-RNS scalar hardware

Current measured functional-unit usage:

```text
NTT forward=18, inverse=68, total=86
BaseConversion=10
Automorphism=200
EWU total=53460992
DoubleRNS add=0, mul=2
```

## Step 5: EvalMod

Source function: `SlimBootstrapperHW.EvalMod`.

Purpose:

- Approximate modular reduction on encrypted CKKS slots.
- This removes the large modulus multiple introduced by the coefficient-domain
  bootstrapping representation and maps values back into the CKKS message
  interval.

Input ciphertexts:

- `ctReal` from C2S.
- `ctImag` from C2S, if full complex repacking is active.
- The current bootstrap evaluates EvalMod independently on real and imaginary
  branches.

Plaintexts and constants needed:

- EvalMod polynomial coefficients from `btp.Ref.Mod1Parameters.Mod1Poly`.
- Optional inverse polynomial coefficients if `Mod1InvPoly` is configured.
  The current default has `mod1_inv_degree = 0`, so no inverse polynomial is
  used.
- Scalar offset for cosine EvalMod.
- Scalar constants derived from `sqrt(2*pi)` and double-angle reconstruction.
- Relinearization key for ciphertext squaring.

Operation process:

1. Check that the ciphertext has at least `Mod1Parameters.LevelQ`.
2. If it has more levels, copy it down to exactly the EvalMod level.
3. Set the working scale to `Mod1Parameters.ScalingFactor()`.
4. Compute the target polynomial evaluation scale by walking the primes that
   will be consumed by the polynomial depth and double-angle rounds.
5. For `mod1.CosDiscrete` and `mod1.CosContinuous`, add the EvalMod offset:

   ```text
   offset = -0.5 / ((B - A) * IntervalShrinkFactor)
   ```

6. Adjust polynomial coefficients when no inverse polynomial is used.
7. Evaluate the EvalMod polynomial with Lattigo's polynomial evaluator
   interface, but backed by this repository's `hardwareCKKSEvaluator`.
   Therefore ciphertext additions, multiplications, relinearization, scalar
   operations, and rescale are routed through hardware-modeled operations.
8. Run `DoubleAngle` rounds. Each round performs:

   ```text
   ct = ct * ct
   ct = ct + ct
   ct = ct - sqrt2pi_constant
   ct = Rescale(ct)
   ```

9. If an inverse polynomial exists, evaluate it. The default config skips this.
10. Reset the output scale to the CKKS default scale.

Hardware operations used:

- scalar add through `addDoubleRNSScalarUnit`
- polynomial evaluation through `hardwareCKKSEvaluator`
- `EvalTensor`
- `Relinearize`
- `KeySwitchQP`
- `ModDownQPtoQNTTUnit`
- `RescaleCiphertext`
- scalar constants through double-RNS add/mul units

Current measured usage for one EvalMod branch:

```text
NTT forward=52, inverse=64, total=116
BaseConversion=24
Automorphism=0
EWU total=12304384
DoubleRNS add=14, mul=27
```

Because the current C2S returns both real and imaginary branches, this cost is
paid twice.

## Step 6: Recombine

Source location: final block of `SlimBootstrapperHW.Bootstrap`.

Purpose:

- Recombine the independently reduced real and imaginary branches into one
  complex CKKS ciphertext.

Input ciphertexts:

- `ctReal = EvalMod(real branch)`
- `ctImag = EvalMod(imag branch)`

Plaintexts and constants needed:

- Scalar plaintext `i`.

Operation process:

1. Multiply the imaginary branch by `i`.
2. Add it to the real branch:

   ```text
   ctOut = ctReal + i * ctImag
   ```

3. Set output scale to the residual/default CKKS scale.

Hardware operations used:

- double-RNS scalar multiply for `i`
- ciphertext add

Current measured functional-unit usage:

```text
NTT total=0
BaseConversion=0
Automorphism=0
EWU total=81920
DoubleRNS add=0, mul=2
```

## Functional Unit Mapping

The current bootstrapping implementation routes core arithmetic through these
hardware-modeled units:

| Unit | Main source | Role in bootstrap |
| --- | --- | --- |
| `NTTUnit` | `functional_unit.go` | Forward/inverse NTT for key switching, mod-down, and ModUp |
| `BaseConversionUnit` | `functional_unit.go` | CRT base conversion in Q/P key switching and mod-down |
| `AutomorphismGaloisUnit` | `functional_unit.go` | Slot rotations and conjugation |
| `ElementWiseUnit(OpTensor)` | `functional_unit.go` | ciphertext-ciphertext tensor product |
| `ElementWiseUnit(OpAccQ)` | `functional_unit.go` | Q-side key-switch accumulation |
| `ElementWiseUnit(OpAccP)` | `functional_unit.go` | P-side key-switch and plaintext-diagonal accumulation |
| `ElementWiseUnit(OpModD)` | `functional_unit.go` | QP-to-Q mod-down subtract step |
| `ElementWiseUnit(OpMAD)` | `functional_unit.go` | scalar multiply, scalar multiply/add, relinearization combine |
| `addDoubleRNSScalarUnit` | `hardware_evaluator.go` | add complex scalar plaintext to CKKS ciphertext |
| `mulDoubleRNSScalarUnit` | `hardware_evaluator.go` | multiply CKKS ciphertext by complex scalar plaintext |

Some setup objects still come from Lattigo:

- DFT diagonal plaintext generation with `dft.NewMatrixFromLiteral`.
- EvalMod polynomial coefficient construction inside Lattigo
  `Mod1Parameters`.
- Polynomial evaluator scheduling.

The arithmetic executed on ciphertexts by that schedule is routed through
`hardwareCKKSEvaluator`.

## Full Measured Functional Unit Usage

Generated by:

```powershell
go run . -mode=bootstrap-usage -config .\configs\bootstrap_config.example.json -out .\reports\bootstrap_usage_report.csv -format csv -by-step=true
```

| Stage | NTT fwd | NTT inv | NTT total | BaseConv | Auto | EWU total | DoubleRNS add | DoubleRNS mul |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| SlotsToCoeffs | 12 | 87 | 99 | 6 | 300 | 17326080 | 0 | 0 |
| ScaleDown | 0 | 0 | 0 | 0 | 0 | 16384 | 0 | 0 |
| ModUp | 2 | 2 | 4 | 0 | 0 | 278528 | 0 | 0 |
| CoeffsToSlots | 18 | 68 | 86 | 10 | 200 | 53460992 | 0 | 2 |
| EvalMod(real) | 52 | 64 | 116 | 24 | 0 | 12304384 | 14 | 27 |
| EvalMod(imag) | 52 | 64 | 116 | 24 | 0 | 12304384 | 14 | 27 |
| Recombine | 0 | 0 | 0 | 0 | 0 | 81920 | 0 | 2 |
| Total | 136 | 285 | 421 | 64 | 500 | 95772672 | 28 | 58 |

## Source Map

| Concept | Code location |
| --- | --- |
| Bootstrap order | `bootstrapping_hw.go`, `SlimBootstrapperHW.Bootstrap` |
| S2C/C2S DFT execution | `bootstrapping_hw.go`, `dft` and `LinearTransform` |
| S2C/C2S matrix generation | `bootstrap_usage_report.go`, `ConfigureNaiveDFTMatrices` |
| Parameter construction | `bootstrap_config.go`, `NewBootstrapParametersFromConfig` |
| ScaleDown | `bootstrapping_hw.go`, `ScaleDown` |
| ModUp | `bootstrapping_hw.go`, `ModUp` |
| EvalMod | `bootstrapping_hw.go`, `EvalMod` |
| Hardware CKKS add/mul/rescale interface | `hardware_evaluator.go` |
| Functional units | `functional_unit.go` |
| Basic ciphertext operations | `basic_operation.go` |
| Usage counters | `functional_unit_usage.go` |

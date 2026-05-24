package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

// TestEnv bundles the cryptographic context to pass into modular test functions cleanly.
type TestEnv struct {
	params        ckks.Parameters
	encryptor     *rlwe.Encryptor
	decryptor     *rlwe.Decryptor
	encoder       *ckks.Encoder
	ckkEval       *ckks.Evaluator
	ringQ         *ring.Ring
	ringP         *ring.Ring
	nQ            int
	N             int
	evkQ          [][2]ring.Poly
	evkP          [][2]ring.Poly
	rotKeyQ       [][2]ring.Poly
	rotKeyP       [][2]ring.Poly
	bconv         BasisConvConstants
	slotsToRotate int
}

func main() {
	mode := flag.String("mode", "ops", "run mode: ops or bootstrap-usage")
	configPath := flag.String("config", "", "bootstrap config JSON path; empty uses DefaultBootstrapConfig")
	outputPath := flag.String("out", "", "optional usage report output path (.json or .csv)")
	format := flag.String("format", "", "usage report format: json or csv; empty infers from -out")
	byStep := flag.Bool("by-step", true, "split bootstrap usage by SlotsToCoeffs, EvalMod, Recombine, and other stages")
	verbose := flag.Bool("verbose", false, "print verbose bootstrapping trace")
	flag.Parse()

	if *mode == "bootstrap-usage" {
		report, err := RunBootstrapUsageReport(BootstrapUsageReportOptions{
			ConfigPath: *configPath,
			OutputPath: *outputPath,
			Format:     *format,
			ByStep:     *byStep,
			Verbose:    *verbose,
		})
		if err != nil {
			log.Fatal(err)
		}
		PrintBootstrapUsageReport(report)
		if err := WriteBootstrapUsageReport(report, *outputPath, *format); err != nil {
			log.Fatal(err)
		}
		if *outputPath != "" {
			fmt.Printf("Wrote usage report: %s\n", *outputPath)
		}
		return
	}

	fmt.Println("Initializing CKKS parameters and cryptographic tools...")
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN:            16,
		LogQ:            []int{36, 36, 36, 36, 36, 36},
		LogP:            []int{36},
		LogDefaultScale: 36,
	})
	if err != nil {
		panic(err)
	}

	kgen := rlwe.NewKeyGenerator(params)
	sk := kgen.GenSecretKeyNew()
	pk := kgen.GenPublicKeyNew(sk)
	rlk := kgen.GenRelinearizationKeyNew(sk)

	slotsToRotate := 1
	galEl := params.GaloisElement(slotsToRotate)
	galKey := kgen.GenGaloisKeyNew(galEl, sk)

	encryptor := rlwe.NewEncryptor(params, pk)
	decryptor := rlwe.NewDecryptor(params, sk)
	encoder := ckks.NewEncoder(params)

	evkSet := rlwe.NewMemEvaluationKeySet(rlk, galKey)
	ckkEval := ckks.NewEvaluator(params, evkSet)

	ringQ := params.RingQ()
	ringP := params.RingP()

	nQ := len(ringQ.ModuliChain())
	N := ringQ.N()

	Q := ringQ.ModuliChain()
	P := ringP.ModuliChain()
	bconv := PrecomputeBasisConvConstants(Q, P)
	fmt.Println("Precomputed BasisConvConstants (hardware ROM values).")

	// Extract Relinearization Key
	rlkGadget := rlk.GadgetCiphertext
	nDigits := len(rlkGadget.Value)
	evkQ := make([][2]ring.Poly, nDigits)
	evkP := make([][2]ring.Poly, nDigits)
	for d := 0; d < nDigits; d++ {
		digitVec := rlkGadget.Value[d][0]
		evkQ[d][0] = digitVec[0].Q
		evkQ[d][1] = digitVec[1].Q
		evkP[d][0] = digitVec[0].P
		evkP[d][1] = digitVec[1].P
	}

	// Extract Galois Key (Rotation Key)
	gkGadget := galKey.GadgetCiphertext
	rotKeyQ := make([][2]ring.Poly, nDigits)
	rotKeyP := make([][2]ring.Poly, nDigits)
	for d := 0; d < nDigits; d++ {
		digitVec := gkGadget.Value[d][0]
		rotKeyQ[d][0] = digitVec[0].Q
		rotKeyQ[d][1] = digitVec[1].Q
		rotKeyP[d][0] = digitVec[0].P
		rotKeyP[d][1] = digitVec[1].P
	}

	fmt.Printf("  nQ=%d, nP=%d, nDigits=%d, N=%d\n", nQ, len(P), nDigits, N)

	// Package the environment to pass to test functions
	env := TestEnv{
		params:        params,
		encryptor:     encryptor,
		decryptor:     decryptor,
		encoder:       encoder,
		ckkEval:       ckkEval,
		ringQ:         ringQ,
		ringP:         ringP,
		nQ:            nQ,
		N:             N,
		evkQ:          evkQ,
		evkP:          evkP,
		rotKeyQ:       rotKeyQ,
		rotKeyP:       rotKeyP,
		bconv:         bconv,
		slotsToRotate: slotsToRotate,
	}

	TestHMul(env)
	TestRotate(env)
	TestRescale(env)
}

// deepCopyPoly is pulled out to package level to be used by multiple test functions
func deepCopyPoly(ringQ *ring.Ring, nQ int, src ring.Poly) ring.Poly {
	dst := newPoly(ringQ, nQ)
	for limb := 0; limb < nQ; limb++ {
		copy(dst.Coeffs[limb], src.Coeffs[limb])
	}
	return dst
}

func TestHMul(env TestEnv) {
	fmt.Println("\n=============================================")
	fmt.Println("Generating test plaintexts and encrypting...")
	ptValues1 := make([]float64, env.params.MaxSlots())
	ptValues2 := make([]float64, env.params.MaxSlots())
	ptValues1[0] = 5
	ptValues2[0] = 7

	pt1 := ckks.NewPlaintext(env.params, env.params.MaxLevel())
	if err := env.encoder.Encode(ptValues1, pt1); err != nil {
		panic(err)
	}
	pt2 := ckks.NewPlaintext(env.params, env.params.MaxLevel())
	if err := env.encoder.Encode(ptValues2, pt2); err != nil {
		panic(err)
	}

	ct1Lattigo, _ := env.encryptor.EncryptNew(pt1)
	ct2Lattigo, _ := env.encryptor.EncryptNew(pt2)

	ct1HW := Ciphertext{Value: []ring.Poly{deepCopyPoly(env.ringQ, env.nQ, ct1Lattigo.Value[0]), deepCopyPoly(env.ringQ, env.nQ, ct1Lattigo.Value[1])}}
	ct2HW := Ciphertext{Value: []ring.Poly{deepCopyPoly(env.ringQ, env.nQ, ct2Lattigo.Value[0]), deepCopyPoly(env.ringQ, env.nQ, ct2Lattigo.Value[1])}}

	fmt.Println("Executing Hardware Golden Model: EvalMult (with Relinearize)...")
	ctOutHW := EvalMult(env.ringQ, env.ringP, ct1HW, ct2HW, env.evkQ, env.evkP, env.bconv)

	fmt.Println("Executing Lattigo Reference Model (MulRelin)...")
	ctOutLattigo := rlwe.NewCiphertext(env.params, 1, env.params.MaxLevel())
	if err := env.ckkEval.MulRelin(ct1Lattigo, ct2Lattigo, ctOutLattigo); err != nil {
		panic(err)
	}

	match := VerifyDatapath(ctOutHW, ctOutLattigo, env.N)

	fmt.Println("\n--- RTL Datapath Verification Report ---")
	if match {
		fmt.Println("[PASS] Custom EvalMult matches Lattigo MulRelin.")
		fmt.Printf("Example Coefficient c0[0]: HW=0x%X | REF=0x%X\n",
			ctOutHW.Value[0].Coeffs[0][0], ctOutLattigo.Value[0].Coeffs[0][0])
	} else {
		fmt.Println("[FAIL] Coefficient mismatch detected.")
	}

	fmt.Println("\n--- Decryption Phase ---")

	ctOutHWLattigo := ctOutLattigo
	ctOutHWLattigo.Value[0] = ctOutHW.Value[0]
	ctOutHWLattigo.Value[1] = ctOutHW.Value[1]
	ptOutHW := env.decryptor.DecryptNew(ctOutHWLattigo)
	resValuesHW := make([]float64, env.params.MaxSlots())
	env.encoder.Decode(ptOutHW, resValuesHW)
	fmt.Printf("Decrypted Output Slot [0] (HW model):    %f (expected ≈ 35.0)\n", resValuesHW[0])

}

func TestRotate(env TestEnv) {
	fmt.Println("\n=============================================")
	fmt.Println("Generating test plaintexts and encrypting...")
	ptValues1 := make([]float64, env.params.MaxSlots())
	ptValues1[0] = 5.0
	ptValues1[1] = 10.0
	ptValues1[2] = 15.0

	pt1 := ckks.NewPlaintext(env.params, env.params.MaxLevel())
	env.encoder.Encode(ptValues1, pt1)
	ct1Lattigo, _ := env.encryptor.EncryptNew(pt1)

	ct1HW := Ciphertext{Value: []ring.Poly{
		deepCopyPoly(env.ringQ, env.nQ, ct1Lattigo.Value[0]),
		deepCopyPoly(env.ringQ, env.nQ, ct1Lattigo.Value[1]),
	}}

	fmt.Println("\n--- Testing EvalRotate Phase ---")

	fmt.Println("Executing Hardware Golden Model: EvalRotate...")
	ctRotHW := EvalRotate(env.ringQ, env.ringP, ct1HW, env.slotsToRotate, env.rotKeyQ, env.rotKeyP, env.bconv)

	fmt.Println("Executing Lattigo Reference Model: Rotate...")
	ctRotLattigo := rlwe.NewCiphertext(env.params, 1, env.params.MaxLevel())
	if err := env.ckkEval.Rotate(ct1Lattigo, env.slotsToRotate, ctRotLattigo); err != nil {
		panic(err)
	}

	matchRot := VerifyDatapath(ctRotHW, ctRotLattigo, env.N)
	if matchRot {
		fmt.Println("  [PASS] Custom EvalRotate exactly matches Lattigo Rotate.")
	} else {
		fmt.Println("  [FAIL] Coefficient mismatch detected in EvalRotate.")
	}

	fmt.Println("\n--- Rotation Decryption Phase ---")

	ctRotHWLattigo := ctRotLattigo
	ctRotHWLattigo.Value[0] = ctRotHW.Value[0]
	ctRotHWLattigo.Value[1] = ctRotHW.Value[1]
	ctRotHWLattigo.Scale = ctRotLattigo.Scale

	ptRotHW := env.decryptor.DecryptNew(ctRotLattigo)
	resRotHW := make([]float64, env.params.MaxSlots())
	env.encoder.Decode(ptRotHW, resRotHW)

	fmt.Printf("Decrypted HW Output Slot [0] (Expected ~10.0): %f\n", resRotHW[0])
	fmt.Printf("Decrypted HW Output Slot [1] (Expected ~15.0): %f\n", resRotHW[1])
	fmt.Printf("Decrypted HW Output Slot [N/2-1] (Expected ~5.0):  %f\n", resRotHW[env.params.MaxSlots()-1])
}

func TestRescale(env TestEnv) {
	fmt.Println("\n=============================================")
	fmt.Println("--- Testing Rescale Phase ---")

	// 1. Setup: Encrypt and Multiply to get a scaled ciphertext (Delta^2)
	ptValues := make([]float64, env.params.MaxSlots())
	ptValues[0] = 5.0
	pt1 := ckks.NewPlaintext(env.params, env.params.MaxLevel())
	env.encoder.Encode(ptValues, pt1)

	ct1Lattigo, _ := env.encryptor.EncryptNew(pt1)
	ct2Lattigo, _ := env.encryptor.EncryptNew(pt1)

	// Lattigo: Multiply and Relinearize
	ctMultLattigo := rlwe.NewCiphertext(env.params, 2, env.params.MaxLevel())
	env.ckkEval.Mul(ct1Lattigo, ct2Lattigo, ctMultLattigo)

	ctRelLattigo := rlwe.NewCiphertext(env.params, 1, env.params.MaxLevel())
	env.ckkEval.Relinearize(ctMultLattigo, ctRelLattigo)

	// Hardware Model copy of the Relinearized Ciphertext
	ctRelHW := Ciphertext{Value: []ring.Poly{
		deepCopyPoly(env.ringQ, env.nQ, ctRelLattigo.Value[0]),
		deepCopyPoly(env.ringQ, env.nQ, ctRelLattigo.Value[1]),
	}}

	// 2. Hardware Model Rescale
	fmt.Println("Executing Hardware Golden Model: Rescale...")
	ctResHW := Rescale(env.ringQ, ctRelHW)

	// 3. Lattigo Reference Rescale
	fmt.Println("Executing Lattigo Reference Model: Rescale...")
	ctResLattigo := rlwe.NewCiphertext(env.params, 1, env.params.MaxLevel()-1)
	env.ckkEval.Rescale(ctRelLattigo, ctResLattigo)

	// 4. Comparison
	matchRes := VerifyDatapath(ctResHW, ctResLattigo, env.N)
	if matchRes {
		fmt.Println("  [PASS] Custom Rescale exactly matches Lattigo Rescale.")
	} else {
		fmt.Println("  [FAIL] Coefficient mismatch detected in Rescale.")
	}

	// 5. Decryption (Value should be back to standard scale, ~25.0)
	fmt.Println("\n--- Rescale Decryption Phase ---")
	ctResHWLattigo := ctResLattigo
	ctResHWLattigo.Value[0] = ctResHW.Value[0]
	ctResHWLattigo.Value[1] = ctResHW.Value[1]

	ptResHW := env.decryptor.DecryptNew(ctResHWLattigo)
	resResHW := make([]float64, env.params.MaxSlots())
	env.encoder.Decode(ptResHW, resResHW)

	fmt.Printf("Decrypted HW Output Slot [0] (Expected ~25.0): %f\n", resResHW[0])
}

func VerifyDatapath(hwCt Ciphertext, refCt *rlwe.Ciphertext, N int) bool {
	if len(hwCt.Value) != len(refCt.Value) {
		fmt.Printf("Degree mismatch: HW has %d polys, REF has %d polys\n",
			len(hwCt.Value), len(refCt.Value))
		return false
	}
	errorCount := 0
	for degree := 0; degree < len(refCt.Value); degree++ {
		numLimbs := len(refCt.Value[degree].Coeffs)
		for limb := 0; limb < numLimbs; limb++ {
			for i := 0; i < N; i++ {
				valHW := hwCt.Value[degree].Coeffs[limb][i]
				valRef := refCt.Value[degree].Coeffs[limb][i]
				if valHW != valRef {
					if errorCount < 10 {
						fmt.Printf("Mismatch at degree=%d limb=%d index=%d: HW=0x%X REF=0x%X\n",
							degree, limb, i, valHW, valRef)
					}
					errorCount++
				}
			}
		}
	}
	if errorCount > 0 {
		fmt.Printf("Total mismatches: %d\n", errorCount)
		return false
	}
	return true
}

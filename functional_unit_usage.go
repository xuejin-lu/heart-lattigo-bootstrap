package main

import "fmt"

type FunctionalUnitUsage struct {
	NTTForward       uint64
	NTTInverse       uint64
	BaseConversion   uint64
	Automorphism     uint64
	DoublePrimeScale uint64
	ElementWise      [5]uint64
	DoubleRNSAdd     uint64
	DoubleRNSMul     uint64
}

var activeFunctionalUnitUsage *FunctionalUnitUsage

func ResetFunctionalUnitUsage() {
	activeFunctionalUnitUsage = &FunctionalUnitUsage{}
}

func CurrentFunctionalUnitUsage() FunctionalUnitUsage {
	if activeFunctionalUnitUsage == nil {
		return FunctionalUnitUsage{}
	}
	return *activeFunctionalUnitUsage
}

func ClearFunctionalUnitUsage() {
	activeFunctionalUnitUsage = nil
}

func countNTTUnit(isNTT int) {
	if activeFunctionalUnitUsage == nil {
		return
	}
	if isNTT == 1 {
		activeFunctionalUnitUsage.NTTForward++
	} else {
		activeFunctionalUnitUsage.NTTInverse++
	}
}

func countBaseConversionUnit() {
	if activeFunctionalUnitUsage != nil {
		activeFunctionalUnitUsage.BaseConversion++
	}
}

func countAutomorphismUnit() {
	if activeFunctionalUnitUsage != nil {
		activeFunctionalUnitUsage.Automorphism++
	}
}

func countElementWiseUnit(opCode int) {
	if activeFunctionalUnitUsage != nil && opCode >= 0 && opCode < len(activeFunctionalUnitUsage.ElementWise) {
		activeFunctionalUnitUsage.ElementWise[opCode]++
	}
}

func countDoublePrimeScalingUnit() {
	if activeFunctionalUnitUsage != nil {
		activeFunctionalUnitUsage.DoublePrimeScale++
	}
}

func countDoubleRNSAddUnit() {
	if activeFunctionalUnitUsage != nil {
		activeFunctionalUnitUsage.DoubleRNSAdd++
	}
}

func countDoubleRNSMulUnit() {
	if activeFunctionalUnitUsage != nil {
		activeFunctionalUnitUsage.DoubleRNSMul++
	}
}

func (u FunctionalUnitUsage) String() string {
	return fmt.Sprintf(
		"NTT forward=%d inverse=%d total=%d; BaseConversion=%d; Automorphism=%d; DoublePrimeScaling=%d; EWU Tensor=%d AccQ=%d AccP=%d ModD=%d MAD=%d total=%d; DoubleRNS add=%d mul=%d",
		u.NTTForward,
		u.NTTInverse,
		u.NTTForward+u.NTTInverse,
		u.BaseConversion,
		u.Automorphism,
		u.DoublePrimeScale,
		u.ElementWise[OpTensor],
		u.ElementWise[OpAccQ],
		u.ElementWise[OpAccP],
		u.ElementWise[OpModD],
		u.ElementWise[OpMAD],
		u.ElementWise[OpTensor]+u.ElementWise[OpAccQ]+u.ElementWise[OpAccP]+u.ElementWise[OpModD]+u.ElementWise[OpMAD],
		u.DoubleRNSAdd,
		u.DoubleRNSMul,
	)
}

func (u FunctionalUnitUsage) ElementWiseTotal() uint64 {
	return u.ElementWise[OpTensor] +
		u.ElementWise[OpAccQ] +
		u.ElementWise[OpAccP] +
		u.ElementWise[OpModD] +
		u.ElementWise[OpMAD]
}

func (u FunctionalUnitUsage) NTTTotal() uint64 {
	return u.NTTForward + u.NTTInverse
}

func (u FunctionalUnitUsage) Add(v FunctionalUnitUsage) FunctionalUnitUsage {
	u.NTTForward += v.NTTForward
	u.NTTInverse += v.NTTInverse
	u.BaseConversion += v.BaseConversion
	u.Automorphism += v.Automorphism
	u.DoublePrimeScale += v.DoublePrimeScale
	for i := range u.ElementWise {
		u.ElementWise[i] += v.ElementWise[i]
	}
	u.DoubleRNSAdd += v.DoubleRNSAdd
	u.DoubleRNSMul += v.DoubleRNSMul
	return u
}

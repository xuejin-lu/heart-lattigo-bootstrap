package main

import (
	"math/big"
	"testing"
)

func TestFIX001P3MulAliasProductBound(t *testing.T) {
	data := targetScaleCenteredData{
		Q01Half: big.NewInt(20),
		Values:  [][]*big.Int{{big.NewInt(-3), big.NewInt(2)}, {big.NewInt(0)}},
	}
	evidence := mulAliasBound(data, 2)
	if evidence.B != "3" || evidence.ProductBound != "18" || !evidence.UniquenessPass {
		t.Fatalf("bound evidence = %+v, want B=3, product bound=18, pass", evidence)
	}
}

func TestFIX001P3MulAliasContractRequiresBothZeroViews(t *testing.T) {
	contract := MulAliasContractEvidence{MaintainedC1Zero: true, C1: MulAliasZeroCheck{CoefficientDomainExactlyZero: true, NTTMontgomeryExactlyZero: true}}
	if !contract.MaintainedC1Zero || !contract.C1.CoefficientDomainExactlyZero || !contract.C1.NTTMontgomeryExactlyZero {
		t.Fatalf("zero-secret contract evidence unexpectedly failed: %+v", contract)
	}
}

package main

import (
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func TestQPREFIXAudit002CenteredQ012CRT(t *testing.T) {
	q := []uint64{17, 19, 23}
	values := []*big.Int{big.NewInt(123), big.NewInt(-321)}
	rows := make([][]uint64, 3)
	for limb, modulus := range q {
		rows[limb] = make([]uint64, len(values))
		for i, value := range values {
			rows[limb][i] = new(big.Int).Mod(new(big.Int).Set(value), new(big.Int).SetUint64(modulus)).Uint64()
		}
	}
	maximum, err := qprefixAudit002MaxCenteredQ012(rows, q)
	require.NoError(t, err)
	require.Equal(t, "321", maximum.String())
}

func TestQPREFIXAudit002RescaleRecurrenceRounding(t *testing.T) {
	raw := qprefixAudit002Checkpoint{Name: "raw", Level: 4, Components: []qprefixAudit002Component{
		{MaxAbsCoefficient: "1000"}, {MaxAbsCoefficient: "501"},
	}}
	post := qprefixAudit002Checkpoint{Name: "post", Level: 3, Components: []qprefixAudit002Component{
		{MaxAbsCoefficient: "143", ProvenMaxAbs: "143"}, {MaxAbsCoefficient: "72", ProvenMaxAbs: "72"},
	}}
	record := qprefixAudit002BuildRescale(0, raw, post, 7, rlwe.NewScale(1050), rlwe.NewScale(150), []*big.Int{big.NewInt(143), big.NewInt(72)})
	require.True(t, record.Pass)
	require.Equal(t, []string{"143", "72"}, record.PredictedMaxAbs)
	require.True(t, record.ScaleTransitionExact)
}

func TestQPREFIXAudit002C2SRawRescale(t *testing.T) {
	if os.Getenv("QPREFIX_AUDIT_002_INTEGRATION") != "1" {
		t.Skip("set QPREFIX_AUDIT_002_INTEGRATION=1 to run the current LogN13 C2S audit")
	}
	cfg, err := LoadBootstrapConfig("configs/bootstrap_config.logN13.json")
	require.NoError(t, err)
	// This is the accepted public-control profile: it differs from the checked-in
	// LogN13 config only by residual q0=56 (rather than q0=55).
	cfg.Q0 = []int{56}
	result, err := RunQPREFIXAudit002(cfg, ".", "../lattigo")
	require.NoError(t, err)
	require.Equal(t, "QPREFIX_C2S_CAPACITY_PROVEN", result.Classification, result.StopReason)
	require.True(t, result.InitialInputProof.Pass)
	require.Len(t, result.LinearTransforms, 4)
	require.Len(t, result.Rescales, 4)
	require.True(t, result.HelperEquivalence.Pass)
	for _, proof := range result.LinearTransforms {
		require.True(t, proof.Pass, "group %d factor %d", proof.Group, proof.FactorIndex)
	}
	for _, rescale := range result.Rescales {
		require.True(t, rescale.Pass, "group %d", rescale.Group)
	}
	for _, restore := range result.Restores {
		require.True(t, restore.Pass, "group %d", restore.Group)
	}
}

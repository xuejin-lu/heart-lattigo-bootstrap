package perfmeasure

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

// CiphertextState is the stable public-boundary metadata record shared by
// perfprobe and fastdiag. Keep field order and JSON names stable because its
// serialized form is fingerprinted in historical experiment artifacts.
type CiphertextState struct {
	Level       int     `json:"level"`
	Degree      int     `json:"degree"`
	Scale       string  `json:"scale"`
	ScaleLog2   float64 `json:"scale_log2"`
	LogRows     int     `json:"log_rows"`
	LogCols     int     `json:"log_cols"`
	QPrefixRows int     `json:"q_prefix_rows"`
	PrefixQ     string  `json:"prefix_q"`
}

// CiphertextStateOf describes the ciphertext at the supplied authoritative
// Q-prefix width. It does not inspect physical backing rows.
func CiphertextStateOf(ct *rlwe.Ciphertext, q []uint64, prefixRows int) CiphertextState {
	if ct == nil {
		return CiphertextState{}
	}
	state := CiphertextState{
		Level: ct.Level(), Degree: ct.Degree(), Scale: ct.Scale.Value.Text('e', 80),
		ScaleLog2: ct.Scale.Log2(), LogRows: ct.LogDimensions.Rows, LogCols: ct.LogDimensions.Cols,
	}
	if q != nil && ct.Level() >= 0 {
		rows := min(prefixRows, len(q))
		if rows > 0 {
			product := big.NewInt(1)
			for _, prime := range q[:rows] {
				product.Mul(product, new(big.Int).SetUint64(prime))
			}
			state.QPrefixRows, state.PrefixQ = rows, product.String()
		}
	}
	return state
}

// CiphertextStateFingerprint hashes the stable JSON representation used by
// existing perfprobe input metadata evidence.
func CiphertextStateFingerprint(state CiphertextState) string {
	data, err := json.Marshal(state)
	if err != nil {
		// CiphertextState contains only JSON-safe scalar fields.
		panic(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

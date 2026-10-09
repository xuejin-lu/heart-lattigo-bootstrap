package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"os"
	"runtime"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

const (
	logN           = 13
	logSlots       = 4
	scaleLog2      = 45
	standardCommit = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
	fastCommit     = "2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac"
	runSchema      = "fast-dropin-above-cap-batch-016-run.v1"
	pairSchema     = "fast-dropin-above-cap-batch-016-evidence.v1"
	runPassed      = "level5_public_chain_passed"
	gateTolerance  = 1e-6
)

type profile struct {
	LogN      int   `json:"log_n"`
	LogQ      []int `json:"log_q_bits"`
	LogP      []int `json:"log_p_bits"`
	ScaleLog2 int   `json:"scale_log2"`
	LogSlots  int   `json:"log_slots"`
	Level     int   `json:"logical_level"`
}
type metric struct {
	Finite      bool    `json:"finite"`
	ComplexRMSE float64 `json:"complex_rmse"`
	MaxComplex  float64 `json:"max_complex_error"`
}
type complexValue struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}
type keyLookups struct {
	Relinearization int `json:"relinearization"`
	Galois          int `json:"galois"`
	GaloisList      int `json:"galois_key_list"`
}
type outputState struct {
	Level                    int     `json:"level"`
	Degree                   int     `json:"degree"`
	ScaleLog2                float64 `json:"scale_log2"`
	IsNTT                    bool    `json:"is_ntt"`
	IsMontgomery             bool    `json:"is_montgomery"`
	C1ZeroOnAuthoritativeQ   bool    `json:"c1_zero_on_authoritative_q"`
	RowsPerComponent         []int   `json:"rows_per_component"`
	RowsAbovePrefixAllocated [][]int `json:"rows_above_qprefix_allocated"`
}
type capacityEvidence struct {
	Rows          int      `json:"rows"`
	PrefixProduct string   `json:"prefix_product"`
	IndependentB  []string `json:"independent_component_bounds"`
	ObservedMax   []string `json:"observer_centered_representative_max_abs"`
	StrictFit     bool     `json:"independent_strict_fit"`
}
type checkpoint struct {
	ID        string            `json:"id"`
	API       string            `json:"api"`
	Decoder   string            `json:"decoder"`
	State     outputState       `json:"state"`
	Plaintext metric            `json:"plaintext_oracle"`
	Capacity  *capacityEvidence `json:"capacity,omitempty"`
	Decoded   []complexValue    `json:"decoded,omitempty"`
	Status    string            `json:"status"`
}
type runEvidence struct {
	SchemaVersion  string       `json:"schema_version"`
	Status         string       `json:"status"`
	CreatedUTC     string       `json:"created_utc"`
	GoVersion      string       `json:"go_version"`
	OS             string       `json:"os"`
	Architecture   string       `json:"architecture"`
	PrimaryCommit  string       `json:"primary_commit"`
	PrimaryDirty   bool         `json:"primary_dirty"`
	BackendCommit  string       `json:"backend_commit"`
	BackendRef     string       `json:"backend_ref"`
	BackendDirty   bool         `json:"backend_dirty"`
	Fast           bool         `json:"fast_zero_secret"`
	KeyLookups     keyLookups   `json:"key_lookups"`
	Profile        profile      `json:"profile"`
	ProfileSHA256  string       `json:"profile_sha256"`
	QPSHA256       string       `json:"effective_qp_sha256"`
	Q              []uint64     `json:"q_primes"`
	P              []uint64     `json:"p_primes"`
	RunnerSHA256   string       `json:"runner_sha256"`
	InputSHA256    string       `json:"input_sha256"`
	Lifecycle      string       `json:"lifecycle"`
	Checkpoints    []checkpoint `json:"checkpoints"`
	BootstrapCalls int          `json:"bootstrap_calls"`
}
type pairedCheckpoint struct {
	ID                 string  `json:"id"`
	Level              int     `json:"level"`
	ScaleLog2          float64 `json:"scale_log2"`
	StandardRMSE       float64 `json:"standard_plaintext_rmse"`
	StandardMax        float64 `json:"standard_plaintext_max"`
	FastRMSE           float64 `json:"fast_plaintext_rmse"`
	FastMax            float64 `json:"fast_plaintext_max"`
	FastVsStandardRMSE float64 `json:"fast_vs_standard_rmse"`
	FastVsStandardMax  float64 `json:"fast_vs_standard_max"`
}
type runSummary struct {
	BackendCommit string       `json:"backend_commit"`
	BackendRef    string       `json:"backend_ref"`
	BackendDirty  bool         `json:"backend_dirty"`
	Fast          bool         `json:"fast_zero_secret"`
	KeyLookups    keyLookups   `json:"key_lookups"`
	Checkpoints   []checkpoint `json:"checkpoints"`
}
type combinedEvidence struct {
	SchemaVersion  string             `json:"schema_version"`
	Status         string             `json:"status"`
	CreatedUTC     string             `json:"created_utc"`
	PrimaryCommit  string             `json:"primary_commit_during_runs"`
	PrimaryDirty   bool               `json:"primary_dirty_during_runs"`
	GoVersion      string             `json:"go_version"`
	OS             string             `json:"os"`
	Architecture   string             `json:"architecture"`
	Standard       runSummary         `json:"standard"`
	Fast           runSummary         `json:"fast"`
	Paired         []pairedCheckpoint `json:"paired_checkpoints"`
	Profile        profile            `json:"profile"`
	ProfileSHA256  string             `json:"profile_sha256"`
	EffectiveQPSHA string             `json:"effective_qp_sha256"`
	Q              []uint64           `json:"q_primes"`
	P              []uint64           `json:"p_primes"`
	RunnerSHA256   string             `json:"runner_sha256"`
	InputSHA256    string             `json:"input_sha256"`
	BootstrapCalls int                `json:"bootstrap_calls"`
	DecodeBoundary string             `json:"decode_boundary"`
	GateTolerance  float64            `json:"fixed_max_error_gate"`
}
type trackedKeySet struct {
	inner rlwe.EvaluationKeySet
	count keyLookups
}

func (k *trackedKeySet) GetGaloisKey(g uint64) (*rlwe.GaloisKey, error) {
	k.count.Galois++
	return k.inner.GetGaloisKey(g)
}
func (k *trackedKeySet) GetGaloisKeysList() []uint64 {
	k.count.GaloisList++
	return k.inner.GetGaloisKeysList()
}
func (k *trackedKeySet) GetRelinearizationKey() (*rlwe.RelinearizationKey, error) {
	k.count.Relinearization++
	return k.inner.GetRelinearizationKey()
}

type capacityObservation struct {
	Rows          int
	PrefixProduct string
	MaxAbs        []string
	StrictFit     bool
}
type capacityAudit interface {
	Observe(name string, ct *rlwe.Ciphertext, rows int) (capacityObservation, error)
}
type derivedBounds struct {
	add, mul, rescale, rotate *big.Int
}

func main() {
	if err := runCLI(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCLI() error {
	out := flag.String("out", "", "output run or paired evidence JSON")
	primaryCommit := flag.String("primary-commit", "", "Primary revision during run")
	primaryDirty := flag.Bool("primary-dirty", false, "whether Primary was dirty")
	backendCommit := flag.String("backend-commit", "", "Lattigo backend revision")
	backendRef := flag.String("backend-ref", "", "Lattigo backend ref")
	backendDirty := flag.Bool("backend-dirty", false, "whether backend was dirty")
	combineStandard := flag.String("combine-standard", "", "Standard run JSON")
	combineFast := flag.String("combine-fast", "", "Fast run JSON")
	flag.Parse()
	if *combineStandard != "" || *combineFast != "" {
		if *combineStandard == "" || *combineFast == "" || *out == "" {
			return errors.New("combine mode requires both run files and -out")
		}
		return combine(*combineStandard, *combineFast, *out)
	}
	if *out == "" || *primaryCommit == "" || *backendCommit == "" || *backendRef == "" {
		return errors.New("-out, -primary-commit, -backend-commit and -backend-ref are required")
	}
	result, err := executeComposition()
	result.PrimaryCommit, result.PrimaryDirty = *primaryCommit, *primaryDirty
	result.BackendCommit, result.BackendRef, result.BackendDirty = *backendCommit, *backendRef, *backendDirty
	if hash, hashErr := sharedRunnerSHA256(); hashErr == nil {
		result.RunnerSHA256 = hash
	} else if err == nil {
		err = fmt.Errorf("hash runner sources: %w", hashErr)
	}
	if writeErr := writeJSON(*out, result); writeErr != nil {
		return writeErr
	}
	return err
}

func executeComposition() (runEvidence, error) {
	used := frozenProfile()
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: used.LogN, LogQ: used.LogQ, LogP: used.LogP,
		LogDefaultScale: used.ScaleLog2, Xs: ring.Ternary{H: 192},
	})
	if err != nil {
		return runEvidence{}, fmt.Errorf("construct six-Q LogN13 profile: %w", err)
	}
	if params.MaxLevel() != 5 || len(params.Q()) != 6 || len(params.P()) != 1 {
		return runEvidence{}, fmt.Errorf("profile mismatch: MaxLevel=%d Q=%d P=%d", params.MaxLevel(), len(params.Q()), len(params.P()))
	}
	isFast := hasFastZeroSecretCapability(params)
	audit := makeCapacityAudit(params)
	if isFast != (audit != nil) {
		return runEvidence{}, errors.New("backend adapter does not match the parameter capability")
	}
	profileJSON, _ := json.Marshal(used)
	qpJSON, _ := json.Marshal(struct{ Q, P []uint64 }{params.Q(), params.P()})
	valuesA, valuesB := deterministicInputs(1 << logSlots)
	boundA, boundB := encodedInputBound(valuesA, math.Exp2(scaleLog2)), encodedInputBound(valuesB, math.Exp2(scaleLog2))
	bounds, err := deriveBounds(params, boundA, boundB)
	if err != nil {
		return runEvidence{}, fmt.Errorf("Checkpoint A capacity preflight: %w", err)
	}
	keys := &trackedKeySet{inner: rlwe.NewMemEvaluationKeySet(nil)}
	keygen := ckks.NewKeyGenerator(params)
	sk := keygen.GenSecretKeyNew()
	relinKey := keygen.GenRelinearizationKeyNew(sk)
	galoisKey := keygen.GenGaloisKeyNew(params.GaloisElementForRotation(1), sk)
	keys.inner = rlwe.NewMemEvaluationKeySet(relinKey, galoisKey)
	evaluator := ckks.NewEvaluator(params, keys)
	encoder := ckks.NewEncoder(params)
	decryptor := rlwe.NewDecryptor(params, sk)
	result := runEvidence{
		SchemaVersion: runSchema, Status: "running", CreatedUTC: time.Now().UTC().Format(time.RFC3339Nano),
		GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH,
		Fast: isFast, Profile: used, ProfileSHA256: hashBytes(profileJSON),
		QPSHA256: hashBytes(qpJSON), Q: params.Q(), P: params.P(),
		InputSHA256: hashInputs(valuesA, valuesB), Lifecycle: lifecycleFor(isFast), BootstrapCalls: 0,
	}
	ctA, err := encrypt(params, encoder, sk, valuesA)
	if err != nil {
		return result, fmt.Errorf("EncryptNew input A: %w", err)
	}
	ctB, err := encrypt(params, encoder, sk, valuesB)
	if err != nil {
		return result, fmt.Errorf("EncryptNew input B: %w", err)
	}
	if isFast && (!zeroC1Prefix(ctA, 4) || !zeroC1Prefix(ctB, 4)) {
		return result, errors.New("Fast EncryptNew violated zero-c1 on q0..q3")
	}
	if err := record(&result, params, encoder, decryptor, audit, "input-a", "EncryptNew", ctA, valuesA, boundA, isFast, false); err != nil {
		return result, err
	}
	if err := record(&result, params, encoder, decryptor, audit, "input-b", "EncryptNew", ctB, valuesB, boundB, isFast, false); err != nil {
		return result, err
	}
	addWant := zip(valuesA, valuesB, func(a, b complex128) complex128 { return a + b })
	sum, err := evaluator.AddNew(ctA, ctB)
	if err != nil {
		return result, fmt.Errorf("AddNew: %w", err)
	}
	if err := record(&result, params, encoder, decryptor, audit, "add", "AddNew", sum, addWant, bounds.add, isFast, isFast); err != nil {
		return result, err
	}
	mulWant := zip(addWant, valuesB, func(a, b complex128) complex128 { return a * b })
	product, err := evaluator.MulRelinNew(sum, ctB)
	if err != nil {
		return result, fmt.Errorf("MulRelinNew: %w", err)
	}
	if err := record(&result, params, encoder, decryptor, audit, "mul-relin", "MulRelinNew", product, mulWant, bounds.mul, isFast, isFast); err != nil {
		return result, err
	}
	rescaled := ckks.NewCiphertext(params, 1, 4)
	if err := evaluator.Rescale(product, rescaled); err != nil {
		return result, fmt.Errorf("Rescale by q5: %w", err)
	}
	if err := record(&result, params, encoder, decryptor, audit, "rescale", "Rescale", rescaled, mulWant, bounds.rescale, isFast, isFast); err != nil {
		return result, err
	}
	rotated, err := evaluator.RotateNew(rescaled, 1)
	if err != nil {
		return result, fmt.Errorf("RotateNew(1): %w", err)
	}
	if err := record(&result, params, encoder, decryptor, audit, "rotate", "RotateNew", rotated, rotate(mulWant, 1), bounds.rotate, isFast, isFast); err != nil {
		return result, err
	}
	result.KeyLookups = keys.count
	if isFast && (keys.count.Relinearization != 0 || keys.count.Galois != 0) {
		return result, fmt.Errorf("Fast public chain looked up per-operation key material: %+v", keys.count)
	}
	result.Status = runPassed
	return result, nil
}

func deriveBounds(params ckks.Parameters, a, b *big.Int) (derivedBounds, error) {
	product := big.NewInt(1)
	for row := 0; row < 4; row++ {
		product.Mul(product, new(big.Int).SetUint64(params.Q()[row]))
	}
	q0 := new(big.Int).SetUint64(params.Q()[0])
	if !strictCapacity(a, q0) || !strictCapacity(b, q0) {
		return derivedBounds{}, fmt.Errorf("encoded inputs are not uniquely centered at q0: A=%s B=%s q0=%s", a, b, q0)
	}
	out := derivedBounds{
		add: new(big.Int).Add(a, b),
		mul: new(big.Int).Mul(new(big.Int).Lsh(big.NewInt(1), logN), new(big.Int).Mul(new(big.Int).Add(a, b), b)),
	}
	q5 := new(big.Int).SetUint64(params.Q()[5])
	rounding := new(big.Int).Rsh(new(big.Int).Sub(new(big.Int).Set(q5), big.NewInt(1)), 1)
	out.rescale = new(big.Int).Quo(new(big.Int).Add(new(big.Int).Set(out.mul), rounding), q5)
	out.rotate = new(big.Int).Set(out.rescale)
	for name, bound := range map[string]*big.Int{
		"input-a": a, "input-b": b, "add": out.add, "mul-relin": out.mul, "rescale": out.rescale, "rotate": out.rotate,
	} {
		if !strictCapacity(bound, product) {
			return out, fmt.Errorf("%s: 2B=%s is not < actual S_Q0123=%s", name, new(big.Int).Lsh(new(big.Int).Set(bound), 1), product)
		}
	}
	return out, nil
}

func encodedInputBound(values []complex128, scale float64) *big.Int {
	maximum := 0.0
	for _, value := range values {
		maximum = math.Max(maximum, cmplx.Abs(value))
	}
	// Embed uses a normalized inverse FFT; twice the scaled maximum plus one
	// covers Float64 twiddle/roundoff and fixed-point nearest rounding.
	return big.NewInt(int64(math.Ceil(2*maximum*scale + 1)))
}

func strictCapacity(bound, product *big.Int) bool {
	return bound != nil && product != nil && bound.Sign() >= 0 && product.Sign() > 0 &&
		new(big.Int).Lsh(new(big.Int).Set(bound), 1).Cmp(product) < 0
}

func encrypt(params ckks.Parameters, encoder *ckks.Encoder, sk *rlwe.SecretKey, values []complex128) (*rlwe.Ciphertext, error) {
	pt := ckks.NewPlaintext(params, 5)
	pt.Scale = rlwe.NewScale(math.Exp2(scaleLog2))
	pt.IsNTT = true
	pt.LogDimensions = ring.Dimensions{Cols: logSlots}
	if err := encoder.Encode(values, pt); err != nil {
		return nil, err
	}
	return rlwe.NewEncryptor(params, sk).EncryptNew(pt)
}

func record(result *runEvidence, params ckks.Parameters, encoder *ckks.Encoder, decryptor *rlwe.Decryptor, audit capacityAudit, id, api string, ct *rlwe.Ciphertext, want []complex128, bound *big.Int, fast, requireCompact bool) error {
	rows := qPrefixRows(ct.Level())
	var capacity *capacityEvidence
	var decoded []complex128
	decoder := "native Standard DecryptNew/Decode"
	if fast {
		var err error
		capacity, err = observeCapacity(audit, params, id, ct, rows, bound)
		if err != nil {
			return err
		}
		pt, err := materializeFastC0ForMeasurement(params, ct, rows, bound)
		if err != nil {
			return fmt.Errorf("%s measurement-only decoder: %w", id, err)
		}
		decoded = make([]complex128, len(want))
		if err := encoder.Decode(pt, decoded); err != nil {
			return fmt.Errorf("%s measurement-only Decode: %w", id, err)
		}
		decoder = "bounded measurement-only centered q0..q3 lift of zero-secret c0; no full-active-Q decrypt"
	} else {
		decoded = make([]complex128, len(want))
		if err := encoder.Decode(decryptor.DecryptNew(ct), decoded); err != nil {
			return fmt.Errorf("%s Standard DecryptNew/Decode: %w", id, err)
		}
	}
	measurement := compare(want, decoded)
	state := summarize(ct, rows)
	level := 5
	expectedScale := rlwe.NewScale(math.Exp2(scaleLog2))
	switch api {
	case "MulRelinNew":
		expectedScale = expectedScale.Mul(rlwe.NewScale(math.Exp2(scaleLog2)))
	case "Rescale", "RotateNew":
		level = 4
		expectedScale = expectedScale.Mul(rlwe.NewScale(math.Exp2(scaleLog2))).Div(rlwe.NewScale(params.Q()[5]))
	}
	status := "passed"
	if !measurement.Finite || measurement.MaxComplex > gateTolerance || state.Level != level || state.Degree != 1 ||
		math.Abs(state.ScaleLog2-expectedScale.Log2()) > 1e-9 || !state.IsNTT || state.IsMontgomery {
		status = "failed_numerical_or_state_gate"
	}
	if fast && (!state.C1ZeroOnAuthoritativeQ ||
		(requireCompact && (!compactRowsMatch(state, rows) || hasRowsAbovePrefix(state)))) {
		status = "failed_fast_row_or_zero_secret_contract"
	}
	wire := make([]complexValue, len(decoded))
	for i, value := range decoded {
		wire[i] = complexValue{real(value), imag(value)}
	}
	result.Checkpoints = append(result.Checkpoints, checkpoint{
		ID: id, API: api, Decoder: decoder, State: state, Plaintext: measurement,
		Capacity: capacity, Decoded: wire, Status: status,
	})
	if status != "passed" {
		return fmt.Errorf("%s status=%s level=%d scaleLog2=%.12g RMSE=%.9g max=%.9g gate=%.1e",
			id, status, state.Level, state.ScaleLog2, measurement.ComplexRMSE, measurement.MaxComplex, gateTolerance)
	}
	return nil
}

func observeCapacity(audit capacityAudit, params ckks.Parameters, name string, ct *rlwe.Ciphertext, rows int, bound *big.Int) (*capacityEvidence, error) {
	prefix := big.NewInt(1)
	for i := 0; i < rows; i++ {
		prefix.Mul(prefix, new(big.Int).SetUint64(params.Q()[i]))
	}
	independentBounds := make([]string, len(ct.Value))
	for i := range independentBounds {
		independentBounds[i] = "0"
		if i == 0 {
			independentBounds[i] = bound.String()
		}
		v, _ := new(big.Int).SetString(independentBounds[i], 10)
		if !strictCapacity(v, prefix) {
			return nil, fmt.Errorf("%s component %d fails independent 2B<S_Q gate", name, i)
		}
	}
	snapshot, err := audit.Observe(name, ct, rows)
	if err != nil {
		return nil, fmt.Errorf("%s existing capacity observer: %w", name, err)
	}
	if snapshot.Rows != rows || snapshot.PrefixProduct != prefix.String() || !snapshot.StrictFit || len(snapshot.MaxAbs) != len(ct.Value) {
		return nil, fmt.Errorf("%s observer/profile mismatch: %+v", name, snapshot)
	}
	for i, text := range snapshot.MaxAbs {
		observed, ok := new(big.Int).SetString(text, 10)
		limit, _ := new(big.Int).SetString(independentBounds[i], 10)
		if !ok || observed.Cmp(limit) > 0 {
			return nil, fmt.Errorf("%s component %d centered representative %q exceeds independent B=%s", name, i, text, limit)
		}
	}
	return &capacityEvidence{Rows: rows, PrefixProduct: prefix.String(), IndependentB: independentBounds, ObservedMax: snapshot.MaxAbs, StrictFit: true}, nil
}

func materializeFastC0ForMeasurement(params ckks.Parameters, ct *rlwe.Ciphertext, rows int, bound *big.Int) (*rlwe.Plaintext, error) {
	if ct == nil || ct.MetaData == nil || len(ct.Value) != 2 || ct.Level() < 0 || ct.Level() > params.MaxLevel() {
		return nil, errors.New("invalid ciphertext shape or metadata")
	}
	if rows != qPrefixRows(ct.Level()) || rows < 1 || rows > len(params.Q()) || !strictCapacity(bound, prefixProduct(params.Q(), rows)) {
		return nil, fmt.Errorf("rows/bound do not satisfy Level %d Q-prefix contract", ct.Level())
	}
	if !zeroC1Prefix(ct, rows) {
		return nil, errors.New("c1 is nonzero or lacks authoritative-row backing")
	}
	coeff := ring.NewPoly(params.N(), rows-1)
	for row := 0; row < rows; row++ {
		if row >= len(ct.Value[0].Coeffs) || len(ct.Value[0].Coeffs[row]) != params.N() {
			return nil, fmt.Errorf("c0 lacks q%d backing", row)
		}
		for i, value := range ct.Value[0].Coeffs[row] {
			if value >= params.Q()[row] {
				return nil, fmt.Errorf("c0 q%d coefficient %d is not canonical", row, i)
			}
		}
		if ct.IsNTT {
			params.RingQ().SubRings[row].INTT(ct.Value[0].Coeffs[row], coeff.Coeffs[row])
		} else {
			copy(coeff.Coeffs[row], ct.Value[0].Coeffs[row])
		}
		if ct.IsMontgomery {
			params.RingQ().SubRings[row].IMForm(coeff.Coeffs[row], coeff.Coeffs[row])
		}
	}
	pt := ckks.NewPlaintext(params, ct.Level())
	pt.Scale = ct.Scale
	pt.IsNTT, pt.IsMontgomery = false, false
	pt.IsBatched, pt.LogDimensions = ct.IsBatched, ct.LogDimensions
	for coefficient := 0; coefficient < params.N(); coefficient++ {
		x, _, err := centeredCRT(coeff, params.Q(), rows, coefficient)
		if err != nil {
			return nil, err
		}
		if new(big.Int).Abs(new(big.Int).Set(x)).Cmp(bound) > 0 {
			return nil, fmt.Errorf("coefficient %d centered lift %s exceeds independent B=%s", coefficient, x, bound)
		}
		for row := 0; row <= ct.Level(); row++ {
			qi := new(big.Int).SetUint64(params.Q()[row])
			residue := new(big.Int).Mod(new(big.Int).Set(x), qi)
			if residue.Sign() < 0 {
				residue.Add(residue, qi)
			}
			pt.Value.Coeffs[row][coefficient] = residue.Uint64()
		}
	}
	return pt, nil
}

func centeredCRT(coeff ring.Poly, q []uint64, rows, coefficient int) (*big.Int, *big.Int, error) {
	x, product := new(big.Int), big.NewInt(1)
	for row := 0; row < rows; row++ {
		qi := new(big.Int).SetUint64(q[row])
		residue := new(big.Int).SetUint64(coeff.Coeffs[row][coefficient])
		delta := new(big.Int).Sub(residue, new(big.Int).Mod(new(big.Int).Set(x), qi))
		delta.Mod(delta, qi)
		inverse := new(big.Int).ModInverse(new(big.Int).Mod(new(big.Int).Set(product), qi), qi)
		if inverse == nil {
			return nil, nil, fmt.Errorf("CRT inverse unavailable for q%d", row)
		}
		delta.Mul(delta, inverse).Mod(delta, qi)
		x.Add(x, new(big.Int).Mul(product, delta))
		product.Mul(product, qi)
	}
	if x.Cmp(new(big.Int).Rsh(new(big.Int).Set(product), 1)) > 0 {
		x.Sub(x, product)
	}
	return x, product, nil
}

func prefixProduct(q []uint64, rows int) *big.Int {
	product := big.NewInt(1)
	for i := 0; i < rows; i++ {
		product.Mul(product, new(big.Int).SetUint64(q[i]))
	}
	return product
}

func summarize(ct *rlwe.Ciphertext, rows int) outputState {
	s := outputState{
		Level: ct.Level(), Degree: ct.Degree(), ScaleLog2: ct.Scale.Log2(),
		IsNTT: ct.IsNTT, IsMontgomery: ct.IsMontgomery, C1ZeroOnAuthoritativeQ: zeroC1Prefix(ct, rows),
	}
	for _, poly := range ct.Value {
		active, dormant := 0, []int{}
		for row := 0; row <= ct.Level() && row < len(poly.Coeffs); row++ {
			if len(poly.Coeffs[row]) > 0 {
				if row < rows {
					active++
				} else {
					dormant = append(dormant, row)
				}
			}
		}
		s.RowsPerComponent = append(s.RowsPerComponent, active)
		s.RowsAbovePrefixAllocated = append(s.RowsAbovePrefixAllocated, dormant)
	}
	return s
}

func zeroC1Prefix(ct *rlwe.Ciphertext, rows int) bool {
	if ct == nil || len(ct.Value) < 2 {
		return false
	}
	for row := 0; row < rows; row++ {
		if row >= len(ct.Value[1].Coeffs) || len(ct.Value[1].Coeffs[row]) == 0 {
			return false
		}
		for _, value := range ct.Value[1].Coeffs[row] {
			if value != 0 {
				return false
			}
		}
	}
	return true
}

func qPrefixRows(level int) int { return min(level+1, 4) }
func compactRowsMatch(s outputState, rows int) bool {
	return len(s.RowsPerComponent) == 2 && s.RowsPerComponent[0] == rows && s.RowsPerComponent[1] == rows
}
func hasRowsAbovePrefix(s outputState) bool {
	for _, rows := range s.RowsAbovePrefixAllocated {
		if len(rows) != 0 {
			return true
		}
	}
	return false
}

func combine(standardPath, fastPath, outPath string) error {
	var standard, fast runEvidence
	if err := readJSON(standardPath, &standard); err != nil {
		return err
	}
	if err := readJSON(fastPath, &fast); err != nil {
		return err
	}
	params, err := frozenParameters()
	if err != nil {
		return err
	}
	profileJSON, _ := json.Marshal(frozenProfile())
	qpJSON, _ := json.Marshal(struct{ Q, P []uint64 }{params.Q(), params.P()})
	a, b := deterministicInputs(1 << logSlots)
	runnerHash, err := sharedRunnerSHA256()
	if err != nil {
		return err
	}
	if err := validateRun(standard, false, params, hashBytes(profileJSON), hashBytes(qpJSON), hashInputs(a, b), runnerHash); err != nil {
		return fmt.Errorf("Standard run: %w", err)
	}
	if err := validateRun(fast, true, params, hashBytes(profileJSON), hashBytes(qpJSON), hashInputs(a, b), runnerHash); err != nil {
		return fmt.Errorf("Fast run: %w", err)
	}
	if standard.PrimaryCommit != fast.PrimaryCommit || standard.PrimaryDirty || fast.PrimaryDirty ||
		standard.GoVersion != fast.GoVersion || standard.OS != fast.OS || standard.Architecture != fast.Architecture {
		return errors.New("paired runs do not share clean Primary/runtime provenance")
	}
	evidence := combinedEvidence{
		SchemaVersion: pairSchema, Status: "paired_level5_public_chain_passed",
		CreatedUTC:    time.Now().UTC().Format(time.RFC3339Nano),
		PrimaryCommit: standard.PrimaryCommit, Standard: summarizeRun(standard), Fast: summarizeRun(fast),
		GoVersion: standard.GoVersion, OS: standard.OS, Architecture: standard.Architecture,
		Profile: standard.Profile, ProfileSHA256: standard.ProfileSHA256, EffectiveQPSHA: standard.QPSHA256,
		Q: standard.Q, P: standard.P,
		RunnerSHA256: standard.RunnerSHA256, InputSHA256: standard.InputSHA256, BootstrapCalls: 0,
		DecodeBoundary: "Fast is decoded only at a measurement boundary. A compositional independent bound proves 2B<S_Q0123; c0 is reconstructed from q0..q3, centered, then extended into a fresh logical-Level plaintext for the common Encoder. q4/q5 are not read by this materializer or the Fast public operation outputs. Native Fast full-active-Q DecryptNew is not used.",
		GateTolerance:  gateTolerance,
	}
	for i, sc := range standard.Checkpoints {
		fc := fast.Checkpoints[i]
		if sc.ID != fc.ID || sc.State.Level != fc.State.Level || sc.State.Degree != fc.State.Degree ||
			math.Abs(sc.State.ScaleLog2-fc.State.ScaleLog2) > 1e-9 {
			return fmt.Errorf("checkpoint state mismatch at %d: %s vs %s", i, sc.ID, fc.ID)
		}
		pair := compare(valuesFromEvidence(sc.Decoded), valuesFromEvidence(fc.Decoded))
		if !pair.Finite || pair.MaxComplex > gateTolerance {
			return fmt.Errorf("%s Fast-vs-Standard failed: RMSE=%.9g max=%.9g threshold=%.1e", sc.ID, pair.ComplexRMSE, pair.MaxComplex, gateTolerance)
		}
		evidence.Paired = append(evidence.Paired, pairedCheckpoint{
			ID: sc.ID, Level: sc.State.Level, ScaleLog2: sc.State.ScaleLog2,
			StandardRMSE: sc.Plaintext.ComplexRMSE, StandardMax: sc.Plaintext.MaxComplex,
			FastRMSE: fc.Plaintext.ComplexRMSE, FastMax: fc.Plaintext.MaxComplex,
			FastVsStandardRMSE: pair.ComplexRMSE, FastVsStandardMax: pair.MaxComplex,
		})
	}
	if fast.KeyLookups.Relinearization != 0 || fast.KeyLookups.Galois != 0 {
		return fmt.Errorf("Fast public operations unexpectedly looked up keys: %+v", fast.KeyLookups)
	}
	return writeJSON(outPath, evidence)
}

func validateRun(run runEvidence, isFast bool, params ckks.Parameters, profileHash, qpHash, inputHash, runnerHash string) error {
	if run.SchemaVersion != runSchema || run.Status != runPassed || run.BootstrapCalls != 0 ||
		run.Fast != isFast || run.PrimaryCommit == "" || run.PrimaryDirty {
		return errors.New("invalid status, backend capability, Primary provenance, or Bootstrap count")
	}
	if run.GoVersion == "" || run.OS == "" || run.Architecture == "" {
		return errors.New("missing runtime metadata")
	}
	if isFast {
		if run.BackendCommit != fastCommit || run.BackendRef != "fast-qprefix" || run.BackendDirty {
			return errors.New("Fast backend pin/ref/clean state mismatch")
		}
	} else if run.BackendCommit != standardCommit || run.BackendRef != "pinned-standard-"+standardCommit || run.BackendDirty {
		return errors.New("Standard backend pin/ref/clean state mismatch")
	}
	if _, err := time.Parse(time.RFC3339Nano, run.CreatedUTC); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}
	p, _ := json.Marshal(run.Profile)
	if run.ProfileSHA256 != profileHash || hashBytes(p) != profileHash {
		return errors.New("profile hash mismatch")
	}
	qp, _ := json.Marshal(struct{ Q, P []uint64 }{run.Q, run.P})
	if run.QPSHA256 != qpHash || hashBytes(qp) != qpHash {
		return errors.New("actual Q/P hash mismatch")
	}
	if run.InputSHA256 != inputHash || run.RunnerSHA256 != runnerHash || run.Lifecycle != lifecycleFor(isFast) {
		return errors.New("input, runner, or lifecycle mismatch")
	}
	inputA, inputB := deterministicInputs(1 << logSlots)
	boundA := encodedInputBound(inputA, math.Exp2(scaleLog2))
	boundB := encodedInputBound(inputB, math.Exp2(scaleLog2))
	bounds, err := deriveBounds(params, boundA, boundB)
	if err != nil {
		return fmt.Errorf("frozen profile capacity proof: %w", err)
	}
	ids := []string{"input-a", "input-b", "add", "mul-relin", "rescale", "rotate"}
	levels := []int{5, 5, 5, 5, 4, 4}
	independentBounds := []*big.Int{boundA, boundB, bounds.add, bounds.mul, bounds.rescale, bounds.rotate}
	baseScale := rlwe.NewScale(math.Exp2(scaleLog2))
	if len(run.Checkpoints) != len(ids) {
		return fmt.Errorf("expected %d checkpoints, got %d", len(ids), len(run.Checkpoints))
	}
	for i, cp := range run.Checkpoints {
		if cp.ID != ids[i] || cp.Status != "passed" || cp.State.Level != levels[i] || cp.State.Degree != 1 ||
			len(cp.Decoded) != 1<<logSlots || !cp.Plaintext.Finite || cp.Plaintext.MaxComplex > gateTolerance {
			return fmt.Errorf("checkpoint %d failed identity/state/numerical gate", i)
		}
		if !cp.State.IsNTT || cp.State.IsMontgomery {
			return fmt.Errorf("%s unexpected NTT/Montgomery state", cp.ID)
		}
		expectedScale := baseScale
		if i == 3 {
			expectedScale = baseScale.Mul(baseScale)
		} else if i >= 4 {
			expectedScale = baseScale.Mul(baseScale).Div(rlwe.NewScale(params.Q()[5]))
		}
		if math.Abs(cp.State.ScaleLog2-expectedScale.Log2()) > 1e-9 {
			return fmt.Errorf("%s Scale progression mismatch", cp.ID)
		}
		for j, x := range cp.Decoded {
			if !finite(x.Real) || !finite(x.Imag) {
				return fmt.Errorf("%s sample %d nonfinite", cp.ID, j)
			}
		}
		if isFast {
			rows := qPrefixRows(cp.State.Level)
			if cp.Capacity == nil || cp.Capacity.Rows != rows || !cp.Capacity.StrictFit ||
				len(cp.State.RowsPerComponent) != 2 || cp.State.RowsPerComponent[0] < rows ||
				cp.State.RowsPerComponent[1] < rows || !cp.State.C1ZeroOnAuthoritativeQ {
				return fmt.Errorf("%s lacks Fast capacity/row/zero-c1 evidence", cp.ID)
			}
			prefix := prefixProduct(run.Q, rows)
			if cp.Capacity.PrefixProduct != prefix.String() ||
				len(cp.Capacity.IndependentB) != 2 || len(cp.Capacity.ObservedMax) != 2 {
				return fmt.Errorf("%s capacity evidence has the wrong prefix or component count", cp.ID)
			}
			for component := 0; component < 2; component++ {
				wantBound := big.NewInt(0)
				if component == 0 {
					wantBound = independentBounds[i]
				}
				bound, okBound := new(big.Int).SetString(cp.Capacity.IndependentB[component], 10)
				observed, okObserved := new(big.Int).SetString(cp.Capacity.ObservedMax[component], 10)
				if !okBound || !okObserved || bound.Cmp(wantBound) != 0 ||
					!strictCapacity(bound, prefix) || observed.Cmp(bound) > 0 {
					return fmt.Errorf("%s component %d capacity evidence differs from the independent bound", cp.ID, component)
				}
			}
			if i >= 2 && (!compactRowsMatch(cp.State, rows) || hasRowsAbovePrefix(cp.State)) {
				return fmt.Errorf("%s public operation output is not compact at its Q-prefix", cp.ID)
			}
		} else if cp.Capacity != nil {
			return fmt.Errorf("Standard checkpoint %s has Fast-only evidence", cp.ID)
		}
	}
	return nil
}

func summarizeRun(run runEvidence) runSummary {
	cp := make([]checkpoint, len(run.Checkpoints))
	copy(cp, run.Checkpoints)
	for i := range cp {
		cp[i].Decoded = nil
	}
	return runSummary{run.BackendCommit, run.BackendRef, run.BackendDirty, run.Fast, run.KeyLookups, cp}
}

func deterministicInputs(slots int) ([]complex128, []complex128) {
	a, b := make([]complex128, slots), make([]complex128, slots)
	for i := range a {
		a[i] = complex(float64(i%7-3)/32, float64((2*i)%5-2)/64)
		b[i] = complex(float64(i%5-2)/16, float64(i%3-1)/32)
	}
	return a, b
}
func zip(a, b []complex128, f func(complex128, complex128) complex128) []complex128 {
	out := make([]complex128, len(a))
	for i := range a {
		out[i] = f(a[i], b[i])
	}
	return out
}
func rotate(values []complex128, k int) []complex128 {
	out := make([]complex128, len(values))
	shift := k % len(values)
	for i := range out {
		out[i] = values[(i+shift)%len(values)]
	}
	return out
}
func compare(want, got []complex128) metric {
	out := metric{}
	if len(want) == 0 || len(want) != len(got) {
		return out
	}
	var squares float64
	for i := range want {
		if !finite(real(want[i])) || !finite(imag(want[i])) || !finite(real(got[i])) || !finite(imag(got[i])) {
			return metric{}
		}
		d := cmplx.Abs(got[i] - want[i])
		if !finite(d) {
			return metric{}
		}
		squares += d * d
		out.MaxComplex = math.Max(out.MaxComplex, d)
	}
	out.ComplexRMSE = math.Sqrt(squares / float64(len(want)))
	out.Finite = finite(out.ComplexRMSE) && finite(out.MaxComplex)
	return out
}
func frozenProfile() profile {
	return profile{logN, []int{55, 39, 40, 39, 39, 39}, []int{60}, scaleLog2, logSlots, 5}
}
func frozenParameters() (ckks.Parameters, error) {
	p := frozenProfile()
	return ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: p.LogN, LogQ: p.LogQ, LogP: p.LogP, LogDefaultScale: p.ScaleLog2, Xs: ring.Ternary{H: 192},
	})
}
func lifecycleFor(fast bool) string {
	if fast {
		return "same public CKKS frontend/profile; zero-secret key capability; Encode/EncryptNew(c0=encoded pt,c1=0); AddNew/MulRelinNew/Rescale/RotateNew; measurement-only bounded q0..q3 lift and common Encoder.Decode"
	}
	return "same public CKKS frontend/profile; native Standard keygen/Encode/EncryptNew; AddNew/MulRelinNew/Rescale/RotateNew; native DecryptNew/Decode"
}
func hasFastZeroSecretCapability(p ckks.Parameters) bool {
	_, ok := any(p).(interface{ FastCKKSZeroSecretSimulation() })
	return ok
}
func sharedRunnerSHA256() (string, error) {
	paths := []string{
		"tools/fast-dropin-above-cap-numerical-batch-016/main.go",
		"tools/fast-dropin-above-cap-numerical-batch-016/capacity_adapter_fast.go",
		"tools/fast-dropin-above-cap-numerical-batch-016/capacity_adapter_standard.go",
	}
	h := sha256.New()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		_, _ = h.Write([]byte(path))
		_, _ = h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func hashInputs(inputs ...[]complex128) string {
	h := sha256.New()
	var b [16]byte
	for _, values := range inputs {
		binary.LittleEndian.PutUint64(b[:8], uint64(len(values)))
		_, _ = h.Write(b[:8])
		for _, v := range values {
			binary.LittleEndian.PutUint64(b[:8], math.Float64bits(real(v)))
			binary.LittleEndian.PutUint64(b[8:], math.Float64bits(imag(v)))
			_, _ = h.Write(b[:])
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}
func hashBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func valuesFromEvidence(v []complexValue) []complex128 {
	out := make([]complex128, len(v))
	for i := range v {
		out[i] = complex(v[i].Real, v[i].Imag)
	}
	return out
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func readJSON(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

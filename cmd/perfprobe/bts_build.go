package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	bts008SourceManifestSHA = "53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda"
	bts008PrimaryBaseline   = "f3eb665e35f15b91d5f290626fbb7a3bb360a95c"
	bts008StandardCommit    = "5dbffbdea05394de2ca3a432ed5318aa832e3f40"
	bts008FastCommit        = "ee308c9743000b5f281a84a871e7c4bd59b176c4"
	bts008NestedModulePath  = "github.com/xuejin-lu/heart-lattigo-bootstrap/btsmeasurement"
)

type btsBuildOptions struct {
	backend       string
	primaryRoot   string
	secondaryRoot string
	binaryOut     string
	workRoot      string
}

type btsBuildManifest struct {
	SchemaVersion                  string            `json:"schema_version"`
	Task                           string            `json:"task"`
	Backend                        string            `json:"backend"`
	BackendCommit                  string            `json:"backend_commit"`
	BackendRef                     string            `json:"backend_ref"`
	BackendClean                   bool              `json:"backend_clean"`
	BackendRoot                    string            `json:"backend_root"`
	PrimaryCommit                  string            `json:"primary_commit"`
	PrimaryDirty                   bool              `json:"primary_dirty"`
	PrimaryMeasurementSourceSHA256 map[string]string `json:"primary_measurement_source_sha256"`
	OriginalSourceRoot             string            `json:"original_source_root"`
	OriginalSourceManifest         string            `json:"original_source_manifest_sha256"`
	MeasurementAdapter             string            `json:"measurement_adapter_path"`
	MeasurementAdapterSHA256       string            `json:"measurement_adapter_sha256"`
	BinaryPath                     string            `json:"binary_path"`
	BinarySHA256                   string            `json:"binary_sha256"`
	GoVersion                      string            `json:"go_version"`
	OS                             string            `json:"os"`
	Arch                           string            `json:"arch"`
	BuildCommand                   []string          `json:"build_command"`
	GoBuildInfo                    string            `json:"go_build_info"`
	BuiltAt                        time.Time         `json:"built_at"`
}

func runBTSBuild(args []string) error {
	flags := flag.NewFlagSet("perfprobe bts-build", flag.ContinueOnError)
	var opts btsBuildOptions
	flags.StringVar(&opts.backend, "backend", "", "standard or fast")
	flags.StringVar(&opts.primaryRoot, "primary-root", "", "Primary repository root")
	flags.StringVar(&opts.secondaryRoot, "secondary-root", "", "clean pinned Standard or Fast Lattigo checkout")
	flags.StringVar(&opts.binaryOut, "binary-out", "", "new executable output path")
	flags.StringVar(&opts.workRoot, "work-root", "", "new temporary build-work directory (must not already exist)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || opts.primaryRoot == "" || opts.secondaryRoot == "" || opts.binaryOut == "" || opts.workRoot == "" {
		return errors.New("--primary-root, --secondary-root, --binary-out, and --work-root are required")
	}
	if opts.backend != "standard" && opts.backend != "fast" {
		return fmt.Errorf("unsupported BTS backend %q", opts.backend)
	}
	primaryRoot, err := filepath.Abs(opts.primaryRoot)
	if err != nil {
		return err
	}
	secondaryRoot, err := filepath.Abs(opts.secondaryRoot)
	if err != nil {
		return err
	}
	binaryOut, err := filepath.Abs(opts.binaryOut)
	if err != nil {
		return err
	}
	workRoot, err := filepath.Abs(opts.workRoot)
	if err != nil {
		return err
	}
	if _, err := os.Stat(workRoot); err == nil {
		return fmt.Errorf("refusing to reuse existing build work root %s", workRoot)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := ensureBTSBuildOutputAbsent(binaryOut); err != nil {
		return err
	}
	if err := os.MkdirAll(workRoot, 0o700); err != nil {
		return err
	}

	wantBackend := bts008StandardCommit
	if opts.backend == "fast" {
		wantBackend = bts008FastCommit
	}
	backendRoot, backendCommit, backendRef, err := cleanSecondaryState(secondaryRoot, wantBackend)
	if err != nil {
		return fmt.Errorf("validate %s checkout: %w", opts.backend, err)
	}
	manifestSHA, err := btsOriginalSourceManifest(primaryRoot)
	if err != nil {
		return err
	}
	if manifestSHA != bts008SourceManifestSHA {
		return fmt.Errorf("external/BTS source manifest=%s, want pinned %s", manifestSHA, bts008SourceManifestSHA)
	}
	primaryCommit, err := gitOutput(primaryRoot, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	primaryStatus, err := gitOutput(primaryRoot, "status", "--porcelain")
	if err != nil {
		return err
	}
	measurementSourceSHAs, err := hashPrimaryMeasurementSources(primaryRoot)
	if err != nil {
		return err
	}
	adapterPath := filepath.Join(primaryRoot, "cmd", "perfprobe", "testdata", "bts", "overlay_main.go")
	adapterSHA, err := sha256File(adapterPath)
	if err != nil {
		return err
	}

	btsRoot := filepath.Join(primaryRoot, "external", "BTS")
	originalMod, err := os.ReadFile(filepath.Join(btsRoot, "go.mod"))
	if err != nil {
		return err
	}
	modText, err := rewriteBTSModfile(string(originalMod), primaryRoot, backendRoot)
	if err != nil {
		return err
	}
	modPath := filepath.Join(workRoot, "bts.mod")
	if err := os.WriteFile(modPath, []byte(modText), 0o600); err != nil {
		return err
	}
	sumData, err := os.ReadFile(filepath.Join(btsRoot, "go.sum"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workRoot, "bts.sum"), sumData, 0o600); err != nil {
		return err
	}
	overlayPath := filepath.Join(workRoot, "overlay.json")
	overlayData, err := json.Marshal(struct {
		Replace map[string]string `json:"Replace"`
	}{Replace: map[string]string{filepath.Join(btsRoot, "main.go"): adapterPath}})
	if err != nil {
		return err
	}
	if err := os.WriteFile(overlayPath, overlayData, 0o600); err != nil {
		return err
	}
	cachePath := filepath.Join(workRoot, "gocache")
	if err := os.MkdirAll(cachePath, 0o700); err != nil {
		return err
	}
	buildArgs := []string{"build", "-mod=mod", "-modfile=" + modPath, "-overlay=" + overlayPath, "-o", binaryOut, "."}
	command := exec.Command("go", buildArgs...)
	command.Dir = btsRoot
	command.Env = setEnvironment(os.Environ(), map[string]string{"GOWORK": "off", "GOCACHE": cachePath})
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	fmt.Printf("BTS %s build: %s\n", opts.backend, strings.Join(append([]string{"go"}, buildArgs...), " "))
	if err := command.Run(); err != nil {
		return fmt.Errorf("build BTS %s adapter: %w", opts.backend, err)
	}
	binarySHA, err := sha256File(binaryOut)
	if err != nil {
		return err
	}
	infoCommand := exec.Command("go", "version", "-m", binaryOut)
	infoBytes, err := infoCommand.Output()
	if err != nil {
		return fmt.Errorf("read built binary module provenance: %w", err)
	}
	info := string(infoBytes)
	if !strings.Contains(info, backendRoot) || !strings.Contains(info, primaryRoot) {
		return fmt.Errorf("binary provenance does not bind both selected backend and Primary measurement helper; go version -m:\n%s", info)
	}
	manifest := btsBuildManifest{
		SchemaVersion: "bts-notrace-warm-and-numerical-008-build.v1", Task: "BTS-NOTRACE-WARM-AND-NUMERICAL-008",
		Backend: opts.backend, BackendCommit: backendCommit, BackendRef: backendRef, BackendClean: true,
		BackendRoot: backendRoot, PrimaryCommit: primaryCommit, PrimaryDirty: primaryStatus != "",
		PrimaryMeasurementSourceSHA256: measurementSourceSHAs,
		OriginalSourceRoot:             btsRoot, OriginalSourceManifest: manifestSHA,
		MeasurementAdapter: adapterPath, MeasurementAdapterSHA256: adapterSHA,
		BinaryPath: binaryOut, BinarySHA256: binarySHA, GoVersion: runtime.Version(),
		OS: runtime.GOOS, Arch: runtime.GOARCH, BuildCommand: append([]string{"go"}, buildArgs...),
		GoBuildInfo: info, BuiltAt: time.Now().UTC(),
	}
	manifestPath := binaryOut + ".manifest.json"
	if err := writeExclusiveJSON(manifestPath, manifest); err != nil {
		return err
	}
	fmt.Printf("BTS %s build ready: commit=%s binary_sha256=%s manifest=%s\n", opts.backend, backendCommit, binarySHA, manifestPath)
	return nil
}

func hashPrimaryMeasurementSources(primaryRoot string) (map[string]string, error) {
	result := make(map[string]string)
	for _, directory := range []string{"internal/perfmeasure", "internal/numericalmetrics"} {
		paths, err := filepath.Glob(filepath.Join(primaryRoot, directory, "*.go"))
		if err != nil {
			return nil, err
		}
		if len(paths) == 0 {
			return nil, fmt.Errorf("no Go sources found for imported Primary package %s", directory)
		}
		sort.Strings(paths)
		for _, path := range paths {
			digest, err := sha256File(path)
			if err != nil {
				return nil, err
			}
			relative, err := filepath.Rel(primaryRoot, path)
			if err != nil {
				return nil, err
			}
			result[filepath.ToSlash(relative)] = digest
		}
	}
	return result, nil
}

func rewriteBTSModfile(original, primaryRoot, secondaryRoot string) (string, error) {
	if !strings.HasPrefix(original, "module mys2c\n") {
		return "", errors.New("unexpected external/BTS go.mod module declaration")
	}
	module := "module " + bts008NestedModulePath + "\n"
	updated := strings.Replace(original, "module mys2c\n", module, 1)
	updated += "\nrequire github.com/xuejin-lu/heart-lattigo-bootstrap v0.0.0\n"
	updated += "replace github.com/xuejin-lu/heart-lattigo-bootstrap => " + filepath.ToSlash(primaryRoot) + "\n"
	updated += "replace github.com/tuneinsight/lattigo/v6 => " + filepath.ToSlash(secondaryRoot) + "\n"
	return updated, nil
}

func btsOriginalSourceManifest(primaryRoot string) (string, error) {
	btsRoot := filepath.Join(primaryRoot, "external", "BTS")
	status, err := gitOutput(primaryRoot, "status", "--porcelain", "--untracked-files=all", "--", "external/BTS")
	if err != nil {
		return "", err
	}
	if status != "" {
		return "", fmt.Errorf("external/BTS has local changes; refusing measurement build:\n%s", status)
	}
	changed, err := gitOutput(primaryRoot, "diff", "--name-only", bts008PrimaryBaseline, "--", "external/BTS")
	if err != nil {
		return "", err
	}
	if changed != "" {
		return "", fmt.Errorf("external/BTS differs from its approved Batch008 baseline:\n%s", changed)
	}
	goFiles, err := filepath.Glob(filepath.Join(btsRoot, "*.go"))
	if err != nil {
		return "", err
	}
	paths := append(goFiles, filepath.Join(btsRoot, "go.mod"), filepath.Join(btsRoot, "go.sum"), filepath.Join(btsRoot, "run.sh"))
	sort.Strings(paths)
	if len(paths) != 13 {
		return "", fmt.Errorf("external/BTS source file count=%d, want 13", len(paths))
	}
	hash := sha256.New()
	for _, path := range paths {
		fileSHA, err := sha256File(path)
		if err != nil {
			return "", err
		}
		relative := filepath.ToSlash(filepath.Join("external", "BTS", filepath.Base(path)))
		if _, err := fmt.Fprintf(hash, "%s  %s\n", fileSHA, relative); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func ensureBTSBuildOutputAbsent(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("refusing to overwrite existing build output %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	if _, err := os.Lstat(path + ".manifest.json"); err == nil {
		return fmt.Errorf("refusing to overwrite existing build manifest %s", path+".manifest.json")
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(filepath.Dir(path), 0o700)
}

func setEnvironment(existing []string, replacements map[string]string) []string {
	result := make([]string, 0, len(existing)+len(replacements))
	for _, item := range existing {
		key, _, found := strings.Cut(item, "=")
		if _, replace := replacements[key]; found && replace {
			continue
		}
		result = append(result, item)
	}
	for key, value := range replacements {
		result = append(result, key+"="+value)
	}
	return result
}

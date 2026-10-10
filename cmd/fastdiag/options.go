package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

const supportedProfile = "p93-q55"
const logN16Profile = "logn16-q55"
const publicE32Profile = "logn13-e32-public"

type options struct {
	mode           string
	profile        string
	traceScopes    []string
	baseline       string
	candidate      string
	output         string
	e32Manifest    string
	fastVectors    string
	warmup         int
	repetitions    int
	standardTrials int
}

func parseArgs(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, errors.New("使用方式：fastdiag trace|compare|numerical [options]，或 fastdiag offline-e32 --raw ... --sidecar ... --manifest ... --vectors ...")
	}
	opts := options{mode: args[0], profile: supportedProfile, warmup: 1, repetitions: 5, standardTrials: 3}
	if opts.mode != "trace" && opts.mode != "compare" && opts.mode != "numerical" {
		return options{}, fmt.Errorf("未知子命令 %q；請使用 trace、compare 或 numerical", opts.mode)
	}
	flags := flag.NewFlagSet("fastdiag "+opts.mode, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	profile := flags.String("profile", supportedProfile, "固定 diagnostic workload profile")
	trace := flags.String("trace", "", "逗號分隔的 trace scopes：stage,power,rescale 或 all")
	baseline := flags.String("baseline", "", "compare 的 Secondary baseline ref")
	candidate := flags.String("candidate", "", "compare 的 Secondary candidate ref")
	output := flags.String("out", "", "輸出 JSON 路徑（summary 使用相同 basename）")
	e32Manifest := flags.String("e32-manifest", "", "public E32 trace fixture manifest (logn13-e32-public only)")
	fastVectors := flags.String("fast-vectors", "", "matching uninstrumented Fast repeatability vectors (logn13-e32-public only)")
	warmup := flags.Int("warmup", 1, "trace warmup 次數")
	repetitions := flags.Int("repetitions", 5, "trace measured repetitions")
	standardTrials := flags.Int("standard-trials", 3, "numerical mode 的獨立 Standard key/evaluation-key trials（3–10）")
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("不預期的位置參數：%s", strings.Join(flags.Args(), " "))
	}
	if *profile != supportedProfile && *profile != logN16Profile && *profile != publicE32Profile {
		return options{}, fmt.Errorf("不支援 profile %q；允許 %s、%s 或 %s", *profile, supportedProfile, logN16Profile, publicE32Profile)
	}
	if *warmup < 0 || *warmup > 100 {
		return options{}, errors.New("--warmup 必須介於 0 與 100")
	}
	if *repetitions < 1 || *repetitions > 10 {
		return options{}, errors.New("--repetitions 必須介於 1 與 10，以限制 raw trace 大小")
	}
	standardTrialsSet := false
	warmupSet, repetitionsSet := false, false
	flags.Visit(func(flag *flag.Flag) {
		switch flag.Name {
		case "standard-trials":
			standardTrialsSet = true
		case "warmup":
			warmupSet = true
		case "repetitions":
			repetitionsSet = true
		}
	})
	var scopes []string
	if opts.mode == "numerical" {
		if *standardTrials < 3 || *standardTrials > 10 {
			return options{}, errors.New("--standard-trials 必須介於 3 與 10")
		}
		if *trace != "" || *baseline != "" || *candidate != "" {
			return options{}, errors.New("numerical 子命令不接受 --trace、--baseline 或 --candidate")
		}
	} else {
		if standardTrialsSet {
			return options{}, errors.New("--standard-trials 僅供 numerical 子命令使用")
		}
		parsedScopes, parseErr := parseScopes(*trace)
		if parseErr != nil {
			return options{}, parseErr
		}
		scopes = parsedScopes
		if len(scopes) == 0 {
			return options{}, errors.New("需要 --trace scope：stage、power、rescale 或 all")
		}
		if opts.mode == "trace" {
			if *baseline != "" || *candidate != "" {
				return options{}, errors.New("trace 子命令不接受 --baseline 或 --candidate")
			}
		} else if *baseline == "" || *candidate == "" {
			return options{}, errors.New("compare 子命令需要 --baseline 與 --candidate")
		}
	}
	if *profile == publicE32Profile {
		if opts.mode != "trace" {
			return options{}, errors.New("logn13-e32-public 僅供 trace 子命令使用")
		}
		if *e32Manifest == "" || *fastVectors == "" {
			return options{}, errors.New("logn13-e32-public trace 需要 --e32-manifest 與 --fast-vectors")
		}
		if strings.Join(scopes, ",") != "stage,power,rescale" {
			return options{}, errors.New("logn13-e32-public trace 固定使用 --trace stage,power,rescale")
		}
		if warmupSet && *warmup != 0 || repetitionsSet && *repetitions != 1 {
			return options{}, errors.New("logn13-e32-public 固定執行一次 cold 與一次 traced warm call，不接受 warmup/repetitions 變更")
		}
	} else if *e32Manifest != "" || *fastVectors != "" {
		return options{}, errors.New("--e32-manifest 與 --fast-vectors 僅供 logn13-e32-public trace 使用")
	}
	opts.profile, opts.traceScopes, opts.baseline, opts.candidate = *profile, scopes, *baseline, *candidate
	opts.output, opts.warmup, opts.repetitions = *output, *warmup, *repetitions
	opts.e32Manifest, opts.fastVectors = *e32Manifest, *fastVectors
	if opts.profile == publicE32Profile {
		opts.warmup, opts.repetitions = 0, 1
	}
	opts.standardTrials = *standardTrials
	return opts, nil
}

func parseScopes(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	selected := make(map[string]bool)
	for _, raw := range strings.Split(value, ",") {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if name == "all" {
			selected["stage"], selected["power"], selected["rescale"] = true, true, true
			continue
		}
		switch name {
		case "stage", "power", "rescale":
			selected[name] = true
		default:
			return nil, fmt.Errorf("未知 trace scope %q；允許 stage、power、rescale 或 all", name)
		}
	}
	var scopes []string
	for _, name := range []string{"stage", "power", "rescale"} {
		if selected[name] {
			scopes = append(scopes, name)
		}
	}
	if len(scopes) == 0 {
		return nil, errors.New("至少選擇一個 trace scope")
	}
	return scopes, nil
}

func (opts options) traceCSV() string { return strings.Join(opts.traceScopes, ",") }

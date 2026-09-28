package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

const supportedProfile = "p93-q55"

type options struct {
	mode        string
	profile     string
	traceScopes []string
	baseline    string
	candidate   string
	output      string
	warmup      int
	repetitions int
}

func parseArgs(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, errors.New("使用方式：fastdiag trace|compare [options]")
	}
	opts := options{mode: args[0], profile: supportedProfile, warmup: 1, repetitions: 5}
	if opts.mode != "trace" && opts.mode != "compare" {
		return options{}, fmt.Errorf("未知子命令 %q；請使用 trace 或 compare", opts.mode)
	}
	flags := flag.NewFlagSet("fastdiag "+opts.mode, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	profile := flags.String("profile", supportedProfile, "固定 diagnostic workload profile")
	trace := flags.String("trace", "", "逗號分隔的 trace scopes：stage,power,rescale 或 all")
	baseline := flags.String("baseline", "", "compare 的 Secondary baseline ref")
	candidate := flags.String("candidate", "", "compare 的 Secondary candidate ref")
	output := flags.String("out", "", "輸出 JSON 路徑（summary 使用相同 basename）")
	warmup := flags.Int("warmup", 1, "trace warmup 次數")
	repetitions := flags.Int("repetitions", 5, "trace measured repetitions")
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("不預期的位置參數：%s", strings.Join(flags.Args(), " "))
	}
	if *profile != supportedProfile {
		return options{}, fmt.Errorf("不支援 profile %q；目前只支援 %s", *profile, supportedProfile)
	}
	if *warmup < 0 || *warmup > 100 {
		return options{}, errors.New("--warmup 必須介於 0 與 100")
	}
	if *repetitions < 1 || *repetitions > 10 {
		return options{}, errors.New("--repetitions 必須介於 1 與 10，以限制 raw trace 大小")
	}
	scopes, err := parseScopes(*trace)
	if err != nil {
		return options{}, err
	}
	if len(scopes) == 0 {
		return options{}, errors.New("需要 --trace scope：stage、power、rescale 或 all")
	}
	if opts.mode == "trace" {
		if *baseline != "" || *candidate != "" {
			return options{}, errors.New("trace 子命令不接受 --baseline 或 --candidate")
		}
	} else {
		if *baseline == "" || *candidate == "" {
			return options{}, errors.New("compare 子命令需要 --baseline 與 --candidate")
		}
	}
	opts.profile, opts.traceScopes, opts.baseline, opts.candidate = *profile, scopes, *baseline, *candidate
	opts.output, opts.warmup, opts.repetitions = *output, *warmup, *repetitions
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

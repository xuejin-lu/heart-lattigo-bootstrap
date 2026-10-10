package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseArgsAndScopeValidation(t *testing.T) {
	opts, err := parseArgs([]string{"trace", "--trace", "stage, power", "--profile", "p93-q55", "--warmup", "0", "--repetitions", "3"})
	require.NoError(t, err)
	require.Equal(t, []string{"stage", "power"}, opts.traceScopes)
	require.Equal(t, 0, opts.warmup)
	require.Equal(t, 3, opts.repetitions)

	_, err = parseArgs([]string{"trace", "--trace", "stage,typo"})
	require.ErrorContains(t, err, "未知 trace scope")
	_, err = parseArgs([]string{"trace", "--trace", "stage", "--profile", "q0-56"})
	require.ErrorContains(t, err, "不支援 profile")
	_, err = parseArgs([]string{"compare", "--trace", "all", "--baseline", "a"})
	require.ErrorContains(t, err, "需要 --baseline 與 --candidate")
	_, err = parseArgs([]string{"trace", "--trace", "all", "--repetitions", "0"})
	require.ErrorContains(t, err, "--repetitions")
	_, err = parseArgs([]string{"trace", "--trace", "stage", "--repetitions", "11"})
	require.ErrorContains(t, err, "raw trace 大小")
	_, err = parseArgs([]string{"compare", "--baseline", "base", "--candidate", "head"})
	require.ErrorContains(t, err, "需要 --trace scope")

	opts, err = parseArgs([]string{"compare", "--trace", "all", "--baseline", "base", "--candidate", "head"})
	require.NoError(t, err)
	require.Equal(t, []string{"stage", "power", "rescale"}, opts.traceScopes)

	opts, err = parseArgs([]string{"numerical", "--profile", "p93-q55", "--standard-trials", "3"})
	require.NoError(t, err)
	require.Equal(t, 3, opts.standardTrials)
	require.Empty(t, opts.traceScopes)
	opts, err = parseArgs([]string{"numerical", "--profile", logN16Profile, "--standard-trials", "3"})
	require.NoError(t, err)
	require.Equal(t, logN16Profile, opts.profile)
	opts, err = parseArgs([]string{"trace", "--profile", publicE32Profile, "--trace", "stage,power,rescale", "--e32-manifest", "fixture.json", "--fast-vectors", "fast-vectors.json"})
	require.NoError(t, err)
	require.Equal(t, 0, opts.warmup)
	require.Equal(t, 1, opts.repetitions)
	_, err = parseArgs([]string{"trace", "--profile", publicE32Profile, "--trace", "stage", "--e32-manifest", "fixture.json", "--fast-vectors", "fast-vectors.json"})
	require.ErrorContains(t, err, "固定使用")
	_, err = parseArgs([]string{"compare", "--profile", publicE32Profile, "--trace", "all", "--baseline", "a", "--candidate", "b", "--e32-manifest", "fixture.json", "--fast-vectors", "fast-vectors.json"})
	require.ErrorContains(t, err, "僅供 trace")
	_, err = parseArgs([]string{"trace", "--profile", publicE32Profile, "--trace", "all", "--e32-manifest", "fixture.json", "--fast-vectors", "fast-vectors.json", "--warmup", "1"})
	require.ErrorContains(t, err, "固定執行")
	_, err = parseArgs([]string{"numerical", "--standard-trials", "2"})
	require.ErrorContains(t, err, "介於 3 與 10")
	_, err = parseArgs([]string{"trace", "--trace", "stage", "--standard-trials", "3"})
	require.ErrorContains(t, err, "僅供 numerical")
}

func TestParseScopesRejectsEmptyAndAcceptsAll(t *testing.T) {
	_, err := parseScopes(" , ")
	require.ErrorContains(t, err, "至少選擇")
	scopes, err := parseScopes("all")
	require.NoError(t, err)
	require.Equal(t, "stage,power,rescale", strings.Join(scopes, ","))
}

//go:build !perf_standard

package main

import "errors"

func verifyAZeroBootstrapKeys(backend backendAdapter) error {
	return errors.New("Standard A=0 key audit requires a perf_standard build")
}

//go:build !fix001t3diag

package main

import "fmt"

func runFIX001T3(BootstrapConfig, string, string, string) error {
	return fmt.Errorf("FIX-001-DIAG-T3 runner requires -tags fix001t3diag")
}

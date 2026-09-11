//go:build !fix001diag

package main

import "fmt"

func runFIX001DiagPoly(BootstrapConfig, string, string, string) error {
	return fmt.Errorf("FIX-001-DIAG-POLY runner requires -tags fix001diag")
}

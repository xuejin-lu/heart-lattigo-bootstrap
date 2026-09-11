//go:build !fix001t3capacity

package main

import "fmt"

func runFIX001T3Capacity(BootstrapConfig, string, string, string) error {
	return fmt.Errorf("FIX-001-DIAG-T3-CAPACITY runner requires -tags fix001t3capacity")
}

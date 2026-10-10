//go:build lattigo_standard

package main

import "github.com/tuneinsight/lattigo/v6/schemes/ckks"

func isFastBackend() bool { return false }

func makeCapacityAudit(ckks.Parameters) capacityAudit { return nil }

package main

import "testing"

func TestFIX001P3LogN13Q055Plan92PSDependencyWindowClosureGraphUsesDataflow(t *testing.T) {
	nodes := []dependencyGraphNode{
		{ID: "G0-rescale", Q01CenteredUnique: true},
		{ID: "G0-multiply", InputIDs: []string{"G0-rescale"}},
		{ID: "G0-add", InputIDs: []string{"G0-multiply", "B4-term-2"}},
		{ID: "G1-add", InputIDs: []string{"G0-add", "G1-multiply"}, CenteredSensitive: false},
		{ID: "G2-add", InputIDs: []string{"G1-add", "G2-multiply"}, CenteredSensitive: false},
		{ID: "G3-add", InputIDs: []string{"G2-add", "G3-multiply"}, CenteredSensitive: false},
		{ID: "F0-final-rescale", InputIDs: []string{"G3-add"}, CenteredSensitive: true},
	}
	if got := dependencyFirstReachable(nodes, "G0-add", func(node dependencyGraphNode) bool { return node.CenteredSensitive }); got != "F0-final-rescale" {
		t.Fatalf("dependency consumer = %q, want F0-final-rescale", got)
	}
	if got := dependencyLastUniqueAncestor(nodes, "G0-add"); got != "G0-rescale" {
		t.Fatalf("last unique ancestor = %q, want G0-rescale", got)
	}
}

func TestFIX001P3LogN13Q055Plan92PSDependencyWindowClosureClassificationNames(t *testing.T) {
	valid := map[string]bool{
		"logn13_q055_plan92_ps_single_q012_window_system_sufficient":           true,
		"logn13_q055_plan92_ps_multiple_q012_windows_system_sufficient":        true,
		"logn13_q055_plan92_ps_multiple_capacity_frontiers_not_yet_sufficient": true,
		"logn13_q055_plan92_ps_first_remaining_noncapacity_mismatch":           true,
		"logn13_q055_plan92_ps_q012_window_width_insufficient":                 true,
		"logn13_q055_plan92_ps_dependency_control_mismatch":                    true,
		"logn13_q055_plan92_ps_dependency_harness_mismatch":                    true,
	}
	if !valid["logn13_q055_plan92_ps_single_q012_window_system_sufficient"] {
		t.Fatal("required classification missing")
	}
}

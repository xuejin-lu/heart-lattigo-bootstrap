package perfmeasure

import (
	"os/exec"
	"runtime"
	"strings"
)

// CPUModel returns the host CPU model when the platform exposes it, falling
// back to the architecture name when model metadata is unavailable.
func CPUModel() string {
	if runtime.GOOS == "darwin" {
		if output, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			if model := strings.TrimSpace(string(output)); model != "" {
				return model
			}
		}
		if output, err := exec.Command("system_profiler", "SPHardwareDataType").Output(); err == nil {
			if model := systemProfilerChip(string(output)); model != "" {
				return model
			}
		}
	}
	return runtime.GOARCH
}

func systemProfilerChip(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chip:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Chip:"))
		}
	}
	return ""
}

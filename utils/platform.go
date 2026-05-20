package utils

import (
	"os"
	"runtime"
	"strings"
)

// IsWSL reports whether the process is running inside Windows Subsystem for Linux.
// Checks WSL_DISTRO_NAME env var first (always set by WSL), then falls back to
// reading /proc/version for "microsoft".
func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

// Platform returns the effective platform: "windows", "darwin", "wsl", or "linux".
func Platform() string {
	if IsWSL() {
		return "wsl"
	}
	return runtime.GOOS
}

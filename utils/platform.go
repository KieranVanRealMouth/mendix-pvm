package utils

import (
	"os"
	"runtime"
	"strings"
)

// IsWSL reports whether the process is running inside Windows Subsystem for Linux.
// /proc/version contains "microsoft" on WSL kernels.
func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
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

package steelcut

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const PageSize = 4096 // size of each memory page in bytes

type MacOSHost struct {
	*UnixHost
	PackageManager PackageManager
}

// Reboot restarts the macOS host.
func (h MacOSHost) Reboot() error {
	_, err := h.RunCommand("sudo reboot")
	return err
}

// Shutdown turns off the macOS host.
func (h MacOSHost) Shutdown() error {
	_, err := h.RunCommand("sudo shutdown -h now")
	return err
}

// Pass through the package management operations to the PackageManager.
func (h MacOSHost) ListPackages() ([]string, error) {
	return h.PackageManager.ListPackages(h.UnixHost)
}

func (h MacOSHost) AddPackage(pkg string) error {
	return h.PackageManager.AddPackage(h.UnixHost, pkg)
}

func (h MacOSHost) RemovePackage(pkg string) error {
	return h.PackageManager.RemovePackage(h.UnixHost, pkg)
}

func (h MacOSHost) UpgradePackage(pkg string) error {
	return h.PackageManager.UpgradePackage(h.UnixHost, pkg)
}

func (h MacOSHost) UpgradeAllPackages() ([]Update, error) {
	return h.PackageManager.UpgradeAll(h.UnixHost)
}

func (h MacOSHost) CheckUpdates() ([]Update, error) {
	// Implement the update check for macOS hosts.
	return []Update{}, nil
}

func (h MacOSHost) RunCommand(cmd string) (string, error) {
	return h.UnixHost.RunCommand(cmd)
}

// CPUUsage retrieves the CPU usage for the macOS host.
var idleRegex = regexp.MustCompile(`(\d+\.\d+)% idle`)

func (h MacOSHost) CPUUsage() (float64, error) {
	output, err := h.RunCommand("top -l 1 | grep 'CPU usage'")
	if err != nil {
		return 0, err
	}

	matches := idleRegex.FindStringSubmatch(output)
	if len(matches) != 2 {
		return 0, fmt.Errorf("unexpected CPU usage output: %s", output)
	}

	idle, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing idle percentage: %v", err)
	}

	return 100 - idle, nil
}

// DiskUsage retrieves the disk usage for the macOS host.
func (h MacOSHost) DiskUsage() (float64, error) {
	output, err := h.RunCommand("df / | tail -1 | awk '{print $5}'")
	if err != nil {
		return 0, err
	}

	output = strings.TrimSuffix(output, "%")
	usage, err := strconv.ParseFloat(output, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing disk usage: %v", err)
	}

	return usage, nil
}

// RunningProcesses retrieves a list of running processes on the macOS host.
func (h MacOSHost) RunningProcesses() ([]string, error) {
	output, err := h.RunCommand("ps -axco command")
	if err != nil {
		return nil, err
	}

	return strings.Split(output, "\n"), nil
}

// MemoryUsage retrieves the active memory usage for the macOS host.
var activeMemoryRegex = regexp.MustCompile(`Pages active:\s+(\d+)`)

func (h MacOSHost) MemoryUsage() (float64, error) {
	output, err := h.RunCommand("vm_stat | grep 'Pages active'")
	if err != nil {
		return 0, err
	}

	matches := activeMemoryRegex.FindStringSubmatch(output)
	if len(matches) != 2 {
		return 0, fmt.Errorf("could not parse memory info: %s", output)
	}

	activeMemory, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("error parsing active memory pages: %v", err)
	}

	return float64(activeMemory*PageSize) / 1024 / 1024, nil
}

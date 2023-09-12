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
	commandOptions := CommandOptions{UseSudo: true}
	_, err := h.RunCommand("reboot", commandOptions)
	return err
}

// Shutdown turns off the macOS host.
func (h MacOSHost) Shutdown() error {
	commandOptions := CommandOptions{UseSudo: true}
	_, err := h.RunCommand("shutdown -h now", commandOptions)
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

func (h MacOSHost) RunCommand(cmd string, commandOptions CommandOptions) (string, error) {
	return h.UnixHost.RunCommand(cmd, commandOptions)
}

// CPUUsage retrieves the CPU usage for the macOS host.
var idleRegex = regexp.MustCompile(`(\d+\.\d+)% idle`)

func (h MacOSHost) CPUUsage() (float64, error) {
	commandOptions := CommandOptions{UseSudo: false}
	output, err := h.RunCommand("top -l 1 | grep 'CPU usage'", commandOptions)
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
	commandOptions := CommandOptions{UseSudo: false}
	output, err := h.RunCommand("df / | tail -1 | awk '{print $5}'", commandOptions)
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
	commandOptions := CommandOptions{UseSudo: false}
	output, err := h.RunCommand("ps -axco command", commandOptions)
	if err != nil {
		return nil, err
	}

	return strings.Split(output, "\n"), nil
}

// MemoryUsage retrieves the active memory usage for the macOS host.
var activeMemoryRegex = regexp.MustCompile(`Pages active:\s+(\d+)`)

func (h MacOSHost) MemoryUsage() (float64, error) {
	commandOptions := CommandOptions{UseSudo: false}
	output, err := h.RunCommand("vm_stat | grep 'Pages active'", commandOptions)
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

func (h *MacOSHost) Info() (HostInfo, error) {
	// Fetch CPU Usage
	cpuUsage, err := h.CPUUsage()
	if err != nil {
		return HostInfo{}, err
	}

	// Fetch Memory Usage
	memoryUsage, err := h.MemoryUsage()
	if err != nil {
		return HostInfo{}, err
	}

	// Fetch Disk Usage
	diskUsage, err := h.DiskUsage()
	if err != nil {
		return HostInfo{}, err
	}

	// Fetch Running Processes
	runningProcesses, err := h.RunningProcesses()
	if err != nil {
		return HostInfo{}, err
	}

	// Construct and return the HostInfo struct
	return HostInfo{
		CPUUsage:         cpuUsage,
		DiskUsage:        diskUsage,
		MemoryUsage:      memoryUsage,
		RunningProcesses: runningProcesses,
	}, nil
}

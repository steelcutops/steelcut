package steelcut

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type MacOSHost struct {
	*UnixHost
	PackageManager PackageManager
}

func (h MacOSHost) Reboot() error {
	_, err := h.RunCommand("sudo reboot")
	return err
}

func (h MacOSHost) Shutdown() error {
	_, err := h.RunCommand("sudo shutdown -h now")
	return err
}

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

func (h MacOSHost) CPUUsage() (float64, error) {
	output, err := h.RunCommand("top -l 1 | grep 'CPU usage'")
	if err != nil {
		return 0, err
	}

	// The output should look like: CPU usage: 7.78% user, 12.22% sys, 80.0% idle
	// Split the output by ',' to get the idle percentage.
	parts := strings.Split(output, ",")
	if len(parts) < 3 {
		return 0, fmt.Errorf("unexpected output from top command: %s", output)
	}

	// The idle percentage is the third part. Trim the spaces and '%' and parse to float.
	idleStr := strings.TrimSpace(parts[2])
	idleStr = strings.TrimSuffix(idleStr, "% idle")

	idle, err := strconv.ParseFloat(idleStr, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing idle percentage: %v", err)
	}

	// The CPU usage is 100% minus the idle percentage.
	cpuUsage := 100 - idle

	return cpuUsage, nil
}

func (h MacOSHost) DiskUsage() (float64, error) {
	output, err := h.RunCommand("df / | tail -1 | awk '{print $5}'")
	if err != nil {
		return 0, err
	}

	// The output is the percentage of disk space used.
	// You'll need to remove the trailing % and convert to a float64.
	output = strings.TrimSuffix(output, "%")
	usage, err := strconv.ParseFloat(output, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing disk usage: %v", err)
	}

	return usage, nil
}

func (h MacOSHost) RunningProcesses() ([]string, error) {
	output, err := h.RunCommand("ps -axco command")
	if err != nil {
		return nil, err
	}

	// The output is one line per running process.
	// You can split it into a slice.
	processes := strings.Split(output, "\n")
	return processes, nil
}

func (h MacOSHost) MemoryUsage() (float64, error) {
	output, err := h.RunCommand("vm_stat | grep 'Pages active'")
	if err != nil {
		return 0, err
	}

	// output looks like "Pages active:                        837161.".
	// So we need to extract the number.
	re := regexp.MustCompile("[0-9]+")
	match := re.FindString(output)
	if match == "" {
		return 0, fmt.Errorf("could not parse memory info: %s", output)
	}

	activeMemory, err := strconv.Atoi(match)
	if err != nil {
		return 0, fmt.Errorf("error parsing active memory pages: %v", err)
	}

	// Each page is 4096 bytes.
	// The total memory is the number of active pages * the size of each page.
	totalMemory := float64(activeMemory) * 4096
	return totalMemory / 1024 / 1024, nil
}

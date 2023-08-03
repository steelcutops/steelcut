package steelcut

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type LinuxHost struct {
	*UnixHost
	PackageManager PackageManager
}

func (h LinuxHost) ListPackages() ([]string, error) {
	return h.PackageManager.ListPackages(h.UnixHost)
}

func (h LinuxHost) AddPackage(pkg string) error {
	return h.PackageManager.AddPackage(h.UnixHost, pkg)
}

func (h LinuxHost) RemovePackage(pkg string) error {
	return h.PackageManager.RemovePackage(h.UnixHost, pkg)
}

func (h LinuxHost) UpgradePackage(pkg string) error {
	return h.PackageManager.UpgradePackage(h.UnixHost, pkg)
}

func (h LinuxHost) CheckUpdates() ([]Update, error) {
	// Implement the update check for Linux hosts.
	return []Update{}, nil
}

func (h LinuxHost) RunCommand(cmd string) (string, error) {
	return h.UnixHost.RunCommand(cmd)
}

func (h LinuxHost) CPUUsage() (float64, error) {
	output, err := h.RunCommand("cat /proc/stat")
	if err != nil {
		return 0, err
	}

	// Find the line that starts with 'cpu'
	var cpuLine string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			cpuLine = line
			break
		}
	}

	if cpuLine == "" {
		return 0, fmt.Errorf("could not find cpu info in /proc/stat")
	}

	// Parse the line to calculate CPU usage.
	// The format is: cpu  user nice system idle iowait irq softirq steal guest guest_nice
	fields := strings.Fields(cpuLine)
	if len(fields) < 10 {
		return 0, fmt.Errorf("unexpected format of /proc/stat")
	}

	var total float64
	for _, field := range fields[1:] { // Skip the "cpu" prefix
		value, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return 0, fmt.Errorf("error parsing /proc/stat: %v", err)
		}
		total += value
	}

	// CPU usage is the proportion of total time not spent idle
	idle, err := strconv.ParseFloat(fields[4], 64) // "idle" is the 5th field
	if err != nil {
		return 0, fmt.Errorf("error parsing /proc/stat: %v", err)
	}
	usage := 100 * (1 - idle/total)

	return usage, nil
}

func (h LinuxHost) MemoryUsage() (float64, error) {
	output, err := h.RunCommand("cat /proc/meminfo")
	if err != nil {
		return 0, err
	}

	// Parse the output to calculate memory usage.
	// The relevant lines are MemTotal: and MemAvailable:.
	lines := strings.Split(output, "\n")
	var memTotal, memAvailable float64
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			memTotal, err = strconv.ParseFloat(fields[1], 64)
			if err != nil {
				return 0, fmt.Errorf("error parsing MemTotal: %v", err)
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			memAvailable, err = strconv.ParseFloat(fields[1], 64)
			if err != nil {
				return 0, fmt.Errorf("error parsing MemAvailable: %v", err)
			}
		}
	}

	memUsage := 100 * (1 - memAvailable/memTotal)

	return memUsage, nil
}

func (h LinuxHost) DiskUsage() (float64, error) {
	output, err := h.RunCommand("df --output=pcent /")
	if err != nil {
		return 0, err
	}

	// Get the last line of the output
	lines := strings.Split(strings.TrimSpace(output), "\n")
	lastLine := lines[len(lines)-1]

	// The last line is the percentage of disk space used
	lastLine = strings.TrimSpace(lastLine)      // Remove leading/trailing whitespace
	lastLine = strings.TrimRight(lastLine, "%") // Remove trailing %
	usage, err := strconv.ParseFloat(lastLine, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing disk usage: %v", err)
	}

	return usage, nil
}

func (h LinuxHost) RunningProcesses() ([]string, error) {
	output, err := h.RunCommand("ps aux")
	if err != nil {
		return nil, err
	}

	// The output is one line per running process.
	lines := strings.Split(output, "\n")
	return lines[1:], nil // Skip the header line
}

func (h LinuxHost) Reboot() error {
	_, err := h.RunCommand("sudo reboot")
	return err
}

func (h LinuxHost) Shutdown() error {
	_, err := h.RunCommand("sudo shutdown -h now")
	return err
}

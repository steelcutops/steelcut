package steelcut

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const PageSize = 4096 // size of each memory page in bytes

type MacOSHost struct {
	Options        *HostOptions
	PackageManager PackageManager
}

func (h *MacOSHost) Hostname() string {
	return h.Options.HostString
}

func (h *UnixHost) CheckUpdates() ([]Update, error) {
	return []Update{}, nil
}

func (h UnixHost) IsReachable() error {
	if h.isLocal() {
		return nil
	}

	if err := h.ping(); err != nil {
		return err
	}
	return h.sshable()
}

func (h UnixHost) ping() error {
	cmd := "ping -c 1 " + h.Hostname()
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Errorf("ping test failed: %v", err)
	}
	log.Printf("Ping test passed for host '%s'\n", h.Hostname())
	return nil
}

func (h UnixHost) sshable() error {
	if h.isLocal() {
		return nil
	}

	config, err := h.getSSHConfig()
	if err != nil {
		return err
	}

	timeout := 60 * time.Second
	client, err := h.SSHClient.Dial("tcp", h.Hostname()+":22", config, timeout)
	if err != nil {
		return fmt.Errorf("SSH test failed: %v", err)
	}
	client.Close()
	log.Printf("SSH test passed for host '%s'\n", h.Hostname())
	return nil
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

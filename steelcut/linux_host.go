package steelcut

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// LinuxHost represents a Linux host with package management capabilities.
type LinuxHost struct {
	*UnixHost
	PackageManager PackageManager
}

// ListPackages retrieves a list of installed packages on the Linux host.
func (h LinuxHost) ListPackages() ([]string, error) {
	return h.PackageManager.ListPackages(h.UnixHost)
}

// AddPackage installs the given package on the Linux host.
func (h LinuxHost) AddPackage(pkg string) error {
	return h.PackageManager.AddPackage(h.UnixHost, pkg)
}

// RemovePackage uninstalls the given package from the Linux host.
func (h LinuxHost) RemovePackage(pkg string) error {
	return h.PackageManager.RemovePackage(h.UnixHost, pkg)
}

// UpgradePackage upgrades the given package to the latest version on the Linux host.
func (h LinuxHost) UpgradePackage(pkg string) error {
	return h.PackageManager.UpgradePackage(h.UnixHost, pkg)
}

// UpgradeAllPackages upgrades all installed packages to their latest versions on the Linux host.
func (h LinuxHost) UpgradeAllPackages() ([]Update, error) {
	return h.PackageManager.UpgradeAll(h.UnixHost)
}

// CheckUpdates checks for available updates for the OS and returns them as a slice of Update objects.
func (h LinuxHost) CheckUpdates() ([]Update, error) {
	log.Printf("Checking for OS updates on %s", h.Hostname())
	updates, err := h.PackageManager.CheckOSUpdates(h.UnixHost)
	if err != nil {
		log.Printf("Error checking OS updates: %v", err)
		return nil, err
	}

	// Create a slice to hold the parsed updates
	var parsedUpdates []Update

	for _, update := range updates {
		// Assuming the update string contains the package name and version, separated by a space
		parts := strings.SplitN(update, " ", 2)
		if len(parts) == 2 {
			packageName := parts[0]
			version := parts[1]
			parsedUpdates = append(parsedUpdates, Update{
				PackageName: packageName,
				Version:     version,
			})
		}
	}

	return parsedUpdates, nil
}

// RunCommand runs the given command on the Linux host and returns the output.
func (h LinuxHost) RunCommand(cmd string, commandOptions CommandOptions) (string, error) {
	return h.UnixHost.RunCommand(cmd, commandOptions)
}

// CPUUsage calculates the CPU usage as a percentage on the Linux host.
func (h LinuxHost) CPUUsage() (float64, error) {
	commandOptions := CommandOptions{UseSudo: false}
	output, err := h.RunCommand("cat /proc/stat", commandOptions)
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

// MemoryUsage calculates the memory usage as a percentage on the Linux host.
func (h LinuxHost) MemoryUsage() (float64, error) {
	commandOptions := CommandOptions{UseSudo: false}
	output, err := h.RunCommand("cat /proc/meminfo", commandOptions)
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

// DiskUsage calculates the disk usage as a percentage for the root directory on the Linux host.
func (h LinuxHost) DiskUsage() (float64, error) {
	commandOptions := CommandOptions{UseSudo: false}
	output, err := h.RunCommand("df --output=pcent /", commandOptions)
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

// RunningProcesses retrieves a list of running processes on the Linux host.
func (h LinuxHost) RunningProcesses() ([]string, error) {
	commandOptions := CommandOptions{UseSudo: false}
	output, err := h.RunCommand("ps aux", commandOptions)
	if err != nil {
		return nil, err
	}

	// The output is one line per running process.
	lines := strings.Split(output, "\n")
	return lines[1:], nil // Skip the header line
}

// Reboot restarts the Linux host.
func (h LinuxHost) Reboot() error {
	commandOptions := CommandOptions{UseSudo: true}
	_, err := h.RunCommand("reboot", commandOptions)
	return err
}

// Shutdown powers off the Linux host.
func (h LinuxHost) Shutdown() error {
	commandOptions := CommandOptions{UseSudo: true}
	_, err := h.RunCommand("shutdown -h now", commandOptions)
	return err
}

// UTMP represents an entry in the wtmp file
type UTMP struct {
	Type int16     // Type of login
	_    [2]byte   // Alignment padding
	PID  int32     // Process ID
	Line [32]byte  // Device name
	ID   [4]byte   // Terminal ID
	User [32]byte  // Username
	Host [256]byte // Hostname
	_    [64]byte  // Additional fields
}

// ListUserSessions retrieves a list of user sessions from the wtmp log file.
func (h LinuxHost) ListUserSessions() ([]string, error) {
	file, err := os.Open("/var/log/wtmp")
	if err != nil {
		return nil, fmt.Errorf("error opening wtmp file: %v", err)
	}
	defer file.Close()

	var userSessions []string
	for {
		var utmp UTMP
		err := binary.Read(file, binary.LittleEndian, &utmp)
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("error reading wtmp entry: %v", err)
			}
			break
		}

		session := fmt.Sprintf("Type: %d, PID: %d, Line: %s, User: %s, Host: %s",
			utmp.Type,
			utmp.PID,
			string(utmp.Line[:]),
			string(utmp.User[:]),
			string(utmp.Host[:]))
		userSessions = append(userSessions, session)
	}
	return userSessions, nil
}

// EnableService enables a systemd service on the Linux host.
func (h LinuxHost) EnableService(serviceName string) error {
	commandOptions := CommandOptions{UseSudo: true}
	_, err := h.RunCommand(fmt.Sprintf("systemctl enable %s", serviceName), commandOptions)
	return err
}

// StartService starts a systemd service on the Linux host.
func (h LinuxHost) StartService(serviceName string) error {
	commandOptions := CommandOptions{UseSudo: true}
	_, err := h.RunCommand(fmt.Sprintf("systemctl start %s", serviceName), commandOptions)
	return err
}

// StopService stops a systemd service on the Linux host.
func (h LinuxHost) StopService(serviceName string) error {
	commandOptions := CommandOptions{UseSudo: true}
	_, err := h.RunCommand(fmt.Sprintf("systemctl stop %s", serviceName), commandOptions)
	return err
}

// RestartService restarts a systemd service on the Linux host.
func (h LinuxHost) RestartService(serviceName string) error {
	commandOptions := CommandOptions{UseSudo: true}
	_, err := h.RunCommand(fmt.Sprintf("systemctl restart %s", serviceName), commandOptions)
	return err
}

// CheckServiceStatus checks the status of a systemd service on the Linux host.
func (h LinuxHost) CheckServiceStatus(serviceName string) (string, error) {
	commandOptions := CommandOptions{UseSudo: false}
	return h.RunCommand(fmt.Sprintf("systemctl status %s", serviceName), commandOptions)
}

func (h *LinuxHost) Info() (HostInfo, error) {
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

package steelcut

import (
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Update struct {
	// Fields for the Update struct
}
type SystemReporter interface {
	CPUUsage() (float64, error)
	MemoryUsage() (float64, error)
	DiskUsage() (float64, error)
	RunningProcesses() ([]string, error)
}

type Host interface {
	CheckUpdates() ([]Update, error)
	RunCommand(cmd string) (string, error)
	ListPackages() ([]string, error)
	AddPackage(pkg string) error
	RemovePackage(pkg string) error
	UpgradePackage(pkg string) error
	Reboot() error
	Shutdown() error
	SystemReporter
}

type UnixHost struct {
	Hostname      string
	User          string
	Password      string
	KeyPassphrase string
	OS            string
}

type MacOSHost struct {
	UnixHost
	PackageManager PackageManager
}

type LinuxHost struct {
	UnixHost
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
	output, err := h.RunCommand("cat /proc/stat | grep '^cpu '")
	if err != nil {
		return 0, err
	}

	// Parse the output to calculate CPU usage.
	// The format is: cpu  user nice system idle iowait irq softirq steal guest guest_nice
	fields := strings.Fields(output)
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
	output, err := h.RunCommand("df --output=pcent / | tail -n 1")
	if err != nil {
		return 0, err
	}

	// The output is the percentage of disk space used.
	output = strings.TrimSpace(output)      // Remove leading/trailing whitespace
	output = strings.TrimRight(output, "%") // Remove trailing %
	usage, err := strconv.ParseFloat(output, 64)
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

type HostOption func(*UnixHost)

func WithUser(user string) HostOption {
	return func(host *UnixHost) {
		host.User = user
	}
}

func WithPassword(password string) HostOption {
	return func(host *UnixHost) {
		host.Password = password
	}
}

func WithKeyPassphrase(keyPassphrase string) HostOption {
	return func(host *UnixHost) {
		host.KeyPassphrase = keyPassphrase
	}
}

func WithOS(os string) HostOption {
	return func(host *UnixHost) {
		host.OS = os
	}
}

func NewHost(hostname string, options ...HostOption) (Host, error) {
	host := &UnixHost{
		Hostname: hostname,
	}

	for _, option := range options {
		option(host)
	}

	// If the username has not been specified, use the current user's username.
	if host.User == "" {
		currentUser, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %v", err)
		}
		host.User = currentUser.Username
	}

	// If the OS has not been specified, determine it.
	if host.OS == "" {
		os, err := determineOS(host)
		if err != nil {
			return nil, err
		}
		host.OS = os
	}

	switch host.OS {
	case "Linux":
		// Determine the package manager.
		// Here we just guess based on the contents of /etc/os-release.
		osRelease, _ := host.RunCommand("cat /etc/os-release")
		if strings.Contains(osRelease, "ID=ubuntu") || strings.Contains(osRelease, "ID=debian") {
			return LinuxHost{*host, AptPackageManager{}}, nil
		} else {
			// Assume Red Hat/CentOS/Fedora if not Debian/Ubuntu.
			return LinuxHost{*host, YumPackageManager{}}, nil
		}
	case "Darwin":
		return MacOSHost{*host, BrewPackageManager{}}, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", host.OS)
	}

}

func determineOS(host *UnixHost) (string, error) {
	output, err := host.RunCommand("uname")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func (h UnixHost) RunCommand(cmd string) (string, error) {
	log.Printf("Running command '%s' on host '%s' with user '%s'\n", cmd, h.Hostname, h.User)
	// If the hostname is "localhost" or "127.0.0.1", run the command locally.
	if h.Hostname == "localhost" || h.Hostname == "127.0.0.1" {
		parts := strings.Fields(cmd)
		head := parts[0]
		parts = parts[1:]

		out, err := exec.Command(head, parts...).Output()
		if err != nil {
			return "", err
		}

		return string(out), nil
	}

	// Otherwise, run the command over SSH.
	var authMethod ssh.AuthMethod

	if h.Password != "" {
		log.Println("Using password authentication")
		authMethod = ssh.Password(h.Password)
	} else {
		log.Println("Using public key authentication")
		var keyManager SSHKeyManager
		if h.KeyPassphrase != "" {
			keyManager = FileSSHKeyManager{}
		} else {
			keyManager = AgentSSHKeyManager{}
		}

		keys, err := keyManager.ReadPrivateKeys(h.KeyPassphrase)
		if err != nil {
			return "", err
		}

		authMethod = ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			return keys, nil
		})
	}

	config := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", h.Hostname+":22", config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

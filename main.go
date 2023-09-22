package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/steelcutops/steelcut/steelcut/commandmanager"
	"github.com/steelcutops/steelcut/steelcut/filemanager"
	"github.com/steelcutops/steelcut/steelcut/host"
	"github.com/steelcutops/steelcut/steelcut/hostgroup"

	"golang.org/x/term"
	"gopkg.in/ini.v1"
)

var programLevel = new(slog.LevelVar)

type HostInfo struct {
	CPUUsage         float64                   `json:"cpuUsage"`
	DiskUsageDetails filemanager.DiskUsageInfo `json:"diskUsageDetails"`
	MemoryUsage      int64                     `json:"memoryUsage"`
	RunningProcesses []string                  `json:"runningProcesses"`
}

type flags struct {
	CheckHealth        bool
	CPUThreshold       float64
	Concurrency        int
	Debug              bool
	DiskThreshold      float64
	ExecCommand        string
	Hostnames          hostnamesValue
	InfoDump           bool
	IniFilePath        string
	KeyPassPrompt      bool
	ListPackages       bool
	ListUpgradable     bool
	LogFileName        string
	MemoryThreshold    int64
	Monitor            bool
	MonitorInterval    time.Duration
	PasswordPrompt     bool
	ScriptPath         string
	SudoPasswordPrompt bool
	UpgradePackages    bool
	Username           string
}

type hostnamesValue []string

func (h *hostnamesValue) String() string {
	return strings.Join(*h, ",")
}

func (h *hostnamesValue) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func readHostsFromFile(filePath string) (map[string][]string, error) {
	cfg, err := ini.Load(filePath)
	if err != nil {
		return nil, err
	}

	hosts := make(map[string][]string)

	for _, section := range cfg.Sections() {
		name := section.Name()
		for _, key := range section.Keys() {
			hosts[name] = append(hosts[name], key.String())
		}
	}

	return hosts, nil
}

func checkHostHealth(host *host.Host) error {
	// Ping the host
	result, err := host.NetworkManager.Ping(host.Hostname)
	if err != nil || !result.Success {
		return fmt.Errorf("host %s is not reachable: %v", host.Hostname, err)
	}

	slog.Info("Host %s is healthy with RTT: %f ms", host.Hostname, result.RTT)
	return nil
}

func parseFlags() *flags {
	f := &flags{}
	flag.BoolVar(&f.CheckHealth, "check-health", false, "Perform a basic health check on the host")
	flag.BoolVar(&f.Debug, "debug", false, "Enable debug log level")
	flag.BoolVar(&f.InfoDump, "info", false, "Dump information about the hosts")
	flag.BoolVar(&f.KeyPassPrompt, "keypass", false, "Passphrase for decrypting SSH keys")
	flag.BoolVar(&f.ListPackages, "list", false, "List all packages")
	flag.BoolVar(&f.ListUpgradable, "upgradable", false, "List all upgradable packages")
	flag.DurationVar(&f.MonitorInterval, "monitor-interval", 5*time.Second, "Interval between monitoring checks")
	flag.BoolVar(&f.Monitor, "monitor", false, "Enable host monitoring")
	flag.BoolVar(&f.PasswordPrompt, "password", false, "Use a password for SSH connection")
	flag.BoolVar(&f.SudoPasswordPrompt, "sudo-password", false, "Prompt for sudo password")
	flag.BoolVar(&f.UpgradePackages, "upgrade", false, "Upgrade all packages")
	flag.Float64Var(&f.CPUThreshold, "cpu-threshold", 80.0, "Threshold for CPU usage in percent")
	flag.Float64Var(&f.DiskThreshold, "disk-threshold", 80.0, "Threshold for disk usage in percent")
	flag.Int64Var(&f.MemoryThreshold, "memory-threshold", 80, "Threshold for memory usage in percent")
	flag.IntVar(&f.Concurrency, "concurrency", 10, "Maximum number of concurrent host connections")
	flag.StringVar(&f.ExecCommand, "exec", "", "Execute command on the host")
	flag.StringVar(&f.IniFilePath, "ini", "", "Path to INI file with host configurations")
	flag.StringVar(&f.LogFileName, "log", "slog.txt", "Log file name")
	flag.StringVar(&f.ScriptPath, "script", "", "Path to script file to be executed on the host")
	flag.StringVar(&f.Username, "username", "", "Username to use for SSH connection")
	flag.Var(&f.Hostnames, "hostname", "Hostname to connect to")

	flag.Parse()

	return f
}

func monitorHosts(hg *hostgroup.HostGroup, f *flags) {
	for {
		hg.RLock()
		for _, host := range hg.Hosts {
			hostLogger := slog.With("host", host.Hostname) // Setting up contextual logger
			hostLogger.Debug("Monitoring host")
			hostInfo, err := getHostInfo(host)
			if err != nil {
				hostLogger.Error("Failed to get host info", "error", err)
				continue
			}

			if hostInfo.CPUUsage > f.CPUThreshold {
				hostLogger.Info("CPU usage exceeded threshold", "Usage", hostInfo.CPUUsage, "Threshold", f.CPUThreshold)
			} else {
				hostLogger.Debug("CPU usage is within threshold", "Usage", hostInfo.CPUUsage, "Threshold", f.CPUThreshold)
			}

			if hostInfo.MemoryUsage > f.MemoryThreshold {
				hostLogger.Warn("Memory usage exceeded threshold", "Usage", hostInfo.MemoryUsage, "Threshold", f.MemoryThreshold)
			} else {
				hostLogger.Debug("Memory usage is within threshold", "Usage", hostInfo.MemoryUsage, "Threshold", f.MemoryThreshold)
			}

			if float64(hostInfo.DiskUsageDetails.Available) > f.DiskThreshold {
				hostLogger.Warn("Disk usage exceeded threshold", "Usage", hostInfo.DiskUsageDetails.Available, "Threshold", f.DiskThreshold)
			} else {
				hostLogger.Debug("Disk usage is within threshold", "Usage", hostInfo.DiskUsageDetails.Available, "Threshold", f.DiskThreshold)
			}
		}
		hg.RUnlock()

		time.Sleep(f.MonitorInterval) // Delay between monitoring checks
	}
}

func readScriptFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func executeScript(host *host.Host, script string) error {
	// Use the context with a reasonable timeout; you can adjust this as needed.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create the CommandConfig
	config := commandmanager.CommandConfig{
		Command: script,
		Sudo:    false,
	}

	// Use the CommandManager embedded in the host.Host struct
	result, err := host.CommandManager.Run(ctx, config)
	if err != nil {
		return err
	}

	// Print the output
	fmt.Printf("Output of script on host %s:\n%s\n", host.Hostname, result.STDOUT)

	// Check for errors in STDERR if necessary
	if result.STDERR != "" {
		fmt.Printf("Errors: %s\n", result.STDERR)
	}

	return nil
}

func processHosts(hg *hostgroup.HostGroup, action func(h *host.Host) error, maxConcurrency int) error {
	sem := make(chan struct{}, maxConcurrency) // Create semaphore with buffer size equal to max concurrency
	errCh := make(chan error, len(hg.Hosts))
	var wg sync.WaitGroup

	hg.RLock()
	for _, hst := range hg.Hosts {
		wg.Add(1) // Increment wait group counter
		go func(h *host.Host) {
			defer wg.Done()          // Decrement wait group counter when done
			sem <- struct{}{}        // Acquire token
			defer func() { <-sem }() // Release token

			if err := action(h); err != nil {
				errCh <- fmt.Errorf("error while processing host %s: %w", h.Hostname, err)
			}
		}(hst)
	}
	hg.RUnlock()

	wg.Wait()    // Wait for all goroutines to complete
	close(errCh) // Close error channel when done

	var result *multierror.Error // Initialize multierror
	for err := range errCh {
		// Append each error into multierror
		result = multierror.Append(result, err)
	}

	if result != nil {
		// Log all errors
		for _, err := range result.Errors {
			slog.Error("Host processing error", "error", err)
		}
		return result // Return the multierror
	}

	return nil
}

func dumpHostInfo(host *host.Host) error {
	hostInfo, err := getHostInfo(host)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(hostInfo, "", "  ")
	if err != nil {
		return err
	}

	// Use the hostname to create a unique file name for each host
	fileName := fmt.Sprintf("host_info_%s.json", host.Hostname)

	err = os.WriteFile(fileName, b, 0644)
	return err
}

func listAllPackages(host *host.Host) error {
	packages, err := host.PackageManager.ListPackages()
	if err != nil {
		return fmt.Errorf("failed to list packages: %v", err)
	}
	fmt.Println("Packages:")
	for _, pkg := range packages {
		fmt.Println(pkg)
	}
	return nil
}

func listUpgradablePackages(host *host.Host) error {
	upgradable, err := host.PackageManager.CheckOSUpdates()
	if err != nil {
		return fmt.Errorf("failed to check OS updates: %v", err)
	}
	fmt.Println("Upgradable packages:")
	for _, pkg := range upgradable {
		fmt.Println(pkg)
	}
	return nil
}

func upgradeAllPackages(host *host.Host) error {
	_, err := host.PackageManager.UpgradeAll()
	if err != nil {
		return fmt.Errorf("failed to upgrade packages: %v", err)
	}
	slog.Info("Upgraded packages on host", "host", host.Hostname)
	return nil
}

func addHosts(hostnames []string, hostGroup *hostgroup.HostGroup, options ...host.HostOption) {
	for _, hostname := range hostnames {
		slog.Debug("Adding host", "host", hostname)
		server, err := host.NewHost(hostname, options...)
		if err != nil {
			slog.Error("Failed to create new host", "host", hostname, "error", err)

			continue
		}

		hostGroup.AddHost(server)
	}
}

func executeCommandOnHost(host *host.Host, command string) error {
	// Use the context with a reasonable timeout; you can adjust this as needed.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create the CommandConfig
	config := commandmanager.CommandConfig{
		Command: command,
		Sudo:    false,
	}

	if host.SSHClient == nil {
		slog.Error("SSHClient is nil in executeCommandOnHost")
	} else {
		slog.Debug("SSHClient is available in executeCommandOnHost")
	}

	// Use the CommandManager embedded in the host.Host struct
	result, err := host.CommandManager.Run(ctx, config)
	if err != nil {
		return err
	}

	// Print the output
	fmt.Printf("Output of command on host %s:\n%s\n", host.Hostname, result.STDOUT)

	// Check for errors in STDERR if necessary
	if result.STDERR != "" {
		fmt.Printf("Errors: %s\n", result.STDERR)
	}

	return nil
}

func getHostInfo(host *host.Host) (HostInfo, error) {
	cpuUsage, err := host.HostManager.CPUUsage()
	if err != nil {
		return HostInfo{}, err
	}

	memoryUsage, err := host.HostManager.FreeMemory()
	if err != nil {
		return HostInfo{}, err
	}

	diskUsageDetails, err := host.FileManager.DiskUsage("/")
	if err != nil {
		return HostInfo{}, err
	}

	runningProcesses, err := host.HostManager.Processes()
	if err != nil {
		return HostInfo{}, err
	}

	return HostInfo{
		CPUUsage:         cpuUsage,
		DiskUsageDetails: diskUsageDetails,
		MemoryUsage:      memoryUsage,
		RunningProcesses: runningProcesses,
	}, nil
}

func main() {
	f := parseFlags()
	configureLogger(f)

	password, keyPass := readPasswords(f)
	options := buildHostOptions(f, password, keyPass)

	hostGroup := initializeHosts(f, options)

	if f.CheckHealth {
		err := processHosts(hostGroup, checkHostHealth, f.Concurrency)
		if err != nil {
			slog.Error("Error during Health Check", "error", err)
		}
	}

	if f.ExecCommand != "" {
		err := processHosts(hostGroup, func(host *host.Host) error {
			return executeCommandOnHost(host, f.ExecCommand)
		}, f.Concurrency)
		if err != nil {
			slog.Error("Error during ExecCommand", "error", err)
		}
	}

	if f.InfoDump {
		err := processHosts(hostGroup, dumpHostInfo, f.Concurrency)
		if err != nil {
			slog.Error("Error during InfoDump", "error", err)
		}
	}

	if f.ListPackages {
		err := processHosts(hostGroup, listAllPackages, f.Concurrency)
		if err != nil {
			slog.Error("Error during ListPackages", "error", err)
		}
	}

	if f.ListUpgradable {
		err := processHosts(hostGroup, listUpgradablePackages, f.Concurrency)
		if err != nil {
			slog.Error("Error during ListUpgradable", "error", err)
		}
	}

	if f.UpgradePackages {
		err := processHosts(hostGroup, upgradeAllPackages, f.Concurrency)
		if err != nil {
			slog.Error("Error during UpgradePackages", "error", err)
		}
	}

	if f.ScriptPath != "" {
		script, err := readScriptFile(f.ScriptPath)
		if err != nil {
			slog.Error("Failed to read script file", "error", err)
		}
		err = processHosts(hostGroup, func(host *host.Host) error {
			return executeScript(host, script)
		}, f.Concurrency)
		if err != nil {
			slog.Error("Error during Script execution", "error", err)
		}
	}

	if f.Monitor {
		monitorHosts(hostGroup, f)
	}
}

func configureLogger(f *flags) {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))

	if f.Debug {
		slog.Debug("Debug mode enabled")
		programLevel.Set(slog.LevelDebug)
	} else {
		slog.Info("Debug mode disabled")
		programLevel.Set(slog.LevelInfo)
	}
}

func readPasswords(f *flags) (password, keyPass string) {
	if f.PasswordPrompt {
		fmt.Print("Enter the password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			slog.Error("Failed to read password", "error", err)
		}
		password = string(passwordBytes)
		fmt.Println()
	}

	if f.KeyPassPrompt {
		fmt.Print("Enter the key passphrase: ")
		keyPassBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			slog.Error("Failed to read key passphrase", "error", err)
		}
		keyPass = string(keyPassBytes)
		fmt.Println()
	}
	return
}

func buildHostOptions(f *flags, password, keyPass string) []host.HostOption {
	var options []host.HostOption
	if f.Username != "" {
		options = append(options, host.WithUser(f.Username))
	}
	if password != "" {
		options = append(options, host.WithPassword(password))
	}
	if keyPass != "" {
		options = append(options, host.WithKeyPassphrase(keyPass))
	}
	if f.SudoPasswordPrompt {
		fmt.Print("Enter the sudo password: ")
		sudoPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			slog.Error("Failed to read sudo password: %v", err)
		}
		sudoPassword := string(sudoPasswordBytes)
		fmt.Println()
		if sudoPassword != "" {
			options = append(options, host.WithSudoPassword(sudoPassword))
		}
	}
	options = append(options, host.WithSSHClient(&host.RealSSHClient{}))
	slog.Debug("SSHClient set in options")
	return options
}

func initializeHosts(f *flags, options []host.HostOption) *hostgroup.HostGroup {
	hostGroup := hostgroup.NewHostGroup()

	if f.IniFilePath != "" {
		hostsMap, err := readHostsFromFile(f.IniFilePath)
		if err != nil {
			slog.Error("Failed to read INI file", "error", err)
		}
		for group, hosts := range hostsMap {
			slog.Error("Adding hosts from group", "group", group)
			addHosts(hosts, hostGroup, options...)
		}
	}
	if len(f.Hostnames) == 0 {
		f.Hostnames = append(f.Hostnames, "localhost")
	}
	addHosts(f.Hostnames, hostGroup, options...)

	return hostGroup
}

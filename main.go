package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	"encoding/json"

	"github.com/m-217/steelcut/steelcut"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

type HostInfo struct {
	CPUUsage         float64  `json:"cpuUsage"`
	DiskUsage        float64  `json:"diskUsage"`
	MemoryUsage      float64  `json:"memoryUsage"`
	RunningProcesses []string `json:"runningProcesses"`
}

type flags struct {
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
	MemoryThreshold    float64
	Monitor            bool
	PasswordPrompt     bool
	ScriptPath         string
	SudoPasswordPrompt bool
	UpgradePackages    bool
	Username           string
}

type hostnamesValue []string

var (
	logger = logrus.New()
)

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

func parseFlags() *flags {
	f := &flags{}
	flag.BoolVar(&f.Debug, "debug", false, "Enable debug log level")
	flag.BoolVar(&f.InfoDump, "info", false, "Dump information about the hosts")
	flag.BoolVar(&f.KeyPassPrompt, "keypass", false, "Passphrase for decrypting SSH keys")
	flag.BoolVar(&f.ListPackages, "list", false, "List all packages")
	flag.BoolVar(&f.ListUpgradable, "upgradable", false, "List all upgradable packages")
	flag.BoolVar(&f.Monitor, "monitor", false, "Enable host monitoring")
	flag.BoolVar(&f.PasswordPrompt, "password", false, "Use a password for SSH connection")
	flag.BoolVar(&f.SudoPasswordPrompt, "sudo-password", false, "Prompt for sudo password")
	flag.BoolVar(&f.UpgradePackages, "upgrade", false, "Upgrade all packages")
	flag.Float64Var(&f.CPUThreshold, "cpu-threshold", 80.0, "Threshold for CPU usage in percent")
	flag.Float64Var(&f.DiskThreshold, "disk-threshold", 80.0, "Threshold for disk usage in percent")
	flag.Float64Var(&f.MemoryThreshold, "memory-threshold", 80.0, "Threshold for memory usage in percent")
	flag.IntVar(&f.Concurrency, "concurrency", 10, "Maximum number of concurrent host connections")
	flag.StringVar(&f.ExecCommand, "exec", "", "Execute command on the host")
	flag.StringVar(&f.IniFilePath, "ini", "", "Path to INI file with host configurations")
	flag.StringVar(&f.LogFileName, "log", "log.txt", "Log file name")
	flag.StringVar(&f.ScriptPath, "script", "", "Path to script file to be executed on the host")
	flag.StringVar(&f.Username, "username", "", "Username to use for SSH connection")
	flag.Var(&f.Hostnames, "hostname", "Hostname to connect to")

	flag.Parse()

	return f
}

func monitorHosts(hg *steelcut.HostGroup, f *flags) {
	for {
		hg.RLock()
		for _, host := range hg.Hosts {
			hostInfo, err := getHostInfo(host)
			if err != nil {
				logger.Error(err)
				continue
			}

			if hostInfo.CPUUsage > f.CPUThreshold {
				logger.Warnf("CPU usage on host %s exceeded threshold: %.2f%%", host.Hostname(), hostInfo.CPUUsage)
			}

			if hostInfo.MemoryUsage > f.MemoryThreshold {
				logger.Warnf("Memory usage on host %s exceeded threshold: %.2f%%", host.Hostname(), hostInfo.MemoryUsage)
			}

			if hostInfo.DiskUsage > f.DiskThreshold {
				logger.Warnf("Disk usage on host %s exceeded threshold: %.2f%%", host.Hostname(), hostInfo.DiskUsage)
			}
		}
		hg.RUnlock()

		time.Sleep(5 * time.Second) // Delay between monitoring checks
	}
}

func readScriptFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func executeScript(host steelcut.Host, script string) error {
	result, err := host.RunCommand(script)
	if err != nil {
		return err
	}
	fmt.Printf("Output of script on host %s:\n%s\n", host.Hostname(), result)
	return nil
}

func processHosts(hg *steelcut.HostGroup, action func(host steelcut.Host) error, maxConcurrency int) error {
	sem := make(chan struct{}, maxConcurrency) // Create semaphore with buffer size equal to max concurrency
	errCh := make(chan error, len(hg.Hosts))

	hg.RLock()
	for _, host := range hg.Hosts {
		go func(h steelcut.Host) {
			sem <- struct{}{}        // Acquire token
			defer func() { <-sem }() // Release token
			if err := action(h); err != nil {
				errCh <- err
			}
		}(host)
	}
	hg.RUnlock()

	for i := 0; i < maxConcurrency; i++ {
		sem <- struct{}{}
	}

	close(errCh)
	close(sem) // Close the semaphore channel when done

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("Host processing error: %v", err)
		}
		return fmt.Errorf("some hosts encountered errors")
	}

	return nil
}

func dumpHostInfo(host steelcut.Host) error {
	hostInfo, err := getHostInfo(host)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(hostInfo, "", "  ")
	if err != nil {
		return err
	}

	// Use the hostname to create a unique file name for each host
	fileName := fmt.Sprintf("host_info_%s.json", host.Hostname())

	err = os.WriteFile(fileName, b, 0644)
	return err
}

func listAllPackages(host steelcut.Host) error {
	packages, err := host.ListPackages()
	if err != nil {
		return fmt.Errorf("failed to list packages: %v", err)
	}
	fmt.Println("Packages:")
	for _, pkg := range packages {
		fmt.Println(pkg)
	}
	return nil
}

func listUpgradablePackages(host steelcut.Host) error {
	upgradable, err := host.CheckUpdates()
	if err != nil {
		return fmt.Errorf("failed to check OS updates: %v", err)
	}
	fmt.Println("Upgradable packages:")
	for _, pkg := range upgradable {
		fmt.Println(pkg)
	}
	return nil
}

func upgradeAllPackages(host steelcut.Host) error {
	_, err := host.UpgradeAllPackages()
	if err != nil {
		return fmt.Errorf("failed to upgrade packages: %v", err)
	}
	fmt.Println("Successfully upgraded all packages.")
	return nil
}

func addHosts(hostnames []string, hostGroup *steelcut.HostGroup, options ...steelcut.HostOption) {
	for _, host := range hostnames {
		log.Printf("Adding host %s", host)
		server, err := steelcut.NewHost(host, options...)
		if err != nil {
			log.Printf("Failed to create new host: %v", err)
			continue
		}

		if err := server.IsReachable(); err != nil {
			log.Printf("Host %s is not reachable, skipping: %v", host, err)
			continue
		}

		hostGroup.AddHost(server)
	}
}

func getHostInfo(host steelcut.Host) (HostInfo, error) {
	cpuUsage, err := host.CPUUsage()
	if err != nil {
		return HostInfo{}, err
	}

	memoryUsage, err := host.MemoryUsage()
	if err != nil {
		return HostInfo{}, err
	}

	diskUsage, err := host.DiskUsage()
	if err != nil {
		return HostInfo{}, err
	}

	runningProcesses, err := host.RunningProcesses()
	if err != nil {
		return HostInfo{}, err
	}

	return HostInfo{
		CPUUsage:         cpuUsage,
		DiskUsage:        diskUsage,
		MemoryUsage:      memoryUsage,
		RunningProcesses: runningProcesses,
	}, nil
}

func main() {
	f := parseFlags()

	file, err := os.OpenFile(f.LogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatal(err)
	}
	defer file.Close()

	logger.SetOutput(file)
	if f.Debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	var password, keyPass string
	if f.PasswordPrompt {
		fmt.Print("Enter the password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Failed to read password: %v", err)
		}
		password = string(passwordBytes)
		fmt.Println()
	}

	if f.KeyPassPrompt {
		fmt.Print("Enter the key passphrase: ")
		keyPassBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Failed to read key passphrase: %v", err)
		}
		keyPass = string(keyPassBytes)
		fmt.Println()
	}

	var options []steelcut.HostOption
	if f.Username != "" {
		options = append(options, steelcut.WithUser(f.Username))
	}
	if password != "" {
		options = append(options, steelcut.WithPassword(password))
	}
	if keyPass != "" {
		options = append(options, steelcut.WithKeyPassphrase(keyPass))
	}

	if len(f.Hostnames) == 0 {
		f.Hostnames = append(f.Hostnames, "localhost")
	}

	var sudoPassword string
	if f.SudoPasswordPrompt {
		fmt.Print("Enter the sudo password: ")
		sudoPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Failed to read sudo password: %v", err)
		}
		sudoPassword = string(sudoPasswordBytes)
		fmt.Println()
	}

	if sudoPassword != "" {
		options = append(options, steelcut.WithSudoPassword(sudoPassword))
	}

	hostGroup := steelcut.NewHostGroup()

	client := &steelcut.RealSSHClient{}
	options = append(options, steelcut.WithSSHClient(client))

	if f.IniFilePath != "" {
		hostsMap, err := readHostsFromFile(f.IniFilePath)
		if err != nil {
			log.Fatalf("Failed to read INI file: %v", err)
		}

		for group, hosts := range hostsMap {
			log.Printf("Adding hosts from group %s", group)
			addHosts(hosts, hostGroup, options...)
		}
	}

	addHosts(f.Hostnames, hostGroup, options...)

	if f.InfoDump {
		processHosts(hostGroup, dumpHostInfo, f.Concurrency)
	}

	if f.Monitor {
		go monitorHosts(hostGroup, f)
		stopChan := make(chan struct{})
		<-stopChan
	}

	if f.ListPackages {
		processHosts(hostGroup, listAllPackages, f.Concurrency)
	}

	if f.ListUpgradable {
		processHosts(hostGroup, listUpgradablePackages, f.Concurrency)
	}

	if f.UpgradePackages {
		processHosts(hostGroup, upgradeAllPackages, f.Concurrency)
	}

	if f.ExecCommand != "" {
		results := hostGroup.RunCommandOnAll(f.ExecCommand)
		for _, result := range results {
			if result.Error != nil {
				logger.Error(result.Error)
			} else {
				fmt.Printf("Output of command on host %s:\n%s\n", result.Host, result.Result)
			}
		}
	}

	if f.ScriptPath != "" {
		script, err := readScriptFile(f.ScriptPath)
		if err != nil {
			log.Fatalf("Failed to read script file: %v", err)
		}
		processHosts(hostGroup, func(host steelcut.Host) error {
			return executeScript(host, script)
		}, f.Concurrency)
	}

}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	"encoding/json"

	"github.com/m-217/steelcut/steelcut"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

type HostInfo struct {
	CPUUsage         float64  `json:"cpuUsage"`
	MemoryUsage      float64  `json:"memoryUsage"`
	DiskUsage        float64  `json:"diskUsage"`
	RunningProcesses []string `json:"runningProcesses"`
}

var (
	logFileName     string
	debug           bool
	logger          = logrus.New()
	infoDump        bool
	listPackages    bool
	listUpgradable  bool
	upgradePackages bool
	execCommand     string
)

type hostnamesValue []string

func (h *hostnamesValue) String() string {
	return strings.Join(*h, ",")
}

func (h *hostnamesValue) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func init() {
	flag.StringVar(&logFileName, "log", "log.txt", "Log file name")
	flag.BoolVar(&debug, "debug", false, "Enable debug log level")
	flag.BoolVar(&infoDump, "info", false, "Dump information about the hosts")
	flag.BoolVar(&listPackages, "list", false, "List all packages")
	flag.BoolVar(&listUpgradable, "upgradable", false, "List all upgradable packages")
	flag.BoolVar(&upgradePackages, "upgrade", false, "Upgrade all packages")
	flag.StringVar(&execCommand, "exec", "", "Execute command on the host")
}

type SSHClientImpl struct{}

func (s *SSHClientImpl) Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	return ssh.Dial(network, addr, config)
}

func processHosts(hg *steelcut.HostGroup, action func(host steelcut.Host) error) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(hg.Hosts))

	hg.RLock() // Lock for reading
	for _, host := range hg.Hosts {
		wg.Add(1)
		go func(h steelcut.Host) {
			defer wg.Done()
			if err := action(h); err != nil {
				errCh <- err
			}
		}(host)
	}
	hg.RUnlock()

	wg.Wait()
	close(errCh)

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d error(s): %v", len(errors), errors)
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

	err = os.WriteFile("host_info.json", b, 0644)
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

func main() {
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatal(err)
	}
	defer file.Close()

	logger.SetOutput(file)
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	var hostnames hostnamesValue
	flag.Var(&hostnames, "hostname", "Hostname to connect to")

	username := flag.String("username", "", "Username to use for SSH connection")
	passwordPrompt := flag.Bool("password", false, "Use a password for SSH connection")
	keyPassPrompt := flag.Bool("keypass", false, "Passphrase for decrypting SSH keys")
	sudoPasswordPrompt := flag.Bool("sudo-password", false, "Prompt for sudo password")

	flag.Parse()

	var password, keyPass string
	if *passwordPrompt {
		fmt.Print("Enter the password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Failed to read password: %v", err)
		}
		password = string(passwordBytes)
		fmt.Println()
	}
	if *keyPassPrompt {
		fmt.Print("Enter the key passphrase: ")
		keyPassBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Failed to read key passphrase: %v", err)
		}
		keyPass = string(keyPassBytes)
		fmt.Println()
	}

	var options []steelcut.HostOption
	if *username != "" {
		options = append(options, steelcut.WithUser(*username))
	}
	if password != "" {
		options = append(options, steelcut.WithPassword(password))
	}
	if keyPass != "" {
		options = append(options, steelcut.WithKeyPassphrase(keyPass))
	}

	if len(hostnames) == 0 {
		hostnames = append(hostnames, "localhost")
	}

	var sudoPassword string
	if *sudoPasswordPrompt {
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

	client := &SSHClientImpl{}
	options = append(options, steelcut.WithSSHClient(client))

	// Iterate over the hostnames and add them to the host group
	for _, host := range hostnames {
		log.Printf("Adding host %s", host)
		server, err := steelcut.NewHost(host, options...)
		if err != nil {
			log.Fatalf("Failed to create new host: %v", err)
		}
		hostGroup.AddHost(server)
	}

	if infoDump {
		processHosts(hostGroup, dumpHostInfo)
	}

	if listPackages {
		processHosts(hostGroup, listAllPackages)
	}

	if listUpgradable {
		processHosts(hostGroup, listUpgradablePackages)
	}

	if upgradePackages {
		processHosts(hostGroup, upgradeAllPackages)
	}

	if execCommand != "" {
		results, errs := hostGroup.RunCommandOnAll(execCommand)
		for i, result := range results {
			fmt.Printf("Output of command on host %s:\n%s\n", hostnames[i], result)
		}
		for _, err := range errs {
			logger.Error(err)
		}
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
		MemoryUsage:      memoryUsage,
		DiskUsage:        diskUsage,
		RunningProcesses: runningProcesses,
	}, nil
}

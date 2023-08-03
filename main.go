package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
)

func init() {
	flag.StringVar(&logFileName, "log", "log.txt", "Log file name")
	flag.BoolVar(&debug, "debug", false, "Enable debug log level")
	flag.BoolVar(&infoDump, "info", false, "Dump information about the hosts")
	flag.BoolVar(&listPackages, "list", false, "List all packages")
	flag.BoolVar(&listUpgradable, "upgradable", false, "List all upgradable packages")
	flag.BoolVar(&upgradePackages, "upgrade", false, "Upgrade all packages")
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

	hostname := flag.String("hostname", "", "Hostname to connect to")
	username := flag.String("username", "", "Username to use for SSH connection")
	passwordPrompt := flag.Bool("password", false, "Use a password for SSH connection")
	keyPassPrompt := flag.Bool("keypass", false, "Passphrase for decrypting SSH keys")

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

	if *hostname == "" {
		*hostname = "localhost"
	}

	hostGroup := steelcut.NewHostGroup()

	hosts := []string{*hostname, "localhost"}
	for _, host := range hosts {
		server, err := steelcut.NewHost(host, options...)
		if err != nil {
			log.Fatalf("Failed to create new host: %v", err)
		}
		hostGroup.AddHost(server)
	}

	if infoDump {
		for _, host := range hostGroup.Hosts {
			hostInfo, err := getHostInfo(host)
			if err != nil {
				log.Fatalf("Failed to get host info: %v", err)
			}

			b, err := json.MarshalIndent(hostInfo, "", "  ")
			if err != nil {
				log.Fatalf("Failed to serialize host info: %v", err)
			}

			err = os.WriteFile("host_info.json", b, 0644)
			if err != nil {
				log.Fatalf("Failed to write host info to file: %v", err)
			}
		}
	}

	if listPackages {
		for _, host := range hostGroup.Hosts {
			packages, err := host.ListPackages()
			if err != nil {
				log.Fatalf("Failed to list packages: %v", err)
			}
			for _, pkg := range packages {
				fmt.Println(pkg)
			}
		}
	}

	if listUpgradable {
		for _, host := range hostGroup.Hosts {
			upgradable, err := host.CheckUpdates()
			if err != nil {
				log.Fatalf("Failed to check OS updates: %v", err)
			}
			for _, pkg := range upgradable {
				fmt.Println(pkg)
			}
		}
	}

	results, errors := hostGroup.RunCommandOnAll("uname -a")

	if len(errors) > 0 {
		for _, err := range errors {
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	for i, result := range results {
		fmt.Printf("Result for host %d: %s\n", i+1, result)
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

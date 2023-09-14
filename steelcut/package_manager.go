package steelcut

import (
	"fmt"
	"strings"
)

// PackageManager interface defines the methods that package manager implementations must provide.
type PackageManager interface {
	ListPackages(*UnixHost) ([]string, error)
	AddPackage(*UnixHost, string) error
	RemovePackage(*UnixHost, string) error
	UpgradePackage(*UnixHost, string) error
	CheckOSUpdates(host *UnixHost) ([]string, error)
	UpgradeAll(*UnixHost) ([]Update, error)
}

// Update represents a package update.
type Update struct {
	PackageName string
	Version     string
}

type YumPackageManager struct {
	Executor CommandExecutor
}

var registeredPackageManagers = map[OSType]func(host *UnixHost, cmdOptions CommandOptions) Host{
	LinuxUbuntu: func(host *UnixHost, cmdOptions CommandOptions) Host {
		return configureLinuxHost(host, cmdOptions, "apt")
	},
	LinuxDebian: func(host *UnixHost, cmdOptions CommandOptions) Host {
		return configureLinuxHost(host, cmdOptions, "apt")
	},
	LinuxFedora: func(host *UnixHost, cmdOptions CommandOptions) Host {
		return configureLinuxHost(host, cmdOptions, "dnf")
	},
	LinuxRedHat: func(host *UnixHost, cmdOptions CommandOptions) Host {
		return configureLinuxHost(host, cmdOptions, "yum")
	},
	Darwin: func(host *UnixHost, cmdOptions CommandOptions) Host {
		return configureMacHost(host, cmdOptions)
	},
}

func (pm YumPackageManager) ListPackages(host *UnixHost) ([]string, error) {
	output, err := pm.Executor.RunCommand("yum list installed", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, err
	}
	return strings.Split(output, "\n"), nil
}

func (pm YumPackageManager) AddPackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("yum install -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

func (pm YumPackageManager) RemovePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("yum remove -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

func (pm YumPackageManager) UpgradePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("yum upgrade -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

func (pm YumPackageManager) CheckOSUpdates(host *UnixHost) ([]string, error) {
	log.Info("Checking for YUM OS updates")
	output, err := pm.Executor.RunCommand("yum check-update", CommandOptions{UseSudo: true})
	if err != nil {
		log.Error("Error checking YUM updates: %v", err)
		return nil, err
	}

	updates := strings.Split(output, "\n")
	log.Info("YUM Updates available: %v", updates)
	return updates, nil
}

// UpgradeAll upgrades all the packages to their latest versions.
func (pm YumPackageManager) UpgradeAll(host *UnixHost) ([]Update, error) {
	output, err := pm.Executor.RunCommand("yum update -y", CommandOptions{UseSudo: true})
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade all packages: %v, Output: %s", err, output)
	}
	updates := parseUpdates(output)
	return updates, nil
}

type AptPackageManager struct {
	Executor CommandExecutor
}

// ListPackages returns the installed packages.
func (pm AptPackageManager) ListPackages(host *UnixHost) ([]string, error) {
	output, err := pm.Executor.RunCommand("apt list --installed", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, err
	}

	packages := strings.Split(output, "\n")
	return packages, nil
}

// AddPackage adds a package to the host.
func (pm AptPackageManager) AddPackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(
		fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y -q %s", pkg),
		CommandOptions{UseSudo: true},
	)
	return err
}

// RemovePackage removes a package from the host.
func (pm AptPackageManager) RemovePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(
		fmt.Sprintf("apt-get remove -y -q %s", pkg),
		CommandOptions{UseSudo: true},
	)
	return err
}

// UpgradePackage upgrades a package to the latest version.
func (pm AptPackageManager) UpgradePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(
		fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get upgrade -y -q %s", pkg),
		CommandOptions{UseSudo: true},
	)
	return err
}

// UpgradeAll upgrades all the packages to their latest versions.
func (pm AptPackageManager) UpgradeAll(host *UnixHost) ([]Update, error) {
	if pm.Executor == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	output, err := pm.Executor.RunCommand(
		"DEBIAN_FRONTEND=noninteractive apt-get upgrade -y -q",
		CommandOptions{UseSudo: true},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade all packages: %v, Output: %s", err, output)
	}

	updates := pm.parseAptUpdates(output)

	return updates, nil
}

// CheckOSUpdates checks for OS updates.
func (pm AptPackageManager) CheckOSUpdates(host *UnixHost) ([]string, error) {
	_, err := pm.Executor.RunCommand(
		"apt-get update -q",
		CommandOptions{UseSudo: true},
	)
	if err != nil {
		log.Error("Failed to update apt: %v", err)
		return nil, fmt.Errorf("failed to update apt: %w", err)
	}

	output, err := pm.Executor.RunCommand(
		"apt list --upgradable -q",
		CommandOptions{UseSudo: false},
	)
	if err != nil {
		return nil, err
	}

	updates := strings.Split(output, "\n")
	return updates, nil
}

func (pm AptPackageManager) parseAptUpdates(output string) []Update {
	lines := strings.Split(output, "\n")
	var updates []Update

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}

		if parts[3] != "[upgradable" {
			continue
		}

		packageName := strings.Split(parts[0], "/")[0]
		version := parts[1]
		update := Update{
			PackageName: packageName,
			Version:     version,
		}

		updates = append(updates, update)
	}

	return updates
}

type DnfPackageManager struct {
	Executor CommandExecutor
}

func (pm DnfPackageManager) ListPackages(host *UnixHost) ([]string, error) {
	output, err := pm.Executor.RunCommand("dnf list installed", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, err
	}
	return strings.Split(output, "\n"), nil
}

func (pm DnfPackageManager) AddPackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("dnf install -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

func (pm DnfPackageManager) RemovePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("dnf remove -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

func (pm DnfPackageManager) UpgradePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("dnf upgrade -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

func (pm DnfPackageManager) CheckOSUpdates(host *UnixHost) ([]string, error) {
	log.Info("Checking for DNF OS updates")
	output, err := pm.Executor.RunCommand("dnf check-update", CommandOptions{UseSudo: true})
	if err != nil {
		log.Info("Error checking DNF updates: %v", err)
		return nil, err
	}

	updates := strings.Split(output, "\n")
	log.Info("DNF Updates available: %v", updates)
	return updates, nil
}

func (pm DnfPackageManager) UpgradeAll(host *UnixHost) ([]Update, error) {
	output, err := pm.Executor.RunCommand("dnf upgrade -y", CommandOptions{UseSudo: true})
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade all packages using DNF: %v, Output: %s", err, output)
	}
	updates := parseUpdates(output)
	return updates, nil
}

type BrewPackageManager struct {
	Executor CommandExecutor
}

// ListPackages returns the installed packages.
func (pm BrewPackageManager) ListPackages(host *UnixHost) ([]string, error) {
	output, err := pm.Executor.RunCommand("brew list --version", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, err
	}

	packages := strings.Split(output, "\n")
	return packages, nil
}

// AddPackage adds a package to the host.
func (pm BrewPackageManager) AddPackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("brew install %s", pkg), CommandOptions{UseSudo: false})
	return err
}

func (pm BrewPackageManager) RemovePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("brew uninstall %s", pkg), CommandOptions{UseSudo: false})
	return err
}

// UpgradePackage upgrades a package to the latest version.
func (pm BrewPackageManager) UpgradePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("brew upgrade %s", pkg), CommandOptions{UseSudo: false})
	return err
}

// CheckOSUpdates checks for OS updates.
func (pm BrewPackageManager) CheckOSUpdates(host *UnixHost) ([]string, error) {
	output, err := pm.Executor.RunCommand("brew outdated", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, err
	}

	updates := strings.Split(output, "\n")
	return updates, nil
}

// UpgradeAll upgrades all the packages to their latest versions.
func (pm BrewPackageManager) UpgradeAll(host *UnixHost) ([]Update, error) {
	// We explicitly don't want to run as root here, as brew will complain
	output, err := pm.Executor.RunCommand("brew upgrade", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade all packages: %v, Output: %s", err, output)
	}
	updates := pm.parseUpdates(output)
	return updates, nil
}

// parseUpdates parses the output of `brew upgrade` to get the list of packages that will be upgraded.
func (pm BrewPackageManager) parseUpdates(output string) []Update {
	lines := strings.Split(output, "\n")
	var updates []Update

	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue
		}

		update := Update{
			PackageName: parts[0],
			Version:     parts[1],
		}

		updates = append(updates, update)
	}

	return updates
}

// parseUpdates is a common function used to parse package update information.
func parseUpdates(output string) []Update {
	lines := strings.Split(output, "\n")
	var updates []Update

	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue
		}

		update := Update{
			PackageName: parts[0],
			Version:     parts[1],
		}

		updates = append(updates, update)
	}

	return updates
}

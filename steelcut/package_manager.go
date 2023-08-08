package steelcut

import (
	"fmt"
	"strings"
)

type PackageManager interface {
	ListPackages(*UnixHost) ([]string, error)
	AddPackage(*UnixHost, string) error
	RemovePackage(*UnixHost, string) error
	UpgradePackage(*UnixHost, string) error
}

type Update struct {
	PackageName string
	Version     string
}

type YumPackageManager struct{}

func (pm YumPackageManager) ListPackages(host *UnixHost) ([]string, error) {
	output, err := host.RunCommand("yum list installed")
	if err != nil {
		return nil, err
	}

	packages := strings.Split(output, "\n")
	return packages, nil
}

func (pm YumPackageManager) AddPackage(host *UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("yum install -y %s", pkg))
	return err
}

func (pm YumPackageManager) RemovePackage(host *UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("yum remove -y %s", pkg))
	return err
}

func (pm YumPackageManager) UpgradePackage(host *UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("yum upgrade -y %s", pkg))
	return err
}

func (pm YumPackageManager) CheckOSUpdates(host *UnixHost) ([]string, error) {
	output, err := host.RunCommand("yum check-update")
	if err != nil {
		return nil, err
	}

	updates := strings.Split(output, "\n")
	return updates, nil
}

func (pm YumPackageManager) UpgradeOS(host *UnixHost) ([]Update, error) {
	output, err := host.RunCommand("yum upgrade -y")
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade OS: %v", err)
	}
	updates := parseUpdates(output)
	return updates, nil
}

type AptPackageManager struct{}

func (pm AptPackageManager) ListPackages(host *UnixHost) ([]string, error) {
	output, err := host.RunCommand("apt list --installed")
	if err != nil {
		return nil, err
	}

	packages := strings.Split(output, "\n")
	return packages, nil
}

func (pm AptPackageManager) AddPackage(host *UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("apt install -y %s", pkg))
	return err
}

func (pm AptPackageManager) RemovePackage(host *UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("apt remove -y %s", pkg))
	return err
}

func (pm AptPackageManager) UpgradePackage(host *UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("apt upgrade -y %s", pkg))
	return err
}

func (pm AptPackageManager) CheckOSUpdates(host *UnixHost) ([]string, error) {
	_, err := host.RunCommand("apt update")
	if err != nil {
		return nil, err
	}

	output, err := host.RunCommand("apt list --upgradable")
	if err != nil {
		return nil, err
	}

	updates := strings.Split(output, "\n")
	return updates, nil
}

func (pm AptPackageManager) UpgradeOS(host *UnixHost) ([]Update, error) {
	output, err := host.RunCommand("apt upgrade -y")
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade OS: %v", err)
	}
	updates := parseUpdates(output)
	return updates, nil
}

type BrewPackageManager struct{}

func (pm BrewPackageManager) ListPackages(host *UnixHost) ([]string, error) {
	output, err := host.RunCommand("brew list --version")
	if err != nil {
		return nil, err
	}

	packages := strings.Split(output, "\n")
	return packages, nil
}

func (pm BrewPackageManager) AddPackage(host *UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("brew install %s", pkg))
	return err
}

func (pm BrewPackageManager) RemovePackage(host *UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("brew uninstall %s", pkg))
	return err
}

func (pm BrewPackageManager) UpgradePackage(host *UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("brew upgrade %s", pkg))
	return err
}

func (pm BrewPackageManager) CheckOSUpdates(host *UnixHost) ([]string, error) {
	output, err := host.RunCommand("brew outdated")
	if err != nil {
		return nil, err
	}

	updates := strings.Split(output, "\n")
	return updates, nil
}

func (pm BrewPackageManager) UpgradeOS(host *UnixHost) ([]Update, error) {
	output, err := host.RunCommand("brew upgrade")
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade OS: %v", err)
	}
	updates := pm.parseUpdates(output)
	return updates, nil
}

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

package steelcut

import (
	"fmt"
	"strings"
)

type PackageManager interface {
	ListPackages(host UnixHost) ([]string, error)
	AddPackage(host UnixHost, pkg string) error
	RemovePackage(host UnixHost, pkg string) error
	UpgradePackage(host UnixHost, pkg string) error
}

type YumPackageManager struct{}

func (pm YumPackageManager) ListPackages(host UnixHost) ([]string, error) {
	output, err := host.RunCommand("yum list installed")
	if err != nil {
		return nil, err
	}

	packages := strings.Split(output, "\n")
	return packages, nil
}

func (pm YumPackageManager) AddPackage(host UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("yum install -y %s", pkg))
	return err
}

func (pm YumPackageManager) RemovePackage(host UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("yum remove -y %s", pkg))
	return err
}

func (pm YumPackageManager) UpgradePackage(host UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("yum upgrade -y %s", pkg))
	return err
}

type AptPackageManager struct{}

func (pm AptPackageManager) ListPackages(host UnixHost) ([]string, error) {
	output, err := host.RunCommand("apt list --installed")
	if err != nil {
		return nil, err
	}

	packages := strings.Split(output, "\n")
	return packages, nil
}

func (pm AptPackageManager) AddPackage(host UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("apt install -y %s", pkg))
	return err
}

func (pm AptPackageManager) RemovePackage(host UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("apt remove -y %s", pkg))
	return err
}

func (pm AptPackageManager) UpgradePackage(host UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("apt upgrade -y %s", pkg))
	return err
}

type BrewPackageManager struct{}

func (pm BrewPackageManager) ListPackages(host UnixHost) ([]string, error) {
	output, err := host.RunCommand("brew list")
	if err != nil {
		return nil, err
	}

	packages := strings.Split(output, "\n")
	return packages, nil
}

func (pm BrewPackageManager) AddPackage(host UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("brew install %s", pkg))
	return err
}

func (pm BrewPackageManager) RemovePackage(host UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("brew uninstall %s", pkg))
	return err
}

func (pm BrewPackageManager) UpgradePackage(host UnixHost, pkg string) error {
	_, err := host.RunCommand(fmt.Sprintf("brew upgrade %s", pkg))
	return err
}

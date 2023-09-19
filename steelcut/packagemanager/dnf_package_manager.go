package packagemanager

import (
	"context"
	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
	"strings"
)

type DnfPackageManager struct {
	CommandManager cm.CommandManager
}

func (dpm *DnfPackageManager) ListPackages() ([]string, error) {
	output, err := dpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "dnf",
		Args:    []string{"list", "installed"},
	})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output.STDOUT, "\n")
	var packages []string
	for _, line := range lines[1:] { // Skipping the header line
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}
	return packages, nil
}

func (dpm *DnfPackageManager) AddPackage(pkg string) error {
	_, err := dpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "dnf",
		Args:    []string{"install", "-y", pkg},
	})
	return err
}

func (dpm *DnfPackageManager) RemovePackage(pkg string) error {
	_, err := dpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "dnf",
		Args:    []string{"remove", "-y", pkg},
	})
	return err
}

func (dpm *DnfPackageManager) UpgradePackage(pkg string) error {
	_, err := dpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "dnf",
		Args:    []string{"upgrade", "-y", pkg},
	})
	return err
}

func (dpm *DnfPackageManager) CheckOSUpdates() ([]string, error) {
	output, err := dpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "dnf",
		Args:    []string{"list", "upgrades"},
	})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output.STDOUT, "\n")
	var updates []string
	for _, line := range lines[1:] { // Skipping the header line
		parts := strings.Fields(line)
		if len(parts) > 0 {
			updates = append(updates, parts[0])
		}
	}
	return updates, nil
}

func (dpm *DnfPackageManager) UpgradeAll() ([]string, error) {
	_, err := dpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "dnf",
		Args:    []string{"upgrade", "-y"},
	})
	if err != nil {
		return nil, err
	}
	return dpm.CheckOSUpdates()
}

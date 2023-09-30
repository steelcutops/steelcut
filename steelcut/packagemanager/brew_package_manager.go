package packagemanager

import (
	"context"
	"strings"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type BrewPackageManager struct {
	CommandManager cm.CommandManager
}

func (bpm *BrewPackageManager) ListPackages() ([]string, error) {
	output, err := bpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "brew",
		Args:    []string{"list"},
	})
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(output.STDOUT), "\n"), nil
}

func (bpm *BrewPackageManager) AddPackage(pkg string) error {
	_, err := bpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "brew",
		Args:    []string{"install", pkg},
	})
	return err
}

func (bpm *BrewPackageManager) RemovePackage(pkg string) error {
	_, err := bpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "brew",
		Args:    []string{"uninstall", pkg},
	})
	return err
}

func (bpm *BrewPackageManager) UpgradePackage(pkg string) error {
	_, err := bpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "brew",
		Args:    []string{"upgrade", pkg},
	})
	return err
}

func (bpm *BrewPackageManager) CheckOSUpdates() ([]string, error) {
	output, err := bpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "brew",
		Args:    []string{"outdated"},
	})
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(output.STDOUT), "\n"), nil
}

func (bpm *BrewPackageManager) UpgradeAll() ([]string, error) {
	_, err := bpm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "brew",
		Args:    []string{"upgrade"},
	})
	if err != nil {
		return nil, err
	}
	return bpm.CheckOSUpdates()
}

func (bpm *BrewPackageManager) EnsurePackagePresent(pkg string) error {
	packages, err := bpm.ListPackages()
	if err != nil {
		return err
	}

	for _, installedPkg := range packages {
		if installedPkg == pkg {
			// Package is already installed; return without taking action
			return nil
		}
	}

	// Package is not installed; proceed with installation
	return bpm.AddPackage(pkg)
}

func (bpm *BrewPackageManager) EnsurePackageAbsent(pkg string) error {
	packages, err := bpm.ListPackages()
	if err != nil {
		return err
	}

	for _, installedPkg := range packages {
		if installedPkg == pkg {
			// Package is installed; proceed with removal
			return bpm.RemovePackage(pkg)
		}
	}

	// Package is not installed; return without taking action
	return nil
}

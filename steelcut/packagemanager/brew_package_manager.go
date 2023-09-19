package packagemanager

import (
	"context"
	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
	"strings"
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

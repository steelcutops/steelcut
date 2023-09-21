package packagemanager

import (
	"context"
	"strings"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type YumPackageManager struct {
	CommandManager cm.CommandManager
}

func (ypm *YumPackageManager) ListPackages() ([]string, error) {
	output, err := ypm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "yum",
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

func (ypm *YumPackageManager) AddPackage(pkg string) error {
	_, err := ypm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "yum",
		Sudo:    true,
		Args:    []string{"install", "-y", pkg},
	})
	return err
}

func (ypm *YumPackageManager) RemovePackage(pkg string) error {
	_, err := ypm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "yum",
		Sudo:    true,
		Args:    []string{"remove", "-y", pkg},
	})
	return err
}

func (ypm *YumPackageManager) UpgradePackage(pkg string) error {
	_, err := ypm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "yum",
		Sudo:    true,
		Args:    []string{"update", "-y", pkg},
	})
	return err
}

func (ypm *YumPackageManager) CheckOSUpdates() ([]string, error) {
	output, err := ypm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "yum",
		Args:    []string{"list", "updates"},
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

func (ypm *YumPackageManager) UpgradeAll() ([]string, error) {
	_, err := ypm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "yum",
		Sudo:    true,
		Args:    []string{"update", "-y"},
	})
	if err != nil {
		return nil, err
	}
	return ypm.CheckOSUpdates()
}

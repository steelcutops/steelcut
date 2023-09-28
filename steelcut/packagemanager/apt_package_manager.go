package packagemanager

import (
	"context"
	"strings"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type AptPackageManager struct {
	CommandManager cm.CommandManager
}

func (apm *AptPackageManager) ListPackages() ([]string, error) {
	output, err := apm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "dpkg",
		Args:    []string{"--get-selections"},
	})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output.STDOUT, "\n")
	var packages []string
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}
	return packages, nil
}

func (apm *AptPackageManager) AddPackage(pkg string) error {
	_, err := apm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apt-get",
		Sudo:    true,
		Env:     []string{"DEBIAN_FRONTEND=noninteractive"},
		Args:    []string{"install", "-y", "-o", "Dpkg::Options::=--force-confdef", "-o", "Dpkg::Options::=--force-confold", pkg},
	})
	return err
}

func (apm *AptPackageManager) RemovePackage(pkg string) error {
	_, err := apm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apt-get",
		Sudo:    true,
		Args:    []string{"remove", "-y", pkg},
	})
	return err
}

func (apm *AptPackageManager) UpgradePackage(pkg string) error {
	_, err := apm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apt-get",
		Sudo:    true,
		Env:     []string{"DEBIAN_FRONTEND=noninteractive"},
		Args:    []string{"install", "--only-upgrade", "-y", "-o", "Dpkg::Options::=--force-confdef", "-o", "Dpkg::Options::=--force-confold", pkg},
	})
	return err
}

func (apm *AptPackageManager) CheckOSUpdates() ([]string, error) {
	_, err := apm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apt-get",
		Sudo:    true,
		Args:    []string{"update"},
	})
	if err != nil {
		return nil, err
	}

	output, err := apm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apt",
		Args:    []string{"list", "--upgradable"},
	})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output.STDOUT, "\n")
	var updates []string
	for _, line := range lines {
		if strings.Contains(line, "upgradable from") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				updates = append(updates, parts[0])
			}
		}
	}
	return updates, nil
}

func (apm *AptPackageManager) UpgradeAll() ([]string, error) {
	_, err := apm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apt-get",
		Sudo:    true,
		Env:     []string{"DEBIAN_FRONTEND=noninteractive"},
		Args:    []string{"dist-upgrade", "-y", "-o", "Dpkg::Options::=--force-confdef", "-o", "Dpkg::Options::=--force-confold"},
	})
	if err != nil {
		return nil, err
	}
	return apm.CheckOSUpdates()
}

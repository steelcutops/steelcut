package packagemanager

import (
	"context"
	"strings"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type ApkPackageManager struct {
	CommandManager cm.CommandManager
}

func (apkm *ApkPackageManager) ListPackages() ([]string, error) {
	output, err := apkm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apk",
		Args:    []string{"info"},
	})
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(output.STDOUT), "\n"), nil
}

func (apkm *ApkPackageManager) AddPackage(pkg string) error {
	_, err := apkm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apk",
		Args:    []string{"add", pkg},
	})
	return err
}

func (apkm *ApkPackageManager) RemovePackage(pkg string) error {
	_, err := apkm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apk",
		Args:    []string{"del", pkg},
	})
	return err
}

func (apkm *ApkPackageManager) UpgradePackage(pkg string) error {
	return apkm.AddPackage(pkg)
}

func (apkm *ApkPackageManager) CheckOSUpdates() ([]string, error) {
	_, err := apkm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apk",
		Args:    []string{"update"},
	})
	if err != nil {
		return nil, err
	}

	output, err := apkm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apk",
		Args:    []string{"version", "-v", "-l", "<"},
	})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output.STDOUT, "\n")
	var updates []string
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 0 {
			updates = append(updates, parts[0])
		}
	}
	return updates, nil
}

func (apkm *ApkPackageManager) UpgradeAll() ([]string, error) {
	_, err := apkm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "apk",
		Args:    []string{"upgrade"},
	})
	if err != nil {
		return nil, err
	}
	return apkm.CheckOSUpdates()
}

func (apkm *ApkPackageManager) EnsurePackagePresent(pkg string) error {
	packages, err := apkm.ListPackages()
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
	return apkm.AddPackage(pkg)
}

func (apkm *ApkPackageManager) EnsurePackageAbsent(pkg string) error {
	packages, err := apkm.ListPackages()
	if err != nil {
		return err
	}

	for _, installedPkg := range packages {
		if installedPkg == pkg {
			// Package is installed; proceed with removal
			return apkm.RemovePackage(pkg)
		}
	}
	// Package is not installed; return without taking action
	return nil
}

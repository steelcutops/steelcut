package packagemanager

import (
	"context"
	"strings"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type ApkPackageManager struct {
	CommandManager cm.CommandManager
}

func (apkm *ApkPackageManager) ListPackages(ctx context.Context) ([]string, error) {
	output, err := apkm.CommandManager.Run(ctx, cm.CommandConfig{
		Command: "apk",
		Args:    []string{"info"},
	})
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(output.STDOUT), "\n"), nil
}

func (apkm *ApkPackageManager) AddPackage(ctx context.Context, pkg string) error {
	_, err := apkm.CommandManager.Run(ctx, cm.CommandConfig{
		Command: "apk",
		Args:    []string{"add", pkg},
	})
	return err
}

func (apkm *ApkPackageManager) RemovePackage(ctx context.Context, pkg string) error {
	_, err := apkm.CommandManager.Run(ctx, cm.CommandConfig{
		Command: "apk",
		Args:    []string{"del", pkg},
	})
	return err
}

// APK doesn't have an explicit command for upgrading a single package.
// To upgrade a specific package, you would typically use the 'add' command,
// which will also upgrade packages.
func (apkm *ApkPackageManager) UpgradePackage(ctx context.Context, pkg string) error {
	return apkm.AddPackage(ctx, pkg)
}

func (apkm *ApkPackageManager) CheckOSUpdates(ctx context.Context) ([]string, error) {
	_, err := apkm.CommandManager.Run(ctx, cm.CommandConfig{
		Command: "apk",
		Args:    []string{"update"},
	})
	if err != nil {
		return nil, err
	}

	output, err := apkm.CommandManager.Run(ctx, cm.CommandConfig{
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

func (apkm *ApkPackageManager) UpgradeAll(ctx context.Context) error {
	_, err := apkm.CommandManager.Run(ctx, cm.CommandConfig{
		Command: "apk",
		Args:    []string{"upgrade"},
	})
	return err
}

package servicemanager

import (
	"context"
	"strings"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type LinuxServiceManager struct {
	CommandManager cm.CommandManager
}

func (lsm *LinuxServiceManager) EnableService(serviceName string) error {
	_, err := lsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "systemctl",
		Args:    []string{"enable", serviceName},
	})
	return err
}

func (lsm *LinuxServiceManager) DisableService(serviceName string) error {
	_, err := lsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "systemctl",
		Args:    []string{"disable", serviceName},
	})
	return err
}

func (lsm *LinuxServiceManager) StartService(serviceName string) error {
	_, err := lsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "systemctl",
		Args:    []string{"start", serviceName},
	})
	return err
}

func (lsm *LinuxServiceManager) StopService(serviceName string) error {
	_, err := lsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "systemctl",
		Args:    []string{"stop", serviceName},
	})
	return err
}

func (lsm *LinuxServiceManager) RestartService(serviceName string) error {
	_, err := lsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "systemctl",
		Args:    []string{"restart", serviceName},
	})
	return err
}

func (lsm *LinuxServiceManager) ReloadService(serviceName string) error {
	_, err := lsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "systemctl",
		Args:    []string{"reload", serviceName},
	})
	return err
}

func (lsm *LinuxServiceManager) CheckServiceStatus(serviceName string) (ServiceStatus, error) {
	output, err := lsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "systemctl",
		Args:    []string{"is-active", serviceName},
	})
	if err != nil {
		return "", err
	}
	switch strings.TrimSpace(output.STDOUT) {
	case "active":
		return Active, nil
	case "inactive":
		return Inactive, nil
	case "failed":
		return Failed, nil
	default:
		return "", err
	}
}

func (lsm *LinuxServiceManager) IsServiceEnabled(serviceName string) (bool, error) {
	output, err := lsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "systemctl",
		Args:    []string{"is-enabled", serviceName},
	})
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output.STDOUT) == "enabled", nil
}

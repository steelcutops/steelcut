package servicemanager

import (
	"context"

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

func (lsm *LinuxServiceManager) CheckServiceStatus(serviceName string) (string, error) {
	output, err := lsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "systemctl",
		Args:    []string{"is-active", serviceName},
	})
	if err != nil {
		return "", err
	}
	return output.STDOUT, nil
}

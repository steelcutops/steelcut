package servicemanager

import (
	"context"
	"fmt"
	"strings"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type DarwinServiceManager struct {
	CommandManager cm.CommandManager
}

func (dsm *DarwinServiceManager) EnableService(serviceName string) error {
	_, err := dsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "launchctl",
		Args:    []string{"bootstrap", "system", fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName)},
	})
	return err
}

func (dsm *DarwinServiceManager) DisableService(serviceName string) error {
	_, err := dsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "launchctl",
		Args:    []string{"bootout", "system", fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName)},
	})
	return err
}

func (dsm *DarwinServiceManager) StartService(serviceName string) error {
	_, err := dsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "launchctl",
		Args:    []string{"kickstart", "-k", fmt.Sprintf("system/%s", serviceName)},
	})
	return err
}

func (dsm *DarwinServiceManager) StopService(serviceName string) error {
	// To stop a service in Darwin without unloading it, we can simply use the 'kickstart -k' command without bootout
	_, err := dsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "launchctl",
		Args:    []string{"kickstart", "-k", fmt.Sprintf("system/%s", serviceName)},
	})
	return err
}

func (dsm *DarwinServiceManager) RestartService(serviceName string) error {
	return dsm.StartService(serviceName) // Since 'kickstart -k' restarts the service
}

func (dsm *DarwinServiceManager) ReloadService(serviceName string) error {
	// Darwin doesn't have a direct reload for launchctl services, but a restart can serve a similar purpose.
	return dsm.RestartService(serviceName)
}

func (dsm *DarwinServiceManager) CheckServiceStatus(serviceName string) (ServiceStatus, error) {
	output, err := dsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "launchctl",
		Args:    []string{"print", fmt.Sprintf("system/%s", serviceName)},
	})
	if err != nil {
		return "", err
	}
	if strings.Contains(output.STDOUT, "running") {
		return Active, nil
	} else {
		return Inactive, nil
	}
}

func (dsm *DarwinServiceManager) IsServiceEnabled(serviceName string) (bool, error) {
	// On Darwin, determining if a service is enabled is tricky. The service's plist presence in /Library/LaunchDaemons
	// doesn't guarantee it's enabled. This is a basic check and might not be 100% accurate.
	output, err := dsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "launchctl",
		Args:    []string{"print", fmt.Sprintf("system/%s", serviceName)},
	})
	if err != nil {
		return false, err
	}
	return strings.Contains(output.STDOUT, serviceName), nil
}

package servicemanager

import (
	"context"
	"fmt"
	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
	"strings"
)

type DarwinServiceManager struct {
	CommandManager cm.CommandManager
}

func (dsm *DarwinServiceManager) EnableService(serviceName string) error {
	// On macOS, "enabling" is akin to loading the service.
	_, err := dsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "launchctl",
		Args:    []string{"bootstrap", "system", fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName)},
	})
	return err
}

func (dsm *DarwinServiceManager) StartService(serviceName string) error {
	// Starting the service after it's loaded.
	_, err := dsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "launchctl",
		Args:    []string{"kickstart", "-k", fmt.Sprintf("system/%s", serviceName)},
	})
	return err
}

func (dsm *DarwinServiceManager) StopService(serviceName string) error {
	// Stopping the service. This doesn't unload it.
	_, err := dsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "launchctl",
		Args:    []string{"bootout", "system", fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName)},
	})
	return err
}

func (dsm *DarwinServiceManager) RestartService(serviceName string) error {
	// Restart by stopping and starting.
	err := dsm.StopService(serviceName)
	if err != nil {
		return err
	}
	return dsm.StartService(serviceName)
}

func (dsm *DarwinServiceManager) CheckServiceStatus(serviceName string) (string, error) {
	output, err := dsm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "launchctl",
		Args:    []string{"print", fmt.Sprintf("system/%s", serviceName)},
	})
	if err != nil {
		return "", err
	}
	// Parse the output for the specific status. This is a simplified check; more detailed parsing might be needed.
	if strings.Contains(output.STDOUT, "running") {
		return "running", nil
	} else {
		return "stopped", nil
	}
}

package environmentmanager

import (
	"context"
	"strings"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type UnixEnvironmentManager struct {
	CommandManager cm.CommandManager
}

func (e *UnixEnvironmentManager) Get(key string) (string, error) {
	output, err := e.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "printenv",
		Args:    []string{key},
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output.STDOUT), nil
}

func (e *UnixEnvironmentManager) Set(key, value string) error {
	_, err := e.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "export",
		Args:    []string{key + "=" + value},
	})
	return err
}

func (e *UnixEnvironmentManager) Unset(key string) error {
	_, err := e.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "unset",
		Args:    []string{key},
	})
	return err
}

func (e *UnixEnvironmentManager) List() (map[string]string, error) {
	output, err := e.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "printenv",
	})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output.STDOUT, "\n")
	envs := make(map[string]string)

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envs[parts[0]] = parts[1]
		}
	}

	return envs, nil
}

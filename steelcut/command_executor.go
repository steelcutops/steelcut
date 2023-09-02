package steelcut

import (
	"errors"
)

type DefaultCommandExecutor struct {
	Host    Host
	Options CommandOptions
}

func (dce DefaultCommandExecutor) RunCommand(command string, options CommandOptions) (string, error) {
	if options == (CommandOptions{}) { // if no specific options provided
		options = dce.Options // use the defaults
	}

	return dce.RunCommandWithOverride(command, options)
}

func (dce DefaultCommandExecutor) RunCommandWithOverride(command string, overrideOptions CommandOptions) (string, error) {
	if dce.Host == nil {
		return "", errors.New("host is not set in command executor")
	}

	finalOptions := dce.Options // Start with default options.

	// Override with provided options if necessary.
	if overrideOptions.UseSudo {
		finalOptions.UseSudo = overrideOptions.UseSudo
	}
	if overrideOptions.SudoPassword != "" {
		finalOptions.SudoPassword = overrideOptions.SudoPassword
	}

	return dce.Host.RunCommand(command, finalOptions)
}

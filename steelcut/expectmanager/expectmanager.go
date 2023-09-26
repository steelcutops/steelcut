package expectmanager

import (
	"context"
	"errors"
	"strings"
	"time"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type Expectation struct {
	Pattern     string
	Response    string
	Timeout     time.Duration
	OnTimeout   func() error
	OnSuccess   func() error
}

type ExpectManager struct {
	commandManager cm.CommandManager
}

func NewExpectManager(cm cm.CommandManager) *ExpectManager {
	return &ExpectManager{commandManager: cm}
}

// Interact initiates the interaction process and processes the expectations.
func (em *ExpectManager) Interact(ctx context.Context, cmdConfig cm.CommandConfig, expectations []Expectation) error {
	// Execute the command
	result, err := em.commandManager.Run(ctx, cmdConfig)
	if err != nil {
		return err
	}

	// For each expectation
	for _, exp := range expectations {
		// Using a timeout for each expectation
		expectationCtx, cancel := context.WithTimeout(ctx, exp.Timeout)
		defer cancel()

		// Check the output for the expected pattern
		if strings.Contains(result.STDOUT, exp.Pattern) {
			// If pattern found, run the success function if provided
			if exp.OnSuccess != nil {
				err := exp.OnSuccess()
				if err != nil {
					return err
				}
			}

			// Send the response
			_, err := em.commandManager.Run(expectationCtx, cm.CommandConfig{
				Command: "echo",
				Args:    []string{exp.Response},
			})
			if err != nil {
				return err
			}
		} else {
			// If timeout reached, run the timeout function if provided
			if exp.OnTimeout != nil {
				err := exp.OnTimeout()
				if err != nil {
					return err
				}
			} else {
				return errors.New("expectation pattern not found within the timeout")
			}
		}
	}

	return nil
}

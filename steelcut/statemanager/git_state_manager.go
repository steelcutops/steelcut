package statemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitStateManager struct {
	repoPath string
}

func NewGitStateManager(repoPath string) (*GitStateManager, error) {
	// Check if repoPath exists and is a git repository
	// Optionally: If not, initialize it as a new Git repo
	return &GitStateManager{repoPath: repoPath}, nil
}

func (g *GitStateManager) Save(ctx context.Context, state State) (string, error) {
	// Serialize state data to JSON
	data, err := json.Marshal(state.Data)
	if err != nil {
		return "", err
	}

	// Save to a file named after the resource_id
	filePath := filepath.Join(g.repoPath, state.ResourceID+".json")
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", err
	}

	// Commit the change
	err = g.gitCommit(fmt.Sprintf("Update state for %s by %s: %s", state.ResourceID, state.ChangedBy, state.Description))
	if err != nil {
		return "", err
	}

	return state.ID, nil
}

func (g *GitStateManager) Get(ctx context.Context, resourceID string, version ...int) (State, error) {
	// In the Git context, version can be mapped to commit hashes or tags.
	// This example just focuses on the latest version (HEAD).

	filePath := filepath.Join(g.repoPath, resourceID+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return State{}, err
	}

	var stateData map[string]interface{}
	err = json.Unmarshal(data, &stateData)
	if err != nil {
		return State{}, err
	}

	return State{ResourceID: resourceID, Data: stateData}, nil
}

func (g *GitStateManager) gitCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = g.repoPath
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (g *GitStateManager) List(ctx context.Context, filter ...string) ([]State, error) {
	files, err := os.ReadDir(g.repoPath)
	if err != nil {
		return nil, err
	}

	var states []State

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".json") {
			state, err := g.Get(ctx, strings.TrimSuffix(file.Name(), ".json"))
			if err != nil {
				return nil, err
			}
			states = append(states, state)
		}
	}

	return states, nil
}

func (g *GitStateManager) Delete(ctx context.Context, resourceID string) error {
	statePath := filepath.Join(g.repoPath, resourceID+".json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return fmt.Errorf("state with resource ID %s not found", resourceID)
	}

	err := os.Remove(statePath)
	if err != nil {
		return err
	}

	err = g.gitCommit(fmt.Sprintf("Deleted state for %s", resourceID))
	return err
}

func (g *GitStateManager) Exists(ctx context.Context, resourceID string) (bool, error) {
	statePath := filepath.Join(g.repoPath, resourceID+".json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

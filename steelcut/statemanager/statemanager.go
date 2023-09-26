package statemanager

import (
	"context"
	"time"
)

// State represents the data structure for a particular state.
type State struct {
	ID          string
	ResourceID  string                 // Unique identifier for a resource (server, network device, etc.)
	Version     int                    // Incremented each time this state is updated
	Timestamp   time.Time              // When this state was last updated
	Data        map[string]interface{} // The actual state data
	ChangedBy   string                 // Identifier for the user/tool that made the change
	Description string                 // Description or reason for the state change
}

// StateManager provides the interface for managing and interacting with states.
type StateManager interface {
	// Save stores the state and returns an ID for the stored state.
	Save(ctx context.Context, state State) (string, error)

	// Get retrieves the state for a given resource ID and version.
	// If version is omitted, it retrieves the latest state.
	Get(ctx context.Context, resourceID string, version ...int) (State, error)

	// List returns a list of states, possibly filtered by some criteria.
	List(ctx context.Context, filter map[string]interface{}) ([]State, error)

	// Lock locks a state for a given resource ID to prevent concurrent modifications.
	Lock(ctx context.Context, resourceID string) error

	// Unlock releases the lock for a given resource ID.
	Unlock(ctx context.Context, resourceID string) error

	// Subscribe returns a channel through which state changes are broadcasted.
	Subscribe(ctx context.Context) (<-chan State, error)

	// Validate checks if the stored state matches the actual environment and returns discrepancies.
	Validate(ctx context.Context, resourceID string) (map[string]interface{}, error)
}

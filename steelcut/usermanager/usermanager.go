package usermanager

// User represents an individual user account on the system.
type User struct {
	Username string // user login name
	UID      int    // user ID
	GID      int    // group ID
	Comment  string // user full name or comment
	HomeDir  string // user home directory
	Shell    string // user's shell
}

// UserManager encompasses operations related to user management.
type UserManager interface {
	// Fetches the details of a user based on username
	GetUser(username string) (User, error)

	// Adds a new user
	AddUser(user User) error

	// Modifies an existing user
	ModifyUser(user User) error

	// Deletes a user based on username
	DeleteUser(username string) error

	// Lists all users
	ListUsers() ([]User, error)
}

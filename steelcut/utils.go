package steelcut

import (
	"fmt"
	"os/user"
)

func setDefaultUserIfEmpty(host *UnixHost) error {
	if host.User != "" {
		return nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %v", err)
	}
	host.User = currentUser.Username
	return nil
}

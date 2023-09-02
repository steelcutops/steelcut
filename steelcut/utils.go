package steelcut

import (
	"fmt"
	"os/user"
	"strings"
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

func isOsType(os string, types ...string) bool {
	for _, t := range types {
		if strings.HasPrefix(os, t) {
			return true
		}
	}
	return false
}

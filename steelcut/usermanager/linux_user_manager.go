package usermanager

import (
	"context"
	"errors"
	"strconv"
	"strings"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type LinuxUserManager struct {
	CommandManager cm.CommandManager
}

func (l *LinuxUserManager) GetUser(username string) (User, error) {
	output, err := l.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "getent",
		Args:    []string{"passwd", username},
	})
	if err != nil {
		return User{}, err
	}

	parts := strings.Split(output.STDOUT, ":")
	if len(parts) < 7 {
		return User{}, errors.New("unexpected format")
	}

	uid, _ := strconv.Atoi(parts[2])
	gid, _ := strconv.Atoi(parts[3])

	return User{
		Username: parts[0],
		UID:      uid,
		GID:      gid,
		Comment:  parts[4],
		HomeDir:  parts[5],
		Shell:    parts[6],
	}, nil
}

func (l *LinuxUserManager) AddUser(user User) error {
	_, err := l.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "useradd",
		Args: []string{
			"-m",
			"-u", strconv.Itoa(user.UID),
			"-g", strconv.Itoa(user.GID),
			"-c", user.Comment,
			"-d", user.HomeDir,
			"-s", user.Shell,
			user.Username,
		},
	})
	return err
}

func (l *LinuxUserManager) ModifyUser(user User) error {
	_, err := l.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "usermod",
		Args: []string{
			"-u", strconv.Itoa(user.UID),
			"-g", strconv.Itoa(user.GID),
			"-c", user.Comment,
			"-d", user.HomeDir,
			"-s", user.Shell,
			user.Username,
		},
	})
	return err
}

func (l *LinuxUserManager) DeleteUser(username string) error {
	_, err := l.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "userdel",
		Args:    []string{"-r", username},
	})
	return err
}

func (l *LinuxUserManager) ListUsers() ([]User, error) {
	output, err := l.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "getent",
		Args:    []string{"passwd"},
	})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output.STDOUT, "\n")
	users := []User{}

	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}

		uid, _ := strconv.Atoi(parts[2])
		gid, _ := strconv.Atoi(parts[3])

		users = append(users, User{
			Username: parts[0],
			UID:      uid,
			GID:      gid,
			Comment:  parts[4],
			HomeDir:  parts[5],
			Shell:    parts[6],
		})
	}
	return users, nil
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/m-217/steelcut/steelcut"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

var (
	logFileName string
	debug       bool
	logger      = logrus.New()
)

func init() {
	flag.StringVar(&logFileName, "log", "log.txt", "Log file name")
	flag.BoolVar(&debug, "debug", false, "Enable debug log level")
}

func main() {
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatal(err)
	}
	defer file.Close()

	logger.SetOutput(file)
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	hostname := flag.String("hostname", "", "Hostname to connect to")
	username := flag.String("username", "", "Username to use for SSH connection")
	passwordPrompt := flag.Bool("password", false, "Use a password for SSH connection")
	keyPassPrompt := flag.Bool("keypass", false, "Passphrase for decrypting SSH keys")

	flag.Parse()

	var password, keyPass string
	if *passwordPrompt {
		fmt.Print("Enter the password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Failed to read password: %v", err)
		}
		password = string(passwordBytes)
		fmt.Println()
	}
	if *keyPassPrompt {
		fmt.Print("Enter the key passphrase: ")
		keyPassBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Failed to read key passphrase: %v", err)
		}
		keyPass = string(keyPassBytes)
		fmt.Println()
	}

	var options []steelcut.HostOption
	if *username != "" {
		options = append(options, steelcut.WithUser(*username))
	}
	if password != "" {
		options = append(options, steelcut.WithPassword(password))
	}
	if keyPass != "" {
		options = append(options, steelcut.WithKeyPassphrase(keyPass))
	}

	if *hostname == "" {
		*hostname = "localhost"
	}

	hostGroup := steelcut.NewHostGroup()

	hosts := []string{*hostname, "localhost"}
	for _, host := range hosts {
		server, err := steelcut.NewHost(host, options...)
		if err != nil {
			log.Fatalf("Failed to create new host: %v", err)
		}
		hostGroup.AddHost(server)
	}

	results, err := hostGroup.RunCommandOnAll("uname -a")
	if err != nil {
		log.Fatal(err)
	}

	for i, result := range results {
		fmt.Printf("Result for host %d: %s\n", i+1, result)
	}
}

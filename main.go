package main

import (
	"fmt"
	"log"

	"github.com/m-217/steelcut/steelcut"
)

func main() {
	// Create a LinuxHost from the steelcut library
	host := steelcut.UnixHost{
		Hostname: "localhost",
	}

	// Use the RunCommand method from the steelcut library
	output, err := host.RunCommand("uname -a")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Output:", output)
}

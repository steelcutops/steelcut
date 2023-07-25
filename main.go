package main

import (
	"fmt"
	"log"

	"github.com/m-217/steelcut/steelcut"
)

func main() {
	host, err := steelcut.NewHost("localhost")
	if err != nil {
		log.Fatal(err)
	}

	output, err := host.RunCommand("uname -a")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Output:", output)
}

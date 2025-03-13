package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/azyphon/markparsr"
)

func main() {
	// Command line flags
	readmePath := flag.String("path", "README.md", "Path to the README.md file to validate")
	flag.Parse()

	// Create validator
	validator, err := markparsr.NewReadmeValidator(*readmePath)
	if err != nil {
		fmt.Printf("Error creating validator: %v\n", err)
		os.Exit(1)
	}

	// Run validation
	errors := validator.Validate()
	if len(errors) > 0 {
		fmt.Printf("Found %d validation errors:\n", len(errors))
		for i, err := range errors {
			fmt.Printf("  %d. %v\n", i+1, err)
		}
		os.Exit(1)
	}

	fmt.Println("Readme validation successful! No errors found.")
}

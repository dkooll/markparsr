package markparsr_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/azyphon/markparsr"
)

// Example of using markparsr as a standalone tool
func Example_standalone() {
	readmePath := "README.md"

	validator, err := markparsr.NewReadmeValidator(readmePath)
	if err != nil {
		fmt.Printf("Error creating validator: %v\n", err)
		os.Exit(1)
	}

	errors := validator.Validate()
	if len(errors) > 0 {
		fmt.Printf("Found %d validation errors\n", len(errors))
		os.Exit(1)
	}

	fmt.Println("Validation successful")
}

// Example of using markparsr in a test
func Example_test() {
	func() {
		readmePath := filepath.Join("..", "README.md")

		validator, err := markparsr.NewReadmeValidator(readmePath)
		if err != nil {
			fmt.Printf("Failed to create validator\n")
			return
		}

		errors := validator.Validate()
		if len(errors) > 0 {
			fmt.Printf("Found documentation errors\n")
			return
		}

		fmt.Println("Documentation valid")
	}()
}

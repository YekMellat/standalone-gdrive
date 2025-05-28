package fs

import (
	"fmt"
)

// Option is describes an option for the config system
type Option struct {
	Name     string      // name of the option in snake_case
	Help     string      // Help text - may contain newlines, will be wrapped
	Provider string      // Set to filter on provider
	Default  interface{} // default value, nil => ""
	Examples []OptionExample
}

// OptionExamples is a slice of examples
type OptionExamples []OptionExample

// OptionExample describes an example for an Option
type OptionExample struct {
	Value string
	Help  string
}

// OptionDefinition defines the possible options for a backend
type OptionDefinition []Option

// Register registers an Fs with the given name
func Register(info *RegInfo) {
	// Just a stub for compatibility
	fmt.Printf("Registered backend %q: %s\n", info.Name, info.Description)
}

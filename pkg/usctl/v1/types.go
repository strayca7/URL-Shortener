package v1

import (
	"strings"
)

const (
	Version = "v0.0.3beta"
)

var (
	resources = []string{"role", "rolebinding", "user", "publicurl", "userurl", "url"}
)

// command represents the supported commands, such as get and apply.
type command struct {
	Name        string
	Description string
	Handler     func(...string)
}

// cli manages the terminal's status.
type cli struct {
	// Commands map stores the command of usctl.
	// Every command is a command type. It contains the command's name, description,
	// and its own handler that will be used to execute the command logic.
	Commands map[string]command

	// debug indicates whether the debug mode is enabled.
	Debug   bool
	Help    bool
	Version bool
}

// Special fields of apply command.
type applyFlags struct {
	outputFormat string
	filePaths    []string
	showHelp     bool
	// Only print the resource list to the terminal but not send to API Server.
	dryRun bool
}

type getFlags struct {
	namespace     string
	allNamespaces bool
}

// Custom []string type.
// To satisfied the pflag.Value interface.
type stringSlice []string

func (s *stringSlice) String() string {
	if s == nil {
		return ""
	}
	return strings.Join(*s, " ")
}

func (s *stringSlice) Set(value string) error {
	paths := strings.SplitSeq(value, " ")
	for path := range paths {
		trimmedPath := strings.TrimSpace(path)
		if trimmedPath != "" {
			*s = append(*s, trimmedPath)
		}
	}
	return nil
}

func (s *stringSlice) Type() string {
	return "stringSlice"
}

type TypedResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

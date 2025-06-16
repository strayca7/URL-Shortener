package v1

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/pflag"
)

func PrintGlobalHelp() {
	fmt.Println("URL Shortener Command Line Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  usctl [command] [flags]")

	fmt.Println("\nAvailable Commands:")
	longestName := 0
	for _, cmd := range CLI.Commands {
		if len(cmd.Name) > longestName {
			longestName = len(cmd.Name)
		}
	}

	var cmdNames []string
	for name := range CLI.Commands {
		cmdNames = append(cmdNames, name)
	}
	sort.Strings(cmdNames)

	for _, name := range cmdNames {
		cmd := CLI.Commands[name]
		padding := strings.Repeat(" ", longestName-len(cmd.Name)+2)
		fmt.Printf("  %s%s%s\n", cmd.Name, padding, cmd.Description)
	}

	fmt.Println("\nCommon Flags:")
	fmt.Println("  -h, --help                   Print help information")
	fmt.Println("  -v  --version                Print version information and quit")

	fmt.Println("\nUse \"usctl [command] --help\" for more information about a command")

	fmt.Println("\nExamples:")
	fmt.Println("  # Apply a configuration file")
	fmt.Println("  usctl apply -f config.yaml")
	fmt.Println("\n  # Get all short URLs")
	fmt.Println("  usctl get shorturl")
}

func PrintHelp() {
	fmt.Printf("\n")
	fmt.Println("Run 'usctl --help' for more information")
}

func PrintApplyHelp(fs *pflag.FlagSet) {
	fmt.Println("Apply a configuration to a resource")

	fmt.Println("\nUsage:")
	fmt.Printf("  usctl apply [flags] -f FILENAME\n")

	fmt.Println("\nExamples:")
	fmt.Printf("  # Apply the configuration in manifest.yaml\n")
	fmt.Printf("  usctl apply -f manifest.yaml\n\n")
	fmt.Printf("  # Apply the configurations in multiple files\n")
	fmt.Printf("  usctl apply -f \"manifest.yaml service.yaml\"\n\n")
	fmt.Printf("  # Apply with JSON output format\n")
	fmt.Printf("  usctl apply -f manifest.yaml -o json\n\n")
	fmt.Printf("  # Apply in dry-run mode\n")
	fmt.Printf("  usctl apply -f manifest.yaml --dry-run\n")

	fmt.Println("\nFlags:")
	fs.VisitAll(func(f *pflag.Flag) {
		var usage string
		if f.Shorthand == "" {
			usage = fmt.Sprintf("      --%s", f.Name)
		} else {
			usage = fmt.Sprintf("  -%s, --%s", f.Shorthand, f.Name)
		}

		if len(usage) < 20 {
			usage += strings.Repeat(" ", 20-len(usage))
		}

		if f.DefValue != "" {
			usage += fmt.Sprintf("[default: %s]  ", f.DefValue)
		}

		usage += f.Usage

		fmt.Println(usage)
	})
}

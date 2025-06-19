package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"url-shortener/pkg/apis/rbac"
	"url-shortener/pkg/usctl"
	"url-shortener/pkg/usctl/praser"

	"github.com/spf13/pflag"
)

var CLI = &cli{
	Commands: map[string]command{
		"apply": {
			Name:        "apply",
			Description: "Apply configuration to resources",
			Handler:     applyHandler,
		},
		"get": {
			Name:        "get",
			Description: "Get all specified resources. Only one can be specified at a time.",
			Handler:     getHandler,
		},
	},
	Debug:   false,
	Help:    false,
	Version: false,
}

// CLIHandler is a external used method.
// It is the entrance of usctl. It receives all arguments and parse them.
func CLIHandler(args ...string) {
	usctlCmd := pflag.NewFlagSet("usctl", pflag.ExitOnError)
	usctlCmd.BoolVarP(&CLI.Help, "help", "h", false, "Print help information and quit")
	usctlCmd.BoolVarP(&CLI.Version, "version", "v", false, "Print version information and quit")

	usctlCmd.Usage = func() {
		PrintGlobalHelp()
		// usctlCmd.PrintDefaults()
	}

	usctlCmd.Parse(args[:2])
	if CLI.Help {
		PrintGlobalHelp()
		os.Exit(0)
	}
	if CLI.Version {
		fmt.Println("usctl version: " + Version)
		os.Exit(0)
	}

	// If the second argument is not a flag, then handle them in CLI.Command
	cmd, exists := CLI.Commands[args[1]]
	if !exists {
		fmt.Fprintln(os.Stderr, "usctl: unknown command: "+args[1])
		PrintHelp()
		os.Exit(1)
	}
	cmd.Handler(args[2:]...)
}

// applyHandler is a Command handler for apply operation.
func applyHandler(args ...string) {
	var files stringSlice
	applyFlags := &applyFlags{}

	applyCmd := pflag.NewFlagSet("apply", pflag.ExitOnError)
	applyCmd.VarP(&files, "file", "f", "YAML file path, separate multiple files with Spaces")
	applyCmd.StringVarP(&applyFlags.outputFormat, "output", "o", "", "Output format YAML or JSON")
	applyCmd.BoolVarP(&applyFlags.showHelp, "help", "h", false, "Print help message")
	applyCmd.BoolVarP(&applyFlags.dryRun, "dry-run", "", false, "Only print the manifest list but not send to API Server")

	applyCmd.Parse(args)

	if applyFlags.showHelp {
		PrintApplyHelp(applyCmd)
		return
	}

	applyFlags.filePaths = files

	if len(applyFlags.filePaths) == 0 {
		fmt.Fprintln(os.Stderr, "error: must specify one of -f")
		PrintHelp()
		os.Exit(1)
	}
	// fmt.Println(files)
	if err := processApplyCommand(applyFlags); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func processApplyCommand(flags *applyFlags) error {
	for _, file := range flags.filePaths {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("%w: \""+file+"\"", usctl.ErrFileNotFound)
		}
		praser := praser.ResourceParser{}
		KindAndName, resources, printResources, err := praser.Parse(data, flags.outputFormat)
		if err != nil {
			return fmt.Errorf("%w: \""+file+"\"", err)
		}

		if flags.outputFormat != "" {
			switch flags.outputFormat {
			case "yaml":
				fmt.Println(string(printResources))
			case "json":
				fmt.Println(string(resources))
			default:
				fmt.Fprintf(os.Stderr, "usctl: unknown output type: %s\n", flags.outputFormat)
				PrintHelp()
				os.Exit(1)
			}
		}
		fmt.Println(KindAndName + " created")
	}
	return nil
}

func getHandler(args ...string) {
	getFlags := &getFlags{}

	getCmd := pflag.NewFlagSet("get", pflag.ExitOnError)
	getCmd.BoolVarP(&getFlags.allNamespaces, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified explicitly.")

	getCmd.Parse(args)

	kinds := getCmd.Args()
	for _, kind := range kinds {
		if kind == "role" {
			handleGetRoles()
			continue
		} else if kind == "rolebinding" {
			handleGetRoleBindings()
			continue
		} else if kind == "user" {
			fmt.Println("user")
			continue
		} else {
			fmt.Fprintln(os.Stderr, "invalid resource type")
			os.Exit(1)
		}
	}

}

// wait for check
func handleGetRoles() {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	req, err := http.NewRequest("GET", "http://localhost:8080/rbac/v1/role", nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	var body TypedResponse[[]rbac.Role]
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	roles := body.Data

	if !body.Success {
		fmt.Fprintln(os.Stderr, body.Error)
		os.Exit(1)
	}

	p := NewTablePrinter(os.Stdout)
	p.PrintHeader("Namespace", "Name")
	for _, role := range roles {
		p.PrintRow(role.Namespace, role.Name)
	}
	p.Flush()
}

func handleGetRoleBindings() {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	req, err := http.NewRequest("GET", "http://localhost:8080/rbac/v1/rolebinding", nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	var body TypedResponse[[]rbac.RoleBinding]
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	roleBindings := body.Data

	if !body.Success {
		fmt.Fprintln(os.Stderr, body.Error)
		os.Exit(1)
	}

	p := NewTablePrinter(os.Stdout)
	p.PrintHeader("Namespace", "Name")
	for _, roleBinding := range roleBindings {
		p.PrintRow(roleBinding.Namespace, roleBinding.Name)
	}
	p.Flush()
}

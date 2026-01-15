package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/oroddlokken/preflag/preprocessor"
	_ "github.com/oroddlokken/preflag/preprocessor/emojiexample"
	_ "github.com/oroddlokken/preflag/preprocessor/url2hostname"
)

// Version information - set via ldflags at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func shellQuote(s string) string {
	if strings.ContainsAny(s, " \t\n\"'\\$`!") {
		// Use single quotes, escaping any single quotes in the string
		escaped := strings.ReplaceAll(s, "'", "'\"'\"'")
		return "'" + escaped + "'"
	}
	return s
}

func printVersion(name string) {
	fmt.Printf("%s %s\n", name, version)
	if commit != "none" {
		fmt.Printf("  commit: %s\n", commit)
	}
	if date != "unknown" {
		fmt.Printf("  built:  %s\n", date)
	}
	fmt.Println("\nCopyright (c) 2026 Ove Ragnar Oddløkken")
	fmt.Println("Licensed under the MIT License")
}

func printLicense() {
	licenseText := `MIT License

Copyright (c) 2026 Ove Ragnar Oddløkken

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`
	fmt.Println(licenseText)
}

// CommandPreprocessors maps commands to their preprocessors
type CommandPreprocessors map[string][]string

// parseWithReflagInitArgs parses --init-with-reflag arguments
// Arguments starting with + or - are reflag arguments
// Arguments with : are preflag arguments (command:preprocessor)
func parseWithReflagInitArgs(args []string) (shell string, cmdPreprocessors CommandPreprocessors, reflagArgs []string) {
	shell = "bash"
	cmdPreprocessors = make(CommandPreprocessors)

	if len(args) == 0 {
		return
	}

	startIdx := 0
	// First arg could be shell
	if args[0] == "bash" || args[0] == "zsh" || args[0] == "fish" {
		shell = args[0]
		startIdx = 1
	}

	// Parse arguments
	for i := startIdx; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "+") || strings.HasPrefix(arg, "-") {
			// Reflag argument
			reflagArgs = append(reflagArgs, arg)
		} else if strings.Contains(arg, ":") {
			// Preflag argument (command:preprocessor)
			parts := strings.SplitN(arg, ":", 2)
			cmd := parts[0]
			preprocessorStr := parts[1]

			var preprocessors []string
			if strings.Contains(preprocessorStr, ",") {
				preprocessors = strings.Split(preprocessorStr, ",")
			} else {
				preprocessors = []string{preprocessorStr}
			}

			cmdPreprocessors[cmd] = preprocessors
		}
	}
	return
}

// parseInitArgs parses --init arguments, returning shell type and command:preprocessor mappings
func parseInitArgs(args []string) (shell string, cmdPreprocessors CommandPreprocessors) {
	shell = "bash"
	cmdPreprocessors = make(CommandPreprocessors)

	if len(args) == 0 {
		return
	}

	startIdx := 0
	// First arg could be shell
	if args[0] == "bash" || args[0] == "zsh" || args[0] == "fish" {
		shell = args[0]
		startIdx = 1
	}

	// Parse command:preprocessor mappings
	for i := startIdx; i < len(args); i++ {
		mapping := args[i]
		if strings.Contains(mapping, ":") {
			parts := strings.SplitN(mapping, ":", 2)
			cmd := parts[0]
			preprocessorStr := parts[1]

			var preprocessors []string
			if strings.Contains(preprocessorStr, ",") {
				preprocessors = strings.Split(preprocessorStr, ",")
			} else {
				preprocessors = []string{preprocessorStr}
			}

			cmdPreprocessors[cmd] = preprocessors
		}
	}
	return
}

func printWithReflagInit(shell string, cmdPreprocessors CommandPreprocessors, reflagArgs []string) {
	if len(cmdPreprocessors) == 0 {
		fmt.Fprintln(os.Stderr, "preflag --init-with-reflag: no command:preprocessor mappings specified")
		fmt.Fprintln(os.Stderr, "Usage: preflag --init-with-reflag [shell] command:preprocessor [+reflag|-reflag] ...")
		fmt.Fprintln(os.Stderr, "Example: preflag --init-with-reflag zsh dig:url2hostname +dig2doggo")
		os.Exit(1)
	}

	// Sort commands for consistent output
	var commands []string
	for cmd := range cmdPreprocessors {
		commands = append(commands, cmd)
	}
	sort.Strings(commands)

	switch shell {
	case "fish":
		fmt.Fprintln(os.Stderr, "preflag --init-with-reflag: fish shell not yet supported for chaining")
		os.Exit(1)
	default: // bash, zsh
		fmt.Println("# preflag + reflag chaining setup")
		fmt.Println()

		// Step 1: Initialize preflag
		fmt.Println("# Step 1: Initialize preflag")
		for _, cmd := range commands {
			preprocessors := cmdPreprocessors[cmd]
			preprocessorStr := strings.Join(preprocessors, ",")
			fmt.Printf("__preflag_orig_%s() { command %s \"$@\"; }\n", cmd, cmd)
			fmt.Printf("__preflag_%s_impl() {\n", cmd)
			fmt.Printf("    local args\n")
			fmt.Printf("    args=$(preflag %s \"$@\")\n", preprocessorStr)
			fmt.Printf("    if declare -f __preflag_%s_next >/dev/null 2>&1; then\n", cmd)
			fmt.Printf("        eval \"__preflag_%s_next $args\"\n", cmd)
			fmt.Printf("    else\n")
			fmt.Printf("        eval \"__preflag_orig_%s $args\"\n", cmd)
			fmt.Printf("    fi\n")
			fmt.Printf("}\n")
			fmt.Printf("%s() { __preflag_%s_impl \"$@\"; }\n\n", cmd, cmd)
		}

		// Step 2: Initialize reflag (this will overwrite the command functions)
		fmt.Println("# Step 2: Initialize reflag (overwrites command functions)")
		reflagArgsStr := strings.Join(reflagArgs, " ")
		fmt.Printf("eval \"$(reflag --init %s %s)\"\n\n", shell, reflagArgsStr)

		// Step 3: Create next functions to capture reflag's logic
		fmt.Println("# Step 3: Create next functions to capture reflag's logic")
		for _, cmd := range commands {
			// Extract target command from reflag args
			// Look for +cmd2target pattern
			targetCmd := ""
			for _, arg := range reflagArgs {
				if strings.HasPrefix(arg, "+") {
					translatorName := strings.TrimPrefix(arg, "+")
					// Check if this translator starts with our command
					if strings.HasPrefix(translatorName, cmd+"2") {
						parts := strings.SplitN(translatorName, "2", 2)
						if len(parts) == 2 {
							targetCmd = parts[1]
							break
						}
					}
				}
			}

			if targetCmd != "" {
				fmt.Printf("__preflag_%s_next() { eval \"$(reflag %s %s \"$@\")\"; }\n", cmd, cmd, targetCmd)
			}
		}
		fmt.Println()

		// Step 4: Restore preflag wrappers
		fmt.Println("# Step 4: Restore preflag wrappers")
		for _, cmd := range commands {
			fmt.Printf("%s() { __preflag_%s_impl \"$@\"; }\n", cmd, cmd)
		}
	}
}

func printInit(shell string, cmdPreprocessors CommandPreprocessors) {
	if len(cmdPreprocessors) == 0 {
		fmt.Fprintln(os.Stderr, "preflag --init: no command:preprocessor mappings specified")
		fmt.Fprintln(os.Stderr, "Usage: preflag --init [shell] command:preprocessor1,preprocessor2 ...")
		fmt.Fprintln(os.Stderr, "Example: preflag --init zsh ping:url2hostname whois:url2hostname")
		os.Exit(1)
	}

	// Sort commands for consistent output
	var commands []string
	for cmd := range cmdPreprocessors {
		commands = append(commands, cmd)
	}
	sort.Strings(commands)

	switch shell {
	case "fish":
		for _, cmd := range commands {
			preprocessors := cmdPreprocessors[cmd]
			preprocessorStr := strings.Join(preprocessors, ",")
			fmt.Printf("alias %s='preflag %s'\n", cmd, preprocessorStr)
		}
	default: // bash, zsh
		for _, cmd := range commands {
			preprocessors := cmdPreprocessors[cmd]
			preprocessorStr := strings.Join(preprocessors, ",")
			fmt.Printf("__preflag_orig_%s() { command %s \"$@\"; }\n", cmd, cmd)
			fmt.Printf("__preflag_%s_impl() {\n", cmd)
			fmt.Printf("    local args\n")
			fmt.Printf("    args=$(preflag %s \"$@\")\n", preprocessorStr)
			fmt.Printf("    if declare -f __preflag_%s_next >/dev/null 2>&1; then\n", cmd)
			fmt.Printf("        eval \"__preflag_%s_next $args\"\n", cmd)
			fmt.Printf("    else\n")
			fmt.Printf("        eval \"__preflag_orig_%s $args\"\n", cmd)
			fmt.Printf("    fi\n")
			fmt.Printf("}\n")
			fmt.Printf("%s() { __preflag_%s_impl \"$@\"; }\n\n", cmd, cmd)
		}
	}
}

func printUsage() {
	fmt.Println("preflag - preprocess command-line arguments")
	fmt.Println()
	fmt.Println("Quick setup:")
	fmt.Println("  eval \"$(preflag --init zsh ping:url2hostname whois:url2hostname)\" >> ~/.zshrc")
	fmt.Println()
	fmt.Println("Quick setup with reflag chaining:")
	fmt.Println("  eval \"$(preflag --init-with-reflag zsh dig:url2hostname +dig2doggo)\" >> ~/.zshrc")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  preflag --init [shell] command:preprocessor1,preprocessor2 ...")
	fmt.Println("      Generate shell aliases/functions for commands with preprocessors")
	fmt.Println("      shell: bash (default), zsh, fish")
	fmt.Println("      command:preprocessor: map a command to one or more preprocessors (comma-separated)")
	fmt.Println()
	fmt.Println("  preflag --init-with-reflag [shell] command:preprocessor [+reflag|-reflag] ...")
	fmt.Println("      Generate complete preflag + reflag chaining setup")
	fmt.Println("      shell: bash (default), zsh")
	fmt.Println("      command:preprocessor: preflag mappings (contains ':')")
	fmt.Println("      +reflag/-reflag: reflag translators to enable/disable (starts with '+' or '-')")
	fmt.Println()
	fmt.Println("  preflag <preprocessors> [args...]")
	fmt.Println("      Preprocess arguments using specified preprocessors")
	fmt.Println("      preprocessors: comma-separated list (e.g., url2hostname,preprocessor2)")
	fmt.Println("      Outputs shell-quoted arguments ready for eval")
	fmt.Println()
	fmt.Println("  preflag --list")
	fmt.Println("      List all available preprocessors")
	fmt.Println()
	fmt.Println("  preflag --version")
	fmt.Println("      Print version information")
	fmt.Println()
	fmt.Println("  preflag --help")
	fmt.Println("      Print this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  preflag --init zsh ping:url2hostname whois:url2hostname")
	fmt.Println("      → Creates ping() and whois() functions in zsh")
	fmt.Println()
	fmt.Println("  preflag --init-with-reflag zsh dig:url2hostname +dig2doggo")
	fmt.Println("      → Creates complete chaining setup: dig URL → hostname → doggo")
	fmt.Println()
	fmt.Println("  preflag url2hostname https://vg.no")
	fmt.Println("      → vg.no")
}

// processArgs applies specified preprocessors to the arguments
func processArgs(cmd string, preprocessorNames []string, args []string) []string {
	result := args
	for _, name := range preprocessorNames {
		p := preprocessor.GetByName(name)
		if p != nil {
			result = p.Process(result)
		}
	}
	return result
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		return
	}

	// Handle preflag's own flags
	switch args[0] {
	case "--version", "-v":
		printVersion("preflag")
		return
	case "--license":
		printLicense()
		return
	case "--help", "-h":
		printUsage()
		return
	case "--list", "-l":
		preprocessor.PrintTable(os.Stdout)
		return
	case "--init":
		shell, preprocessors := parseInitArgs(args[1:])
		printInit(shell, preprocessors)
		return
	case "--init-with-reflag":
		shell, preprocessors, reflagArgs := parseWithReflagInitArgs(args[1:])
		printWithReflagInit(shell, preprocessors, reflagArgs)
		return
	default:
		// New simplified format: preflag preprocessor1,preprocessor2 args...
		// First arg should contain preprocessor names
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "preflag: missing preprocessors")
			os.Exit(1)
		}

		var preprocessorNames []string
		var cmdArgs []string

		// Parse preprocessor list (format: "preprocessor1,preprocessor2 args...")
		if strings.Contains(args[0], ",") {
			// Multiple preprocessors
			preprocessorNames = strings.Split(args[0], ",")
			cmdArgs = args[1:]
		} else {
			// Single preprocessor name
			preprocessorNames = []string{args[0]}
			cmdArgs = args[1:]
		}

		processed := processArgs("", preprocessorNames, cmdArgs)
		// Output shell-quoted arguments
		quoted := make([]string, len(processed))
		for i, arg := range processed {
			quoted[i] = shellQuote(arg)
		}
		fmt.Println(strings.Join(quoted, " "))
		return
	}

}

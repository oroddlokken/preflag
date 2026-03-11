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

// parseInitArgs parses --init arguments, returning shell type, command:preprocessor mappings, and reflag args
// Arguments starting with + are reflag translator arguments
// Arguments with : are preflag arguments (command:preprocessor)
func parseInitArgs(args []string) (shell string, cmdPreprocessors CommandPreprocessors, reflagArgs []string) {
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
		if strings.HasPrefix(arg, "+") {
			// Reflag translator argument
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

func printInit(shell string, cmdPreprocessors CommandPreprocessors, reflagArgs []string) {
	// Validate that at least one feature is specified
	if len(cmdPreprocessors) == 0 && len(reflagArgs) == 0 {
		fmt.Fprintln(os.Stderr, "preflag --init: no features specified")
		fmt.Fprintln(os.Stderr, "Usage: preflag --init [shell] [command:preprocessor...] [+translator...]")
		fmt.Fprintln(os.Stderr, "Example: preflag --init zsh ping:url2hostname +cat2bat")
		os.Exit(1)
	}

	// Sort commands for consistent output
	var commands []string
	for cmd := range cmdPreprocessors {
		commands = append(commands, cmd)
	}
	sort.Strings(commands)

	// Handle reflag-only case: use reflag --init directly instead
	if len(cmdPreprocessors) == 0 && len(reflagArgs) > 0 {
		fmt.Fprintln(os.Stderr, "preflag --init: reflag-only commands should use 'reflag --init' directly")
		fmt.Fprintln(os.Stderr, "Example: eval \"$(reflag --init zsh)\"")
		os.Exit(1)
	}

	// Handle preflag-only case (no reflag args)
	if len(reflagArgs) == 0 {
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
		return
	}

	// Handle combined preflag+reflag case
	switch shell {
	default: // bash, zsh
		for _, cmd := range commands {
			preprocessors := cmdPreprocessors[cmd]
			preprocessorStr := strings.Join(preprocessors, ",")

			// Extract target command from reflag args for this command
			targetCmd := ""
			for _, arg := range reflagArgs {
				if strings.HasPrefix(arg, "+") {
					translatorName := strings.TrimPrefix(arg, "+")
					if strings.HasPrefix(translatorName, cmd+"2") {
						parts := strings.SplitN(translatorName, "2", 2)
						if len(parts) == 2 {
							targetCmd = parts[1]
							break
						}
					}
				}
			}

			fmt.Printf("__preflag_orig_%s() { command %s \"$@\"; }\n", cmd, cmd)

			if targetCmd != "" {
				// Chain: preprocess args, then translate via reflag
				fmt.Printf("%s() {\n", cmd)
				fmt.Printf("    local args\n")
				fmt.Printf("    args=$(preflag %s \"$@\")\n", preprocessorStr)
				fmt.Printf("    eval \"$(reflag %s %s $args)\"\n", cmd, targetCmd)
				fmt.Printf("}\n")
			} else {
				// Preprocess only, call original command
				fmt.Printf("%s() {\n", cmd)
				fmt.Printf("    local args\n")
				fmt.Printf("    args=$(preflag %s \"$@\")\n", preprocessorStr)
				fmt.Printf("    eval \"__preflag_orig_%s $args\"\n", cmd)
				fmt.Printf("}\n")
			}
		}
	}
}

func printUsage() {
	fmt.Println("preflag - preprocess command-line arguments")
	fmt.Println()
	fmt.Println("Quick setup:")
	fmt.Println("  eval \"$(preflag --init zsh ping:url2hostname whois:url2hostname)\" >> ~/.zshrc")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  preflag --init [shell] [command:preprocessor...] [+translator...]")
	fmt.Println("      Generate shell aliases/functions for commands")
	fmt.Println("      shell: bash (default), zsh, fish")
	fmt.Println("      command:preprocessor: preflag mappings (contains ':')")
	fmt.Println("      +translator: reflag translators (starts with '+')")
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
	fmt.Println("  preflag --init zsh nslookup:url2hostname")
	fmt.Println("      → Creates nslookup() function with url2hostname preprocessor")
	fmt.Println()
	fmt.Println("  preflag --init zsh dig:url2hostname +dig2doggo")
	fmt.Println("      → Creates dig() function with chained preflag+reflag: URL → hostname → doggo")
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
		shell, preprocessors, reflagArgs := parseInitArgs(args[1:])
		printInit(shell, preprocessors, reflagArgs)
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

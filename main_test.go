package main

import (
	"bytes"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/oroddlokken/preflag/preprocessor"
	_ "github.com/oroddlokken/preflag/preprocessor"
	_ "github.com/oroddlokken/preflag/preprocessor/emojiexample"
	_ "github.com/oroddlokken/preflag/preprocessor/url2hostname"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with space", "'with space'"},
		{"with\ttab", "'with\ttab'"},
		{"with\nnewline", "'with\nnewline'"},
		{"with'quote", "'with'\"'\"'quote'"},
		{"with\"double", "'with\"double'"},
		{"with$dollar", "'with$dollar'"},
		{"with`backtick", "'with`backtick'"},
		{"with\\backslash", "'with\\backslash'"},
		{"with!exclaim", "'with!exclaim'"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := shellQuote(tt.input)
			if result != tt.expected {
				t.Errorf("shellQuote(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseInitArgs(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		expectedShell       string
		expectedCmdPreprocs map[string][]string
		expectedReflagArgs  []string
	}{
		{
			name:                "no args defaults to bash with no mappings",
			args:                []string{},
			expectedShell:       "bash",
			expectedCmdPreprocs: map[string][]string{},
			expectedReflagArgs:  []string{},
		},
		{
			name:                "shell only with no mappings",
			args:                []string{"fish"},
			expectedShell:       "fish",
			expectedCmdPreprocs: map[string][]string{},
			expectedReflagArgs:  []string{},
		},
		{
			name:          "shell with single command:preprocessor",
			args:          []string{"bash", "ping:url2hostname"},
			expectedShell: "bash",
			expectedCmdPreprocs: map[string][]string{
				"ping": {"url2hostname"},
			},
			expectedReflagArgs: []string{},
		},
		{
			name:          "shell with multiple preprocessors for one command",
			args:          []string{"zsh", "ping:url2hostname,preprocessor2"},
			expectedShell: "zsh",
			expectedCmdPreprocs: map[string][]string{
				"ping": {"url2hostname", "preprocessor2"},
			},
			expectedReflagArgs: []string{},
		},
		{
			name:          "multiple command:preprocessor mappings",
			args:          []string{"bash", "ping:url2hostname", "whois:url2hostname"},
			expectedShell: "bash",
			expectedCmdPreprocs: map[string][]string{
				"ping":  {"url2hostname"},
				"whois": {"url2hostname"},
			},
			expectedReflagArgs: []string{},
		},
		{
			name:          "command:preprocessor without shell",
			args:          []string{"ping:url2hostname"},
			expectedShell: "bash",
			expectedCmdPreprocs: map[string][]string{
				"ping": {"url2hostname"},
			},
			expectedReflagArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell, cmdPreprocs, reflagArgs := parseInitArgs(tt.args)
			if shell != tt.expectedShell {
				t.Errorf("parseInitArgs(%v) shell = %q, want %q", tt.args, shell, tt.expectedShell)
			}
			if len(cmdPreprocs) != len(tt.expectedCmdPreprocs) {
				t.Errorf("parseInitArgs(%v) got %d mappings, want %d", tt.args, len(cmdPreprocs), len(tt.expectedCmdPreprocs))
			}
			for cmd, expectedPreprocs := range tt.expectedCmdPreprocs {
				if gotPreprocs, ok := cmdPreprocs[cmd]; !ok {
					t.Errorf("parseInitArgs(%v) missing command %q", tt.args, cmd)
				} else if !slices.Equal(gotPreprocs, expectedPreprocs) {
					t.Errorf("parseInitArgs(%v) command %q: got %v, want %v", tt.args, cmd, gotPreprocs, expectedPreprocs)
				}
			}
			if !slices.Equal(reflagArgs, tt.expectedReflagArgs) {
				t.Errorf("parseInitArgs(%v) reflag args = %v, want %v", tt.args, reflagArgs, tt.expectedReflagArgs)
			}
		})
	}
}

func TestPreprocessorRegistry(t *testing.T) {
	// url2hostname should be registered via init()
	p := preprocessor.GetByName("url2hostname")
	if p == nil {
		t.Fatal("url2hostname preprocessor not registered")
	}

	if p.Name() != "url2hostname" {
		t.Errorf("Name() = %q, want %q", p.Name(), "url2hostname")
	}

	// Test List
	names := preprocessor.List()
	if !slices.Contains(names, "url2hostname") {
		t.Error("url2hostname not found in List()")
	}
}

func TestParseInitArgsWithReflag(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		expectedShell       string
		expectedCmdPreprocs map[string][]string
		expectedReflagArgs  []string
	}{
		{
			name:                "no args defaults to bash with no mappings",
			args:                []string{},
			expectedShell:       "bash",
			expectedCmdPreprocs: map[string][]string{},
			expectedReflagArgs:  []string{},
		},
		{
			name:                "shell only with no mappings",
			args:                []string{"zsh"},
			expectedShell:       "zsh",
			expectedCmdPreprocs: map[string][]string{},
			expectedReflagArgs:  []string{},
		},
		{
			name:          "single preflag mapping with single reflag arg",
			args:          []string{"zsh", "dig:url2hostname", "+dig2doggo"},
			expectedShell: "zsh",
			expectedCmdPreprocs: map[string][]string{
				"dig": {"url2hostname"},
			},
			expectedReflagArgs: []string{"+dig2doggo"},
		},
		{
			name:          "multiple preflag mappings with multiple reflag args",
			args:          []string{"bash", "dig:url2hostname", "grep:url2hostname", "+dig2doggo", "+grep2rg"},
			expectedShell: "bash",
			expectedCmdPreprocs: map[string][]string{
				"dig":  {"url2hostname"},
				"grep": {"url2hostname"},
			},
			expectedReflagArgs: []string{"+dig2doggo", "+grep2rg"},
		},
		{
			name:          "preflag mapping with multiple preprocessors",
			args:          []string{"zsh", "dig:url2hostname,preprocessor2", "+dig2doggo"},
			expectedShell: "zsh",
			expectedCmdPreprocs: map[string][]string{
				"dig": {"url2hostname", "preprocessor2"},
			},
			expectedReflagArgs: []string{"+dig2doggo"},
		},
		{
			name:          "no shell specified defaults to bash",
			args:          []string{"dig:url2hostname", "+dig2doggo"},
			expectedShell: "bash",
			expectedCmdPreprocs: map[string][]string{
				"dig": {"url2hostname"},
			},
			expectedReflagArgs: []string{"+dig2doggo"},
		},
		{
			name:          "mixed order of preflag and reflag args",
			args:          []string{"zsh", "+dig2doggo", "dig:url2hostname", "grep:url2hostname", "+grep2rg"},
			expectedShell: "zsh",
			expectedCmdPreprocs: map[string][]string{
				"dig":  {"url2hostname"},
				"grep": {"url2hostname"},
			},
			expectedReflagArgs: []string{"+dig2doggo", "+grep2rg"},
		},
		{
			name:                "only reflag args no preflag mappings",
			args:                []string{"bash", "+dig2doggo", "+ls2eza"},
			expectedShell:       "bash",
			expectedCmdPreprocs: map[string][]string{},
			expectedReflagArgs:  []string{"+dig2doggo", "+ls2eza"},
		},
		{
			name:          "only preflag mappings no reflag args",
			args:          []string{"zsh", "dig:url2hostname", "grep:url2hostname"},
			expectedShell: "zsh",
			expectedCmdPreprocs: map[string][]string{
				"dig":  {"url2hostname"},
				"grep": {"url2hostname"},
			},
			expectedReflagArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell, cmdPreprocs, reflagArgs := parseInitArgs(tt.args)

			if shell != tt.expectedShell {
				t.Errorf("parseInitArgs(%v) shell = %q, want %q", tt.args, shell, tt.expectedShell)
			}

			if len(cmdPreprocs) != len(tt.expectedCmdPreprocs) {
				t.Errorf("parseInitArgs(%v) got %d preflag mappings, want %d", tt.args, len(cmdPreprocs), len(tt.expectedCmdPreprocs))
			}

			for cmd, expectedPreprocs := range tt.expectedCmdPreprocs {
				if gotPreprocs, ok := cmdPreprocs[cmd]; !ok {
					t.Errorf("parseInitArgs(%v) missing command %q", tt.args, cmd)
				} else if !slices.Equal(gotPreprocs, expectedPreprocs) {
					t.Errorf("parseInitArgs(%v) command %q: got %v, want %v", tt.args, cmd, gotPreprocs, expectedPreprocs)
				}
			}

			if !slices.Equal(reflagArgs, tt.expectedReflagArgs) {
				t.Errorf("parseInitArgs(%v) reflag args = %v, want %v", tt.args, reflagArgs, tt.expectedReflagArgs)
			}
		})
	}
}

func TestPrintInit(t *testing.T) {
	tests := []struct {
		name             string
		shell            string
		cmdPreprocessors CommandPreprocessors
		reflagArgs       []string
		expectedStrings  []string
		notExpected      []string
	}{
		{
			name:  "combined preflag+reflag chains preprocessing and translation",
			shell: "bash",
			cmdPreprocessors: CommandPreprocessors{
				"dig": {"url2hostname"},
			},
			reflagArgs: []string{"+dig2doggo"},
			expectedStrings: []string{
				"__preflag_orig_dig() { command dig \"$@\"; }",
				"args=$(preflag url2hostname \"$@\")",
				"eval \"$(reflag dig doggo $args)\"",
			},
			notExpected: []string{
				"reflag --init",
			},
		},
		{
			name:  "multiple commands with matching and non-matching reflag args",
			shell: "bash",
			cmdPreprocessors: CommandPreprocessors{
				"dig":  {"url2hostname"},
				"grep": {"url2hostname"},
			},
			reflagArgs: []string{"+dig2doggo", "+grep2rg"},
			expectedStrings: []string{
				"__preflag_orig_dig() { command dig \"$@\"; }",
				"__preflag_orig_grep() { command grep \"$@\"; }",
				"eval \"$(reflag dig doggo $args)\"",
				"eval \"$(reflag grep rg $args)\"",
			},
		},
		{
			name:  "command with multiple preprocessors",
			shell: "zsh",
			cmdPreprocessors: CommandPreprocessors{
				"dig": {"url2hostname", "preprocessor2"},
			},
			reflagArgs: []string{"+dig2doggo"},
			expectedStrings: []string{
				"args=$(preflag url2hostname,preprocessor2 \"$@\")",
			},
		},
		{
			name:  "no matching reflag translator falls back to original command",
			shell: "bash",
			cmdPreprocessors: CommandPreprocessors{
				"ping": {"url2hostname"},
			},
			reflagArgs: []string{"+dig2doggo"},
			expectedStrings: []string{
				"__preflag_orig_ping() { command ping \"$@\"; }",
				"eval \"__preflag_orig_ping $args\"",
			},
			notExpected: []string{
				"reflag ping",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printInit(tt.shell, tt.cmdPreprocessors, tt.reflagArgs)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			for _, expected := range tt.expectedStrings {
				if !strings.Contains(output, expected) {
					t.Errorf("printInit() output missing expected string:\n%q\nGot output:\n%s", expected, output)
				}
			}

			for _, notExp := range tt.notExpected {
				if strings.Contains(output, notExp) {
					t.Errorf("printInit() output contains unexpected string:\n%q\nGot output:\n%s", notExp, output)
				}
			}
		})
	}
}

func TestPrintInitTargetCommandExtraction(t *testing.T) {
	tests := []struct {
		name           string
		cmdPreprocessors CommandPreprocessors
		reflagArgs     []string
		expectedReflag string
	}{
		{
			name:             "dig2doggo extracts doggo as target",
			cmdPreprocessors: CommandPreprocessors{"dig": {"url2hostname"}},
			reflagArgs:       []string{"+dig2doggo"},
			expectedReflag:   "eval \"$(reflag dig doggo $args)\"",
		},
		{
			name:             "grep2rg extracts rg as target",
			cmdPreprocessors: CommandPreprocessors{"grep": {"url2hostname"}},
			reflagArgs:       []string{"+grep2rg"},
			expectedReflag:   "eval \"$(reflag grep rg $args)\"",
		},
		{
			name:             "find2fd extracts fd as target",
			cmdPreprocessors: CommandPreprocessors{"find": {"url2hostname"}},
			reflagArgs:       []string{"+find2fd"},
			expectedReflag:   "eval \"$(reflag find fd $args)\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printInit("bash", tt.cmdPreprocessors, tt.reflagArgs)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, tt.expectedReflag) {
				t.Errorf("printInit() output missing expected reflag call:\n%q\nGot output:\n%s", tt.expectedReflag, output)
			}
		})
	}
}

func TestParseInitArgsEdgeCases(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		expectedShell       string
		expectedCmdPreprocs map[string][]string
		expectedReflagArgs  []string
	}{
		{
			name:          "argument that looks like shell but has colon is treated as preflag mapping",
			args:          []string{"bash:url2hostname"},
			expectedShell: "bash",
			expectedCmdPreprocs: map[string][]string{
				"bash": {"url2hostname"},
			},
			expectedReflagArgs: []string{},
		},
		{
			name:          "multiple colons in mapping uses only first split",
			args:          []string{"cmd:prep1:prep2"},
			expectedShell: "bash",
			expectedCmdPreprocs: map[string][]string{
				"cmd": {"prep1:prep2"},
			},
			expectedReflagArgs: []string{},
		},
		{
			name:                "plus sign without translator name",
			args:                []string{"+"},
			expectedShell:       "bash",
			expectedCmdPreprocs: map[string][]string{},
			expectedReflagArgs:  []string{"+"},
		},
		{
			name:                "minus sign without translator name is ignored",
			args:                []string{"-"},
			expectedShell:       "bash",
			expectedCmdPreprocs: map[string][]string{},
			expectedReflagArgs:  []string{},
		},
		{
			name:          "empty preprocessor list after colon",
			args:          []string{"cmd:"},
			expectedShell: "bash",
			expectedCmdPreprocs: map[string][]string{
				"cmd": {""},
			},
			expectedReflagArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell, cmdPreprocs, reflagArgs := parseInitArgs(tt.args)

			if shell != tt.expectedShell {
				t.Errorf("parseInitArgs(%v) shell = %q, want %q", tt.args, shell, tt.expectedShell)
			}

			if len(cmdPreprocs) != len(tt.expectedCmdPreprocs) {
				t.Errorf("parseInitArgs(%v) got %d preflag mappings, want %d", tt.args, len(cmdPreprocs), len(tt.expectedCmdPreprocs))
			}

			for cmd, expectedPreprocs := range tt.expectedCmdPreprocs {
				if gotPreprocs, ok := cmdPreprocs[cmd]; !ok {
					t.Errorf("parseInitArgs(%v) missing command %q", tt.args, cmd)
				} else if !slices.Equal(gotPreprocs, expectedPreprocs) {
					t.Errorf("parseInitArgs(%v) command %q: got %v, want %v", tt.args, cmd, gotPreprocs, expectedPreprocs)
				}
			}

			if !slices.Equal(reflagArgs, tt.expectedReflagArgs) {
				t.Errorf("parseInitArgs(%v) reflag args = %v, want %v", tt.args, reflagArgs, tt.expectedReflagArgs)
			}
		})
	}
}

func TestPrintInitComplexScenarios(t *testing.T) {
	tests := []struct {
		name             string
		shell            string
		cmdPreprocessors CommandPreprocessors
		reflagArgs       []string
		mustContain      []string
		mustNotContain   []string
	}{
		{
			name:  "three commands with mixed reflag args",
			shell: "bash",
			cmdPreprocessors: CommandPreprocessors{
				"dig":  {"url2hostname"},
				"grep": {"url2hostname"},
				"ls":   {"url2hostname"},
			},
			reflagArgs: []string{"+dig2doggo", "+grep2rg", "+du2dust"},
			mustContain: []string{
				"__preflag_orig_dig()",
				"__preflag_orig_grep()",
				"__preflag_orig_ls()",
				"eval \"$(reflag dig doggo $args)\"",
				"eval \"$(reflag grep rg $args)\"",
			},
			mustNotContain: []string{
				"reflag --init",
				"reflag ls",
			},
		},
		{
			name:  "command with comma-separated preprocessors",
			shell: "zsh",
			cmdPreprocessors: CommandPreprocessors{
				"dig": {"url2hostname", "preprocessor2", "preprocessor3"},
			},
			reflagArgs: []string{"+dig2doggo"},
			mustContain: []string{
				"args=$(preflag url2hostname,preprocessor2,preprocessor3 \"$@\")",
			},
		},
		{
			name:  "no matching reflag translator uses original command",
			shell: "bash",
			cmdPreprocessors: CommandPreprocessors{
				"ping": {"url2hostname"},
				"ssh":  {"url2hostname"},
			},
			reflagArgs: []string{"+dig2doggo", "+grep2rg"},
			mustContain: []string{
				"__preflag_orig_ping()",
				"__preflag_orig_ssh()",
				"eval \"__preflag_orig_ping $args\"",
				"eval \"__preflag_orig_ssh $args\"",
			},
			mustNotContain: []string{
				"reflag ping",
				"reflag ssh",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printInit(tt.shell, tt.cmdPreprocessors, tt.reflagArgs)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			for _, must := range tt.mustContain {
				if !strings.Contains(output, must) {
					t.Errorf("Output missing required string:\n%q\nGot:\n%s", must, output)
				}
			}

			for _, mustNot := range tt.mustNotContain {
				if strings.Contains(output, mustNot) {
					t.Errorf("Output contains forbidden string:\n%q\nGot:\n%s", mustNot, output)
				}
			}
		})
	}
}

func TestPrintInitCommandSorting(t *testing.T) {
	cmdPreprocessors := CommandPreprocessors{
		"zsh":  {"url2hostname"},
		"dig":  {"url2hostname"},
		"grep": {"url2hostname"},
		"awk":  {"url2hostname"},
	}
	reflagArgs := []string{"+dig2doggo"}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printInit("bash", cmdPreprocessors, reflagArgs)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	awkIdx := strings.Index(output, "__preflag_orig_awk()")
	digIdx := strings.Index(output, "__preflag_orig_dig()")
	grepIdx := strings.Index(output, "__preflag_orig_grep()")
	zshIdx := strings.Index(output, "__preflag_orig_zsh()")

	if awkIdx == -1 || digIdx == -1 || grepIdx == -1 || zshIdx == -1 {
		t.Fatal("Not all commands found in output")
	}

	if !(awkIdx < digIdx && digIdx < grepIdx && grepIdx < zshIdx) {
		t.Error("Commands are not sorted alphabetically in output")
	}
}

func TestProcessArgs(t *testing.T) {
	tests := []struct {
		name          string
		cmd           string
		preprocessors []string
		args          []string
		expected      []string
	}{
		{
			name:          "ping with URL using url2hostname",
			cmd:           "ping",
			preprocessors: []string{"url2hostname"},
			args:          []string{"https://vg.no"},
			expected:      []string{"vg.no"},
		},
		{
			name:          "ping with flags and URL using url2hostname",
			cmd:           "ping",
			preprocessors: []string{"url2hostname"},
			args:          []string{"-c", "4", "https://github.com/user"},
			expected:      []string{"-c", "4", "github.com"},
		},
		{
			name:          "ssh with URL using url2hostname",
			cmd:           "ssh",
			preprocessors: []string{"url2hostname"},
			args:          []string{"https://example.com/path"},
			expected:      []string{"example.com"},
		},
		{
			name:          "no preprocessors passes through",
			cmd:           "unknowncmd",
			preprocessors: []string{},
			args:          []string{"https://example.com"},
			expected:      []string{"https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processArgs(tt.cmd, tt.preprocessors, tt.args)
			if !slices.Equal(result, tt.expected) {
				t.Errorf("processArgs(%q, %v, %v) = %v, want %v", tt.cmd, tt.preprocessors, tt.args, result, tt.expected)
			}
		})
	}
}

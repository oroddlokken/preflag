package url2hostname

import (
	"slices"
	"testing"
)

func TestExtractHostname(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple hostname",
			input:    "vg.no",
			expected: "vg.no",
		},
		{
			name:     "https URL",
			input:    "https://vg.no",
			expected: "vg.no",
		},
		{
			name:     "http URL",
			input:    "http://example.com",
			expected: "example.com",
		},
		{
			name:     "URL with path",
			input:    "https://github.com/user/repo",
			expected: "github.com",
		},
		{
			name:     "URL with port",
			input:    "https://example.com:8080/path",
			expected: "example.com",
		},
		{
			name:     "URL with query string",
			input:    "https://example.com/path?query=value",
			expected: "example.com",
		},
		{
			name:     "URL with authentication",
			input:    "https://user:pass@example.com/path",
			expected: "example.com",
		},
		{
			name:     "ftp URL",
			input:    "ftp://ftp.example.com/file",
			expected: "ftp.example.com",
		},
		{
			name:     "flag is preserved",
			input:    "-c",
			expected: "-c",
		},
		{
			name:     "long flag is preserved",
			input:    "--count=5",
			expected: "--count=5",
		},
		{
			name:     "IP address preserved",
			input:    "192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "URL without scheme with path",
			input:    "www.example.com/path/to/page",
			expected: "www.example.com",
		},
		{
			name:     "localhost URL",
			input:    "http://localhost:3000/api",
			expected: "localhost",
		},
		{
			name:     "absolute path preserved",
			input:    "/path/to/file",
			expected: "/path/to/file",
		},
		{
			name:     "number preserved",
			input:    "5",
			expected: "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHostname(tt.input)
			if result != tt.expected {
				t.Errorf("extractHostname(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPreprocessorProcess(t *testing.T) {
	p := &Preprocessor{}

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "ping with URL",
			args:     []string{"https://vg.no"},
			expected: []string{"vg.no"},
		},
		{
			name:     "ping with flags and URL",
			args:     []string{"-c", "4", "https://github.com/user"},
			expected: []string{"-c", "4", "github.com"},
		},
		{
			name:     "multiple URLs",
			args:     []string{"https://a.com", "https://b.com"},
			expected: []string{"a.com", "b.com"},
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: []string{},
		},
		{
			name:     "only flags",
			args:     []string{"-c", "4", "-W", "1"},
			expected: []string{"-c", "4", "-W", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.Process(tt.args)
			if !slices.Equal(result, tt.expected) {
				t.Errorf("Process(%v) = %v, want %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestPreprocessorMetadata(t *testing.T) {
	p := &Preprocessor{}

	t.Run("Name", func(t *testing.T) {
		if got := p.Name(); got != "url2hostname" {
			t.Errorf("Name() = %q, want %q", got, "url2hostname")
		}
	})
}

package emojiexample

import (
	"slices"
	"testing"
)

func TestTransformArg(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase a",
			input:    "apple",
			expected: "🅰️🅿️🅿️le",
		},
		{
			name:     "uppercase A",
			input:    "APPLE",
			expected: "🅰️🅿️🅿️LE",
		},
		{
			name:     "lowercase b",
			input:    "banana",
			expected: "🅱️🅰️n🅰️n🅰️",
		},
		{
			name:     "uppercase B",
			input:    "BANANA",
			expected: "🅱️🅰️N🅰️N🅰️",
		},
		{
			name:     "lowercase c",
			input:    "car",
			expected: "©️🅰️®️",
		},
		{
			name:     "uppercase C",
			input:    "CAR",
			expected: "©️🅰️®️",
		},
		{
			name:     "lowercase m",
			input:    "map",
			expected: "Ⓜ️🅰️🅿️",
		},
		{
			name:     "uppercase M",
			input:    "MAP",
			expected: "Ⓜ️🅰️🅿️",
		},
		{
			name:     "lowercase o",
			input:    "box",
			expected: "🅱️🅾️❌",
		},
		{
			name:     "uppercase O",
			input:    "BOX",
			expected: "🅱️🅾️❌",
		},
		{
			name:     "lowercase p",
			input:    "park",
			expected: "🅿️🅰️®️k",
		},
		{
			name:     "uppercase P",
			input:    "PARK",
			expected: "🅿️🅰️®️K",
		},
		{
			name:     "lowercase r",
			input:    "red",
			expected: "®️ed",
		},
		{
			name:     "uppercase R",
			input:    "RED",
			expected: "®️ED",
		},
		{
			name:     "lowercase x",
			input:    "box",
			expected: "🅱️🅾️❌",
		},
		{
			name:     "uppercase X",
			input:    "BOX",
			expected: "🅱️🅾️❌",
		},
		{
			name:     "mixed case",
			input:    "ProBox",
			expected: "🅿️®️🅾️🅱️🅾️❌",
		},
		{
			name:     "no matching letters",
			input:    "test",
			expected: "test",
		},
		{
			name:     "flag is preserved",
			input:    "-c",
			expected: "-c",
		},
		{
			name:     "long flag is preserved",
			input:    "--verbose",
			expected: "--verbose",
		},
		{
			name:     "flag with equals",
			input:    "--max=5",
			expected: "--max=5",
		},
		{
			name:     "all transformable letters",
			input:    "abcmoprx",
			expected: "🅰️🅱️©️Ⓜ️🅾️🅿️®️❌",
		},
		{
			name:     "all transformable letters uppercase",
			input:    "ABCMOPRX",
			expected: "🅰️🅱️©️Ⓜ️🅾️🅿️®️❌",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformArg(tt.input)
			if result != tt.expected {
				t.Errorf("transformArg(%q) = %q, want %q", tt.input, result, tt.expected)
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
			name:     "single word",
			args:     []string{"box"},
			expected: []string{"🅱️🅾️❌"},
		},
		{
			name:     "multiple words",
			args:     []string{"car", "map"},
			expected: []string{"©️🅰️®️", "Ⓜ️🅰️🅿️"},
		},
		{
			name:     "command with flags and args",
			args:     []string{"-f", "box"},
			expected: []string{"-f", "🅱️🅾️❌"},
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: []string{},
		},
		{
			name:     "only flags",
			args:     []string{"-a", "-b", "--max"},
			expected: []string{"-a", "-b", "--max"},
		},
		{
			name:     "mixed flags and words",
			args:     []string{"-c", "4", "apple", "--verbose", "box"},
			expected: []string{"-c", "4", "🅰️🅿️🅿️le", "--verbose", "🅱️🅾️❌"},
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
		if got := p.Name(); got != "emojiexample" {
			t.Errorf("Name() = %q, want %q", got, "emojiexample")
		}
	})

	t.Run("Description", func(t *testing.T) {
		if got := p.Description(); got == "" {
			t.Error("Description() should not be empty")
		}
	})
}

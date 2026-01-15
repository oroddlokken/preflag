package emojiexample

import (
	"strings"

	"github.com/oroddlokken/preflag/preprocessor"
)

func init() {
	preprocessor.Register(&Preprocessor{})
}

type Preprocessor struct{}

func (p *Preprocessor) Name() string {
	return "emojiexample"
}

func (p *Preprocessor) Description() string {
	return "Replace letters with emoji equivalents (a→🅰️, b→🅱️, c→©️, m→Ⓜ️, o→🅾️, p→🅿️, r→®️, x→❌)"
}

func (p *Preprocessor) Process(args []string) []string {
	result := make([]string, len(args))
	for i, arg := range args {
		result[i] = transformArg(arg)
	}
	return result
}

func transformArg(arg string) string {
	if strings.HasPrefix(arg, "-") {
		return arg
	}

	replacer := strings.NewReplacer(
		"a", "🅰️",
		"A", "🅰️",
		"b", "🅱️",
		"B", "🅱️",
		"c", "©️",
		"C", "©️",
		"m", "Ⓜ️",
		"M", "Ⓜ️",
		"o", "🅾️",
		"O", "🅾️",
		"p", "🅿️",
		"P", "🅿️",
		"r", "®️",
		"R", "®️",
		"x", "❌",
		"X", "❌",
	)

	return replacer.Replace(arg)
}

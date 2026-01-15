package url2hostname

import (
	"net/url"
	"strings"

	"github.com/oroddlokken/preflag/preprocessor"
)

func init() {
	preprocessor.Register(&Preprocessor{})
}

// Preprocessor extracts hostnames from URLs for network commands
type Preprocessor struct{}

func (p *Preprocessor) Name() string {
	return "url2hostname"
}

func (p *Preprocessor) Description() string {
	return "Extract hostname from URLs"
}

func (p *Preprocessor) Process(args []string) []string {
	result := make([]string, len(args))
	for i, arg := range args {
		result[i] = extractHostname(arg)
	}
	return result
}

// extractHostname attempts to extract a hostname from a URL-like string
// If the argument doesn't look like a URL, it's returned unchanged
func extractHostname(arg string) string {
	// Skip if it looks like a flag
	if strings.HasPrefix(arg, "-") {
		return arg
	}

	// If it contains "://", treat it as a URL
	if strings.Contains(arg, "://") {
		if u, err := url.Parse(arg); err == nil && u.Host != "" {
			// Return just the hostname (strips port if present)
			return u.Hostname()
		}
	}

	// Handle URLs without scheme (e.g., "www.example.com/path")
	// Only if it contains a "/" and looks like a domain
	if strings.Contains(arg, "/") && !strings.HasPrefix(arg, "/") {
		// Try parsing with a scheme added
		if u, err := url.Parse("http://" + arg); err == nil && u.Host != "" {
			host := u.Hostname()
			// Verify it looks like a valid hostname (contains a dot or is localhost)
			if strings.Contains(host, ".") || host == "localhost" {
				return host
			}
		}
	}

	return arg
}

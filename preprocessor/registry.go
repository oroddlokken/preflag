package preprocessor

import (
	"fmt"
	"io"
	"sort"
	"sync"
	"text/tabwriter"
)

var (
	registry = make(map[string]Preprocessor)
	mu       sync.RWMutex
)

// Register adds a preprocessor to the global registry
func Register(p Preprocessor) {
	mu.Lock()
	defer mu.Unlock()
	registry[p.Name()] = p
}

// GetByName returns a preprocessor by its name (e.g., "url2hostname")
// Returns nil if no preprocessor is found
func GetByName(name string) Preprocessor {
	mu.RLock()
	defer mu.RUnlock()
	return registry[name]
}

// List returns all registered preprocessor names
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// MustGetByName returns a preprocessor or panics if not found
func MustGetByName(name string) Preprocessor {
	p := GetByName(name)
	if p == nil {
		panic(fmt.Sprintf("no preprocessor registered with name %s", name))
	}
	return p
}

// PrintTable writes a formatted table of all preprocessors to the given writer
func PrintTable(w io.Writer) {
	mu.RLock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	mu.RUnlock()

	sort.Strings(names)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PREPROCESSOR\tDESCRIPTION")
	for _, name := range names {
		p := registry[name]
		fmt.Fprintf(tw, "%s\t%s\n", name, p.Description())
	}
	tw.Flush()
}

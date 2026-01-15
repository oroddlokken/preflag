package preprocessor

// Preprocessor defines the interface for transforming command arguments
type Preprocessor interface {
	// Name returns the preprocessor identifier (e.g., "url2hostname")
	Name() string

	// Description returns a short description of the preprocessor
	Description() string

	// Process transforms arguments, returning modified args
	// Only non-flag arguments should be transformed
	Process(args []string) []string
}

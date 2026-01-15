# Agent Guide: Adding a New Preprocessor

This guide explains how to add a new preprocessor to preflag. Follow these steps to implement a preprocessor that transforms command-line arguments.

## Architecture Overview

Preflag uses a **registry pattern** where preprocessors:

1. Implement the `Preprocessor` interface
2. Self-register via `init()` functions
3. Are automatically discovered and made available to users

## Step-by-Step Implementation

### 1. Create Preprocessor Directory

Create a new directory under `preprocessor/`:

```bash
mkdir -p preprocessor/yourpreprocessor
```

### 2. Implement the Preprocessor

Create `preprocessor/yourpreprocessor/preprocessor.go`:

```go
package yourpreprocessor

import (
 "github.com/oroddlokken/preflag/preprocessor"
)

func init() {
 preprocessor.Register(&Preprocessor{})
}

type Preprocessor struct{}

func (p *Preprocessor) Name() string {
 return "yourpreprocessor"
}

func (p *Preprocessor) Description() string {
 return "Brief description of what it does"
}

func (p *Preprocessor) Process(args []string) []string {
 result := make([]string, len(args))
 for i, arg := range args {
  result[i] = transformArg(arg)
 }
 return result
}

func transformArg(arg string) string {
 // Skip flags (arguments starting with -)
 if strings.HasPrefix(arg, "-") {
  return arg
 }
 
 // Your transformation logic here
 // Return the transformed argument or the original if no transformation needed
 return arg
}
```

**Key Points:**

- The `init()` function registers your preprocessor automatically
- `Name()` returns the identifier users will use (e.g., `preflag yourpreprocessor arg`)
- `Description()` appears in `preflag --list` output
- `Process()` receives all arguments and returns transformed arguments
- **Always preserve flags** (arguments starting with `-`)

### 3. Write Tests

Create `preprocessor/yourpreprocessor/preprocessor_test.go`:

```go
package yourpreprocessor

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
   name:     "basic transformation",
   input:    "input",
   expected: "output",
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
   name:     "command with args",
   args:     []string{"arg1", "arg2"},
   expected: []string{"transformed1", "transformed2"},
  },
  {
   name:     "command with flags and args",
   args:     []string{"-f", "arg1"},
   expected: []string{"-f", "transformed1"},
  },
  {
   name:     "empty args",
   args:     []string{},
   expected: []string{},
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

 if got := p.Name(); got != "yourpreprocessor" {
  t.Errorf("Name() = %q, want %q", got, "yourpreprocessor")
 }

 if got := p.Description(); got == "" {
  t.Error("Description() should not be empty")
 }
}
```

### 4. Import in Main

Add the import to `main.go` (blank import for side-effects):

```go
import (
 // ... existing imports
 _ "github.com/oroddlokken/preflag/preprocessor/yourpreprocessor"
)
```

### 5. Import in Tests

Add the import to `main_test.go`:

```go
import (
 // ... existing imports
 _ "github.com/oroddlokken/preflag/preprocessor/yourpreprocessor"
)
```

### 6. Run Tests

```bash
# Test your preprocessor
go test ./preprocessor/yourpreprocessor/...

# Test the entire project
go test ./...
```

### 7. Build and Verify

```bash
# Build
make build

# Verify registration
./preflag --list

# Test manually
./preflag yourpreprocessor arg1 arg2
```

## Example: url2hostname Preprocessor

The `url2hostname` preprocessor extracts hostnames from URLs:

**Input:** `https://github.com/user/repo`  
**Output:** `github.com`

**Implementation highlights:**

- Detects URLs by checking for `://` or `/` patterns
- Uses `net/url` package for parsing
- Preserves flags and non-URL arguments unchanged
- Handles edge cases: ports, query strings, authentication, schemes

**File structure:**

```
preprocessor/url2hostname/
├── preprocessor.go       # Implementation
└── preprocessor_test.go  # Tests
```

## Best Practices

1. **Preserve flags:** Always return flag arguments (`-x`, `--flag`) unchanged
2. **Handle edge cases:** Test with empty args, only flags, mixed inputs
3. **Return unchanged on error:** If transformation fails, return original argument
4. **Use descriptive names:** Name should clearly indicate the transformation
5. **Write comprehensive tests:** Cover normal cases, edge cases, and error conditions
6. **Document behavior:** Add comments explaining non-obvious logic
7. **Keep it simple:** Each preprocessor should do one thing well

## Testing Checklist

- [ ] Unit tests for transformation function
- [ ] Tests for `Process()` method
- [ ] Tests for metadata (`Name()`, `Description()`)
- [ ] Tests with flags preserved
- [ ] Tests with empty input
- [ ] Tests with mixed flags and arguments
- [ ] Integration test in `main_test.go` (verify registration)

## Common Patterns

### Pattern 1: URL/String Transformation

Transform string arguments (like `url2hostname`):

```go
func transformArg(arg string) string {
 if strings.HasPrefix(arg, "-") {
  return arg
 }
 // Transform logic
 return transformed
}
```

### Pattern 2: Argument Filtering

Remove or filter certain arguments:

```go
func (p *Preprocessor) Process(args []string) []string {
 var result []string
 for _, arg := range args {
  if shouldKeep(arg) {
   result = append(result, arg)
  }
 }
 return result
}
```

### Pattern 3: Argument Expansion

Expand arguments into multiple arguments:

```go
func (p *Preprocessor) Process(args []string) []string {
 var result []string
 for _, arg := range args {
  expanded := expand(arg)
  result = append(result, expanded...)
 }
 return result
}
```

## Troubleshooting

**Preprocessor not showing in `--list`:**

- Verify `init()` function calls `preprocessor.Register()`
- Check blank import in `main.go`
- Rebuild the binary

**Tests failing:**

- Ensure test imports the preprocessor package
- Check that transformation preserves flags
- Verify edge cases are handled

**Runtime errors:**

- Add nil checks for pointer operations
- Handle empty input gracefully
- Return original argument on parse errors

## Module Dependencies

If your preprocessor needs external packages:

1. Add to `preprocessor/go.mod` (if using workspace)
2. Or add to root `go.mod`
3. Run `go mod tidy`

## Summary

Adding a preprocessor requires:

1. Create directory: `preprocessor/yourpreprocessor/`
2. Implement interface with `init()` registration
3. Write comprehensive tests
4. Import in `main.go` and `main_test.go`
5. Build and verify with `--list`

The registry pattern ensures your preprocessor is automatically discovered and available to users without modifying core code.

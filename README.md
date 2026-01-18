# preflag

![AI Slop Badge](ai-slop-badge.svg)

A tool that preprocesses command-line arguments before they reach the actual command. Currently supports:

- `url2hostname` - Extract hostnames from URLs for network commands

## Why?

Network commands like `ping`, `dig`, `whois`, and `traceroute` expect hostnames or IP addresses—not full URLs. But when you're debugging a website issue or checking DNS records, you often have the full URL in your clipboard. Instead of manually extracting the hostname from `https://example.com/some/path`, preflag does it automatically.

preflag sits between you and the command, transforming arguments before they're executed. It's designed to work seamlessly with [reflag](https://github.com/kluzzebass/reflag) for a complete command transformation pipeline.

## Quick Start

### 1. Install preflag

Choose one of the following:

**Using Go**:

```bash
go install github.com/oroddlokken/preflag@latest
```

**From source**:

```bash
git clone https://github.com/oroddlokken/preflag.git
cd preflag
make build
sudo mv preflag /usr/local/bin/
```

### 2. Set up shell integration

Run this to enable automatic preprocessing:

**bash** (`~/.bashrc`):

```bash
echo 'eval "$(preflag --init bash whois:url2hostname ping:url2hostname dig:url2hostname)"' >> ~/.bashrc && source ~/.bashrc
```

**zsh** (`~/.zshrc`):

```bash
echo 'eval "$(preflag --init zsh whois:url2hostname ping:url2hostname dig:url2hostname)"' >> ~/.zshrc && source ~/.zshrc
```

### 3. Start using URLs directly

```bash
ping https://example.com          # Preprocessed to: ping example.com
dig https://vg.no                 # Preprocessed to: dig vg.no
whois https://github.com/user     # Preprocessed to: whois github.com
```

That's it! Copy-paste URLs directly into network commands.

**Note:** To bypass preflag and use the original command, use `command ping` or `/bin/ping`.

## Usage

### Explicit Mode

Apply preprocessors directly:

```bash
$ preflag url2hostname https://example.com/path
example.com

$ preflag url2hostname https://vg.no https://github.com
vg.no github.com

$ preflag url2hostname -A example.com/path
-A example.com
```

### Shell Integration

Generate shell functions that wrap commands with preprocessors:

```bash
# Preview what would be generated
preflag --init bash ping:url2hostname dig:url2hostname

# Generate for specific commands
preflag --init zsh ping:url2hostname dig:url2hostname whois:url2hostname
```

**Recommended setup:** Add this to your shell config:

```bash
# ~/.bashrc or ~/.zshrc
eval "$(preflag --init zsh ping:url2hostname dig:url2hostname whois:url2hostname)"
```

### List Available Preprocessors

```bash
$ preflag --list
url2hostname: Extract hostname from URLs
```

## url2hostname Preprocessor

The url2hostname preprocessor extracts hostnames from URL-like arguments, making network commands work seamlessly with full URLs.

### How It Works

- **Full URLs**: `https://example.com/path` → `example.com`
- **URLs with ports**: `https://example.com:8080/path` → `example.com`
- **Partial URLs**: `example.com/path` → `example.com`
- **Flags preserved**: `-A https://example.com` → `-A example.com`
- **Non-URLs unchanged**: `example.com` → `example.com`

### Examples

```bash
$ preflag url2hostname https://vg.no
vg.no

$ preflag url2hostname https://github.com/oroddlokken/preflag
github.com

$ preflag url2hostname www.example.com/some/path
www.example.com

$ preflag url2hostname -A https://example.com
-A example.com
```

### Use Cases

Perfect for network diagnostic commands:

```bash
ping https://example.com          # Works like: ping example.com
dig https://vg.no MX              # Works like: dig vg.no MX
whois https://github.com          # Works like: whois github.com
traceroute https://example.com    # Works like: traceroute example.com
nslookup https://example.com      # Works like: nslookup example.com
```

## Chaining with reflag

preflag is designed to work seamlessly with [reflag](https://github.com/kluzzebass/reflag) for complete command transformation. See [CHAINING.md](CHAINING.md) for details.

### Quick Setup with reflag

```bash
# Single command that sets up both preflag and reflag
eval "$(preflag --init zsh dig:url2hostname +dig2doggo)"
```

This creates a complete pipeline:

1. **preflag** preprocesses URLs to hostnames
2. **reflag** translates old flags to modern tool flags
3. Modern tool executes with clean arguments

Example:

```bash
dig https://vg.no +short
# → preflag extracts: vg.no
# → reflag translates: dig → doggo
# → executes: doggo --short vg.no
```

### Execution Flow

```text
User input: dig https://vg.no +short
     ↓
preflag (url2hostname): https://vg.no → vg.no
     ↓
reflag (dig2doggo): dig +short → doggo --short
     ↓
Execute: doggo --short vg.no
```

## Adding New Preprocessors

preflag is designed to be extensible. To add a new preprocessor:

1. Create a new package under `preprocessor/`
2. Implement the `preprocessor.Preprocessor` interface:

   ```go
   type Preprocessor interface {
       Name() string
       Description() string
       Process(args []string) []string
   }
   ```

3. Register it in `init()` using `preprocessor.Register()`

See `preprocessor/url2hostname/` for an example implementation.

### Preprocessor Design Guidelines

- **Preserve flags**: Arguments starting with `-` should typically pass through unchanged
- **Be conservative**: Only transform arguments that clearly match your pattern
- **Return unchanged**: If unsure, return the argument as-is
- **No side effects**: Preprocessors should be pure transformations

## Limitations

**Preprocessing is heuristic-based.** The url2hostname preprocessor uses pattern matching to detect URLs. Edge cases may exist where valid hostnames are incorrectly processed or URLs aren't recognized.

**Flags are preserved but not parsed.** Preprocessors see arguments as strings and don't understand command-specific flag syntax. Complex flag interactions may behave unexpectedly.

**Shell scripts are generally unaffected.** Scripts using `#!/bin/bash` run in non-interactive mode and don't source `~/.bashrc` or `~/.zshrc`, so they won't see the preflag functions. If you encounter issues, bypass with `command ping` or `/bin/ping`.

**This is a convenience tool.** preflag aims to make common workflows smoother, not to provide perfect argument parsing for every edge case.

## Building from Source

```bash
# Using Go
go install github.com/oroddlokken/preflag@latest

# Or build manually
git clone https://github.com/oroddlokken/preflag.git
cd preflag
make build
```

## Advanced Examples

### With reflag Chaining

```bash
# Complete setup with reflag
eval "$(preflag --init zsh dig:url2hostname +dig2doggo)"

# Use modern tools with URLs
dig https://example.com +short
# Executes: doggo --short example.com
```

### Multiple Commands

```bash
# Set up multiple commands at once
eval "$(preflag --init zsh ping:url2hostname dig:url2hostname whois:url2hostname)"

# Now use URLs directly
ping https://example.com
dig https://vg.no MX
whois https://github.com
```

## License

MIT

## Inspiration

preflag was inspired by [reflag](https://github.com/kluzzebass/reflag), a tool for translating old command-line flags to modern tool flags, and goes hand in hand with it.

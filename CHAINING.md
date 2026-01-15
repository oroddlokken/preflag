# Using preflag and reflag Together

When you want to use both `preflag` and `reflag` for the same command, use the `--init-with-reflag` flag for automatic setup.

## Quick Setup (Recommended)

In your shell config (`~/.zshrc` or `~/.bashrc`):

```bash
# Simple one-liner that sets up everything
eval "$(preflag --init-with-reflag zsh dig:url2hostname +dig2doggo)"
```

Arguments:

- Arguments with `:` are preflag mappings (e.g., `dig:url2hostname`)
- Arguments starting with `+` or `-` are reflag translators (e.g., `+dig2doggo`, `-ls2eza`)

Multiple commands example:

```bash
eval "$(preflag --init-with-reflag zsh dig:url2hostname +dig2doggo +grep2rg -ls2eza)"
```

## How It Works

Preflag checks for a `<cmd>_preflag_next` function that you can point to reflag's implementation:

1. **`dig()`** - Preflag wrapper (applies preprocessors, checks for next step)
2. **`dig_preflag_next()`** - Points to reflag's translation logic
3. **`command dig`** - The actual system command

## Manual Setup (Advanced)

If you need more control, you can set up the chain manually:

```bash
# 1. Initialize preflag first
eval "$(preflag --init zsh dig:url2hostname)"

# 2. Initialize reflag second (this overwrites dig())
eval "$(reflag --init zsh +dig2doggo -grep2rg -ls2eza)"

# 3. Capture reflag's dig implementation
__preflag_dig_next() { eval "$(reflag dig doggo "$@")"; }

# 4. Restore preflag's dig wrapper
dig() { __preflag_dig_next "$@"; }
```

## Execution Flow

When you run `dig https://vg.no`:

1. `dig()` function (from preflag) is called
2. Calls `__preflag_dig_impl` with original arguments
3. `__preflag_dig_impl` applies the `url2hostname` preprocessor (transforms `https://vg.no` → `vg.no`)
4. Checks if `__preflag_dig_next` function exists
5. Since it does, calls `__preflag_dig_next` with preprocessed args (`vg.no`)
6. `__preflag_dig_next` runs reflag's translation (`dig` → `doggo`)
7. Result executes: `doggo vg.no`

The chain is: **preflag preprocessing → reflag translation → actual command**

## Example

```bash
# Add to ~/.zshrc
eval "$(preflag --init-with-reflag zsh dig:url2hostname +dig2doggo)"

# Now both tools work together:
# dig https://vg.no
# → preprocesses URL to hostname (preflag)
# → translates dig to doggo (reflag)
# → executes: doggo vg.no
```

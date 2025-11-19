---
title: execx (command execution)
---

Purpose
-------
`internal/execx` provides controlled ways to run external commands, capture stdout/stderr, and set working directories and timeouts.

Key APIs
--------
Key APIs
--------
- `RunCmd(ctx context.Context, dir string, cmd string) error` — execute a command-line string by splitting on spaces and running the program with the resulting args; captures stdout/stderr and returns a formatted error on failure.
- `RunCmdLive(ctx context.Context, dir string, cmd string) error` — like `RunCmd` but streams stdout/stderr to the user's terminal (interactive programs); also attaches Stdin so prompts work.
- `RunCmdOutput(ctx context.Context, dir string, cmd string) (stdout string, stderr string, err error)` — run a command and return captured stdout and stderr as strings; on error stderr is returned and an error describing the failure is provided.
- `OpenBrowser(url string) error` — platform-aware helper to open the given URL in the user's default browser (Windows/macOS/Linux).
 - `RunCmd(ctx context.Context, dir string, cmd string) error` — execute a command-line string by splitting on spaces and running the program with the resulting args; captures stdout/stderr and returns a formatted error on failure.

Functions (implementation details)
----------------------------------
- `RunCmd(ctx context.Context, dir string, cmd string) error`
	- Signature: `func RunCmd(ctx context.Context, dir string, cmd string) error`
	- Purpose: Run a command in the given working directory and capture stdout/stderr.
	- Behavior: Splits the provided `cmd` on spaces (using `strings.Split`), uses the first token as the executable name and the rest as args, and runs it with `exec.CommandContext`.
	- Notes & edge-cases:
	  - Splitting on spaces is simple and will not handle quoted arguments correctly (e.g., commands with spaces inside quoted values). Consider using a tokenizer or running commands through a shell when you need full shell semantics.
	  - Uses in-memory buffers for stdout/stderr — avoid for commands that produce extremely large output.

- `RunCmdLive(ctx context.Context, dir string, cmd string) error`
	- Signature: `func RunCmdLive(ctx context.Context, dir string, cmd string) error`
	- Purpose: Run commands that need interactive I/O or that should stream output to the user's terminal (for example `firebase login` or similar interactive CLIs).
	- Behavior: Splits `cmd` on spaces, runs the executable, and wires command stdout/stderr to `os.Stdout`/`os.Stderr` (via `io.MultiWriter`) while still capturing output into buffers that are returned inside errors on failure. Also sets `c.Stdin = os.Stdin` so interactive prompts work.
	- When to use: Use when the executed program expects to interact with the user or when you want to show streaming progress to the user's terminal.

- `RunCmdOutput(ctx context.Context, dir string, cmd string) (string, string, error)`
	- Signature: `func RunCmdOutput(ctx context.Context, dir string, cmd string) (string, string, error)`
	- Purpose: Run a command and return captured stdout and stderr as strings (instead of printing them). On error, stderr is returned and an error describing the failure is provided.
	- Behavior: Splits `cmd`, runs the command, and returns `(stdout, stderr, err)`; error includes stderr for easier debugging.

- `OpenBrowser(url string) error`
	- Signature: `func OpenBrowser(url string) error`
	- Purpose: Platform-aware helper to open the given URL in the user's default browser.
	- Behavior: Dispatches to `rundll32 url.dll,FileProtocolHandler <url>` on Windows, `open <url>` on macOS, and `xdg-open <url>` on other Unix-like systems. Uses `exec.Command(...).Start()` so it does not block waiting for the browser process to exit.
	- Notes: This is a lightweight convenience helper and does not handle errors beyond what `exec.Command.Start()` returns.

When to use
-----------
- If you need to accept a pasted command line (e.g., `npx create-next-app ...`) or need shell features (expansions, globbing, redirections), use a shell wrapper (not yet present in this package) or pass explicit args to `RunCmd`.
- If you can provide programmatic argv (executable + args slice) prefer the explicit form (safer, avoids shell quoting issues).

Example
-------
```go
// explicit args (recommended)
if err := execx.RunCmd(ctx, projectRoot, "git", "init"); err != nil {
		return err
}

// passing multiple args
if err := execx.RunCmd(ctx, projectRoot, "npx", "create-next-app@latest", "frontend", "--ts", "--use-npm"); err != nil {
		return err
}
```

Notes
-----
- Prefer explicit args where possible. When using the shell, be mindful of platform differences and escaping.
- The current implementation captures stdout/stderr into memory buffers; beware of very large command output.

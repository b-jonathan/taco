Title: [Enhancement] Add helper to run whole shell command strings in `execx`

Type: Feature

Description:
`internal/execx.RunCmd` accepts a command name and args. Add a helper that accepts a single shell command string and executes it safely (e.g., via the shell with `-c` or a proper parser).
`internal/execx.RunCmd` currently accepts a command name and args. Add a convenience helper that accepts a single command line string (for example `sh -c "..."` on Unix-like systems) or a small wrapper that parses a single string into command + args with proper escaping.


Location:
- File: `internal/execx/run.go`
- Approx line: 10

Steps to reproduce (current behavior):
1. Call `execx.RunCmd(ctx, dir, "sh", "-c", "...")` is required to run a whole string.

Expected result (after change):
- `execx` exposes `RunShell(ctx, dir, cmdline string)` that handles shell invocation and quoting concerns.

Environment:
- Cross-platform CLI invocations (note: shell invocation differs on Windows vs Unix)

Checks:
- [ ] I searched existing issues
- [ ] I considered cross-platform behavior (Windows `cmd.exe`/PowerShell)

Suggested labels: enhancement, usability, exec

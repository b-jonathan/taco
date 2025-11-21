---
title: prompt (interactive helpers)
---

Purpose
-------
`internal/prompt` provides wrappers around interactive prompt libraries (survey) and exposes helpers that are safe to skip in non-interactive mode.

Key APIs
--------
- helpers for `CreateSurveyInput`, `CreateSurveySelect`, and `CreateSurveyConfirm` that return defaults or errors depending on TTY presence.

Functions (implementation details)
----------------------------------
- `Lock()` / `Unlock()`
	- Purpose: Acquire/release `TermLock` (a `sync.Mutex`) to serialize terminal/prompt usage when multiple goroutines may interact with the terminal.

- `IsTTY() bool`
	- Purpose: Detect whether stdin is a TTY (interactive). Returns false in non-interactive environments (CI). Use this to gate prompts or provide defaults.

- `CreateSurveyInput(message string, options AskOpts) (string, error)`
	- Purpose: Prompt for a single-line input using `survey.Input`.
	- Behavior: If not a TTY and `options.Default` is nil, returns an error. Otherwise builds the prompt and calls the internal `askOneString` which acquires the `TermLock` and calls `survey.AskOne`.

- `CreateSurveySelect(message string, choices []string, options AskOpts) (string, error)`
	- Purpose: Present a single-choice select. Requires non-empty choices.

- `CreateSurveyMultiSelect(message string, choices []string, options AskOpts) ([]string, error)`
	- Purpose: Present a multi-select. Note: current implementation uses `survey.Select` (single-choice) — if you need multiple selections use `survey.MultiSelect` instead.

- `CreateSurveyConfirm(message string, options AskOpts) (bool, error)`
	- Purpose: Yes/no confirmation prompt.

Internal helpers
----------------
- `askOneString`, `askOneBool`, `askManyString`
	- Acquire `TermLock`, call `survey.AskOne`, and return the result. Ensures only one prompt runs at a time in-process.

AskOpts note
------------
The code references an `AskOpts` type that contains fields such as `Default`, `Help`, `PageSize` and `Validator`. Use that struct to supply defaults and validation options to prompts. Inspect `internal/prompt/types.go` for the exact shape.

Recommendations
---------------
- Ensure calling code provides defaults or guards prompt calls when `IsTTY()` is false so automation/CI runs succeed.
- Fix `CreateSurveyMultiSelect` to use `survey.MultiSelect` if multiple-selection UX is desired.

When to use
-----------
- Use prompt helpers in CLI flows when gathering user input; code should fall back to non-interactive behavior when `prompt.IsTTY()` is false.

Notes
-----
- Keep prompts centralized so non-interactive automation (CI) can be supported easily.

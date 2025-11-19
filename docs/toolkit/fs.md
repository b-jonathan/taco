---
title: fsutil (filesystem helpers)
---

Purpose
-------
`internal/fsutil` contains utilities for safely creating files, writing multiple files, appending unique lines, rendering templates, and a small file-lock helper.

Key APIs
--------
- `EnsureFile(path string) error` — create parent directories and an empty file if the target is missing. Uses `os.MkdirAll` for parent dirs.
- `WriteFile(file FileInfo) error` — ensure the file exists (via `EnsureFile`) and write the provided content.
- `WriteMultipleFiles(files []FileInfo) error` — iterate `WriteFile` for multiple files and return on first error.
- `AppendUniqueLines(path string, lines []string) error` — read the file and append each line only when it doesn't already appear (idempotent append).
- `WithFileLock(path string, fn func() error) error` — acquire a per-path mutex (process-level) to run `fn` with exclusive access; useful for concurrent scaffolding operations.
- `RenderTemplate(tmplPath string) ([]byte, error)` — parse and execute a text/template located under `internal/stacks/templates` and return the rendered bytes.

Functions (implementation details)
----------------------------------
- `EnsureFile(path string) error`
	- Creates parent directories and an empty file using `os.OpenFile(..., os.O_CREATE|os.O_EXCL, ...)`. If the file exists, it returns nil.

- `WriteFile(file FileInfo) error`
	- Ensures the file exists and writes `file.Content` to `file.Path` with 0644 permissions.

- `WriteMultipleFiles(files []FileInfo) error`
	- Convenience wrapper that calls `WriteFile` for each element and propagates the first error.

- `AppendUniqueLines(path string, lines []string) error`
	- Reads the file into memory and appends lines that are not already present. Preserves trailing newline semantics.
	- Edge cases: reads the full file into memory; large files may be problematic.

- `WithFileLock(path string, fn func() error) error`
	- Uses a package-level `sync.Map` to store per-absolute-path `*sync.Mutex` values. Locks the mutex, runs `fn`, unlocks.

- `RenderTemplate(tmplPath string) ([]byte, error)`
	- Loads a template from `internal/stacks/templates/<tmplPath>`, executes it with a nil data context (currently), and returns the bytes.
	- Suggestion: accept a data interface{} parameter if templates need dynamic input.

When to use
-----------
- Use these helpers from stack implementations when scaffolding files.

Example
-------
```go
content, _ := fsutil.RenderTemplate("express/src/index.ts.tmpl")
file := fsutil.FileInfo{ Path: filepath.Join(projectRoot, "backend", "src", "index.ts"), Content: content }
_ = fsutil.WriteFile(file)
```

Notes
-----
- `AppendUniqueLines` is useful for idempotent updates to `.gitignore` or `.env` files.
- Add unit tests for edge cases (concurrent appends, files without trailing newline).

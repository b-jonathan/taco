---
title: nodepkg (package helper)
---

Purpose
-------
`internal/nodepkg` helps programmatically create `package.json` contents and common npm scripts for generated Node projects.

Key APIs
--------
- `InitPackage(dir string, params InitPackageParams) error` — create or update `package.json` with given scripts and metadata.

Functions (implementation details)
----------------------------------
- `InitPackage(dir string, params InitPackageParams) error`
	- Purpose: Read an existing `package.json` in `dir`, merge in scripts from `params` (only adding keys that do not already exist), optionally set `name` and `main`, then write the file back with pretty-printed JSON.
	- Behavior:
		- Reads `package.json` into a `map[string]any`.
		- Ensures `pkg["scripts"]` exists and copies in any scripts from `params.Scripts` that are missing.
		- Sets `pkg["name"]` and `pkg["main"]` when provided.
		- Marshals with `json.MarshalIndent` and writes the file.
	- Error modes & notes:
		- If `package.json` is missing the read will fail; consider allowing creation of a minimal `package.json` when desired.
		- This function intentionally avoids overwriting existing scripts to keep scaffolding idempotent.
		- Define and document `InitPackageParams` shape (Scripts map[string]string, Name, Main) for clarity.


When to use
-----------
- Called by stack generators after writing source files to ensure `package.json` contains expected scripts (dev, build, lint, start).

Notes
-----
- Keep dependency lists separate from package writing; consider centralizing dependency lists per stack for easier testability.

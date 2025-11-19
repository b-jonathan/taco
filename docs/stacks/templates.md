---
title: Templates
---

Templates used by stacks are stored under `internal/stacks/templates` and are rendered with `fsutil.RenderTemplate`.

Notes:
- Templates are plain text files with Go `text/template` syntax. Avoid including sensitive data in templates.
- When adding templates, ensure the path used in `RenderTemplate` matches the template location (for example `express/src/index.ts.tmpl`).

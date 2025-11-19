---
title: Troubleshooting
---

Common problems and fixes when running `taco`.

- "chdir ..\project\frontend: The system cannot find the file specified"
  - Symptom: `create-next-app` (npx) didn't create the `frontend` folder.
  - Fix: Run the `npx create-next-app...` command manually to see its output. Ensure `npx` is available on PATH.

- `golangci-lint` not found
  - Fix: Add `$GOPATH/bin` to your PATH or install golangci-lint globally. Example for Git Bash:
    ```bash
    export PATH="$(go env GOPATH)/bin:$PATH"
    ```

- GitHub token / permissions
  - Ensure `GITHUB_TOKEN` has `repo` scope for private repo creation or `public_repo` for public repos.

If a command fails in scaffold, check the generated logs (stdout/stderr captured by `execx`) and run the failing command manually.

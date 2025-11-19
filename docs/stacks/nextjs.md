---
title: NextJS stack
---

## NextJS stack

What it generates:
- `frontend/` created by `create-next-app` with TypeScript, Tailwind, and app directory.

Key implementation points:
- See `internal/stacks/nextjs/nextjs.go`.
- Because this uses `npx create-next-app`, on Windows prefer shell invocation (RunShell) to preserve flags.
Key implementation points:
- See `internal/stacks/nextjs/nextjs.go`.
- Because this uses `npx create-next-app`, on Windows prefer shell invocation (RunShell) to preserve flags.

Init(), Generate(), Post() details
---------------------------------

Init()
- Ensures the project root exists.
- Runs `npx create-next-app` with flags to scaffold a TypeScript + app-directory Next.js project in `frontend/` (flags include `--ts`, `--app`, `--tailwind`, `--use-npm`, `--disable-git`, etc.).
- Installs dev dependencies such as `eslint`, `@eslint/js`, `globals`, `typescript`, `typescript-eslint`, `@next/eslint-plugin-next`, `eslint-plugin-react-hooks`, `eslint-config-prettier`, `prettier`, and `prettier-plugin-tailwindcss`.

Generate()
- Writes or updates these frontend files from templates:
	- `eslint.config.mjs`
	- `.prettierrc.json`
	- `.prettierignore`
- Updates `package.json` via `internal/nodepkg.InitPackage` to add lint scripts:
	- `lint-check`: `next lint && prettier --check .`
	- `lint-fix`: `(next lint --fix || true) && prettier --write .`

Post()
- Ensures `frontend/.env.local` exists and writes a placeholder:
	- `NEXT_PUBLIC_BACKEND_URL=http://localhost:4000`
- Writes `frontend/src/app/page.tsx` from template.

Validation
- After generation the frontend should contain `package.json`, `src/app/page.tsx`, `.env.local`, and the ESLint/Prettier configs.

Notes
- The stack currently constructs the `npx` command as a single string — on Windows this may require shell invocation to preserve flags. Consider using `execx.RunCmdLive`/`RunCmd` appropriately or a tokenizer to handle quoted flags.

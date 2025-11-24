---
title: Express stack
---

## Express stack

What it generates:
- `backend/` folder with `src/index.ts`, `tsconfig.json`, ESLint and Prettier configs, and `package.json` scripts.

Key implementation points:
- See `internal/stacks/express/express.go`.
- The stack runs `npm init -y` then installs runtime and dev dependencies.
Key implementation points:
- See `internal/stacks/express/express.go`.
- The stack runs `npm init -y` then installs runtime and dev dependencies.

Init(), Generate(), Post() details
---------------------------------
Init()
- Creates `backend/` and `backend/src/` directories.
- Runs `npm init -y` inside `backend/`.
- Installs runtime deps: `express`, `cors`, `dotenv`.
- Installs dev deps: `typescript`, `ts-node`, `@types/node`, `@types/express`, `@types/cors`, `eslint`, `@eslint/js`, `globals`, `typescript-eslint`, `eslint-plugin-n`, `eslint-config-prettier`, `prettier`.

Generate()
- Writes these files under `backend/`:
	- `tsconfig.json` (from `internal/stacks/templates/express/tsconfig.json.tmpl`).
	- `src/index.ts` (basic Express app from template).
	- `eslint.config.mjs` (ESLint config template).
	- `.prettierrc.json` and `.prettierignore` (prettier config and ignore).
- Updates `package.json` by calling `internal/nodepkg.InitPackage` to add scripts (idempotent):
	- `dev`: `tsx watch src/index.ts`
	- `build`: `tsc -p tsconfig.json`
	- `start`: `node dist/index.js`
	- `lint-check`: `eslint . && prettier --check .`
	- `lint-fix`: `eslint . --fix && prettier --write .`

Post()
- Ensures a project-level `.gitignore` exists and appends these entries (idempotent):
	- `backend/node_modules/`
	- `backend/dist/`
	- `backend/.env*`
- Creates `backend/.env` with default values:
	- PORT=4000
	- FRONTEND_ORIGIN=http://localhost:3000

Validation
- After generation the backend should contain `src/index.ts`, `tsconfig.json`, ESLint/Prettier configs, and `package.json` with the scripts above.

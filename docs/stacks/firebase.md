---
title: Firebase stack
---

## Firebase stack

What it generates:
- Adds Firebase authentication wiring to a Next.js frontend. Files are written under `frontend/src/` (components, pages, auth context, and Firebase helpers).

Compatibility
-------------
- Frontend stacks: `nextjs`. The generated templates assume Next.js app directory structure under `frontend/src/`.

Key implementation points
-------------------------
- See `internal/stacks/firebase/firebase.go` and `internal/stacks/firebase/helper.go`.
- The stack is interactive: it requires the Firebase CLI or a `FIREBASE_TOKEN` and may prompt to install the CLI, log in, create a Firebase project, and create a Firebase web app.

Init(), Generate(), Post() details
-----------------------------------------
Init()
- Checks for the Firebase CLI (`firebase`) on PATH. If missing it offers to install it globally via `npm install -g firebase-tools` (interactive prompt).
- If `FIREBASE_TOKEN` is set the stack validates it using `firebase projects:list --non-interactive`; otherwise it detects an active session or prompts the user to run `firebase login` interactively (with browser flow via `RunCmdLive`).
- Creates a Firebase Project named `<appName>-taco` and a Firebase Web App named `<appName>-web` using the Firebase CLI. These operations are executed with `execx.RunCmdLive` to surface interactive steps and progress.
- Prompts the user to open the Firebase Console to enable recommended Authentication providers (Email/Password and Google) and requires confirmation they were enabled before continuing.

Generate()
- Installs the `firebase` JS SDK in the frontend: `npm install firebase`.
- Writes multiple Next.js files from templates into `frontend/src/`:
  - `app/components/Header.tsx`
  - `app/home/page.tsx`
  - `app/login/page.tsx`
  - `app/register/page.tsx`
  - `app/layout.tsx`
  - `context/authContext/index.tsx`
  - `firebase/auth.ts`
  - `firebase/firebase.ts`
- These templates wire a simple auth flow that uses the Firebase Web SDK and a React context for auth state.

Post()
- Appends Firebase-specific ignores to `frontend/.gitignore` (idempotent):
  - `# firebase`
  - `.firebase/`
  - `.firebasehosting.*`
  - `firebase-debug.log`
  - `firestore-debug.log`
  - `ui-debug.log`
- Calls `createCredentials` which runs `firebase apps:sdkconfig web --project <projectID>` to fetch the web app SDK config, extracts the JSON, and appends these environment variables to `frontend/.env.local` (idempotent):
  - `NEXT_PUBLIC_FIREBASE_API_KEY`
  - `NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN`
  - `NEXT_PUBLIC_FIREBASE_PROJECT_ID`
  - `NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET`
  - `NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID`
  - `NEXT_PUBLIC_FIREBASE_APP_ID`

Validation
- After generation you should see the Firebase UI and auth helpers under `frontend/src/`, the `firebase` package in `frontend/package.json`, `frontend/.env.local` populated with NEXT_PUBLIC_FIREBASE_* keys, and `.gitignore` updated with Firebase entries.

Notes / cautions
- The stack performs live operations against your Firebase account (project/app creation). Make sure you want the project created under your account and verify the generated project ID.
- For non-interactive usage, set `FIREBASE_TOKEN` in the environment and ensure the token has the required permissions.

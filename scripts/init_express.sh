#!/usr/bin/env bash
set -euo pipefail

echo "⚡ Initializing Express + TypeScript in $(pwd)"

# --- npm + deps ---
if [ ! -f package.json ]; then
  npm init -y
fi

npm install express detect-port dotenv
npm install --save-dev typescript ts-node ts-node-dev @types/node @types/express

# --- copy template files ---
# assumes your initializer repo has a templates/express-ts/ folder
# with tsconfig.json, src/server.ts, .env.example, .gitignore, etc.
TEMPLATE_DIR="$(dirname "$0")/../express"

echo "📂 Copying template from $TEMPLATE_DIR"

# -R = recursive, -n = no-clobber (don’t overwrite existing files)
cp -Rn "$TEMPLATE_DIR"/. .

# --- package.json scripts patch ---
node - <<'EOF'
const fs = require('fs');
const pkg = JSON.parse(fs.readFileSync('package.json','utf8'));
pkg.type = "commonjs";
pkg.scripts = Object.assign({
  "dev": "ts-node-dev --respawn --transpile-only src/server.ts",
  "build": "tsc",
  "start": "node dist/server.js"
}, pkg.scripts || {});
fs.writeFileSync('package.json', JSON.stringify(pkg, null, 2));
console.log("✔ package.json updated with dev/build/start scripts");
EOF

echo "✅ Express + TS scaffold ready."
echo "   Next:"
echo "     cp .env.example .env"
echo "     npm run dev"

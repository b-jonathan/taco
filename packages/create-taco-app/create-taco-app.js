#!/usr/bin/env node
"use strict";

const path = require("path");
const { execFileSync } = require("child_process");
const os = require("os");

const ext = os.platform() === "win32" ? ".exe" : "";
const binaryPath = path.join(__dirname, "taco" + ext);

const userArgs = process.argv.slice(2);

try {
  execFileSync(binaryPath, ["init", ...userArgs], { stdio: "inherit" });
} catch (err) {
  process.exit(err.status || 1);
}

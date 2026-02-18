const os = require("os");
const fs = require("fs");
const path = require("path");

function resolveBinary() {
  const platform = os.platform();
  const arch = os.arch();

  if (platform === "darwin" && arch === "x64") return "taco-darwin-amd64";
  if (platform === "darwin" && arch === "arm64") return "taco-darwin-arm64";
  if (platform === "linux" && arch === "x64") return "taco-linux-amd64";
  if (platform === "linux" && arch === "arm64") return "taco-linux-arm64";
  if (platform === "win32" && arch === "x64") return "taco-windows-amd64.exe";

  console.error(`Unsupported OS or architecture: ${platform} ${arch}`);
  process.exit(1);
}

function installWrapper() {
  const binName = resolveBinary();
  const binPath = path.join(__dirname, "bin", binName);
  const wrapperPath = path.join(__dirname, "taco");

  if (!fs.existsSync(binPath)) {
    console.error(`Binary not found for your system: ${binPath}`);
    process.exit(1);
  }

  if (os.platform() === "win32") {
    fs.writeFileSync(
      path.join(__dirname, "taco.cmd"),
      `@echo off\r\n"${binPath}" %*\r\n`,
      "utf8"
    );
    if (fs.existsSync(wrapperPath)) fs.unlinkSync(wrapperPath);
  } else {
    fs.writeFileSync(
      wrapperPath,
      `#!/usr/bin/env sh\n"${binPath}" "$@"\n`,
      "utf8"
    );
    fs.chmodSync(wrapperPath, 0o755);
  }

  console.log("Taco CLI wrapper installed");
}

installWrapper();

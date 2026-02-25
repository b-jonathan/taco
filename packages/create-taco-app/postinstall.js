#!/usr/bin/env node
"use strict";

const os = require("os");
const fs = require("fs");
const path = require("path");
const https = require("https");
const http = require("http");

const REPO = "siddharth-mdk/taco";
const VERSION = require("./package.json").version;
const MAX_REDIRECTS = 5;

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

const SUPPORTED = [
  "darwin-amd64",
  "darwin-arm64",
  "linux-amd64",
  "linux-arm64",
  "windows-amd64",
];

function getBinaryName(platform, arch) {
  const goOS = PLATFORM_MAP[platform];
  const goArch = ARCH_MAP[arch];

  if (!goOS || !goArch) return null;

  const key = `${goOS}-${goArch}`;
  if (!SUPPORTED.includes(key)) return null;

  const ext = goOS === "windows" ? ".exe" : "";
  return `taco-${key}${ext}`;
}

function downloadFile(url, redirectCount) {
  if (redirectCount === undefined) redirectCount = 0;

  return new Promise((resolve, reject) => {
    if (redirectCount > MAX_REDIRECTS) {
      return reject(new Error("Too many redirects"));
    }

    const client = url.startsWith("https") ? https : http;

    client
      .get(url, (res) => {
        if (
          res.statusCode >= 300 &&
          res.statusCode < 400 &&
          res.headers.location
        ) {
          return downloadFile(res.headers.location, redirectCount + 1).then(
            resolve,
            reject
          );
        }

        if (res.statusCode !== 200) {
          return reject(
            new Error(`Download failed: HTTP ${res.statusCode} from ${url}`)
          );
        }

        const chunks = [];
        res.on("data", (chunk) => chunks.push(chunk));
        res.on("end", () => resolve(Buffer.concat(chunks)));
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

async function main() {
  const platform = os.platform();
  const arch = os.arch();
  const binaryName = getBinaryName(platform, arch);

  if (!binaryName) {
    console.error(
      `[create-taco-app] Unsupported platform: ${platform}-${arch}\n` +
        `Supported: darwin-x64, darwin-arm64, linux-x64, linux-arm64, win32-x64`
    );
    process.exit(1);
  }

  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryName}`;
  const ext = platform === "win32" ? ".exe" : "";
  const dest = path.join(__dirname, "taco" + ext);

  console.log(
    `[create-taco-app] Downloading taco v${VERSION} for ${platform}-${arch}...`
  );

  try {
    const data = await downloadFile(url);
    fs.writeFileSync(dest, data);
    fs.chmodSync(dest, 0o755);
    console.log(`[create-taco-app] Binary installed successfully`);
  } catch (err) {
    console.error(
      `[create-taco-app] Failed to download binary: ${err.message}`
    );
    console.error(
      `[create-taco-app] Ensure that release v${VERSION} exists at:\n` +
        `  https://github.com/${REPO}/releases/tag/v${VERSION}`
    );
    process.exit(1);
  }
}

main();

#!/usr/bin/env node
"use strict";

const { execFileSync } = require("child_process");
const fs = require("fs");
const https = require("https");
const path = require("path");

const VERSION = require("./package.json").version;
const MAX_REDIRECTS = 5;
const MAX_RETRIES = 3;
const RETRY_BASE_MS = 1000;

const PLATFORM_MAP = { darwin: "darwin", linux: "linux", win32: "windows" };
const ARCH_MAP = { x64: "amd64", arm64: "arm64" };

function getBinaryName() {
  const os = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];
  if (!os || !arch) {
    console.error(
      `Unsupported platform: ${process.platform}-${process.arch}\n` +
        "Supported: darwin-x64, darwin-arm64, linux-x64, linux-arm64, win32-x64, win32-arm64"
    );
    process.exit(1);
  }
  const ext = process.platform === "win32" ? ".exe" : "";
  return { remote: `trabuco-${os}-${arch}${ext}`, local: `trabuco${ext}` };
}

function getBinDir() {
  const dir = path.join(__dirname, "bin");
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  return dir;
}

function versionMatches(binaryPath) {
  try {
    const out = execFileSync(binaryPath, ["version"], {
      encoding: "utf8",
      timeout: 5000,
    }).trim();
    // Output may be "trabuco version v1.5.3" or just "v1.5.3"
    return out.includes(VERSION) || out.includes(`v${VERSION}`);
  } catch {
    return false;
  }
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    followRedirects(url, 0, (err, res) => {
      if (err) return reject(err);
      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`HTTP ${res.statusCode} from ${url}`));
      }
      const file = fs.createWriteStream(dest);
      res.pipe(file);
      file.on("finish", () => file.close(resolve));
      file.on("error", (e) => {
        fs.unlink(dest, () => {});
        reject(e);
      });
    });
  });
}

function followRedirects(url, hops, cb) {
  if (hops >= MAX_REDIRECTS) return cb(new Error("Too many redirects"));
  const get = url.startsWith("https://") ? https.get : require("http").get;
  get(url, (res) => {
    if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
      res.resume();
      return followRedirects(res.headers.location, hops + 1, cb);
    }
    cb(null, res);
  }).on("error", cb);
}

async function downloadWithRetry(url, dest) {
  for (let attempt = 1; attempt <= MAX_RETRIES; attempt++) {
    try {
      await download(url, dest);
      return;
    } catch (err) {
      if (attempt === MAX_RETRIES) throw err;
      const delay = RETRY_BASE_MS * Math.pow(2, attempt - 1);
      console.log(
        `Download attempt ${attempt}/${MAX_RETRIES} failed: ${err.message}. Retrying in ${delay}ms...`
      );
      await new Promise((r) => setTimeout(r, delay));
    }
  }
}

async function main() {
  const { remote, local } = getBinaryName();
  const binDir = getBinDir();
  const binaryPath = path.join(binDir, local);

  // Skip download if binary exists and version matches
  if (fs.existsSync(binaryPath) && versionMatches(binaryPath)) {
    console.log(`trabuco v${VERSION} already installed.`);
    return;
  }

  const url = `https://github.com/arianlopezc/Trabuco/releases/download/v${VERSION}/${remote}`;
  console.log(`Downloading trabuco v${VERSION} for ${process.platform}-${process.arch}...`);

  try {
    await downloadWithRetry(url, binaryPath);
  } catch (err) {
    console.error(
      `Failed to download trabuco binary.\n` +
        `  URL: ${url}\n` +
        `  Error: ${err.message}\n\n` +
        `You can install the CLI manually:\n` +
        `  curl -sSL https://raw.githubusercontent.com/arianlopezc/Trabuco/main/scripts/install.sh | bash`
    );
    process.exit(1);
  }

  if (process.platform !== "win32") {
    fs.chmodSync(binaryPath, 0o755);
  }

  console.log(`trabuco v${VERSION} installed successfully.`);
}

main();

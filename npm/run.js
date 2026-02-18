#!/usr/bin/env node
"use strict";

const { spawn } = require("child_process");
const path = require("path");

const ext = process.platform === "win32" ? ".exe" : "";
const binaryPath = path.join(__dirname, "bin", `trabuco${ext}`);

const fs = require("fs");
if (!fs.existsSync(binaryPath)) {
  console.error(
    "trabuco binary not found. This usually means npm was run with --ignore-scripts.\n" +
      "Run the postinstall script manually:\n\n" +
      "  cd " + path.dirname(__filename) + " && node install.js\n"
  );
  process.exit(1);
}

const child = spawn(binaryPath, ["mcp"], {
  stdio: "inherit",
});

// Forward signals to child process
for (const sig of ["SIGINT", "SIGTERM", "SIGHUP"]) {
  process.on(sig, () => {
    if (!child.killed) child.kill(sig);
  });
}

child.on("close", (code, signal) => {
  if (signal) {
    // Re-raise the signal so the parent sees the correct exit reason
    process.kill(process.pid, signal);
  } else {
    process.exit(code ?? 1);
  }
});

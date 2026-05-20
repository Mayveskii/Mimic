#!/usr/bin/env node

/**
 * bin/mimic.js — npm wrapper for Mimic binary.
 * Spawns the platform-native binary downloaded by scripts/install.js.
 */

const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");

const binaryName = process.platform === "win32" ? "mimic-native.exe" : "mimic-native";
const binaryPath = path.join(__dirname, binaryName);

function run() {
  if (!fs.existsSync(binaryPath)) {
    console.error("[mimic] Native binary not found. Running postinstall...");
    require("../scripts/install.js");
    if (!fs.existsSync(binaryPath)) {
      console.error("[mimic] Failed to install native binary.");
      process.exit(1);
    }
  }

  const child = spawn(binaryPath, process.argv.slice(2), {
    stdio: "inherit",
    env: process.env,
  });

  child.on("exit", (code) => {
    process.exit(code ?? 0);
  });
}

run();

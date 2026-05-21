#!/usr/bin/env node

/**
 * Wrapper script for @mayveskii/mimic
 * Transparently forwards all arguments to the downloaded native binary.
 */

const { spawn } = require('child_process');
const path = require('path');

const binaryPath = path.join(__dirname, 'mimic');

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  windowsHide: true,
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code ?? 0);
  }
});

child.on('error', (err) => {
  console.error(`[mimic] Failed to run binary: ${err.message}`);
  process.exit(1);
});

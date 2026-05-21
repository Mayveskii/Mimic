#!/usr/bin/env node

/**
 * Test script for @mayveskii/mimic
 * Verifies the binary exists and is executable
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

function errorExit(message) {
  console.error(`[mimic] Test FAILED: ${message}`);
  process.exit(1);
}

function success(message) {
  console.log(`[mimic] Test PASSED: ${message}`);
}

const binaryPath = path.join(__dirname, '..', 'bin', 'mimic');

// 1. Check file exists
if (!fs.existsSync(binaryPath)) {
  errorExit(`Binary not found at ${binaryPath}`);
}
success(`Binary exists at ${binaryPath}`);

// 2. Check file is executable
try {
  fs.accessSync(binaryPath, fs.constants.X_OK);
  success('Binary is executable');
} catch (err) {
  errorExit(`Binary is not executable: ${err.message}`);
}

// 3. Run binary --version (or --help as fallback) to verify it works
try {
  const output = execSync(`"${binaryPath}" --version`, { encoding: 'utf8', timeout: 5000 });
  success(`Binary runs and reports version: ${output.trim()}`);
} catch (err) {
  // If --version fails, try --help as a weaker check
  try {
    const output = execSync(`"${binaryPath}" --help`, { encoding: 'utf8', timeout: 5000 });
    success(`Binary runs (version check failed, --help works)`);
  } catch (helpErr) {
    errorExit(`Binary exists but cannot execute: ${err.message}`);
  }
}

console.log('[mimic] All tests passed.');

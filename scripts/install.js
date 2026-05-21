#!/usr/bin/env node

/**
 * Postinstall script for @mayveskii/mimic
 * Downloads the correct binary from GitHub Release
 */

const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { execSync } = require('child_process');

function errorExit(message) {
  console.error(`[mimic] Error: ${message}`);
  process.exit(1);
}

function info(message) {
  console.log(`[mimic] ${message}`);
}

const platform = os.platform();
const arch = os.arch();

if (platform !== 'linux' || arch !== 'x64') {
  errorExit(`Unsupported platform: ${platform}/${arch}. This package currently only supports linux/x64.`);
}

const packageJsonPath = path.join(__dirname, '..', 'package.json');
let version;

try {
  const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
  version = pkg.version;
} catch (err) {
  errorExit(`Failed to read package.json: ${err.message}`);
}

const binDir = path.join(__dirname, '..', 'bin');
const binaryPath = path.join(binDir, 'mimic');
const tarUrl = `https://github.com/Mayveskii/Mimic/releases/download/v${version}/mimic_v${version}_linux_amd64.tar.gz`;
const tarPath = path.join(binDir, `mimic_v${version}_linux_amd64.tar.gz`);

// If binary already exists, skip download
if (fs.existsSync(binaryPath)) {
  info('Binary already exists, skipping download.');
  process.exit(0);
}

info(`Downloading mimic binary v${version} for linux/amd64...`);

try {
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }
} catch (err) {
  errorExit(`Failed to create bin directory: ${err.message}`);
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https
      .get(url, { timeout: 30000 }, (response) => {
        if (response.statusCode === 301 || response.statusCode === 302) {
          const redirectUrl = response.headers.location;
          if (!redirectUrl) {
            reject(new Error('Redirect received but no Location header'));
            return;
          }
          file.close();
          fs.unlinkSync(dest);
          downloadFile(redirectUrl, dest)
            .then(resolve)
            .catch(reject);
          return;
        }

        if (response.statusCode !== 200) {
          file.close();
          fs.unlinkSync(dest);
          reject(new Error(`Download failed with HTTP ${response.statusCode}`));
          return;
        }

        response.pipe(file);
        file.on('finish', () => {
          file.close(resolve);
        });
      })
      .on('error', (err) => {
        fs.unlinkSync(dest);
        reject(err);
      })
      .on('timeout', () => {
        file.close();
        fs.unlinkSync(dest);
        reject(new Error('Download request timed out'));
      });
  });
}

(async () => {
  try {
    await downloadFile(tarUrl, tarPath);
    info('Download complete. Extracting...');

    try {
      execSync(`tar -xzf "${tarPath}" -C "${binDir}"`, { stdio: 'inherit' });
    } catch (err) {
      errorExit(`Failed to extract archive: ${err.message}`);
    }

    // Clean up tar.gz
    try {
      fs.unlinkSync(tarPath);
    } catch (err) {
      info(`Warning: could not remove temporary archive: ${err.message}`);
    }

    // Make binary executable
    try {
      fs.chmodSync(binaryPath, 0o755);
    } catch (err) {
      errorExit(`Failed to make binary executable: ${err.message}`);
    }

    info('Installation complete. Binary available at bin/mimic');
  } catch (err) {
    // Clean up partial download if exists
    if (fs.existsSync(tarPath)) {
      try {
        fs.unlinkSync(tarPath);
      } catch {}
    }
    errorExit(`Network error during download: ${err.message}`);
  }
})();

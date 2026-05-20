const https = require("https");
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");
const { execSync } = require("child_process");

const pkg = require("../package.json");

// Map Node platform/arch to release asset names
const platformMap = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const archMap = {
  x64: "amd64",
  arm64: "arm64",
};

function getAssetName() {
  const platform = platformMap[process.platform];
  const arch = archMap[process.arch];
  if (!platform || !arch) {
    throw new Error(
      `Unsupported platform: ${process.platform} ${process.arch}. ` +
        `Supported: linux/darwin/win32 on x64/arm64.`
    );
  }
  return `mimic-${platform}-${arch}`;
}

function getVersion() {
  // Allow env override for testing/dev installs
  return process.env.MIMIC_VERSION || pkg.version;
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https
      .get(url, { timeout: 30000 }, (res) => {
        if (res.statusCode === 302 || res.statusCode === 301) {
          download(res.headers.location, dest).then(resolve).catch(reject);
          return;
        }
        if (res.statusCode !== 200) {
          reject(new Error(`Download failed: HTTP ${res.statusCode} for ${url}`));
          return;
        }
        res.pipe(file);
        file.on("finish", () => {
          file.close(resolve);
        });
      })
      .on("error", (err) => {
        fs.unlink(dest, () => {});
        reject(err);
      });
  });
}

async function verifySha256(filePath, shaUrl) {
  try {
    const shaData = await new Promise((resolve, reject) => {
      https
        .get(shaUrl, { timeout: 10000 }, (res) => {
          let data = "";
          res.on("data", (chunk) => (data += chunk));
          res.on("end", () => resolve(data));
        })
        .on("error", reject);
    });
    const expected = shaData.trim().split(/\s+/)[0];
    const actual = crypto
      .createHash("sha256")
      .update(fs.readFileSync(filePath))
      .digest("hex");
    if (expected !== actual) {
      throw new Error(
        `SHA256 mismatch: expected ${expected}, got ${actual}`
      );
    }
    console.log("[mimic] SHA256 verified.");
  } catch (e) {
    console.warn(`[mimic] SHA256 verification skipped: ${e.message}`);
  }
}

async function install() {
  const version = getVersion();
  const asset = getAssetName();
  const releaseBase = `https://github.com/Mayveskii/Mimic/releases/download/v${version}`;
  const binUrl = `${releaseBase}/${asset}`;
  const shaUrl = `${releaseBase}/${asset}.sha256`;

  const targetDir = path.join(__dirname, "..", "bin");
  if (!fs.existsSync(targetDir)) {
    fs.mkdirSync(targetDir, { recursive: true });
  }

  const targetPath = path.join(
    targetDir,
    process.platform === "win32" ? "mimic-native.exe" : "mimic-native"
  );

  if (fs.existsSync(targetPath)) {
    console.log("[mimic] Binary already exists.");
    return;
  }

  console.log(`[mimic] Downloading ${asset} v${version}...`);
  await download(binUrl, targetPath);

  await verifySha256(targetPath, shaUrl);

  if (process.platform !== "win32") {
    fs.chmodSync(targetPath, 0o755);
  }

  console.log(`[mimic] Installed to ${targetPath}`);
}

install().catch((err) => {
  console.error(`[mimic] Install failed: ${err.message}`);
  process.exit(1);
});

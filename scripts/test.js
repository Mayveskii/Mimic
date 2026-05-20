const fs = require("fs");
const path = require("path");

const binaryName = process.platform === "win32" ? "mimic-native.exe" : "mimic-native";
const binaryPath = path.join(__dirname, "..", "bin", binaryName);

if (!fs.existsSync(binaryPath)) {
  console.error("[mimic-test] Binary not found:", binaryPath);
  console.error("Run: npm run postinstall");
  process.exit(1);
}

console.log("[mimic-test] Binary present:", binaryPath);
console.log("[mimic-test] OK");

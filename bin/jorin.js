#!/usr/bin/env node
"use strict";

const fs = require("fs");
const os = require("os");
const path = require("path");
const https = require("https");
const { spawn } = require("child_process");

const PACKAGE_JSON = path.join(__dirname, "..", "package.json");
const pkg = JSON.parse(fs.readFileSync(PACKAGE_JSON, "utf8"));

function resolveTarget() {
  let platform = process.platform;
  if (platform === "android") {
    platform = "android";
  }

  switch (platform) {
    case "linux":
    case "darwin":
    case "freebsd":
    case "windows":
    case "android":
      break;
    case "win32":
      platform = "windows";
      break;
    default:
      throw new Error(`Unsupported platform: ${process.platform}`);
  }

  let arch;
  switch (process.arch) {
    case "x64":
      arch = "amd64";
      break;
    case "arm64":
      arch = "arm64";
      break;
    case "arm":
      arch = "arm";
      break;
    case "ia32":
      arch = "386";
      break;
    default:
      throw new Error(`Unsupported architecture: ${process.arch}`);
  }

  if (platform === "windows" && arch === "arm64") {
    throw new Error("Windows ARM64 builds are not available yet.");
  }

  const ext = platform === "windows" ? ".exe" : "";
  return { platform, arch, ext };
}

function resolveVersion() {
  const override = process.env.JORIN_NPX_VERSION;
  if (override) {
    return { tag: override.startsWith("v") ? override : `v${override}`, mode: "tag" };
  }

  if (pkg.version && pkg.version !== "0.0.0") {
    return { tag: `v${pkg.version}`, mode: "tag" };
  }

  return { tag: "latest", mode: "latest" };
}

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function downloadFile(url, destination) {
  return new Promise((resolve, reject) => {
    const request = https.get(url, (response) => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        response.resume();
        return resolve(downloadFile(response.headers.location, destination));
      }

      if (response.statusCode !== 200) {
        response.resume();
        return reject(new Error(`Download failed with status ${response.statusCode}`));
      }

      const file = fs.createWriteStream(destination, { mode: 0o755 });
      response.pipe(file);
      file.on("finish", () => file.close(resolve));
      file.on("error", reject);
    });

    request.on("error", reject);
  });
}

async function main() {
  const { platform, arch, ext } = resolveTarget();
  const { tag, mode } = resolveVersion();

  const baseDir = process.env.JORIN_NPX_DIR || path.join(os.homedir(), ".jorin", "bin");
  ensureDir(baseDir);

  const assetName = `jorin-${platform}-${arch}${ext}`;
  const cacheName = `jorin-${platform}-${arch}-${tag}${ext}`.replace(/[/:]/g, "-");
  const targetPath = path.join(baseDir, cacheName);
  const bundledPath = path.join(__dirname, "..", "dist", assetName);

  const shouldRedownload = process.env.JORIN_NPX_FORCE === "1" || !fs.existsSync(targetPath);
  if (shouldRedownload) {
    if (fs.existsSync(bundledPath)) {
      fs.copyFileSync(bundledPath, targetPath);
    } else {
      const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "jorin-npx-"));
      const tmpPath = path.join(tmpDir, assetName);

      const url =
        mode === "latest"
          ? `https://github.com/dave1010/jorin/releases/latest/download/${assetName}`
          : `https://github.com/dave1010/jorin/releases/download/${tag}/${assetName}`;

      process.stderr.write(`Downloading ${assetName} (${mode === "latest" ? "latest" : tag})...\n`);
      await downloadFile(url, tmpPath);

      if (platform !== "windows") {
        fs.chmodSync(tmpPath, 0o755);
      }

      fs.renameSync(tmpPath, targetPath);
    }
  }

  if (platform !== "windows") {
    fs.chmodSync(targetPath, 0o755);
  }

  const child = spawn(targetPath, process.argv.slice(2), { stdio: "inherit" });
  child.on("exit", (code) => process.exit(code ?? 0));
  child.on("error", (err) => {
    process.stderr.write(`Failed to launch jorin: ${err.message}\n`);
    process.exit(1);
  });
}

main().catch((err) => {
  process.stderr.write(`${err.message}\n`);
  process.exit(1);
});

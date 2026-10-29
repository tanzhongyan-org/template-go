#!/usr/bin/env node
// Pinned gitleaks runner: downloads the release binary for the current
// platform, verifies its checksum, caches it under bin/, and scans the
// worktree. The Makefile passes GITLEAKS_VERSION through.
//
// Usage: node scripts/gitleaks.mjs <version>

import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { arch, platform } from "node:os";
import { join } from "node:path";

const version = process.argv[2];
if (!version || !/^\d+\.\d+\.\d+$/.test(version)) {
  console.error("gitleaks: pass a version like 8.30.1");
  process.exit(2);
}

const PLATFORMS = {
  "linux:x64": "linux_x64.tar.gz",
  "linux:arm64": "linux_arm64.tar.gz",
  "darwin:x64": "darwin_x64.tar.gz",
  "darwin:arm64": "darwin_arm64.tar.gz",
  "win32:x64": "windows_x64.zip",
  "win32:arm64": "windows_arm64.zip",
};

const key = `${platform()}:${arch()}`;
const asset = PLATFORMS[key];
if (!asset) {
  console.error(`gitleaks: unsupported platform ${key}`);
  process.exit(2);
}

const binary = platform() === "win32" ? "bin/gitleaks.exe" : "bin/gitleaks";
const stamp = "bin/.gitleaks-version";

async function download(url) {
  const res = await fetch(url, { signal: AbortSignal.timeout(120000) });
  if (!res.ok) throw new Error(`GET ${url}: ${res.status}`);
  return Buffer.from(await res.arrayBuffer());
}

async function install() {
  const base = `https://github.com/gitleaks/gitleaks/releases/download/v${version}`;
  const archive = `gitleaks_${version}_${asset}`;
  const [data, sums] = await Promise.all([
    download(`${base}/${archive}`),
    download(`${base}/gitleaks_${version}_checksums.txt`),
  ]);
  const want = sums
    .toString("utf8")
    .split("\n")
    .map((line) => line.trim().split(/\s+/))
    .find(([, name]) => name === archive)?.[0];
  if (!want) throw new Error(`gitleaks: no checksum for ${archive}`);
  const got = createHash("sha256").update(data).digest("hex");
  if (got !== want.toLowerCase()) throw new Error("gitleaks: checksum mismatch");
  const file = join("bin", archive);
  writeFileSync(file, data);
  const args = archive.endsWith(".zip") ? ["-xf", file, "-C", "bin"] : ["-xzf", file, "-C", "bin"];
  execFileSync("tar", args, { stdio: "inherit" });
  rmSync(file, { force: true });
  rmSync("bin/LICENSE", { force: true });
  rmSync("bin/README.md", { force: true });
  writeFileSync(stamp, `${version}\n`);
}

mkdirSync("bin", { recursive: true });
const installed = existsSync(stamp) && readFileSync(stamp, "utf8").trim() === version;
if (!installed || !existsSync(binary)) {
  await install();
}

execFileSync(binary, ["detect", "--source", ".", "--no-banner", "--redact", "--exit-code", "1"], {
  stdio: "inherit",
});
console.log("gitleaks: ok");

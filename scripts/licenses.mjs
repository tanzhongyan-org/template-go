#!/usr/bin/env node
// License gate: fails when any dependency license falls outside the D1
// allowlist (MIT, Apache-2.0, ISC, BSD, CC0, OFL-1.1). Reads NOTICE, which
// `make licenses` regenerates from `go tool go-licenses report` first.
//
// Usage: node scripts/licenses.mjs
//
// Unknown licenses fail. LGPL, MPL and EPL fail unless the owner records an
// exception in EXCEPTIONS below. Dev-only Go tools are exempt: the report
// covers the ./... module closure, which never includes tool dependencies.

import { readFileSync } from "node:fs";

const ALLOW = new Set([
  "MIT",
  "Apache-2.0",
  "ISC",
  "BSD-2-Clause",
  "BSD-3-Clause",
  "BSD-3-Clause-Clear",
  "CC0-1.0",
  "OFL-1.1",
  "Unlicense",
]);

// Owner-approved exceptions: module path -> reason. Add LGPL, MPL or EPL
// modules here only with a recorded decision.
const EXCEPTIONS = new Map([
  // ["example.com/mod", "owner exception 2026-10-29: <reason>"],
]);

function modulePath() {
  for (const line of readFileSync("go.mod", "utf8").split("\n")) {
    const m = line.match(/^module\s+(\S+)/);
    if (m) return m[1];
  }
  throw new Error("licenses: no module line in go.mod");
}

const self = modulePath();
const errors = [];
for (const line of readFileSync("NOTICE", "utf8").split("\n")) {
  const trimmed = line.trim();
  if (!trimmed) continue;
  const [pkg, , ...licenses] = trimmed.split(",");
  if (pkg === self) continue;
  if (EXCEPTIONS.has(pkg)) continue;
  if (licenses.length === 0) {
    errors.push(`${pkg}: no license detected`);
    continue;
  }
  for (const license of licenses) {
    if (!ALLOW.has(license)) {
      errors.push(`${pkg}: license ${license} is outside the allowlist`);
    }
  }
}

if (errors.length > 0) {
  for (const e of errors) console.error(`licenses: ${e}`);
  process.exit(1);
}
console.log("licenses: ok");

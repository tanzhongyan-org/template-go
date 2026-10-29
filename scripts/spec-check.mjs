#!/usr/bin/env node
// Spec-diff gate: regenerates the TypeScript client from api/openapi.yaml into
// a temp dir and fails on any difference to the committed clients/ts. The
// Makefile passes the pinned hey-api and TypeScript versions through.
//
// Usage: node scripts/spec-check.mjs <hey-api-version> <typescript-version>

import { execSync } from "node:child_process";
import { mkdtempSync, readFileSync, readdirSync, rmSync, statSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, relative } from "node:path";

const [heyApiVersion, tsVersion] = process.argv.slice(2);
for (const [name, value] of [
  ["hey-api version", heyApiVersion],
  ["TypeScript version", tsVersion],
]) {
  if (!value || !/^\d+\.\d+\.\d+$/.test(value)) {
    console.error(`spec-check: pass ${name} like 0.99.0`);
    process.exit(2);
  }
}

function tree(root) {
  const files = new Map();
  const walk = (dir) => {
    for (const entry of readdirSync(dir)) {
      const full = join(dir, entry);
      if (statSync(full).isDirectory()) {
        walk(full);
      } else {
        files.set(relative(root, full), readFileSync(full));
      }
    }
  };
  walk(root);
  return files;
}

const tmp = mkdtempSync(join(tmpdir(), "spec-check-"));
try {
  execSync(
    `npx --yes --package @hey-api/openapi-ts@${heyApiVersion} --package typescript@${tsVersion} openapi-ts -i ./api/openapi.yaml -o ${join(tmp, "ts")} -c @hey-api/client-fetch`,
    { stdio: "inherit" },
  );
  const fresh = tree(join(tmp, "ts"));
  const committed = tree("clients/ts");
  const diffs = [];
  for (const [name, content] of fresh) {
    if (!committed.has(name)) diffs.push(`only generated: ${name}`);
    else if (!content.equals(committed.get(name))) diffs.push(`changed: ${name}`);
  }
  for (const name of committed.keys()) {
    if (!fresh.has(name)) diffs.push(`only committed: ${name}`);
  }
  if (diffs.length > 0) {
    console.error("spec-check: clients/ts drifted from api/openapi.yaml:");
    for (const d of diffs.slice(0, 20)) console.error(`  ${d}`);
    console.error("Run `make openapi-client` and commit the result.");
    process.exit(1);
  }
  console.log("spec-check: clients/ts matches api/openapi.yaml");
} finally {
  rmSync(tmp, { recursive: true, force: true });
}

#!/usr/bin/env node
// New-dependency gate: every direct module in go.mod must exist on the proxy,
// be at least MIN_AGE_DAYS old, and have a row in the dependency ledger.
// Tool directives resolve to their parent module for the age check but keep
// their full path for the ledger check, matching scripts/repo-policy.mjs.
//
// Usage: node scripts/new-deps.mjs

const MIN_AGE_DAYS = 14;
const MS_PER_DAY = 24 * 60 * 60 * 1000;

// Fast-tracked versions that have not reached MIN_AGE_DAYS yet:
// "module@version" -> reason. Use only for urgent security fixes.
const EXCEPTIONS = new Map([
  // ["example.com/mod@v1.2.4", "CVE-XXXX fast-track, 2026-10-29"],
]);

import { readFileSync } from "node:fs";

function directModules() {
  const text = readFileSync("go.mod", "utf8");
  const versions = new Map();
  const body = text.replace(/^.*\/\/\s*indirect.*$/gm, "");
  for (const [, block] of body.matchAll(/^require\s*\(([\s\S]*?)^\)/gm)) {
    for (const line of block.split("\n")) {
      const m = line.trim().match(/^(\S+)\s+(v\S+)/);
      if (m) versions.set(m[1], m[2]);
    }
  }
  for (const [, name, version] of body.matchAll(/^require\s+(\S+)\s+(v\S+)/gm)) {
    versions.set(name, version);
  }
  // Every require line, direct or indirect, can parent a tool directive.
  const allVersions = new Map(versions);
  for (const [, block] of text.matchAll(/^require\s*\(([\s\S]*?)^\)/gm)) {
    for (const line of block.split("\n")) {
      const m = line.trim().match(/^(\S+)\s+(v\S+)/);
      if (m && !allVersions.has(m[1])) allVersions.set(m[1], m[2]);
    }
  }

  const checks = [];
  for (const [name, version] of versions) {
    checks.push({ ledger: name, module: name, version });
  }
  const tools = [];
  for (const [, block] of text.matchAll(/^tool\s*\(([\s\S]*?)^\)/gm)) {
    for (const line of block.split("\n")) {
      if (line.trim()) tools.push(line.trim());
    }
  }
  for (const [, name] of text.matchAll(/^tool\s+([^\s(]\S*)\s*$/gm)) tools.push(name);
  for (const tool of tools) {
    let parent = null;
    for (const mod of allVersions.keys()) {
      if (tool === mod || tool.startsWith(mod + "/")) {
        if (parent === null || mod.length > parent.length) parent = mod;
      }
    }
    if (parent === null) {
      throw new Error(`new-deps: tool ${tool} has no parent in the require list`);
    }
    checks.push({ ledger: tool, module: parent, version: allVersions.get(parent) });
  }
  return checks;
}

function ledgerRows() {
  const rows = new Set();
  for (const line of readFileSync("docs/dependencies.md", "utf8").split("\n")) {
    const m = line.match(/^\|\s*`([^`]+)`\s*\|/);
    if (m) rows.add(m[1].toLowerCase());
  }
  return rows;
}

async function infoTime(module, version) {
  const res = await fetch(`https://proxy.golang.org/${module}/@v/${version}.info`, {
    signal: AbortSignal.timeout(15000),
  });
  if (!res.ok) return null;
  const info = await res.json();
  return info.Time ? new Date(info.Time) : null;
}

const checks = directModules();
const ledger = ledgerRows();
const errors = [];
for (const { ledger: name, module, version } of checks) {
  if (!ledger.has(name.toLowerCase())) {
    errors.push(`${name}: no row in docs/dependencies.md`);
  }
  const key = `${module}@${version}`;
  if (EXCEPTIONS.has(key)) continue;
  let at;
  try {
    at = await infoTime(module, version);
  } catch (err) {
    errors.push(`${key}: cannot reach the Go proxy (${err.cause?.code ?? err.message})`);
    continue;
  }
  if (at === null) {
    errors.push(`${key}: not found on the Go proxy`);
    continue;
  }
  const ageDays = (Date.now() - at.getTime()) / MS_PER_DAY;
  if (ageDays < MIN_AGE_DAYS) {
    errors.push(`${key}: ${ageDays.toFixed(1)} days old, needs ${MIN_AGE_DAYS}`);
    continue;
  }
  console.log(`${key}: ok (${ageDays.toFixed(0)} days old)`);
}

if (errors.length > 0) {
  for (const e of errors) console.error(`new-deps: ${e}`);
  process.exit(1);
}
console.log("new-deps: ok");

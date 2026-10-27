# 0001. Record architecture decisions

## Context

Choices such as the router, the database access layer and the layout shape every feature. Without a record, later contributors and agents repeat the discussion or undo the choice by accident.

## Decision

We record each irreversible technical choice as a short ADR in `docs/decisions/`, named `<number>-<slug>.md`, in the Context, Decision, Consequences format. We never edit an accepted ADR. A change of mind is a new ADR that states which one it supersedes.

A new top-level directory, or a new directory under `internal/`, needs an ADR and a `layout.allow` change in `repo-policy.json` in the same pull request.

## Consequences

- The reason behind each structural choice is findable with `rg` and `git log`.
- Adding a directory takes one extra file, which keeps the layout deliberate.
- ADRs accumulate; old ones stay readable but may be superseded.

# Docs

Documents live in the repository, sorted by how long they stay true. The routing rules come from `workflow/doc-routing.md` in the shared docs repository.

| Path | Holds | Lifespan |
| --- | --- | --- |
| **decisions** | One ADR per irreversible choice, numbered `docs/decisions/<number>-<slug>.md`. | Permanent. Never edit an ADR; add a new one that supersedes it. |
| **specs** | Functional specs: business rules and acceptance scenarios for a feature. | Until the feature ships. Then fold the rules into `ARCHITECTURE.md` or tests. |
| **plans** | Task checklists for one pull request. | Commit the plan, then `git rm` it on merge. |
| **archive** | Retired specs and old lesson rows. | Permanent record. |
| [dependencies.md](dependencies.md) | Why each dependency exists. | Updated with every dependency change. |

Create the **specs**, **plans** and **archive** folders under `docs` when you first need them. `ARCHITECTURE.md` at the repository root is the living architecture document.

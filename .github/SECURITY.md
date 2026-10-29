# Security policy

## Report a vulnerability

Use GitHub private vulnerability reporting: <https://github.com/tanzhongyan-org/template-go/security/advisories/new>.

<!-- TEMPLATE-TODO(security): confirm private vulnerability reporting is enabled and add a contact address. -->

Do not open a public issue for a vulnerability.

## Supported versions

Only `main` is supported.

## What to expect

| Step | Target |
| --- | --- |
| Acknowledge your report | 3 business days |
| Triage and severity | 10 business days |

## How this repository checks itself

| Layer | Tool | When |
| --- | --- | --- |
| Static analysis | golangci-lint with `gosec`, `bodyclose`, `errorlint`, `noctx`, `sqlclosecheck` | Every pull request (`make lint`) |
| Code scanning | CodeQL with `security-extended` queries, Go and GitHub Actions | Push, pull request, weekly |
| Secrets | gitleaks over full history; staged scan in the pre-commit hook | Every pull request |
| Dependencies | `govulncheck` (`make audit`) and dependency review at high severity | Every pull request |
| Licences | `make licenses`: every dependency licence inside the allowlist, `NOTICE` regenerated | Every pull request |
| Security target | `make security`: gitleaks, govulncheck and the new-dependency check | Every pull request |
| Workflows | zizmor | Every pull request |
| Dynamic scan | OWASP ZAP baseline against the running server | Weekly |

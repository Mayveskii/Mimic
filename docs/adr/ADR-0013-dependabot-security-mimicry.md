# ADR-0013: Dependabot Security Mimicry

## Status

Accepted

## Context

Mimic's primary success metric is **MERGE COUNT FEEL**: patches that are merged into real repositories and feel like human junior-dev work.

Security and dependency patches are high-value merge candidates. Industry systems define the canonical behavior and shared vocabulary for this domain:

- **Dependabot** — automated dependency updates and vulnerability remediation.
- **GitHub Advanced Security** — alert taxonomy for code scanning, secret scanning, and dependency review.
- **GitHub Docs** — canonical platform behavior for workflows, permissions, and security hardening.
- **OWASP** — application security taxonomy (Top 10, ASVS, Cheat Sheets) and remediation guidance.
- **NVD / CVE** — canonical vulnerability identifiers, CVSS severity, CWE classification, and CPE metadata.
- **OSV** — unified, ecosystem-agnostic package vulnerability lookup.
- **GitHub Advisory Database** — CVE to package mapping with patched versions; data source behind Dependabot alerts.
- **OpenSSF Scorecard** — security posture and maintainability hygiene checks.

Mimic must learn from these systems to move from "patch on request" to "autonomous maintainer."

## Decision

Adopt Dependabot, GitHub Advanced Security, GitHub Docs, OWASP, NVD/CVE, OSV, GitHub Advisory Database, and OpenSSF Scorecard as external behavior sources and distillation targets.

- Create `mimicrya/external-behavior-sources.yaml` to track non-Mayveskii behavior sources.
- Add `github/*`, `dependabot/*`, `OWASP/*`, `google/osv.dev`, and `ossf/scorecard` repositories to `mimicrya/repos-manifest.yaml` under a new `security:` section.
- Document the target in `project_context_main/PHILOSOPHY/SECURITY_AUTONOMY.md` and `project_context_main/TODO/GAP_v16Q.md`.

## Consequences

### Positive

- Clear target behavior for Mimic's autonomy stage: Dependabot++ (alert → patch → proof → PR).
- New distillation targets expand mesh coverage into security and dependency management.
- GitHub Docs provides canonical patterns for platform-native automation.
- OWASP/CVE/OSV provide shared vocabulary for classification, severity, and remediation.
- Scorecard defines proactive hygiene targets beyond reactive alerts.

### Negative / Risks

- `dependabot-core` is Ruby, `cli` is Go, `smoke-tests` is Shell — distillation must handle multi-language ecosystems.
- GitHub Docs, OWASP ASVS, and Cheat Sheets are Markdown documentation — distillation pipeline needs text-native slot extraction.
- NVD/OSV/Advisory data is structured JSON, not code — requires dedicated CVE/advisory parsing tools.
- Autonomous security patches require stricter guardrails than user-requested patches.

## Related

- `mimicrya/external-behavior-sources.yaml`
- `mimicrya/repos-manifest.yaml`
- `project_context_main/PHILOSOPHY/SECURITY_AUTONOMY.md`
- `project_context_main/TODO/GAP_v16Q.md`
- ADR-0012: GonkaGate Real Patch Machine

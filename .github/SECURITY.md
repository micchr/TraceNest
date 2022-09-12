# Security Policy

## Supported versions

tracenest is a local analysis tool with no network access, but we still
track security-relevant reports for the current release line.

| Version | Supported          |
|---------|--------------------|
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a vulnerability

Please report suspected vulnerabilities privately using GitHub's
**Report a vulnerability** button on the repository's Security tab
(Security advisories). Do not open a public issue for security reports.

Include:

- the affected version and which CLI you were running,
- a minimal trace file or command that demonstrates the problem,
- the impact you believe it has.

You can expect an acknowledgement within a few days. Because tracenest
parses untrusted trace files without executing them, the most relevant
classes of report are parser crashes, unbounded memory use on crafted
input, and HTML output escaping issues in the report renderer.

## Scope

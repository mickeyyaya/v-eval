# Security Policy

## Scope

v-eval is a research and design prototype. It currently contains documentation, a prompt-workflow skill, a contract template, and a worked example. There is no executable runner yet. Once the Go core exists, this policy covers the released binaries, the plugin manifests, and the skill text.

## Reporting a vulnerability

Please do not open a public issue for a security problem. Use GitHub's private vulnerability reporting on this repository ("Report a vulnerability" under the Security tab). You will receive an acknowledgement within seven days.

Relevant classes of problem for an evaluator, in addition to ordinary software vulnerabilities:

- Instructions embedded in an evaluated artifact that could redirect the evaluator or alter its report.
- Ways for candidate code or supplied logs to be treated as verified evidence without the checks the design requires.
- Paths by which learned material (precedents, rules) could modify the trusted contract or read the locked anchor set.

## Supported versions

Until a first tagged release exists, only the default branch is supported.

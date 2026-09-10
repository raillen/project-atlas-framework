# Security Policy

Prumo generates instructions/configuration for AI-assisted development. Treat agent permissions, context retrieval and orchestrator automation as code with real operational impact.

- Keep arbitrary shell/filesystem/network capabilities disabled unless explicitly required and authorized.
- Separate read-only investigation from mutation where possible.
- Do not expose secrets through Working Context Capsules, generated site output, traces or Project Intelligence.
- Runtime context/index databases are derived state and should be gitignored.
- Do not store provider secrets in `prumo.json`, profiles or model-policy files.
- Deep/unbounded recursive execution is not a default capability.
- Generated documentation dashboards must respect public/internal visibility boundaries.

Report security issues privately to the repository owner rather than opening a public exploit report.

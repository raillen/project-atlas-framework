# Project Atlas — Traycer Adapter

Traycer is an execution/orchestration implementation, not the source of truth. Repository Goals remain canonical.

- Map Project Atlas Phase Goals to Traycer Phases/Tasks.
- Prefer worktrees for independent Goals.
- Use agent-to-agent delegation only inside Goal scope.
- Configure planning, implementation and verification with role-appropriate models.
- Require cross-provider review for high-risk changes when the configured roster permits it.
- Cap same-model retries at two before a quality fallback.
- Do not let Smart/YOLO execution silently weaken locked acceptance criteria; emit a Goal Amendment proposal instead.
- Feed CI, tests, visual evidence and review findings back into Goal evidence.

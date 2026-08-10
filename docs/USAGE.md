# Usage

## 1. Discuss the project normally

Use any chat/model to explore product, architecture and constraints. When decisions stabilize, invoke `ProjectAtlas: finalize` with the framework repository available.

## 2. Create a project profile

Record project type, stack, features, risks, orchestrator/autonomy and an explicit preferred model roster. See the Brasa example.

## 3. Initialize

```bash
atlas init ./project --profile ./project-profile.yaml --non-interactive
```

## 4. Compile adapters

```bash
atlas compile --target codex --path ./project
atlas compile --target traycer --path ./project
```

## 5. Define Goals

Create phase Goals, replace placeholder acceptance criteria, move to PLANNED, review, then LOCK. Agents execute only against locked scope.

## 6. Validate continuously

```bash
atlas validate ./project --schemas ./schemas
```

## 7. Recover anywhere

Give a new model the project repository and ask it to read `ENTRYPOINT.md`. A snapshot can be produced with `atlas snapshot` when repository browsing is inconvenient.

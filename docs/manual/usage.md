# Manual de Uso

## Fluxo recomendado

```text
profile → init → validate → doctor → Goal → context → compile → evidence → review
```

O Atlas organiza estado canônico no repositório. O harness executa ações; o Atlas mantém protocolo, políticas, Goals, evidências e adapters.

## 1. Criar um projeto

```bash
atlas init ./my-project \
  --profile examples/brasa/project-profile.json \
  --non-interactive
```

O profile deve declarar `ai.preferred_models`. O resultado contém `atlas.json`, `.ai/`, `docs/ATLAS.md`, `PROJECT_STATE.md` e histórico derivado.

## 2. Validar e diagnosticar

```bash
atlas validate ./my-project
atlas doctor ./my-project
atlas --json doctor ./my-project
atlas framework-check
```

`validate` verifica estrutura e schemas. `doctor` verifica também versionamento, locks, dependencies, DAGs, workforce, policies, gates e evidence.

## 3. Criar e bloquear Goals

```bash
atlas goal new P00-G01 "Foundation" \
  --phase P00 \
  --objective "Establish a tested project foundation." \
  --path ./my-project

atlas goal list --path ./my-project
atlas goal state P00-G01 PLANNED --path ./my-project
atlas goal state P00-G01 LOCKED --path ./my-project
```

Estados válidos:

```text
DRAFT → PLANNED → LOCKED → EXECUTING → VERIFYING → REVIEWING → DONE
```

Estados podem ir para `BLOCKED` conforme as transições do protocolo. `DONE` exige evidence. Goal bloqueado deve ser alterado com amendment:

```bash
atlas goal amend P00-G01 \
  --file amendment.json \
  --path ./my-project
```

## 4. Planejar contexto

```bash
atlas context plan "debug authentication regression" \
  --path ./my-project \
  --json
```

O planner escolhe uma estratégia e budget conforme o risco sem carregar o repositório inteiro.

## 5. Resolver workforce

```bash
atlas resolve examples/brasa/project-profile.json --json
atlas explain workforce examples/brasa/project-profile.json --json
atlas explain agent architect --json
atlas explain skill clean-code --json
atlas explain recipe web-feature --json
```

A resolução é determinística para os mesmos profile, catálogo e recursos.

## 6. Compilar adapters

```bash
atlas compile --target generic --path ./my-project
atlas compile --target codex --path ./my-project
atlas compile --target claude-code --path ./my-project
atlas compile --target traycer --path ./my-project
```

Targets disponíveis:

```text
generic, chatgpt, claude, kimi, codex, claude-code, traycer
```

Saídas são derivadas. Edite o catálogo/workforce canônico, não o adapter gerado.

## 7. Reports e inteligência

```bash
atlas report add conformance/fixtures/task-report.json --path ./my-project
atlas report summary --path ./my-project --json
```

Reports alimentam `.atlas/history/project-intelligence.json`, que é estado derivado e reconstruível.

## 8. Snapshot e migração

```bash
atlas snapshot ./my-project --output ./my-project-backup.zip
atlas migrate ./my-project --dry-run --json
atlas migrate ./my-project
```

Sempre execute `--dry-run` antes de migrações. A migração cria snapshot prévio quando altera o projeto.

## 9. Saída JSON para automações

```bash
atlas --json version
atlas --json validate ./my-project
atlas --json framework-check
atlas --json compile --target generic --path ./my-project
```

Contrato comum:

```json
{
  "protocol_version": "1",
  "ok": true,
  "data": {},
  "diagnostics": [],
  "warnings": []
}
```

Códigos principais:

| Código | Uso |
|--------|-----|
| `0` | sucesso |
| `1` | validação ou gate falhou |
| `2` | uso/argumentos inválidos |
| `3` | configuração inválida |
| `4` | capability indisponível |
| `5` | erro interno |
| `6` | projeto não encontrado |

## 10. Instalação e estado global

```bash
atlas --home ./atlas-home setup
atlas --home ./atlas-home install connector opencode
atlas --home ./atlas-home uninstall --connectors --purge-cache
```

Use `--home` em CI, testes, devboxes e cenários que não devem tocar `~/.atlas`.

## 11. Python oracle

A implementação Python v0.3 continua disponível somente para conformance e manutenção transitória:

```bash
python -m pip install -e '.[dev]'
pytest
```

Não implemente features novas do v0.4 no runtime Python.

## Command reference

### Core

```text
atlas version
atlas status --path <path>
atlas setup
atlas install connector <id>
atlas uninstall [--connectors] [--purge-cache] [--purge-global-config]
atlas init <path> --profile <profile> --non-interactive
atlas validate [path]
atlas doctor [path] [--json]
atlas framework-check
```

### Protocol

```text
atlas goal new <id> <title> --phase <phase> [--objective <text>] [--path <path>]
atlas goal state <id> <state> [--reason <text>] [--path <path>]
atlas goal amend <id> [--file <path>] [--reason <text>] [--approved-by <actor>] [--path <path>]
atlas goal list [--path <path>]
atlas context plan <task> [--path <path>] [--json]
atlas report add <file> [--path <path>]
atlas report summary [--path <path>]
atlas migrate [path] [--dry-run] [--json]
atlas snapshot [path] [--output <path>]
atlas docs delta propose --goal <goal> [--path <project>] [changed ...] [--json]
atlas docs delta list [--path <project>] [--json]
atlas docs delta show --id <delta> [--path <project>] [--json]
atlas docs delta transition --id <delta> --state <state> [--evidence <id>]... [--path <project>] [--json]
```

### Resolution, explanation, and compiler

```text
atlas resolve <profile> [--json]
atlas explain workforce <profile> [--json]
atlas explain agent <id> [--json]
atlas explain skill <id> [--json]
atlas explain recipe <id> [--json]
atlas explain context <task-id> [--path <path>] [--json]
atlas explain model <role> [--path <path>] [--json]
atlas explain execution <profile> [--path <path>] [--json]
atlas compile --target <target> [--path <path>] [--json]
```

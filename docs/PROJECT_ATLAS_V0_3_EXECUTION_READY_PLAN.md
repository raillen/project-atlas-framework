# Project Atlas Framework v0.3 — Execution-Ready Protocol
## Especificação completa de evolução antes do reboot do Atlas Flow

> **Status:** proposta de implementação
>
> **Objetivo:** transformar o Project Atlas Framework v0.2 em um protocolo agentic realmente executável, portátil, verificável e suficientemente preciso para servir como contrato canônico do novo Atlas Flow em Rust.
>
> **Repositório alvo:** `raillen/project-atlas-framework`
>
> **Projeto de referência posterior:** `raillen/atlas-flow`
>
> **Princípio central:** o Project Atlas Framework define **o protocolo, os contratos e as regras**. O Atlas Flow implementará **o runtime, scheduler, event bus, processos, ACP/MCP, Git worktrees, SQLite operacional, API, UI e desktop host**.
>
> O Project Atlas **não deve virar um orquestrador**. O Atlas Flow **não deve inventar unilateralmente o protocolo**.

---

# 1. Contexto e motivação

O Project Atlas Framework v0.2 já possui uma base conceitual forte:

- Git como memória durável;
- Markdown e JSON como formatos canônicos;
- Goals como unidade de resultado;
- Agents, Skills e Recipes como força de trabalho abstrata;
- Lean Progressive Context / Progressive Context Architecture;
- Project Intelligence;
- adapters substituíveis;
- model policy;
- orquestradores e providers substituíveis;
- evidência acima de afirmação;
- documentação multi-audiência;
- resolução determinística da menor workforce útil.

Porém, várias partes ainda existem mais como **convenção documental** do que como **contrato de máquina executável**.

Exemplos importantes:

- Skills são essencialmente registros de catálogo transformados em um `SKILL.md` muito curto;
- Recipes são pouco mais que uma sequência textual;
- Agents possuem poucas propriedades operacionais;
- Task DAG é mencionado conceitualmente, mas não possui contrato completo;
- gates e evidence ainda são genéricos demais;
- contexto progressivo não possui uma IR formal suficientemente rica;
- permissões de agentes ainda são muito simplificadas;
- model selection e execution backend ainda estão misturados conceitualmente;
- não existe um Run/Attempt/Event Protocol completo;
- não há Conformance Suite abrangente para permitir outra implementação do protocolo;
- documentação de uso, authoring e implementação ainda é insuficiente;
- workflows GitHub e UI/UX ainda não estão maduros como workforce real.

O novo Atlas Flow será refeito do zero em Rust. Antes disso, o framework precisa chegar a uma versão **Execution-Ready**, isto é:

> Um runtime independente deve conseguir implementar Project Atlas apenas lendo schemas, contratos, documentação e fixtures, sem precisar inferir comportamento de prose vaga ou do código Python atual.

---

# 2. Meta da versão

A versão alvo deve ser tratada como:

## Project Atlas Framework v0.3 — Execution-Ready Protocol

Ela deve fechar os contratos necessários para:

```text
Project Atlas Framework
        │
        │ define
        ▼
Goals
Plans
Tasks
DAGs
Agents
Skills
Recipes
Context
Evidence
Gates
Policies
Runs
Attempts
Events
Model routing
Execution routing
Project Intelligence
        │
        │ implementado por
        ▼
Atlas Flow
```

A v0.3 não precisa resolver todo problema futuro.

O objetivo é **fechar o mínimo completo, coerente e testável que permita o Atlas Flow começar sem inventar conceitos ausentes**.

---

# 3. Princípios não negociáveis

Toda implementação desta especificação deve preservar estes princípios.

## 3.1 Git é a memória durável

Conhecimento estável, decisões, Goals, configurações, policies e evidência durável continuam versionados no repositório.

SQLite permanece exclusivamente para:

- índices derivados;
- caches;
- estado operacional;
- contexto temporário;
- runs;
- sessions;
- dados reconstruíveis.

## 3.2 Canonical before generated

Conteúdo gerado nunca substitui fonte canônica.

Adapters, compilados para Codex/Claude/etc., índices, summaries e caches são descartáveis e regeneráveis.

## 3.3 Portabilidade

Nenhuma decisão da v0.3 pode obrigar uso de:

- Codex;
- Claude Code;
- Gemini;
- OpenCode;
- Atlas Flow;
- LiteLLM;
- Bifrost;
- qualquer provider específico.

## 3.4 Least workforce

Selecionar somente Agents, Skills e Recipes necessários.

## 3.5 Least context

Carregar apenas o menor contexto suficiente e expandir progressivamente.

## 3.6 Pointer over payload

Preferir:

- IDs;
- paths;
- symbols;
- ranges;
- evidence pointers;
- references;

em vez de repetir blobs grandes de texto.

## 3.7 Evidence over assertion

Nenhum Agent, LLM ou runtime pode declarar um Goal concluído apenas por confiança textual.

## 3.8 Bounded execution

Retries, fallbacks, context expansion e delegation devem ser limitados.

## 3.9 Protocol first, runtime second

O framework define semântica.

O runtime executa.

---

# 4. Escopo P0 obrigatório antes do Atlas Flow

A v0.3 deve ser considerada pronta somente quando estes blocos estiverem implementados:

1. Skill Package v2.
2. Agent Contract v2.
3. Recipe Contract v2.
4. Clean Engineering workforce.
5. Security workforce.
6. Plan + Task + DAG contracts.
7. Evidence + Gate contracts.
8. Permission + Trust + Approval policies.
9. Context IR / Context Pack / Budget contracts.
10. Model Policy v2.
11. Execution Policy.
12. Run + Attempt + Event Protocol.
13. Versioning + compatibility + migration system.
14. Conformance Suite.
15. `atlas doctor`.
16. `atlas explain`.
17. GitHub issue governance workforce.
18. UI/UX design workforce.
19. Documentation completa de uso e extensão.
20. Projeto/fixture de conformance end-to-end.

Itens como marketplace remoto, cloud sync, registry remoto e routing inteligente avançado ficam fora desta versão.

---

# 5. Workforce v2

## 5.1 Problema atual

As Skills atuais funcionam primariamente como labels roteáveis com uma pequena instrução.

Exemplo conceitual atual:

```json
{
  "id": "code-quality",
  "instructions": "Review complexity, duplication..."
}
```

Depois o compiler gera um `SKILL.md` curto.

Isso não é suficiente para uma capacidade operacional robusta.

## 5.2 Novo conceito

Uma Skill passa a ser um **Skill Package versionado**.

Estrutura padrão:

```text
skills/
└── <skill-id>/
    ├── manifest.json
    ├── SKILL.md
    ├── references/
    ├── templates/
    ├── checks/
    ├── scripts/
    └── examples/
```

Somente:

```text
manifest.json
SKILL.md
```

são obrigatórios.

Os demais diretórios são opcionais.

## 5.3 Filosofia de contexto

Robustez não deve aumentar custo de contexto por padrão.

Primeiro carregar:

```text
manifest.json
SKILL.md
```

Depois carregar somente recursos necessários:

```text
references/api.md
checks/authentication.md
templates/security-report.md
```

A Skill Package deve ser totalmente compatível com LPC/PCA.

---

# 6. Skill Package v2 — manifest

Criar um schema:

```text
schemas/skill.schema.json
```

Estrutura mínima recomendada:

```json
{
  "id": "secure-coding",
  "name": "Secure Coding",
  "version": 1,
  "schema_version": 2,
  "purpose": "Prevent vulnerabilities during implementation.",
  "risk_level": "high",
  "modes": ["implementation", "review"],
  "inputs": [],
  "outputs": [],
  "requires": [],
  "conflicts": [],
  "tools": [],
  "references": [],
  "templates": [],
  "checks": [],
  "scripts": [],
  "required_evidence": [],
  "stop_conditions": [],
  "provenance": {
    "origin": "framework",
    "license": "..."
  }
}
```

Campos desejáveis:

### Identidade

- `id`
- `name`
- `version`
- `schema_version`
- `description`
- `purpose`

### Seleção

- `select.project_types`
- `select.stack_any`
- `select.features_any`
- `select.risk_any`

### Dependências

- `requires`
- `requires_any`
- `conflicts`

### Modos

Exemplos:

```text
implementation
review
audit
research
design
testing
documentation
```

### Capabilities necessárias

A Skill deve poder declarar o que pode precisar:

```json
{
  "capabilities": [
    "filesystem.read",
    "process.spawn"
  ]
}
```

Não significa que a Skill automaticamente recebe essas permissões.

Significa que sua execução precisa passar pela policy do projeto.

### Inputs

Contratos mínimos esperados.

### Outputs

Artifacts ou findings esperados.

### Evidence

O que prova que a Skill foi executada adequadamente.

### Stop conditions

Quando encerrar sem continuar consumindo contexto ou ações.

---

# 7. SKILL.md v2

O `SKILL.md` deve ser operacional e conciso.

Estrutura recomendada:

```markdown
# <Skill Name>

## Purpose

## Use when

## Do not use when

## Required context

## Procedure

## Decision rules

## Evidence required

## Output contract

## Stop conditions

## Escalation rules
```

O arquivo não deve se tornar um tratado enciclopédico.

Conteúdo longo ou especializado fica em `references/`.

---

# 8. Provenance e supply chain das Skills

Toda Skill deve poder registrar:

```text
origin
source
version
license
checksum
modified
```

Escopos:

```text
framework
organization
project
external
```

Objetivos:

- segurança;
- atualização;
- licenciamento;
- reproducibilidade;
- troubleshooting;
- futura distribuição de packages.

Project-local overrides devem ser explícitos.

---

# 9. Escopos da Workforce

O framework deve suportar:

```text
framework
organization
project
```

Precedência:

```text
project override
    ↓
organization
    ↓
framework
```

Mas políticas superiores de segurança não podem ser silenciosamente enfraquecidas.

Uma extensão local pode especializar comportamento.

Ela não deve poder desativar controles críticos sem uma policy explícita e auditável.

---

# 10. Agent Contract v2

## 10.1 Problema

Agents atuais já possuem:

- purpose;
- instructions;
- permissions;
- required skills.

É uma boa fundação, mas ainda insuficiente para execução formal.

## 10.2 Novo schema

Criar:

```text
schemas/agent.schema.json
```

Adicionar:

```text
id
name
version
purpose

inputs
outputs

required_skills
optional_skills

allowed_capabilities
required_capabilities

required_evidence

may_delegate
max_delegation_depth

handoff_to

risk_level

review_requirement

permissions

stop_conditions
```

## 10.3 Exemplo: Implementer

```text
Input
- active Goal
- Task
- approved Plan
- context pack

Output
- implementation
- tests
- evidence

Must not
- alter locked Goal
- weaken acceptance
- self-approve
- merge own work

Handoff
- tester
- reviewer
```

---

# 11. Agents principais v0.3

Manter poucos Agents.

Recomendação:

```text
Explorer
Architect
Implementer
Debugger
Tester
Reviewer
Documentation Maintainer
Release Verifier

Security Architect
Security Reviewer

Quality Reviewer

Design Researcher
UX Architect
Design System Engineer
Accessibility Reviewer

Issue Author
Issue Triager
```

Não transformar toda especialidade em Agent.

Especialidades pequenas devem ser Skills.

---

# 12. Recipe Contract v2

## 12.1 Problema atual

Recipes são majoritariamente sequências textuais.

Exemplo:

```text
Goal -> Implementer -> Tester -> Reviewer
```

Isso não é suficiente para um runtime executar.

## 12.2 Novo formato

Criar:

```text
schemas/recipe.schema.json
```

e canonical packages:

```text
recipes/
└── <recipe-id>/
    ├── recipe.json
    └── RECIPE.md
```

Campos:

```text
id
version
purpose
preconditions
inputs
steps
agents
skills
gates
artifacts
fallbacks
stop_conditions
```

Cada step pode declarar:

```text
id
depends_on
agent_role
skills
input
output
required_evidence
gates
```

## 12.3 Exemplo conceitual

```text
feature-standard

1. explore
2. design/plan
3. implementation
4. tests
5. independent review
6. documentation delta
7. intelligence update
8. context GC
```

---

# 13. Clean Engineering Suite

Criar uma família robusta de Skills.

## 13.1 clean-code

Cobrir:

- naming;
- cohesion;
- coupling;
- module boundaries;
- function responsibility;
- duplication;
- dead code;
- comments;
- side effects;
- mutation;
- API ergonomics;
- testability;
- error clarity;
- maintainability.

Regra:

> Não usar métricas arbitrárias como dogma. Findings devem representar risco concreto de manutenção, clareza ou defeito.

## 13.2 architecture-quality

Cobrir:

- dependency direction;
- cycles;
- boundary leakage;
- transport/domain coupling;
- god modules;
- shared mutable state;
- hidden global state;
- unstable interfaces;
- inappropriate abstraction;
- dependency inversion quando realmente útil.

## 13.3 refactoring

Fluxo:

```text
detect smell
→ establish behavior
→ regression test
→ minimal transformation
→ verify
→ repeat
```

## 13.4 error-handling

Cobrir:

- taxonomy;
- recoverable/fatal;
- retries;
- contextual errors;
- user-facing vs internal errors;
- logging;
- propagation;
- panic/exception boundaries.

## 13.5 dependency-management

Cobrir:

- minimal dependency surface;
- provenance;
- pinning;
- advisories;
- maintenance;
- transitive risk;
- licensing;
- unnecessary libraries.

## 13.6 testing-quality

Cobrir:

- testing pyramid appropriate to project;
- deterministic tests;
- regression tests;
- behavior-oriented assertions;
- test isolation;
- flaky test detection;
- representative fixtures.

## 13.7 concurrency-quality

Especialmente importante para Atlas Flow/Rust.

Cobrir:

- races;
- deadlocks;
- cancellation;
- backpressure;
- leaked tasks;
- unbounded channels;
- lock scope;
- resource cleanup;
- timeouts;
- ownership;
- shutdown.

## 13.8 observability

Cobrir:

- structured logs;
- tracing;
- correlation IDs;
- metrics;
- event provenance;
- actionable errors;
- secrets redaction.

---

# 14. Security Suite

Segurança não deve ser uma única Skill.

Criar packages:

```text
threat-modeling
secure-coding
security-review
auth-security
api-security
web-security
desktop-security
filesystem-security
process-execution-security
supply-chain-security
secrets-security
network-security
plugin-security
mcp-security
acp-security
untrusted-project-security
```

## 14.1 Security Architect

Atua antes da implementação.

Responsabilidades:

- trust boundaries;
- abuse cases;
- attack surface;
- privilege model;
- sandbox design;
- security invariants;
- approval requirements.

## 14.2 Security Reviewer

Atua depois.

Responsabilidades:

- audit;
- attack scenarios;
- validation;
- automated/manual checks;
- findings;
- severity;
- remediation verification.

---

# 15. Permission Policy v2

## 15.1 Problema atual

Permissões do tipo:

```text
write_code
modify_docs
execute_tests
```

são insuficientes.

## 15.2 Capability model

Criar:

```text
schemas/permission-policy.schema.json
```

Capabilities mínimas:

```text
filesystem.read
filesystem.write
filesystem.delete

process.spawn
process.shell
process.kill

network.http
network.listen

git.read
git.branch
git.worktree
git.commit
git.merge
git.rebase
git.push

secrets.read

mcp.invoke

github.read
github.issue
github.pr
github.merge

package.install

artifact.publish
```

## 15.3 Scope

Exemplo:

```json
{
  "capability": "filesystem.write",
  "scope": "project-worktree"
}
```

Ou:

```json
{
  "capability": "network.http",
  "hosts": ["github.com", "crates.io"]
}
```

---

# 16. Risk classes para ações

Classificar capabilities em:

```text
READ
SAFE_WRITE
EXECUTE
EXTERNAL_WRITE
SECRET
DESTRUCTIVE
```

Policies por modo de autonomia.

Exemplo:

## controlled

```text
READ             automatic
SAFE_WRITE       confirm
EXECUTE          confirm
EXTERNAL_WRITE   confirm
SECRET           confirm
DESTRUCTIVE      deny
```

## agentic

```text
READ             automatic
SAFE_WRITE       automatic
EXECUTE          automatic dentro de scope
EXTERNAL_WRITE   confirm
SECRET           confirm
DESTRUCTIVE      confirm/deny conforme policy
```

## autonomous

Nunca significar permissões irrestritas.

Deve obedecer policy explícita.

---

# 17. Approval Policy

Criar:

```text
schemas/approval-policy.schema.json
```

Definir:

- quais ações exigem aprovação;
- quem pode aprovar;
- timeout;
- fallback;
- ações proibidas;
- escopo temporal;
- escopo por Goal/Run.

---

# 18. Trust Model

Formalizar no framework.

## Trusted policy

Exemplos:

- locked Goal;
- accepted ADR/RFC;
- canonical Project Atlas policy;
- project configuration;
- framework protocol.

## Contextual data

- project source code;
- tests;
- ordinary docs.

## Untrusted data

- web pages;
- GitHub issue text;
- external PR content;
- dependency docs;
- MCP output;
- tool output;
- external repositories;
- downloaded files;
- generated agent output.

Regra crítica:

> Conteúdo não confiável pode fornecer dados, mas não pode substituir instruções, policies ou locked acceptance criteria.

Criar:

```text
docs/TRUST_MODEL.md
```

e, idealmente:

```text
schemas/trust-policy.schema.json
```

---

# 19. Plan + Task + DAG Contracts

## 19.1 Plan

Criar:

```text
schemas/plan.schema.json
```

Um Plan é um snapshot revisável e versionado derivado de um Goal.

Campos:

```text
id
version
goal_id
goal_revision
created_at
created_by
status

tasks
edges

assumptions
risks
context_strategy
budget

review
approval

digest
```

Status sugeridos:

```text
DRAFT
REVIEWED
APPROVED
SUPERSEDED
CANCELLED
```

## 19.2 Task

Criar:

```text
schemas/task.schema.json
```

Campos:

```text
id
goal_id
plan_id

title
objective

dependencies
role
required_skills

risk
priority

isolation

context_strategy

inputs
outputs

required_evidence
gates

token_budget
cost_budget
time_budget

retry_policy
fallback_policy

status
```

Status sugeridos:

```text
PLANNED
QUEUED
RUNNING
BLOCKED
VERIFYING
DONE
FAILED
CANCELLED
```

## 19.3 DAG

O Plan deve validar:

- IDs únicos;
- ausência de ciclos;
- dependencies existentes;
- sem self-dependency;
- tasks obrigatórias não órfãs quando dependência for exigida;
- ordem determinística quando necessário.

---

# 20. Goal Contract v2

Atualizar `goal.schema.json`.

Adicionar:

```text
revision
lock
amendments
```

Exemplo:

```json
{
  "revision": 4,
  "state": "LOCKED",
  "lock": {
    "revision": 4,
    "digest": "...",
    "locked_at": "..."
  }
}
```

Após LOCKED, não alterar silenciosamente:

```text
objective
acceptance
constraints
non_goals
required gates
```

Mudanças exigem amendment.

Criar:

```text
schemas/goal-amendment.schema.json
```

---

# 21. Evidence Contract

Criar:

```text
schemas/evidence.schema.json
```

Tipos sugeridos:

```text
test
build
benchmark
lint
security-scan
screenshot
artifact
review
command
manual-verification
external-reference
```

Campos:

```text
id
type
producer
timestamp
status

goal_id
task_id
run_id
attempt_id

command
artifact
path
hash

environment

provenance
confidence

related_acceptance
related_gate

metadata
```

Evidence não deve carregar arbitrariamente blobs enormes quando um pointer/hash for suficiente.

---

# 22. Gate Contract

Criar:

```text
schemas/gate.schema.json
```

Campos:

```text
id
name
type
required
predicate
required_evidence
executor
failure_policy
waiver_policy
status
```

Tipos:

```text
test
build
security
performance
accessibility
review
manual
custom
```

Gate pode ser condicional:

```text
required_when:
  risk_any: ["security"]
```

ou:

```text
required_when:
  features_any: ["ui"]
```

---

# 23. Gate Waivers

Waivers devem ser raros e auditáveis.

Criar:

```text
schemas/gate-waiver.schema.json
```

Campos:

```text
gate_id
reason
approved_by
approved_at
expires_at
scope
risk_acceptance
```

Locked acceptance criteria não podem ser implicitamente anulados por waiver.

---

# 24. Context Protocol

Formalizar LPC/PCA.

Criar:

```text
schemas/context-request.schema.json
schemas/context-plan.schema.json
schemas/context-pack.schema.json
schemas/context-item.schema.json
schemas/context-budget.schema.json
```

---

# 25. ContextItem

Campos:

```text
id
kind
source
locator
range
symbol

reason
priority

estimated_tokens

provenance
trust

freshness
digest
```

Kinds:

```text
file
section
symbol
test
schema
goal
decision
issue
diff
command-output
external-reference
```

---

# 26. ContextRequest

Representa necessidade.

Exemplo:

```json
{
  "task_id": "T-004",
  "intent": "modify model routing",
  "required_topics": [
    "model policy",
    "router implementation"
  ]
}
```

---

# 27. ContextPlan

Define como resolver o pedido:

```text
direct/local
structural/symbol
known pack
progressive retrieval
bounded delegation
stop
```

---

# 28. ContextPack

Pacote concreto fornecido para execução.

Deve ser:

- referenciável;
- mensurável;
- descartável quando runtime;
- reconstruível quando derivado.

Campos:

```text
id
task_id
items
estimated_tokens
actual_tokens
created_at
reason
```

---

# 29. Context Budget

Campos:

```text
max_input_tokens
max_retrieved_tokens
max_injected_tokens
max_intermediate_output_tokens

max_expansions
max_delegations
max_delegation_depth

deadline
cost_budget
```

Stop reasons:

```text
sufficient_evidence
budget_exhausted
blocked
human_required
no_relevant_context
```

---

# 30. Model Policy v2

Separar seleção de inteligência de execução.

`model-policy.json` responde:

> QUAL model/profile usar para uma função?

Campos:

```text
version
selection_rule
cross_provider_review
context_aware_routing

roster
profiles
roles

quality_policy
cost_policy
fallback_policy
```

Model Profile pode declarar:

```text
id
provider
model
capabilities
context_window
strengths
weaknesses
cost_class
```

Não persistir preços temporais como verdade eterna sem provenance.

---

# 31. Execution Policy

Criar:

```text
schemas/execution-policy.schema.json
```

Responde:

> COMO chegar ao modelo/agente escolhido?

Conceitos:

```text
provider
harness
transport
gateway
credential_reference
execution_backend
```

Exemplo:

```text
Codex
→ harness ACP
→ Codex CLI

Claude
→ ACP
→ Claude Code

local-qwen
→ HTTP
→ Bifrost
→ Ollama
```

Backends iniciais abstratos:

```text
direct
gateway
cli
acp
custom
```

O framework não deve tornar LiteLLM ou Bifrost obrigatórios.

---

# 32. Run Contract

Criar:

```text
schemas/run.schema.json
```

Run é uma execução de Task.

Campos:

```text
id
task_id
goal_id
plan_id

status

started_at
finished_at

execution_profile

attempts

budgets

summary

evidence
events
```

Status:

```text
CREATED
QUEUED
RUNNING
BLOCKED
VERIFYING
COMPLETED
FAILED
CANCELLED
```

---

# 33. Attempt Contract

Criar:

```text
schemas/attempt.schema.json
```

Attempt é uma tentativa dentro do Run.

Campos:

```text
id
run_id
number

agent_role
skill_set
model_profile
execution_backend

started_at
finished_at

status

error
retry_reason

tokens
cost
artifacts
evidence
```

Isso permite:

```text
Attempt 1 -> Codex -> failed
Attempt 2 -> Claude -> success
```

sem perder provenance.

---

# 34. Event Protocol

Criar:

```text
schemas/event.schema.json
docs/EVENT_PROTOCOL.md
```

Envelope:

```text
id
type
timestamp
project_id
goal_id
plan_id
task_id
run_id
attempt_id
producer
payload
schema_version
```

Vocabulário inicial:

```text
project.opened

goal.created
goal.locked
goal.amended
goal.completed

plan.created
plan.reviewed
plan.approved
plan.superseded

task.queued
task.started
task.blocked
task.completed
task.failed

agent.spawned
agent.ready
agent.failed
agent.stopped

tool.started
tool.completed
tool.failed

context.created
context.expanded
context.exhausted

evidence.created

gate.started
gate.passed
gate.failed
gate.waived

review.started
review.completed

run.created
run.started
run.completed
run.failed

attempt.started
attempt.completed
attempt.failed
```

Eventos devem ser append-friendly e versionáveis.

---

# 35. Project Intelligence v2

Não tentar fechar métricas perfeitas.

Apenas garantir compatibilidade com Run/Attempt/Event.

Manter:

```text
tokens
cost
effort
quality
debt
tests
documentation
models
```

Adicionar provenance clara:

```text
observed
estimated
allocated
unknown
```

Atlas Flow servirá de laboratório para evoluir esse módulo.

---

# 36. GitHub Governance Suite

Criar Skills:

```text
github-repository
github-issue-create
github-issue-refine
github-issue-triage
github-pr-create
github-pr-review
github-pr-feedback
github-ci-debug
github-release
```

---

# 37. Issue Author Agent

Fluxo:

```text
request
→ repository context
→ active Goal/roadmap
→ search duplicates
→ classify
→ scope
→ acceptance criteria
→ dependencies
→ draft
→ create
```

Categorias mínimas:

```text
bug
feature
technical-debt
security
performance
documentation
research
architecture
ux
```

---

# 38. Issue Triager Agent

Responsabilidades:

```text
duplicate detection
severity
priority
labels
dependencies
milestone
blocked-by
related issues
Goal linkage
```

---

# 39. Project-specific Issue Policy

Criar:

```text
.ai/github/issue-policy.json
```

ou equivalente canônico definido pela arquitetura v0.3.

Campos:

```text
required_sections
allowed_types
labels
priority_scheme
severity_scheme
milestones
goal_linking
duplicate_search
```

Template padrão:

```text
Summary
Context
Problem
Expected outcome
Scope
Out of scope
Acceptance criteria
Technical notes
Dependencies
Risks
Evidence / references
```

Project overrides podem especializar.

---

# 40. GitHub Issue Recipe

Criar Recipe:

```text
github-issue
```

Steps:

```text
1. Resolve repository/project context
2. Search duplicates
3. Classify
4. Draft scope
5. Define acceptance
6. Resolve dependencies
7. Select labels/milestone
8. Create
9. Return canonical issue reference
```

---

# 41. UI/UX Workforce

Criar Agents:

```text
Design Researcher
UX Architect
Design System Engineer
Accessibility Reviewer
```

Interaction Designer pode ser Agent separado apenas se justificar a complexidade; caso contrário pode ser Skill sob UX Architect.

---

# 42. UI/UX Skills

Criar:

```text
design-research
competitor-analysis
website-forensics
interaction-research
visual-reference-research

information-architecture
user-flows
ux-architecture

wireframing
interaction-design

design-system
design-tokens
component-specification

accessibility
keyboard-accessibility
screen-reader
focus-management
contrast
motion-accessibility
zoom-reflow

visual-regression
visual-qa
```

---

# 43. Design Research

Outputs devem separar:

```text
observed evidence
inference
recommendation
```

Artifacts possíveis:

```text
research/
├── references
├── screenshots
├── observed-patterns
├── interaction-map
├── opportunities
└── source-license-metadata
```

Não copiar:

- assets protegidos;
- branding;
- source code proprietário.

---

# 44. UX Architecture

Cobrir:

```text
information architecture
navigation
task flows
screen hierarchy
progressive disclosure
empty states
loading
errors
recovery
```

---

# 45. Wireframes

Wireframe não deve existir somente como imagem.

Formato estruturado mínimo:

```markdown
# Screen

## Purpose

## Primary action

## Secondary actions

## Regions

## Components

## States

## Navigation

## Keyboard

## Responsive behavior

## Accessibility
```

Imagem/SVG pode ser complementar.

---

# 46. Design System

Cobrir:

```text
color
typography
spacing
radius
elevation
motion
icons
components
variants
states
responsive rules
```

Component states:

```text
default
hover
active
focus
disabled
loading
error
success
destructive
```

Design tokens devem preceder valores one-off.

---

# 47. Accessibility

Cobrir antes e depois da implementação.

Checklist mínimo:

- keyboard complete path;
- visible focus;
- semantic naming;
- contrast;
- screen reader;
- zoom/reflow;
- reduced motion;
- error identification;
- forms;
- state not conveyed only by color.

---

# 48. UI Feature Recipe

Criar:

```text
ui-feature
```

Fluxo:

```text
Product requirement
→ Design research
→ User flow
→ Information architecture
→ Wireframe
→ Accessibility constraints
→ Design system impact
→ Interaction spec
→ Implementation
→ Visual tests
→ Accessibility tests
→ UX review
```

---

# 49. Documentation overhaul

O atual `USAGE.md` deve continuar curto como quick start.

Criar documentação orientada a tarefas.

Estrutura recomendada:

```text
docs/
├── getting-started/
│   ├── installation.md
│   ├── first-project.md
│   └── concepts.md
│
├── user-guide/
│   ├── projects.md
│   ├── goals.md
│   ├── plans-and-tasks.md
│   ├── agents.md
│   ├── skills.md
│   ├── recipes.md
│   ├── context.md
│   ├── evidence-and-gates.md
│   ├── models.md
│   ├── execution.md
│   └── project-intelligence.md
│
├── authoring/
│   ├── writing-skills.md
│   ├── writing-agents.md
│   ├── writing-recipes.md
│   ├── project-bundles.md
│   ├── risk-rules.md
│   └── adapters.md
│
├── integration/
│   ├── codex.md
│   ├── claude-code.md
│   ├── gemini.md
│   ├── opencode.md
│   └── generic.md
│
├── reference/
│   ├── cli.md
│   ├── schemas.md
│   ├── events.md
│   ├── project-layout.md
│   └── workforce-resolution.md
│
└── development/
    ├── architecture.md
    ├── compiler.md
    ├── resolver.md
    ├── migrations.md
    ├── testing.md
    └── contributing.md
```

Atualizar `docs/ATLAS.md` como intent router.

---

# 50. Exemplos reais

Criar exemplos:

```text
examples/
├── rust-cli/
├── react-saas/
├── rust-desktop/
├── game-engine/
└── atlas-flow/
```

Cada um deve demonstrar:

```text
project profile
resolved workforce
Goal
Plan
Task DAG
Recipe
Skill selection
Model Policy
Execution Policy
Evidence
Gates
Task report
Documentation delta
```

O exemplo `atlas-flow` pode ser inicialmente um fixture simplificado antes do reboot real.

---

# 51. `atlas doctor`

Adicionar comando:

```bash
atlas doctor ./project
```

Verificações:

```text
framework/protocol compatibility
schema versions
migration requirements
broken references
Goal consistency
Goal lock integrity
Plan validity
DAG cycles
Task references
workforce packages
Skill provenance
policy conflicts
permission scopes
invalid gates
missing evidence references
model policy
execution policy
compiled adapter freshness
```

Output:

- human-readable;
- `--json`.

Exit code deve refletir falha.

---

# 52. `atlas explain`

Adicionar:

```bash
atlas explain workforce
atlas explain agent <id>
atlas explain skill <id>
atlas explain recipe <id>
atlas explain context <task-id>
atlas explain model <role>
atlas explain execution <profile>
```

Exemplo:

```text
security-reviewer selected because:
- project risk: security
- feature: process-execution
- dependency of recipe: security-review
```

Context explain deve mostrar:

```text
why loaded
source
estimated/actual token cost
trust level
```

---

# 53. Resolver explainability

O resolver deve passar a produzir optional trace estruturado.

Exemplo:

```json
{
  "selected": "security-reviewer",
  "reasons": [
    {
      "type": "risk",
      "value": "security"
    }
  ]
}
```

Esse trace é runtime/derived, não precisa ser canônico.

---

# 54. Compiler v2

O compiler atual renderiza itens simples em Markdown.

Reescrever a parte de workforce para compilar packages reais.

Canonical:

```text
src/project_atlas/resources/workforce/
├── agents/
├── skills/
└── recipes/
```

Ou estrutura equivalente aprovada em ADR.

Codex output:

```text
.codex/
├── agents/
└── skills/
    └── secure-coding/
        ├── SKILL.md
        ├── references/
        ├── templates/
        └── scripts/
```

Claude Code equivalente.

Generated output permanece replaceable.

---

# 55. Catalog evolution

O atual catálogo JSON central não deve continuar duplicando todo conteúdo dos packages.

Preferência:

```text
package files = canonical
catalog/index = derived or compact registry
```

Evitar:

```text
skills.json
+
skills/<id>/manifest.json
```

com verdade duplicada.

Se catálogo central continuar existindo, deve funcionar como index gerado.

---

# 56. Protocol versioning

Separar versões.

Exemplo:

```text
framework_version
protocol_version
project_schema_version
workforce_schema_version
event_schema_version
```

`atlas.json` deve declarar compatibilidade.

Exemplo:

```json
{
  "protocol": {
    "version": 3,
    "compatible": ">=3 <4"
  }
}
```

Definir regras de backward compatibility.

---

# 57. Migration system

Criar:

```bash
atlas migrate ./project
atlas migrate ./project --dry-run
```

Workflow:

```text
detect current version
→ determine migration path
→ dry-run
→ backup/snapshot
→ apply deterministic migration
→ validate
→ emit migration report
```

Nunca editar projeto antigo silenciosamente apenas por abri-lo.

---

# 58. Migration packages

Estrutura possível:

```text
src/project_atlas/migrations/
├── v1_to_v2.py
├── v2_to_v3.py
└── ...
```

Cada migration precisa de fixtures e testes.

---

# 59. Conformance Suite

Criar:

```text
conformance/
├── goals/
├── plans/
├── tasks/
├── workforce/
├── context/
├── evidence/
├── gates/
├── policies/
├── runs/
├── events/
├── migrations/
└── adapters/
```

Cada case:

```text
input/
expected/
expected-error.json
```

ou formato equivalente simples.

Objetivo:

```text
Python reference CLI
```

e futuramente:

```text
Rust Atlas Flow
```

devem passar os mesmos vetores.

---

# 60. Golden fixtures

Adicionar casos para:

- valid Goal lifecycle;
- illegal Goal mutation;
- Goal amendment;
- valid DAG;
- cyclic DAG;
- missing task dependency;
- evidence satisfaction;
- gate failure;
- gate waiver;
- permission denied;
- permission allowed by scope;
- untrusted context;
- workforce dependency resolution;
- recipe execution plan;
- context budget exhaustion;
- model routing;
- execution routing;
- retry/fallback attempts;
- event envelope compatibility;
- project migrations.

---

# 61. End-to-end conformance project

Criar um projeto pequeno em:

```text
examples/conformance-project/
```

Fluxo:

```text
bootstrap
→ Goal
→ Plan
→ Tasks
→ workforce resolution
→ context plan
→ evidence/gates
→ task report
→ completion
```

Nenhuma LLM real é necessária.

Usar fake adapters/executors determinísticos.

---

# 62. Fake runtime adapters para testes

Criar abstrações/fakes que permitam simular:

```text
success
failure
timeout
retry
fallback
invalid evidence
tool denial
permission denial
```

Sem chamadas externas.

---

# 63. Quality gates do framework

A v0.3 deve adicionar CI para:

```text
schema validation
workforce package validation
package provenance
resolver determinism
compiler determinism
Goal state machine
Plan DAG validation
Evidence/Gate validation
permission policy
migration fixtures
conformance suite
docs link validation
CLI smoke tests
```

---

# 64. Security gates do próprio framework

Adicionar:

```text
dependency audit
secret scanning
malformed package tests
path traversal tests
unsafe script/package tests
schema fuzz/property tests onde viável
```

Packages de Skills com scripts devem ser tratados como supply-chain-sensitive.

---

# 65. Documentation validation

Validar:

- links;
- ATLAS router;
- schema docs;
- CLI docs;
- examples;
- stale generated adapters;
- missing authoring docs;
- missing migration docs.

---

# 66. Non-goals da v0.3

Não implementar antes do Atlas Flow:

```text
remote skill marketplace
cloud registry
cloud sync
hosted execution
distributed scheduler
deep recursive agents
unbounded auto-routing
automatic model benchmarking service
full graphical UI
remote collaboration
```

---

# 67. Limite Project Atlas vs Atlas Flow

## Project Atlas Framework define

```text
Goal
Plan
Task
DAG
Agent
Skill
Recipe
Context
Evidence
Gate
Policy
Run
Attempt
Event
Model Policy
Execution Policy
```

## Atlas Flow implementará

```text
scheduler
event bus
process supervisor
ACP clients
MCP clients
Git worktrees
SQLite operational state
HTTP
WebSocket
AG-UI
React
Tauri
model gateway integration
UI
```

Nunca mover semântica canônica para o Atlas Flow sem RFC/ADR no framework quando for uma regra portátil.

---

# 68. Estrutura canônica sugerida do framework v0.3

Proposta inicial:

```text
project-atlas-framework/
│
├── FRAMEWORK.md
├── ENTRYPOINT.md
├── README.md
│
├── docs/
│
├── schemas/
│   ├── atlas.schema.json
│   ├── goal.schema.json
│   ├── goal-amendment.schema.json
│   ├── plan.schema.json
│   ├── task.schema.json
│   ├── evidence.schema.json
│   ├── gate.schema.json
│   ├── gate-waiver.schema.json
│   ├── agent.schema.json
│   ├── skill.schema.json
│   ├── recipe.schema.json
│   ├── context-*.schema.json
│   ├── model-policy.schema.json
│   ├── execution-policy.schema.json
│   ├── permission-policy.schema.json
│   ├── approval-policy.schema.json
│   ├── run.schema.json
│   ├── attempt.schema.json
│   └── event.schema.json
│
├── src/project_atlas/
│   ├── ...
│   ├── migrations/
│   └── resources/
│       ├── adapters/
│       └── workforce/
│           ├── agents/
│           ├── skills/
│           └── recipes/
│
├── conformance/
├── examples/
└── tests/
```

A implementação deve produzir um ADR se optar por outra estrutura persistente significativamente diferente.

---

# 69. Ordem obrigatória de implementação

## Phase A — Contracts first

Implementar:

1. protocol version metadata;
2. Goal v2;
3. Plan;
4. Task;
5. Evidence;
6. Gate;
7. Run;
8. Attempt;
9. Event;
10. Permission/Approval;
11. Context;
12. Model/Execution Policy.

### Gate A

Todos os schemas devem:

- validar fixtures positivas;
- rejeitar fixtures negativas;
- estar documentados.

---

## Phase B — Workforce packages

Implementar:

1. Skill Package v2;
2. Agent v2;
3. Recipe v2;
4. package loader;
5. package validator;
6. package resolver;
7. compiler v2.

### Gate B

Codex e Claude Code devem receber Skill Packages compilados com seus recursos.

---

## Phase C — Engineering + Security

Implementar P0 Skills:

```text
clean-code
architecture-quality
refactoring
error-handling
dependency-management
testing-quality
concurrency-quality
observability

threat-modeling
secure-coding
security-review
filesystem-security
process-execution-security
supply-chain-security
secrets-security
mcp-security
acp-security
untrusted-project-security
```

Criar:

```text
Security Architect
Security Reviewer
Quality Reviewer
```

### Gate C

Um exemplo de projeto deve resolver automaticamente Skills de segurança por risco.

---

## Phase D — GitHub + UI/UX

Implementar:

- GitHub issue Skills;
- Issue Author;
- Issue Triager;
- UI/UX workforce;
- recipes correspondentes.

### Gate D

Fixtures devem demonstrar seleção correta sem carregar workforce desnecessária.

---

## Phase E — CLI & explainability

Implementar:

```text
atlas doctor
atlas explain
atlas migrate
```

### Gate E

`atlas doctor` deve falhar em projeto propositalmente inconsistente.

---

## Phase F — Conformance

Implementar:

- golden fixtures;
- fake runtime;
- conformance runner;
- end-to-end project.

### Gate F

Reference implementation Python passa toda suite.

---

## Phase G — Documentation

Completar:

- getting started;
- user guide;
- authoring;
- integration;
- reference;
- development;
- examples.

### Gate G

Uma pessoa/LLM deve conseguir:

```text
install
bootstrap
author Skill
author Agent
author Recipe
create Goal
create Plan
validate
migrate
compile adapter
```

apenas pela documentação.

---

# 70. Critério de conclusão da v0.3

A v0.3 só está pronta quando:

- schemas centrais existem e são testados;
- Task DAG é machine-readable;
- evidence/gates são tipados;
- locked Goal possui integridade;
- permission/trust model existe;
- context protocol é formal;
- model/execution policy estão separados;
- Run/Attempt/Event possuem contratos;
- workforce packages reais substituem prompts mínimos;
- security/clean-code packages possuem conteúdo operacional real;
- GitHub issue workflow funciona;
- UI/UX workflow é suficientemente robusto para iniciar o Atlas Flow frontend;
- migrations são determinísticas;
- conformance suite passa;
- doctor/explain funcionam;
- documentação é utilizável.

---

# 71. Critério para iniciar o novo Atlas Flow

Assim que todos os P0 acima passarem, **parar de adicionar features abstratas ao framework**.

Começar o reboot Rust do Atlas Flow.

A partir desse momento:

```text
Atlas Flow development
        ↓
discovers protocol gap
        ↓
classify gap
        │
        ├── runtime-specific
        │      → implement only in Atlas Flow
        │
        └── portable Project Atlas rule
               → RFC/ADR in framework
```

Essa regra evita framework infinito.

---

# 72. Atlas Flow como reference implementation

O novo Atlas Flow deverá futuramente:

- consumir schemas v0.3;
- passar Conformance Suite;
- não duplicar regras canônicas;
- reportar unsupported protocol version;
- executar migrations somente de modo explícito;
- tratar framework como contrato externo.

O Atlas Flow não deve depender da implementação Python do framework em runtime.

---

# 73. Diretrizes para a LLM implementadora

## 73.1 Não reescrever por reescrever

Preservar código v0.2 que:

- funciona;
- possui testes;
- é compatível com v0.3;
- não conflita com os novos contratos.

## 73.2 Schema first

Para qualquer conceito novo:

```text
semantics
→ schema
→ fixtures
→ validator
→ implementation
→ docs
```

Evitar implementar comportamento antes de definir contrato.

## 73.3 Tests with implementation

Toda feature precisa de:

- happy path;
- invalid path;
- regression;
- deterministic behavior.

## 73.4 No silent weakening

Nunca:

- remover gate porque teste falha;
- relaxar schema apenas para fixture passar;
- modificar locked acceptance sem amendment;
- ignorar incompatibilidade de versão.

## 73.5 Avoid duplicate canonical truth

Se package manifest vira canonical:

- não manter cópia manual equivalente em `skills.json`.

Indices devem ser gerados.

## 73.6 Backward compatibility

Preservar leitura/migração de v0.2 onde razoável.

Não gerar novos projetos em formatos legados.

## 73.7 Token economy

Documentação e Skills podem ser robustas, mas:

- entrypoints curtos;
- progressive loading;
- references sob demanda;
- sem duplicação.

---

# 74. Deliverables esperados da implementação

Ao final, espera-se pelo menos:

```text
schemas atualizados
workforce package loader
workforce package compiler
new workforce packages
new agents
new recipes
new resolver features
doctor
explain
migrate
conformance suite
examples
docs
tests
CHANGELOG
migration guide
```

Atualizar:

```text
README.md
FRAMEWORK.md
ENTRYPOINT.md
docs/ATLAS.md
docs/ARCHITECTURE.md
docs/AI_WORKFORCE.md
docs/QUALITY.md
docs/PORTABILITY.md
SECURITY.md
CHANGELOG.md
```

quando impactados.

---

# 75. CHANGELOG esperado

Registrar v0.3 como mudança significativa.

Separar:

```text
Added
Changed
Deprecated
Removed
Migration
Compatibility
```

Não esconder breaking changes.

---

# 76. ADRs recomendados

Criar ADRs para pelo menos:

```text
Skill Package v2
Execution-Ready Protocol boundaries
Plan/Task/DAG contract
Evidence/Gate contract
Permission/Trust model
Model vs Execution Policy
Run/Attempt/Event Protocol
Protocol Versioning
Conformance Suite
```

Evitar um ADR gigante se decisões forem independentes.

---

# 77. Definition of Done por componente

Cada componente só está DONE quando:

```text
contract/schema
implementation
validation
tests
negative tests
docs
migration impact
examples
```

estão completos.

---

# 78. Definition of Done do projeto v0.3

```text
Fresh clone
→ install
→ tests
→ conformance
→ bootstrap example
→ validate
→ explain workforce
→ doctor
→ compile Codex
→ compile Claude Code
→ migrate v0.2 fixture
```

Tudo deve funcionar de maneira reproduzível.

---

# 79. Requisitos de segurança para scripts em Skill Packages

Se uma Skill contém `scripts/`:

- script não ganha execução automática;
- manifest declara capabilities;
- execution policy precisa autorizar;
- script precisa permanecer sob scope;
- package provenance precisa existir;
- scripts externos/untrusted precisam de tratamento mais restritivo.

No futuro, Atlas Flow poderá sandboxear; o framework já deve definir a policy.

---

# 80. Requisitos de licença

Skills ou referências copiadas/adaptadas externamente devem registrar:

```text
source
license
attribution
modifications
```

Não incorporar material de licença incompatível.

---

# 81. Requisitos de determinismo

Devem ser determinísticos:

- resolver;
- package selection;
- dependency resolution;
- schema validation;
- migration;
- Plan validation;
- DAG ordering quando um tie-break for necessário;
- compiler output dado o mesmo input.

LLMs não fazem parte desses caminhos determinísticos.

---

# 82. Requisitos de explicabilidade

Decisões automáticas precisam ser rastreáveis:

```text
why agent selected
why skill selected
why gate required
why permission denied
why model selected
why fallback occurred
why context expanded
```

Esse requisito é essencial para o futuro Atlas Flow.

---

# 83. Requisitos de future-proofing

Os contratos devem permitir futuros backends sem mudar semântica central:

```text
ACP
CLI
HTTP API
local model
gateway
remote orchestrator
```

E futuras UIs:

```text
Atlas Flow Desktop
Web
CLI
VS Code
mobile/remote
```

---

# 84. Ponto de parada

Após a v0.3 cumprir esta especificação:

> **Não continuar adicionando features de framework antes de iniciar Atlas Flow.**

Project Atlas precisa ser testado sob pressão real.

O Atlas Flow será a implementação de referência que validará:

- clareza dos contratos;
- event model;
- run model;
- context model;
- workforce packages;
- permission model;
- policy separation;
- conformance.

---

# 85. Resultado esperado

Depois desta evolução, a relação deve ser:

```text
PROJECT ATLAS FRAMEWORK
───────────────────────
Portable protocol

Defines:
Goals
Plans
Tasks
DAGs
Workforce
Context
Evidence
Gates
Policies
Runs
Attempts
Events
Models
Execution
Conformance

            │
            ▼

ATLAS FLOW
──────────
Reference runtime

Rust
Tokio
Axum
SQLx
SQLite
ACP
MCP
AG-UI
Git worktrees
React
Tauri
```

O Project Atlas continua utilizável sem Atlas Flow.

O Atlas Flow pode evoluir sem se tornar a fonte de verdade do framework.

Essa separação é parte do contrato v0.3.

---

# 86. Mandato final para a LLM

Implemente o Project Atlas Framework v0.3 conforme esta especificação.

Prioridades:

1. preservar invariantes atuais;
2. tornar os contratos executáveis;
3. substituir Skills/Recipes superficiais por packages operacionais;
4. formalizar Plan/Task/DAG/Evidence/Gate/Context/Run/Attempt/Event;
5. criar permissions/trust policies robustas;
6. separar model selection de execution routing;
7. criar conformance suite;
8. criar doctor/explain/migrate;
9. amadurecer Clean Engineering, Security, GitHub e UI/UX workforce;
10. completar a documentação.

Não introduza dependência obrigatória no Atlas Flow.

Não torne nenhum provider obrigatório.

Não converta o framework em um runtime agentic.

Não enfraqueça Goals, gates ou evidências para manter compatibilidade.

Quando uma decisão de implementação alterar o protocolo, registre ADR/RFC correspondente.

Quando uma questão for apenas detalhe interno do CLI Python, escolha a solução mais simples e mantenível sem expandir desnecessariamente o protocolo.

Ao final, entregue:

- código;
- schemas;
- migrations;
- workforce packages;
- recipes;
- agents;
- conformance fixtures;
- testes;
- documentação;
- exemplos;
- CHANGELOG;
- relatório de validação.

O critério final é simples:

> **Uma implementação independente em Rust deve conseguir implementar Project Atlas v0.3 apenas a partir dos contratos, schemas, fixtures e documentação do framework, sem precisar ler o código Python para descobrir a semântica.**

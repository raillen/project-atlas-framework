# Harness Gap Register (canônico)

Single source of truth for everything missing. Authority: task DoD (§35) +
snapshot `2026-09-11-d41f18bb65f5` (págs. 18, 19, 22–27, 29–35, 40–41) +
code/tests reality. `ACCEPTED != implemented`; environmental and
approval-gated items are labeled, never disguised.

**Disciplina:** todo incremento atualiza os Status aqui e imprime a tabela
no relatório da rodada. IDs são estáveis — nunca reutilizar um ID para outro
item. Novos gaps entram no fim com o próximo número livre.

**Legenda:** ✅ done · 🟡 partial · ⬜ open · 🛑 blocked · ⏸️ deferred ·
🚫 out-of-scope (deste Goal).

## 1. Capability gaps

| ID | Item | Fonte | Status | Evidência / limite | Próximo passo |
|----|------|-------|--------|--------------------|---------------|
| GAP-001 | Budget envelope no path real (flags, consumo, persistência) | DoD §35, HA4 | 🟡 partial | hook existe + testes; `agent run`/daemon não instanciam envelope | fiação CLI + persistência no checkpoint |
| GAP-002 | Evidence/Gates do protocolo no loop | HA2/HA4, pág. 05 | ⬜ open | zero refs a `protocol/evidence` nos paths de run | registrar evidence + gate `--strict` |
| GAP-003 | Team binding real (Runner aninhado por role) | HA10/HA14 | 🟡 partial | `team.Runner` com `Work` injetado; sem binding p/ Runner real | `Work` padrão com runs/checkpoints/Handoff |
| GAP-004 | Egress + segredos no path local | HA-seg, págs. 07/19 | 🟡 partial | `internal/egress` (Allow/SecretProvider/Redactor) existe e não é consultado; container nega rede | consultar egress no exec local; secret refs + redação |
| GAP-005 | Roteamento por custo/latência/privacidade/quota | HA6, pág. 24 | 🟡 partial | Gateway healthy-first; sem score de pricing/latency/data-class | score via `modelregistry` + quota-awareness |
| GAP-006 | Retrieval estrutural (FTS/BM25, símbolos/LSP, repo map, RRF multi-fonte) | HA9, págs. 08/31 | 🟡 partial | `CompileWorkspace` (lista+git+exact); `internal/indexing` desligado | FTS primeiro, LSP depois |
| GAP-007 | Regiões gerenciadas + JSON Patch no doc compiler | HD-base, pág. 26 | 🟡 partial | whole-file + CAS prontos; AST-regions pendentes | regions + Pointer/Patch |
| GAP-008 | Merge/conflict explícito entre worktrees | HA10 | ⬜ open | review existe; merge op não existe | operação explícita com policy |
| GAP-009 | Steering + compaction | HA2, pág. 05 | ⬜ open | `Yield` existe; follow-up e compactação não | enqueue p/ safe point; compact com refs |
| GAP-010 | ACP Agent Server (expor Runtime a editores) | H11, págs. 06/22 | ⬜ open | só lado cliente/probe | mapear session/prompt/permission p/ Run |
| GAP-011 | MCP client real (SDK + transporte) | H5, págs. 06/19 | 🟡 partial | governance por descriptors; sem SDK ligado | depende de GAP-031 (escolha do SDK) |
| GAP-012 | Benchmarks (packing, compilação, Runner) | relatório | ⬜ open | zero medições registradas | `go test -bench` + tabela no relatório |
| GAP-013 | Operação do daemon (PID lock, rotação, unit, stop) | daemon | 🟡 partial | serve/ps/logs ok; sem supervisão | lock + rotação + unit systemd + `stop` |
| GAP-014 | Kill -9 real com side effect pendente | HA4/HA8 | ⬜ open | resume testado via reload, não via SIGKILL | teste de kill + resume sem duplicar |
| GAP-015 | Approvals persistidas | HA3/pág. 07 | 🟡 partial | `perm.Engine.Log` só em memória | persistir resolutions no store do run |
| GAP-016 | Bridge AgentEvent → observability.Event | HA4 | ⬜ open | JSONL próprio; `internal/observability` desligado | projetar/encaminhar eventos |
| GAP-017 | Planning→Build (PlanningSession ⇒ Run) | pág. 25 | 🟡 partial | `PlanningSession` existe e não promove p/ Run | promoção com provenance, sem virar transcript |
| GAP-018 | Memory Atlas cross-project | págs. 25/27-G12 | ⬜ open | zero ocorrências no repo | modelo serializável + gates + fonte do Context |
| GAP-019 | Agent writes KnowledgeDelta-first (G15) | pág. 27-G15 | 🟡 partial | seeding usa `Put` direto; Delta existe p/ promoção global | rotear mutações de run por Validate→Commit |
| GAP-020 | Retention/GC (checkpoints, eventos, knowledge) | pág. 27-G23 | ⬜ open | crescimento ilimitado por run | política + GC sem dangling provenance |
| GAP-021 | Retry com backoff no Gateway | HA6 | 🟡 partial | fallback sim, retry/backoff não | backoff por classe de erro + jitter |
| GAP-022 | Child runs/subagentes com ownership (H14) | H14 | 🟡 partial | team executa roles; runs aninhados com checkpoint não | aninhar Runner + Handoff pai↔filho |
| GAP-023 | Scheduled/background agents (H16) | H16 | ⬜ open | daemon sem scheduler | cron-like mínimo atrás do daemon |
| GAP-024 | Provedores sandbox adicionais (H13) | H13 | 🟡 partial | local/worktree/container-detect | gVisor/strong + remoto |
| GAP-025 | Contratos Local Intel (KnowledgeTask, Router, ResourceManager, Supervisor) | págs. 29–30 | ⬜ open | deferred com workers | contratos primeiro, workers depois |
| GAP-026 | Research Ledger first-class (G11) | pág. 27-G11 | 🟡 partial | ResearchRecord existe como tipo; ledger dedicado não | ledger + API como Decision Ledger |
| GAP-027 | Token-estimate index (G18) | pág. 27-G18 | 🟡 partial | heurística len/4 inline; sem índice versionado | índice medido + pricing |
| GAP-028 | Lint dos agent docs + doc-evals (G19/G21) | pág. 27 | 🟡 partial | testes HD existem; lint de docs de agentes não | lint + eval de docs |
| GAP-029 | Schema evolution/migrations (G20) | pág. 27 | ⬜ open | schemas sem plano de migração | política + testes de migração |
| GAP-030 | Transporte remoto do daemon (+auth) | split gate | ⬜ open | socket local apenas | TLS + token (desenho antes) |
| GAP-031 | Decisão: MCP Go SDK e transports | pág. 19 | 🛑 decision | — | benchmark + ADR |
| GAP-032 | Decisão: subset ACP + matriz oficial | pág. 19 | 🛑 decision | cliente OK; servidor em GAP-010 | definir ordem + registry |
| GAP-033 | Decisão: Docker vs Podman padrão/rootless | pág. 19 | 🛑 decision | detecção honesta pronta | medir + ADR |
| GAP-034 | Decisão: driver SQLite derived runtime | pág. 19 | 🛑 decision | — | medir + ADR |
| GAP-035 | Decisão: isolamento de plugins | pág. 19 | 🛑 decision | — | ADR quando houver plugins |
| GAP-036 | Decisão: budgets de performance + TTL/memory-pressure | pág. 19 | 🛑 decision | — | medir + ADR |
| GAP-037 | Prova live do container | HA5 | 🛑 env | `TestContainerLive` pronto; daemon inacessível aqui | `PRUMO_LIVE_DOCKER=1` onde houver daemon |
| GAP-038 | Chaves live de models | HA6 | 🛑 env | adapters prontos + stub-testados | `PRUMO_MODEL_API_KEY`/BASE_URL |
| GAP-039 | Sends live externos (opencode/codex) | HA7/HA8 | 🛑 approval | **gastam sua quota**; tudo ao redor live-verificado | sua aprovação explícita de spend |
| GAP-040 | Modelos locais + thresholds (benchmark-driven) | pág. 19/29 | 🛑 env+decision | No-LLM first-class mantido | hardware + corpus + aprovação |
| GAP-041 | Bindings não-Go (TS types do IDL) | HA11 | ⬜ open | IDL + SDK Go prontos | gerar types do manifesto |
| GAP-042 | Checkpoint retention no daemon store | HA4 | ⬜ open | ver GAP-020 (caso particular) | incluir na política de retenção |
| GAP-043 | Descoberta de modelos nos adapters reais | HA1 | 🟡 partial | `ModelDiscovery:false` nos dois adapters | listar via `/models` + Anthropic equivalente |
| GAP-044 | Structured-output enforcement | HA1 | 🟡 partial | capability anunciada; sem validação | validar contra schema no adapter |
| GAP-045 | Impact analysis lexical (G5, legado M5) | pág. 27-G5 | 🟡 partial | pré-Harness; fora do path do run | migrar p/ relações tipadas quando tocar M5 |

## 2. Definition of Done (§35) — estado por item

| # | Item DoD | Status |
|---|----------|--------|
| 1 | docs reconciled/promoted | ✅ |
| 2 | AgentRuntime contracts | ✅ |
| 3 | FakeProvider conformance | ✅ |
| 4 | ≥1 real ModelProvider works | ✅ código (live: GAP-038) |
| 5 | NativeAgent end-to-end | ✅ |
| 6 | tools via ToolGateway | ✅ |
| 7 | permission lifecycle | ✅ (persistência: GAP-015) |
| 8 | Environment/Sandbox baseline | ✅ (live: GAP-037) |
| 9 | checkpoint/restart/resume | ✅ (kill real: GAP-014) |
| 10 | duplicate side effects prevented | ✅ |
| 11 | budget enforcement | 🟡 runtime sim, fiação CLI não (GAP-001) |
| 12 | observability/events | ✅ (bridge: GAP-016) |
| 13 | gateway routing/fallback | ✅ (políticas ricas: GAP-005) |
| 14 | ≥1 external AgentProvider works | ✅ nos limites (send: GAP-039) |
| 15 | typed Handoff | ✅ |
| 16 | Context Compiler v2 baseline | ✅ (retrieval rico: GAP-006) |
| 17 | Knowledge Runtime baseline | ✅ (Atlas: GAP-018) |
| 18 | KnowledgeDelta validate/commit | ✅ (writes-first: GAP-019) |
| 19 | Coverage/Readiness baseline | ✅ |
| 20 | Documentation Compiler baseline | ✅ (regions: GAP-007) |
| 21 | multi-agent/worktree baseline | ✅ (binding/merges: GAP-003/008/022) |
| 22 | compatibility/eval suite | ✅ (fuzz/bindings: GAP-041) |
| 23 | `prumo agent` headless usable | ✅ |
| 24 | docs describe reality | ✅ |
| 25 | no P0 contradictions | ✅ (1 não-P0 registrada: regra de imports do `cmd`) |

## 3. Split gate — estado por item

| Item | Status |
|------|--------|
| HA0 contracts | ✅ |
| HA1 FakeProvider + first ModelProvider | ✅ |
| HA2 Native Agent | ✅ |
| HA3 Tool/Permission | ✅ |
| HA4 checkpoint/restart/resume | ✅ |
| HA5 ACI + Sandbox baseline | ✅ c/ limites |
| headless coding Run end-to-end | ✅ |
| versioned public protocol | ✅ (IDL + SDK Go; TS: GAP-041) |
| reconnect/replay | ✅ local (remoto: GAP-030) |
| **Veredito** | **NOT READY** — GAP-030 + GAP-041 + provas live |

## 4. Fora deste Goal (não entra na conta)

Desktop/TUI (prumo-code), execução cloud/microVM (H18), federação A2A (H17), site público/i18n (HD5+), polish visual, `prumo-code` em si.

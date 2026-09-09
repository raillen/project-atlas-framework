# Manual de Instalação

## Escolha o método

| Cenário | Método |
|---------|--------|
| Usuário final | Binário de uma release publicada |
| Desenvolvimento do Atlas | `go build` ou `go run` |
| CI, testes e USB/devbox | `--home` ou `ATLAS_HOME` |
| Compatibilidade v0.3 | Python + `pip install -e '.[dev]'` |

## Desenvolvimento a partir do repositório

```bash
git clone git@github.com:raillen/project-atlas-framework.git
cd project-atlas-framework
go version
go run ./cmd/atlas version
```

Compilar um binário local:

```bash
go build -trimpath -o ./atlas ./cmd/atlas
./atlas version
```

O binário Go não requer Python, Node ou CGO.

## Build de release

O script oficial gera Linux amd64/arm64, macOS amd64/arm64 e Windows amd64/arm64:

```bash
VERSION=0.4.0 sh scripts/release.sh
cat dist/checksums.txt
cat dist/release.json
```

A instalação de release deve verificar o checksum antes de substituir o binário. Assinaturas de release permanecem um gate operacional antes da distribuição pública final.

## Estado global e modo portátil

Por padrão, o Atlas usa `~/.atlas` em sistemas Unix-like. Use `ATLAS_HOME` ou `--home` para isolar o estado:

```bash
ATLAS_HOME="$PWD/.atlas-home" ./atlas setup
./atlas --home "$PWD/.atlas-home" setup
```

O estado global contém:

```text
ATLAS_HOME/
├── config/installation.json
├── cache/
├── logs/
├── connectors/
└── runtimes/
```

Não confunda esse estado com `.atlas/` dentro de um projeto. O segundo pertence ao projeto e pode conter estado derivado local.

## Setup

```bash
atlas --home ~/.atlas setup
```

`setup` é idempotente e registra:

- versão do Atlas;
- caminho do binário;
- harnesses detectados no `PATH`;
- estado de conectores;
- caminhos gerenciados pelo Atlas.

O comando não altera configurações globais de harnesses automaticamente.

## Instalar um conector

```bash
atlas --home ~/.atlas install connector opencode
```

O estado do conector deve possuir um cleanup manifest antes de qualquer remoção automática.

## Verificar instalação

```bash
atlas version
atlas --json version
atlas framework-check
atlas --json framework-check
```

Verificação de projeto:

```bash
atlas validate ./my-project
atlas doctor ./my-project
```

## Python v0.3: somente oracle e desenvolvimento

Para executar a implementação de referência e a suíte Python:

```bash
python3 -m venv .venv
. .venv/bin/activate
python -m pip install -e '.[dev]'
pytest
```

O Python continua necessário para a conformance oracle v0.3, não para o runtime Go.

## Problemas comuns

### `go: command not found`
Instale Go 1.22 ou superior e verifique `go version`.

### `atlas: command not found`
Use o caminho absoluto do binário ou coloque o diretório do binário no `PATH`:

```bash
export PATH="$PWD/bin:$PATH"
```

### `missing atlas.json`
O comando foi executado fora de um projeto Atlas. Informe o caminho do projeto ou execute `atlas init` primeiro.

### Falha de checksum
Não use o artefato. Baixe novamente e compare com `dist/checksums.txt` ou com o checksum publicado pela release.

### Estado global corrompido
Use um home portátil novo para isolar o diagnóstico:

```bash
atlas --home "$PWD/atlas-recovery-home" setup
```

# Manual de Desinstalação

## Regra de segurança

A desinstalação do Atlas nunca remove dados de projetos.

Permanecem intocados:

- `.ai/`;
- `atlas.json`;
- `docs/`;
- Goals, Plans, Tasks, Evidence;
- histórico Git;
- arquivos criados pelo usuário.

Remoção de projeto continua sendo decisão explícita do usuário com as ferramentas normais do sistema operacional e do Git.

## Remoção básica

```bash
atlas uninstall
atlas --home ~/.atlas uninstall
```

Esse comando atualiza o manifest global e remove somente estado de instalação diretamente rastreado pelo Atlas.

## Remover conectores

```bash
atlas uninstall --connectors
```

Um conector só é removido quando possui cleanup manifest com caminhos criados pelo próprio Atlas. Caminhos fora do manifest, arquivos modificados pelo usuário, ou ownership incerto são reportados como leftovers em vez de apagados.

## Limpar cache e estado derivado

```bash
atlas uninstall --purge-cache
```

Remove cache, logs locais, bases derivadas e arquivos temporários sob `ATLAS_HOME`. Esse comando não altera o manifest principal a menos que `--purge-global-config` também seja usado.

## Remover configuração global

```bash
atlas uninstall --purge-global-config
```

Remove `ATLAS_HOME/config`. A execução seguinte cria um estado limpo através de `atlas setup`. Dados de projetos continuam preservados.

## Remoção completa do binário

O CLI não apaga seu próprio executável. Remova o binário instalado pelo procedimento normal do sistema:

```bash
rm /usr/local/bin/atlas
```

Remove também o diretório home somente quando ele não contiver mais nada necessário:

```bash
rm -rf ~/.atlas
```

Confirme antes que o diretório não contenha segredos, chaves, backups locais ou runtime packages ainda necessários.

## Verificação pós-remoção

```bash
command -v atlas || echo "binário removido"
test ! -e ~/.atlas/config/installation.json && echo "configuração global removida"
git -C ./my-project status --short
```

O último comando deve mostrar que o projeto permanece funcional e versionado, mesmo sem o Atlas instalado globalmente.

## Recuperação de erro

Se uma remoção for interrompida:

1. Execute `atlas setup` em um home temporário e novo.
2. Compare os manifests gerados.
3. Restaure apenas backups registrados no cleanup manifest.
4. Nunca restaure um backup sobre um arquivo modificado pelo usuário sem confirmar conteúdo e checksum.

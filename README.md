# Sistema Runner

[![CI & Release](https://github.com/kyriosdata/assinatura/actions/workflows/ci.yml/badge.svg)](https://github.com/kyriosdata/assinatura/actions/workflows/ci.yml)

Projeto acadêmico para a disciplina de **Implementação e Integração** — Bacharelado em Engenharia de Software (UFG 2026).

O Sistema Runner facilita o acesso a operações de assinatura digital simulada e ao Simulador do HubSaúde, abstraindo a complexidade de configuração do ambiente Java.

---

## Componentes

| Componente | Tecnologia | Descrição |
|------------|-----------|-----------|
| `assinatura` | Go 1.24 + Cobra | CLI principal: assinar, validar, gerenciar servidor |
| `simulador` | Go 1.24 + Cobra | CLI dedicado ao Simulador do HubSaúde |
| `assinador.jar` | Java 21 + Maven | Backend que valida parâmetros e simula assinaturas |

---

## Início Rápido

### 1. Baixar o CLI

Acesse a [página de releases](https://github.com/kyriosdata/assinatura/releases) e baixe o binário para sua plataforma:

```bash
# Linux
curl -LO https://github.com/kyriosdata/assinatura/releases/latest/download/assinatura-<versão>-linux-amd64
chmod +x assinatura-<versão>-linux-amd64
mv assinatura-<versão>-linux-amd64 /usr/local/bin/assinatura
```

O Java é provisionado automaticamente pelo CLI caso não esteja disponível.

### 2. Criar uma assinatura

```bash
assinatura sign --content "documento a assinar"
# [OK] Assinatura simulada gerada com sucesso.
# Assinatura: MOCKED_SIGNATURE_BASE64_==
```

### 3. Validar uma assinatura

```bash
assinatura validate \
  --content "documento a assinar" \
  --signature "MOCKED_SIGNATURE_BASE64_=="
# [OK] Assinatura válida.
```

### 4. Usar o modo servidor (menor latência)

```bash
# Iniciar servidor em background
assinatura start

# Operações subsequentes usam HTTP automaticamente
assinatura sign --content "doc 1"
assinatura sign --content "doc 2"

# Encerrar quando terminar
assinatura stop
```

### 5. Gerenciar o Simulador do HubSaúde

```bash
# Inicia e baixa o simulador.jar automaticamente se necessário
assinatura simulador start

assinatura simulador status
assinatura simulador stop
```

---

## Referência de Comandos

| Comando | Descrição |
|---------|-----------|
| `assinatura version` | Exibe a versão do CLI |
| `assinatura sign --content <texto>` | Cria assinatura simulada |
| `assinatura validate --content <texto> --signature <sig>` | Valida assinatura |
| `assinatura start [--port N] [--timeout N]` | Inicia assinador.jar como servidor HTTP |
| `assinatura stop [--port N]` | Encerra o servidor |
| `assinatura status [--port N]` | Exibe status do servidor |
| `assinatura simulador start [--port N] [--source <url>]` | Inicia o Simulador |
| `assinatura simulador stop` | Encerra o Simulador |
| `assinatura simulador status` | Exibe status do Simulador |

Flags globais úteis: `--local` (força modo local em `sign`/`validate`).

---

## Compilar a partir do código-fonte

### Pré-requisitos

- Go 1.24+
- Java 21 + Maven (em `/opt/maven/bin/mvn` ou no `PATH`)

### Go CLI

```bash
cd projetos/assinatura
go mod download
go build ./...          # compila assinatura e simulador
go test ./...           # executa todos os testes
```

### Java backend

```bash
cd projetos/assinador-java
mvn compile             # compila
mvn test                # executa 38 testes
mvn package             # gera assinador.jar em target/
```

---

## CI/CD e Releases

O pipeline em `.github/workflows/ci.yml` dispara a cada push na `main` e:

1. Compila para **Linux**, **Windows** e **macOS** (amd64)
2. Gera checksums **SHA-256** para todos os artefatos
3. Assina com **Cosign** (identidade OIDC + Sigstore transparency log)
4. Publica no **GitHub Releases** quando a versão é nova

Para lançar uma release, incremente `var version` em `projetos/assinatura/cmd/version.go` e faça push na `main`.

### Verificar autenticidade de um artefato

```bash
cosign verify-blob \
  --certificate assinatura-<versão>-linux-amd64.pem \
  --signature assinatura-<versão>-linux-amd64.sig \
  assinatura-<versão>-linux-amd64
```

---

## Documentação

| Documento | Descrição |
|-----------|-----------|
| [Manual do Usuário](docs/manual-usuario.md) | Guia completo de instalação e uso |
| [Especificação](especificacao.md) | Requisitos funcionais e critérios de aceitação |
| [Design (C4)](design.md) | Diagramas de arquitetura |
| [Plano de Sprints](docs/plano-revisitado-v2.md) | Histórias de usuário e rastreabilidade |

---

## Tecnologias

- **Go 1.24** · **Cobra** · **GitHub Actions** · **Cosign / Sigstore**
- **Java 21** · **Maven** · **SunPKCS11** · **JUnit 5**
- **Eclipse Temurin (Adoptium)** — provisionamento automático de JDK
- **PlantUML / C4 Model** — diagramas de arquitetura

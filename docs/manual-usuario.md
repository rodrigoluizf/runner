# Manual do Usuário — Sistema Runner

## 1. Visão Geral

O **Sistema Runner** fornece uma interface de linha de comandos (CLI) para executar operações de assinatura digital simulada e gerenciar o Simulador do HubSaúde, sem exigir configuração manual do ambiente Java.

O sistema é composto por dois CLIs e um backend Java:

| Componente | Descrição |
|------------|-----------|
| `assinatura` | CLI principal: assinar, validar, gerenciar o assinador e o simulador |
| `assinador.jar` | Backend Java que executa as operações de assinatura simulada |
| `simulador.jar` | Simulador do HubSaúde (baixado automaticamente) |

---

## 2. Instalação

### 2.1. Download dos binários

Acesse a [página de releases](https://github.com/kyriosdata/assinatura/releases) e baixe o binário correspondente à sua plataforma:

| Plataforma | Arquivo |
|------------|---------|
| Linux (x64) | `assinatura-<versão>-linux-amd64` |
| Windows (x64) | `assinatura-<versão>-windows-amd64.exe` |
| macOS (x64) | `assinatura-<versão>-darwin-amd64` |

### 2.2. Verificação de integridade (opcional)

Cada release inclui checksums SHA-256 e assinatura via Cosign. Para verificar:

```bash
# Verificar checksum
sha256sum -c assinatura-<versão>-linux-amd64.sha256

# Verificar assinatura com Cosign
cosign verify-blob \
  --certificate assinatura-<versão>-linux-amd64.pem \
  --signature assinatura-<versão>-linux-amd64.sig \
  assinatura-<versão>-linux-amd64
```

### 2.3. Tornar executável (Linux/macOS)

```bash
chmod +x assinatura-<versão>-linux-amd64
mv assinatura-<versão>-linux-amd64 /usr/local/bin/assinatura
```

### 2.4. Dependências

O CLI gerencia o Java automaticamente. Na primeira execução de um comando que precise do Java, o sistema:

1. Verifica se há JDK 21+ no `JAVA_HOME` ou no `PATH`
2. Verifica se há JDK gerenciado em `~/.hubsaude/jdk/`
3. Se não encontrado, baixa o JDK 21 (Eclipse Temurin) automaticamente

O `assinador.jar` deve estar no mesmo diretório do CLI, ou o caminho deve ser definido via variável de ambiente:

```bash
export ASSINADOR_JAR=/caminho/para/assinador.jar
```

---

## 3. Referência de Comandos

### 3.1. `assinatura version`

Exibe a versão atual do CLI.

```
$ assinatura version
0.0.6
```

---

### 3.2. `assinatura sign`

Cria uma assinatura digital simulada para o conteúdo fornecido.

**Parâmetros:**

| Flag | Obrigatório | Descrição |
|------|-------------|-----------|
| `--content` | Sim | Conteúdo a ser assinado |
| `--token` | Não | Token de autenticação do dispositivo criptográfico |
| `--local` | Não | Forçar invocação direta (modo local), ignorando servidor em execução |

**Exemplos:**

```bash
# Assinatura simples
assinatura sign --content "documento a assinar"

# Saída esperada:
# [OK] Assinatura simulada gerada com sucesso.
# Assinatura: MOCKED_SIGNATURE_BASE64_==

# Forçando modo local
assinatura sign --content "documento" --local
```

**Comportamento automático:**
- Se houver um servidor `assinador.jar` em execução na porta padrão (8080), usa-o via HTTP
- Caso contrário, invoca o `assinador.jar` diretamente (modo local)
- Use `--local` para sempre usar o modo local

---

### 3.3. `assinatura validate`

Valida uma assinatura digital simulada.

**Parâmetros:**

| Flag | Obrigatório | Descrição |
|------|-------------|-----------|
| `--content` | Sim | Conteúdo original que foi assinado |
| `--signature` | Sim | Assinatura a ser validada |
| `--local` | Não | Forçar invocação direta (modo local) |

**Exemplos:**

```bash
# Validar assinatura
assinatura validate \
  --content "documento a assinar" \
  --signature "MOCKED_SIGNATURE_BASE64_=="

# Saída para assinatura válida:
# [OK] Assinatura válida.

# Saída para assinatura inválida:
# [FALHA] Assinatura inválida.
```

---

### 3.4. `assinatura start`

Inicia o `assinador.jar` no modo servidor HTTP em background.

**Parâmetros:**

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `--port` | 8080 | Porta em que o servidor irá escutar |
| `--timeout` | 0 | Encerrar automaticamente após N minutos de inatividade (0 = nunca) |

**Exemplos:**

```bash
# Iniciar na porta padrão
assinatura start

# Iniciar em porta customizada com timeout de 30 minutos
assinatura start --port 9090 --timeout 30
```

O PID e porta do processo são registrados em `~/.hubsaude/assinador.json`. Logs são gravados em `~/.hubsaude/assinador.log`.

---

### 3.5. `assinatura stop`

Encerra o `assinador.jar` em execução.

**Parâmetros:**

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `--port` | 8080 | Porta do servidor a encerrar |

**Exemplos:**

```bash
assinatura stop
assinatura stop --port 9090
```

---

### 3.6. `assinatura status`

Exibe o status atual do `assinador.jar`.

**Parâmetros:**

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `--port` | 8080 | Porta do servidor a verificar |

**Exemplos:**

```bash
assinatura status

# Saída quando em execução:
# assinador.jar em execução na porta 8080 (PID 12345)

# Saída quando parado:
# assinador.jar não está em execução na porta 8080
```

---

### 3.7. `assinatura simulador start`

Inicia o Simulador do HubSaúde. O `simulador.jar` é baixado automaticamente se necessário.

**Parâmetros:**

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `--port` | 8443 | Porta do simulador |
| `--source` | — | URL alternativa para download do `simulador.jar` |

**Exemplos:**

```bash
# Iniciar com configuração padrão (baixa automaticamente se necessário)
assinatura simulador start

# Usar porta customizada
assinatura simulador start --port 9443

# Baixar de URL alternativa
assinatura simulador start --source https://meu-servidor.com/simulador.jar
```

O JAR baixado é armazenado em `~/.hubsaude/simulador.jar`. Download é omitido se a versão local já for a mais recente.

---

### 3.8. `assinatura simulador stop`

Encerra o Simulador do HubSaúde.

```bash
assinatura simulador stop
```

---

### 3.9. `assinatura simulador status`

Exibe o status atual do Simulador.

```bash
assinatura simulador status

# Saída quando em execução:
# [OK] Simulador ativo (PID 67890)

# Saída quando parado:
# [INFO] Simulador parado
```

---

## 4. Fluxos Típicos de Uso

### 4.1. Uso pontual (modo local)

```bash
# Executar diretamente sem iniciar servidor
assinatura sign --content "meu documento" --local
```

### 4.2. Uso intenso (modo servidor)

```bash
# 1. Iniciar servidor (uma vez)
assinatura start

# 2. Executar múltiplas operações (reutiliza servidor)
assinatura sign --content "doc 1"
assinatura sign --content "doc 2"
assinatura validate --content "doc 1" --signature "MOCKED_SIGNATURE_BASE64_=="

# 3. Encerrar quando terminar
assinatura stop
```

### 4.3. Gerenciar o Simulador do HubSaúde

```bash
# Iniciar simulador (baixa automaticamente na primeira vez)
assinatura simulador start

# Verificar se está rodando
assinatura simulador status

# Encerrar
assinatura simulador stop
```

---

## 5. Suporte a Dispositivo Criptográfico (PKCS#11)

O `assinador.jar` suporta inicialização do provider PKCS#11 via SunPKCS11. Quando o dispositivo (token USB ou smart card) não está disponível, o sistema continua operando no modo simulado com aviso.

### 5.1. Configuração

Crie um arquivo de configuração PKCS#11, por exemplo `pkcs11.cfg`:

```
name = SoftHSM2
library = /usr/lib/softhsm/libsofthsm2.so
slot = 0
```

### 5.2. Uso com SoftHSM2 (ambiente de desenvolvimento)

```bash
# Instalar SoftHSM2
sudo apt-get install softhsm2

# Inicializar token
softhsm2-util --init-token --slot 0 --label "test" \
  --pin 1234 --so-pin 5678

# Executar operação com PKCS#11
java -jar assinador.jar sign \
  --content "documento" \
  --pkcs11 --pkcs11-config pkcs11.cfg
```

### 5.3. Comportamento quando dispositivo indisponível

Quando o dispositivo PKCS#11 não está conectado ou o arquivo de configuração é inválido, o `assinador.jar` exibe um aviso e continua no modo simulado:

```
Aviso: dispositivo PKCS#11 não disponível. Operando no modo simulado.
```

---

## 6. Estrutura de Arquivos Gerenciados

O Sistema Runner armazena dados em `~/.hubsaude/`:

| Arquivo | Descrição |
|---------|-----------|
| `assinador.json` | PID e porta do `assinador.jar` em execução |
| `assinador.log` | Logs do `assinador.jar` |
| `simulador.json` | PID e porta do `simulador.jar` em execução |
| `simulador.jar` | JAR do Simulador (baixado automaticamente) |
| `simulador.version` | Versão do `simulador.jar` instalado |
| `jdk/` | JDK gerenciado (baixado automaticamente se necessário) |

---

## 7. Resolução de Problemas

### `assinador.jar não encontrado`

Defina a variável de ambiente `ASSINADOR_JAR` ou coloque o JAR no mesmo diretório do CLI:

```bash
export ASSINADOR_JAR=/caminho/para/assinador.jar
```

### `JDK não disponível`

O sistema tenta baixar automaticamente. Se falhar:
- Verifique a conexão com a internet
- Instale o JDK 21 manualmente e defina `JAVA_HOME`

### `Porta X ocupada`

Outro processo está usando a porta. Escolha uma porta diferente com `--port`, ou identifique o processo:

```bash
# Linux/macOS
lsof -i :8080

# Windows
netstat -ano | findstr :8080
```

### Servidor não responde após `start`

Verifique os logs:

```bash
cat ~/.hubsaude/assinador.log
```

---

## 8. Contexto FHIR e Escopo da Simulação

O Sistema Runner é uma **simulação** da integração com a Plataforma HubSaúde da SES-GO. Os parâmetros aceitos pelos comandos `sign` e `validate` mapeiam para os conceitos da [especificação FHIR de segurança](https://fhir.saude.go.gov.br/r4/seguranca/) da seguinte forma:

| Parâmetro CLI | Conceito FHIR | Observação |
|---------------|--------------|------------|
| `--content` | FHIR Bundle + Provenance (JSON) | Na simulação: qualquer texto não-vazio |
| `--token` | Identificador PKCS#11 (slot/token) | Opcional; não validado na simulação |
| `--signature` | JWS JSON Serialization (RFC 7515) | Na simulação: string fixa `MOCKED_SIGNATURE_BASE64_==` |

**O que está fora do escopo desta simulação** (conforme definido na especificação):
- Validação de cadeia de certificados X.509v3 / ICP-Brasil
- Timestamp de referência e políticas de assinatura
- Comunicação real com autoridades certificadoras (OCSP/CRL)
- Integração com TSA (Timestamping Authority)

Para o sistema real, consulte:
- [Caso de Uso: Criar Assinatura](https://fhir.saude.go.gov.br/r4/seguranca/caso-de-uso-criar-assinatura.html)
- [Caso de Uso: Validar Assinatura](https://fhir.saude.go.gov.br/r4/seguranca/caso-de-uso-validar-assinatura.html)

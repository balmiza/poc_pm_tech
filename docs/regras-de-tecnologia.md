# Regras de Tecnologia — Cadastro de Dispositivos

> Documentação técnica do serviço de cadastro de dispositivos (POC). Descreve
> a arquitetura, o modelo de dados, os contratos de cada rota e as decisões de
> implementação. Deve ser lida em conjunto com `docs/regras-de-negocio.md`, que
> descreve o domínio, e com `guia-desenvolvimento.md`, que descreve o fluxo de
> versionamento.

---

## Stack tecnológica

| Aspecto | Decisão |
|---------|---------|
| Linguagem | Go 1.22 |
| Servidor HTTP | `net/http` da biblioteca padrão (sem framework externo) |
| Roteamento | Method-based routing nativo do Go 1.22 (`METHOD /path/{param}`) |
| Serialização | `encoding/json` da biblioteca padrão |
| Persistência | Em memória (`map` + `sync.RWMutex`) — dados não sobrevivem ao reinício |
| Módulo Go | `github.com/balmiza/poc_pm_tech` |
| Dependências externas | Nenhuma (zero `go.sum`) |

---

## Estrutura de diretórios

```
api/
├── cmd/
│   └── server/
│       └── main.go          # Ponto de entrada; wiring das dependências
├── internal/
│   ├── domain/
│   │   ├── device.go        # Entidade Dispositivo, estados e máquina de estados
│   │   └── device_test.go
│   ├── repository/
│   │   ├── repository.go    # Interface DispositivoRepository
│   │   └── memory/
│   │       └── memory.go    # Implementação em memória
│   ├── service/
│   │   ├── service.go       # Casos de uso (regras de negócio)
│   │   └── service_test.go
│   ├── handler/
│   │   └── handler.go       # Camada HTTP: serialização, roteamento, códigos de status
│   └── biometry/
│       ├── biometry.go      # Interface Validador (contrato com API externa)
│       └── mock/
│           └── mock.go      # Mock configurável para testes e desenvolvimento
└── go.mod
```

---

## Arquitetura em camadas

O serviço segue uma arquitetura em camadas com dependências de fora para dentro:

```
handler  →  service  →  domain
                ↓
           repository
                ↓
           biometry (interface para API externa)
```

| Camada | Responsabilidade |
|--------|-----------------|
| `handler` | Recebe requisições HTTP, desserializa entrada, chama o service, serializa resposta e mapeia erros para status HTTP |
| `service` | Orquestra regras de negócio: valida pré-condições, chama repositório e biometria |
| `domain` | Entidade `Dispositivo`, máquina de estados, erros de domínio |
| `repository` | Interface de persistência; a implementação em memória é a única existente nesta POC |
| `biometry` | Interface para validação biométrica; apenas o mock está implementado |

Nenhuma camada interna conhece a camada externa. O `domain` não importa nenhum pacote do projeto.

---

## Modelo de dados

### Entidade `Dispositivo`

Única entidade persistida pelo serviço.

| Campo | Tipo Go | JSON | Descrição |
|-------|---------|------|-----------|
| `DispositivoID` | `string` | `dispositivo_id` | UUID v4 gerado aleatoriamente no momento da associação |
| `ClienteID` | `string` | `cliente_id` | Identificador do cliente proprietário (gerenciado por sistema externo) |
| `Estado` | `Estado` (string) | `estado` | Estado atual no ciclo de vida (`pendente_confirmacao`, `ativo`, `cancelado`) |
| `DataAssociacao` | `time.Time` | `data_associacao` | Timestamp UTC da criação do dispositivo |
| `DataHabilitacao` | `*time.Time` | `data_habilitacao` | Timestamp UTC da ativação; `null` enquanto não habilitado |
| `DataCancelamento` | `*time.Time` | `data_cancelamento` | Timestamp UTC do cancelamento; `null` enquanto não cancelado |

### Estados e transições

```
                    ┌─────────────────────┐
                    │  pendente_confirmacao│ ──── cancelar ────┐
                    └──────────┬──────────┘                    │
                               │ habilitar                     ▼
                               │ (biometria OK)          ┌──────────┐
                               ▼                         │ cancelado│
                         ┌──────────┐                    └──────────┘
                         │  ativo   │ ──── cancelar ─────────┘
                         └──────────┘
```

- Não existe transição de volta de nenhum estado.
- `cancelado` é terminal: para reusar o aparelho é necessária nova associação.
- Cancelar um dispositivo já `cancelado` é **idempotente** (retorna 200 sem alterar o registro).

---

## Variáveis de ambiente

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `PORT` | `8080` | Porta TCP em que o servidor escuta |
| `BIOMETRIA_RESULTADO` | _(ausente)_ | Controla o mock de biometria. `reprovado` faz toda chamada de biometria retornar reprovado; qualquer outro valor (ou ausência) aprova |

---

## Rotas da API

Base URL: `http://localhost:{PORT}`

Todos os bodies são `application/json`. Todos os timestamps seguem RFC 3339 em UTC.

---

### `GET /clientes/{clienteID}/autorizacao`

Verifica se o cliente possui um dispositivo autorizado (ativo).

**Path params**

| Param | Tipo | Descrição |
|-------|------|-----------|
| `clienteID` | string | Identificador do cliente |

**Resposta — sempre `200 OK`**

```json
{
  "cliente_id": "string",
  "autorizado": true,
  "dispositivo_id": "string"
}
```

> `dispositivo_id` é omitido quando `autorizado` é `false`.

| Cenário | `autorizado` | `dispositivo_id` |
|---------|-------------|-----------------|
| Cliente possui dispositivo `ativo` | `true` | ID do dispositivo ativo |
| Cliente sem dispositivo ativo | `false` | ausente |

---

### `POST /clientes/{clienteID}/dispositivos`

Associa um novo dispositivo ao cliente. Cria o registro em `pendente_confirmacao`.

**Path params**

| Param | Tipo | Descrição |
|-------|------|-----------|
| `clienteID` | string | Identificador do cliente |

**Body de entrada**

Nenhum body é necessário.

**Resposta — `201 Created`**

```json
{
  "dispositivo_id": "string",
  "cliente_id": "string",
  "estado": "pendente_confirmacao",
  "data_associacao": "2024-01-15T10:30:00Z"
}
```

**Erros**

| Status | Condição |
|--------|----------|
| `409 Conflict` | Cliente já possui dispositivo `ativo` ou `pendente_confirmacao` |

---

### `GET /dispositivos/{dispositivoID}`

Retorna o estado atual de um dispositivo.

**Path params**

| Param | Tipo | Descrição |
|-------|------|-----------|
| `dispositivoID` | string | UUID do dispositivo |

**Resposta — `200 OK`**

```json
{
  "dispositivo_id": "string",
  "cliente_id": "string",
  "estado": "ativo",
  "data_associacao": "2024-01-15T10:30:00Z",
  "data_habilitacao": "2024-01-15T10:35:00Z",
  "data_cancelamento": null
}
```

> `data_habilitacao` e `data_cancelamento` são omitidos quando `null`.

**Erros**

| Status | Condição |
|--------|----------|
| `404 Not Found` | Dispositivo não encontrado |

---

### `PATCH /dispositivos/{dispositivoID}/habilitar`

Valida a identidade do cliente via biometria facial e ativa o dispositivo.

**Path params**

| Param | Tipo | Descrição |
|-------|------|-----------|
| `dispositivoID` | string | UUID do dispositivo |

**Body de entrada**

Nenhum body é necessário.

**Resposta — `200 OK`**

```json
{
  "dispositivo_id": "string",
  "cliente_id": "string",
  "estado": "ativo",
  "data_associacao": "2024-01-15T10:30:00Z",
  "data_habilitacao": "2024-01-15T10:35:00Z"
}
```

**Erros**

| Status | Condição |
|--------|----------|
| `404 Not Found` | Dispositivo não encontrado |
| `409 Conflict` | Dispositivo não está em `pendente_confirmacao` |
| `422 Unprocessable Entity` | Biometria facial reprovada; dispositivo permanece em `pendente_confirmacao` |

---

### `PATCH /dispositivos/{dispositivoID}/cancelar`

Cancela um dispositivo (soft-delete). É idempotente.

**Path params**

| Param | Tipo | Descrição |
|-------|------|-----------|
| `dispositivoID` | string | UUID do dispositivo |

**Body de entrada**

Nenhum body é necessário.

**Resposta — `200 OK`**

```json
{
  "dispositivo_id": "string",
  "cliente_id": "string",
  "estado": "cancelado",
  "data_associacao": "2024-01-15T10:30:00Z",
  "data_cancelamento": "2024-01-15T11:00:00Z"
}
```

**Erros**

| Status | Condição |
|--------|----------|
| `404 Not Found` | Dispositivo não encontrado |

> Chamar este endpoint em um dispositivo já `cancelado` retorna `200` com o estado atual sem modificar o registro.

---

## Mapeamento de erros de domínio

| Erro de domínio | Status HTTP | Mensagem |
|-----------------|------------|---------|
| `ErrDispositivoNaoEncontrado` | `404` | `dispositivo não encontrado` |
| `ErrClienteJaPossuiDispositivoAtivoPendente` | `409` | `cliente já possui dispositivo ativo ou pendente de confirmação` |
| `ErrTransicaoDeEstadoInvalida` | `409` | `transição de estado inválida para o estado atual` |
| `ErrBiometriaReprovada` | `422` | `biometria facial reprovada` |
| Outros erros internos | `500` | mensagem do erro Go |

**Formato do body de erro (todos os status de erro)**

```json
{
  "erro": "mensagem descritiva do erro"
}
```

---

## Persistência

A implementação atual usa um `map[string]*Dispositivo` protegido por `sync.RWMutex`, sem banco de dados.

- **Leituras** usam `RLock` (concorrentes entre si).
- **Escritas** usam `Lock` (exclusivas).
- Cada operação retorna uma **cópia** da struct (`cp := *d`) para evitar que o chamador mute o estado interno do repositório.
- **Dados são perdidos ao reiniciar o servidor.** Isso é intencional nesta POC.

Para substituir a persistência por um banco real (Postgres, SQLite etc.), basta criar uma nova struct que implemente a interface `repository.DispositivoRepository` e injetá-la em `cmd/server/main.go`. Nenhuma outra camada precisa mudar.

---

## Integração de biometria

A validação biométrica é abstraída pela interface `biometry.Validador`:

```go
type Validador interface {
    ValidarBiometria(clienteID string) (bool, error)
}
```

- O serviço **não armazena** dados biométricos.
- O serviço **apenas recebe** `(true/false, error)` e decide com base nisso.
- A implementação real deve ser injetada em `cmd/server/main.go` no lugar do mock.

### Mock configurável

O mock (`internal/biometry/mock`) é controlado pela variável `BIOMETRIA_RESULTADO`:

| Valor | Comportamento |
|-------|--------------|
| `reprovado` | Toda chamada retorna reprovado |
| qualquer outro (ou ausente) | Toda chamada retorna aprovado |

---

## Como executar

```bash
cd api
go run ./cmd/server
```

Com biometria sempre reprovada:

```bash
cd api
BIOMETRIA_RESULTADO=reprovado go run ./cmd/server
```

Em porta customizada:

```bash
cd api
PORT=9000 go run ./cmd/server
```

---

## Como testar

```bash
cd api
go test ./...
```

Os testes de `internal/domain` e `internal/service` são unitários e não têm dependências externas.

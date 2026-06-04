# poc_pm_tech — Serviço de Cadastro de Dispositivos

API REST em Go para gerenciar o ciclo de vida de dispositivos (celulares) de clientes, com validação de identidade por biometria facial.

## Pré-requisitos

- Go 1.22 ou superior — download em https://go.dev/dl/
- Git

## Como rodar

```bash
# Clone o repositório
git clone https://github.com/balmiza/poc_pm_tech.git
cd poc_pm_tech

# Baixe as dependências (nenhuma externa por enquanto)
go mod tidy

# Inicie o servidor (porta padrão: 8080)
go run ./cmd/server/...

# Para forçar rejeição da biometria (útil para testes manuais):
BIOMETRIA_RESULTADO=reprovado go run ./cmd/server/...

# Para mudar a porta:
PORT=9090 go run ./cmd/server/...
```

## Como testar

```bash
# Roda todos os testes
go test ./...

# Com saída detalhada
go test -v ./...

# Apenas testes de domínio
go test -v ./internal/domain/...

# Apenas testes de serviço
go test -v ./internal/service/...
```

## Endpoints

| Método | Caminho | Descrição |
|--------|---------|-----------|
| `GET` | `/clientes/{clienteID}/autorizacao` | Consulta se o cliente tem dispositivo autorizado |
| `POST` | `/clientes/{clienteID}/dispositivos` | Associa um novo dispositivo ao cliente |
| `GET` | `/dispositivos/{dispositivoID}` | Consulta o estado atual de um dispositivo |
| `PATCH` | `/dispositivos/{dispositivoID}/habilitar` | Habilita o dispositivo (exige biometria aprovada) |
| `PATCH` | `/dispositivos/{dispositivoID}/cancelar` | Cancela o dispositivo (soft-delete) |

### Exemplos com curl

```bash
# Associar dispositivo ao cliente "c123"
curl -X POST http://localhost:8080/clientes/c123/dispositivos

# Consultar autorização
curl http://localhost:8080/clientes/c123/autorizacao

# Habilitar dispositivo (substitua pelo ID retornado na associação)
curl -X PATCH http://localhost:8080/dispositivos/<dispositivo_id>/habilitar

# Cancelar dispositivo
curl -X PATCH http://localhost:8080/dispositivos/<dispositivo_id>/cancelar

# Consultar estado do dispositivo
curl http://localhost:8080/dispositivos/<dispositivo_id>
```

### Códigos HTTP

| Código | Situação |
|--------|----------|
| `200` | Operação bem-sucedida |
| `201` | Dispositivo associado com sucesso |
| `404` | Dispositivo não encontrado |
| `409 Conflict` | Violação de regra: cliente já possui dispositivo ativo/pendente, ou transição de estado inválida |
| `422 Unprocessable Entity` | Biometria reprovada — dispositivo permanece em `pendente_confirmacao` |
| `500` | Erro interno |

## Decisões técnicas

### Linguagem e versão
Go 1.22. A partir dessa versão o `net/http` padrão suporta roteamento com método HTTP e path parameters (`"GET /path/{id}"`), eliminando a necessidade de um roteador externo.

### Persistência: repositório em memória
Adotado para a POC por não exigir nenhuma dependência externa (sem Docker, sem banco instalado). Os dados são perdidos ao reiniciar o servidor, o que é aceitável para validação do domínio.

A interface `DispositivoRepository` está desacoplada da implementação, portanto basta criar `internal/repository/postgres/` (ou SQLite) e trocar a injeção em `main.go` quando houver necessidade de persistência real.

### Dependências externas
Nenhuma. Apenas biblioteca padrão do Go. UUID gerado com `crypto/rand` para evitar `go.sum` vazio e simplificar o setup inicial.

### Biometria: mock configurável
A integração com a API de biometria facial é abstraída pela interface `biometry.Validador` (`internal/biometry/biometry.go`). O servidor usa sempre a implementação mock (`internal/biometry/mock/`), controlável via variável de ambiente:

- `BIOMETRIA_RESULTADO=aprovado` (padrão) — toda chamada aprova
- `BIOMETRIA_RESULTADO=reprovado` — toda chamada rejeita

Para conectar a implementação real, basta criar um struct que implemente a interface e injetá-lo em `main.go`.

### Estrutura de camadas

```
cmd/server/        → ponto de entrada (main, wiring de dependências)
internal/
  domain/          → entidade Dispositivo, estados, transições, erros de domínio
  repository/      → interface + implementação em memória
  biometry/        → interface + mock
  service/         → regras de negócio orquestradas
  handler/         → HTTP handlers (encode/decode JSON, mapeamento de erros)
```

### Convenção de idioma
Código e identificadores em inglês técnico; comentários, mensagens de erro e nomes de domínio em português (refletindo a linguagem ubíqua do negócio — `clienteID`, `EstadoAtivo`, `ErrBiometriaReprovada`).

## Pendências e próximos passos

- [ ] **Push para o remoto:** execute os comandos abaixo após configurar autenticação no GitHub:
  ```bash
  git push -u origin main
  git push -u origin develop
  git push -u origin feature/implementar-api-dispositivos
  ```
- [ ] **Persistência real:** implementar `internal/repository/postgres/` ou `sqlite/` conforme necessidade.
- [ ] **Implementação real de biometria:** criar client HTTP para a API externa e injetá-lo em `main.go`.
- [ ] **Autenticação/autorização:** os endpoints atualmente não exigem autenticação.
- [ ] **Documentação OpenAPI:** gerar especificação `api/openapi.yaml`.
- [ ] **Arquivo `docs/architecture.md`:** detalhamento do ambiente de hospedagem, conforme previsto no `guia-desenvolvimento.md`.
- [ ] **Testes de integração HTTP:** testes de ponta a ponta dos handlers via `httptest`.

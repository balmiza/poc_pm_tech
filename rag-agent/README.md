# rag-agent

Pipeline de RAG (Retrieval-Augmented Generation) para indexar a documentação e o
código da POC de cadastro de dispositivos. Usa embeddings locais via Ollama e
armazena vetores no Postgres com a extensão pgvector.

Esta pasta é totalmente independente do código Go em `api/`.

---

## Pré-requisitos

### 1. Docker Desktop

Necessário para subir o Postgres com pgvector.

- Download: https://www.docker.com/products/docker-desktop
- Após instalar, certifique-se de que o Docker Desktop está rodando antes de
  executar os comandos abaixo.

### 2. Ollama

Necessário para gerar embeddings e rodar o LLM localmente.

- Download: https://ollama.com/download
- Após instalar, baixe os modelos necessários:

```bash
ollama pull nomic-embed-text
ollama pull llama3.1
```

- Verifique que o serviço está rodando: `ollama list`

### 3. Python 3.12+

Já disponível na máquina.

---

## Subindo o ambiente

### 1. Ambiente Python

```bash
# Dentro de rag-agent/
python -m venv .venv
.venv\Scripts\activate        # Windows
# source .venv/bin/activate   # Linux/Mac

pip install -r requirements.txt
```

> O venv já foi criado e as dependências já estão instaladas nesta máquina.
> Basta ativar com `.venv\Scripts\activate` antes de rodar qualquer script.

### 2. Postgres com pgvector

```bash
# Dentro de rag-agent/
docker compose up -d
```

Isso sobe um container Postgres 16 com a extensão pgvector disponível.

Credenciais:

| Parâmetro | Valor |
|-----------|-------|
| Host | `localhost` |
| Porta | `5432` |
| Banco | `rag_poc` |
| Usuário | `rag` |
| Senha | `rag_secret` |

### 3. Validar a conexão

```bash
python validate_connection.py
```

Saída esperada: `OK — pgvector <versão> disponível em localhost:5432`

---

## Estrutura da pasta

```
rag-agent/
├── docker-compose.yml      # Postgres 16 + pgvector via Docker
├── requirements.txt        # Dependências Python (llama-index + integrações)
├── validate_connection.py  # Script de validação da conexão com o banco
├── .gitignore              # Exclui .venv e __pycache__ do versionamento
└── README.md               # Este arquivo
```

---

## Próximo passo

Com o ambiente validado, a **etapa seguinte** é implementar o pipeline de
ingestão, que irá:

1. Ler os arquivos de documentação (`docs/`) e código fonte (`api/`) do
   repositório.
2. Gerar embeddings de cada trecho usando o modelo `nomic-embed-text` via
   Ollama.
3. Armazenar os vetores no Postgres (tabela gerenciada pelo llama-index via
   `PGVectorStore`).

O pipeline de ingestão será criado em uma nova branch de feature a partir de
`develop`, seguindo o gitflow definido em `guia-desenvolvimento.md`.

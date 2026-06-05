"""
Valida que o Postgres com pgvector está acessível.
Execute após `docker compose up -d` dentro de rag-agent/.
"""
import sys

try:
    import psycopg2
except ImportError:
    print("ERRO: psycopg2 não encontrado. Ative o venv e instale as dependências.")
    sys.exit(1)

DSN = "host=localhost port=5432 dbname=rag_poc user=rag password=rag_secret"

try:
    conn = psycopg2.connect(DSN)
    cur = conn.cursor()
    cur.execute("CREATE EXTENSION IF NOT EXISTS vector;")
    cur.execute("SELECT extversion FROM pg_extension WHERE extname = 'vector';")
    row = cur.fetchone()
    conn.commit()
    cur.close()
    conn.close()
    print(f"OK — pgvector {row[0]} disponível em localhost:5432")
except Exception as e:
    print(f"FALHA — {e}")
    sys.exit(1)

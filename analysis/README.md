# analysis

Análise em Python: base normalizada (DuckDB/Parquet), chuva x alagamento (etapa 4) e modelo (etapa 5).

```bash
cd analysis
uv sync
uv run jupyter lab
```

Padronizar os endereços (gera `data/processed/ocorrencias_padronizadas.csv`):

```bash
uv run python -m sp_alagamentos.padronizar
uv run pytest
```

Notebooks em `notebooks/`, com prefixo numérico (`01-exploracao.ipynb`, `02-limiar-chuva.ipynb`…).

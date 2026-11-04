"""Caminhos compartilhados entre o coletor (Go) e a análise (Python)."""

from pathlib import Path

RAIZ = Path(__file__).resolve().parents[3]
DATA = RAIZ / "data"
RAW = DATA / "raw"  # HTML bruto (fora do git)
PROCESSED = DATA / "processed"  # ocorrências normalizadas (Parquet)
EXPORT = DATA / "export"  # JSON estático consumido pelo site gor.dev.br

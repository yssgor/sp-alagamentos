# Infraestrutura — quanto dado é?

Estimativa de 03/11/2026, com base na investigação do CGE (`docs/fontes/cge.md`).
Tamanho das páginas medido em 04/11/2026 pelo workflow `teste-cge`: **12 KB** num dia sem alagamento e **90 KB** num dia com 129 pontos (10/02/2020).

## Resumo

**Pouco dado.** O histórico completo de alagamentos cabe em alguns MB depois de normalizado. O único volume relevante é o HTML bruto do backfill (~dezenas de MB comprimido) e, depois, a chuva do CEMADEN (~centenas de MB). Tudo roda num notebook comum; não há necessidade de servidor potente.

## Alagamentos (CGE)

| Item | Conta | Ordem de grandeza |
|---|---|---|
| Dias de histórico (2005 → hoje) | ~22 anos × 365 | ~8.000 páginas |
| HTML por página | 12 KB (dia vazio) a ~90 KB (dia com 100+ pontos) | — |
| Backfill bruto | ~7.200 dias vazios × 12 KB + ~800 dias com alagamento × ~40 KB | ~120 MB sem compressão, **~20–30 MB em gzip** |
| Tempo de backfill educado | 8.000 req × 5 s | ~11 h (dividir em várias noites) |
| Ocorrências normalizadas | ~80 dias com alagamento/ano × ~20 pontos × 22 anos | **~35–50 mil linhas → < 5 MB em Parquet** |
| Coleta diária | 1 página/dia | ~30 MB/ano bruto, ~3 MB gzip |
| Coleta "ao vivo" a cada 10 min (opcional) | 144 páginas/dia; guardar só quando muda | poucos MB por temporada |

## Chuva (CEMADEN) — etapa 4

| Item | Conta | Ordem de grandeza |
|---|---|---|
| Pluviômetros na cidade | ~100 (a confirmar) | — |
| Linhas horárias 2013–2024 | 100 × 24 × 365 × 12 | ~10 milhões |
| Tamanho | Parquet comprimido | ~100–200 MB |

DuckDB processa isso em segundos num notebook.

## O que isso significa para a decisão "Actions ou VPS"

- O trabalho recorrente é **1 requisição por dia** (ou 144, se houver "ao vivo"). GitHub Actions resolve com folga e de graça.
- O backfill é uma tarefa única: roda no seu computador, em algumas noites.
- VPS só se justifica se quisermos "ao vivo" com horário garantido.
- Onde guardar:
  - base normalizada e JSON do site: **no próprio repositório** (poucos MB);
  - HTML bruto: **fora do git** — Cloudflare R2 (10 GB grátis) ou assets de release do GitHub; localmente em `data/raw/`.

Testado em 04/11/2026: o CGE responde normalmente aos runners do GitHub Actions.

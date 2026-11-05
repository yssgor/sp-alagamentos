# SP Alagamentos

Histórico aberto dos alagamentos da cidade de São Paulo — onde, quando e com quanta chuva a cidade alaga.

O CGE mostra os pontos de alagamento dia a dia, mas não há uma base histórica fácil de analisar. Este projeto coleta essas páginas desde 2005, normaliza as ocorrências e, depois, cruza com a chuva medida pelo CEMADEN.

> **Não é um sistema de alerta.** Para situações de emergência, consulte o [CGE](https://www.cgesp.org) e a Defesa Civil (199).

## Estrutura

```
collector/   coletor em Go — baixa as páginas do CGE e guarda o HTML bruto
  cmd/coletor/        CLI: `dia` e `backfill`
  internal/cge/       cliente HTTP do CGE
  internal/armazem/   snapshots brutos (gzip, idempotente)
  internal/parser/    HTML → ocorrências (etapa 2)
  testdata/           snapshots reais para testar o parser
analysis/    análise em Python (DuckDB, Polars, notebooks)
data/        raw/ (fora do git), processed/ (Parquet), export/ (JSON do site)
docs/        investigação das fontes e decisões de infra
```

## Uso

```bash
# baixar o dia anterior
cd collector
go run ./cmd/coletor dia

# baixar um período (1 requisição a cada 5 s)
go run ./cmd/coletor backfill -de 2020-02-01 -ate 2020-02-29

# transformar o HTML bruto numa tabela (CSV) de ocorrências
go run ./cmd/coletor processar

go test ./...
```

## Documentos

- [PLANO.md](PLANO.md) — objetivo, escopo, etapas e riscos
- [DECISOES.md](DECISOES.md) — registro de decisões
- [docs/fontes/cge.md](docs/fontes/cge.md) — como a página do CGE funciona
- [docs/infra.md](docs/infra.md) — volume de dados e necessidade de infraestrutura

## Fontes e licença

- Alagamentos: CGE – Centro de Gerenciamento de Emergências Climáticas, Prefeitura de São Paulo.
- Chuva: DATA FROM THE CEMADEN/MCTIC OBSERVATIONAL NETWORK.
- Código sob licença [MIT](LICENSE). Os dados pertencem às respectivas fontes.

# data

| Pasta | Conteúdo | No git? |
|---|---|---|
| `raw/cge/AAAA/MM/DD/*.html.gz` | HTML do CGE exatamente como veio, um arquivo por versão diferente | **não** na `main`; o Actions grava na branch `dados` |
| `processed/` | ocorrências normalizadas em Parquet | sim (poucos MB) |
| `export/` | JSON estático para a página do site | na branch `dados` |

Fontes: CGE – Prefeitura de São Paulo; CEMADEN ("DATA FROM THE CEMADEN/MCTIC OBSERVATIONAL NETWORK").

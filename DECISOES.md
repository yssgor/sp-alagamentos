# Decisões — SP Alagamentos

Registro das decisões tomadas nas discussões do projeto. Uma linha por decisão, a mais recente no topo.

| Data | Decisão | Motivo |
|---|---|---|
| 04/11/2026 | MVP roda no GitHub Actions (cron) gravando JSON no repo; página lê o JSON direto do GitHub. Sem VPS | R$ 0; atraso de 10–30 min aceitável com aviso de "não é alerta oficial". Worker da Cloudflare fica como opção futura para ao vivo de verdade |
| 03/11/2026 | Repositório público desde o início, código sob MIT | Portfólio; dados são públicos e sem informação pessoal |
| 03/11/2026 | Monorepo: coletor e parser em Go (`collector/`), análise em Python (`analysis/`) | Go para o coletor resiliente; Python/DuckDB para a ciência de dados |
| 03/11/2026 | Coleta principal = 1 página por dia (dia anterior) + backfill desde 2005; "ao vivo" fica opcional | O CGE permite consultar datas passadas e cada ponto já traz início e fim (ver docs/fontes/cge.md) |
| 03/11/2026 | HTML bruto fora do git (data/raw/); Parquet e JSON do site dentro do repo | Volume pequeno na base normalizada; bruto chega a dezenas de MB (ver docs/infra.md) |
| 03/11/2026 | Projeto escolhido para o portfólio, desenvolvido separado do de ônibus | Dados públicos sem informação pessoal, custo baixo, temporada de chuva começando |

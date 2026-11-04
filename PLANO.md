# SP Alagamentos — histórico de alagamentos da cidade de São Paulo

**Status:** etapa 0 (investigação) em andamento · criado em 03/11/2026
**Projeto irmão:** [sp-onibus](../sp-onibus/PLANO.md)

## 1. Pergunta central

Onde, quando e com quanta chuva São Paulo alaga?

O CGE mostra os pontos de alagamento do dia, mas não há uma base histórica fácil de analisar. Este projeto constrói essa base e, depois, cruza com a chuva medida para responder "quantos milímetros de chuva alagam esta rua?".

## 2. Por que este projeto

- **Lacuna real:** a informação existe, mas espalhada no tempo; ninguém guarda o histórico de forma aberta.
- **Portfólio:** engenharia (coletor resiliente em Go), ciência de dados (chuva x alagamento, modelo com avaliação honesta) e produto (mapa e ranking no gor.dev.br).
- **Público:** moradores, jornalistas, pesquisadores de drenagem urbana, gestores públicos.
- **Sem dados pessoais:** sem risco de LGPD.
- **Timing:** a temporada de chuvas (out–mar) está começando.

## 3. Fontes de dados

| Fonte | O que traz | Acesso | Status |
|---|---|---|---|
| [CGE — página de alagamentos](https://www.cgesp.org/v3/alagamentos.jsp) | Pontos ativos/inativos, transitáveis/intransitáveis, por data | Página HTML, sem API; consulta por data via `?dataBusca=DD/MM/AAAA` | **Verificado:** histórico desde pelo menos 2005; campos e robots em [docs/fontes/cge.md](docs/fontes/cge.md). O RSS é de notícias, não de pontos |
| [CEMADEN — PED](https://plataforma-clima-ambiente-readthedocs.readthedocs.io/en/latest/cemaden_data.html) | Chuva por pluviômetro (horária e diária), série processada 2013–2024 + tempo real | Cadastro com login; token temporário; 12 req/min; CSV por estação | Verificado na documentação. Citar "DATA FROM THE CEMADEN/MCTIC OBSERVATIONAL NETWORK" |
| Estações do próprio CGE | Chuva medida pela rede do CGE | A investigar | A confirmar |
| GeoSampa | Limites de subprefeituras/distritos, rios e córregos | Download | A confirmar quais camadas |

## 4. Escopo

**MVP**
- Coletor que grava os pontos de alagamento periodicamente (e faz backfill, se o histórico do CGE permitir).
- Base normalizada: ponto, endereço, coordenada, início, fim, status (transitável/intransitável).
- Página no gor.dev.br: mapa dos pontos, ranking dos que mais alagam, recortes por temporada e subprefeitura.

**Depois**
- Cruzamento com chuva (CEMADEN): limiar de chuva por ponto.
- Modelo de probabilidade de alagar dado a chuva das últimas horas, com backtesting público.
- Cruzamento com o projeto de ônibus: quanto a chuva atrasa as linhas.
- Artigos no site.

## 5. Arquitetura proposta (rascunho)

```
Coletor Go (a cada 5–10 min)
  → snapshot bruto (HTML/RSS salvo como veio)
  → parser → base normalizada (Parquet / DuckDB)
  → agregações → JSON estático
  → página no Astro (gor.dev.br, Cloudflare Pages)
```

Princípios:
- **Guardar sempre o bruto**, para reprocessar quando o parser mudar.
- **Idempotência:** rodar duas vezes não duplica registros.
- **Alarme de silêncio:** zero registros num dia de chuva forte é suspeito → avisar.

Onde roda o coletor (decisão pendente):
- GitHub Actions com cron — grátis, mas sem garantia de horário exato.
- VPS pequena — ~R$25/mês, mais confiável.

## 6. Etapas

| # | Etapa | Entregável | Estimativa |
|---|---|---|---|
| 0 | Investigação | Respostas da seção 9 (histórico, campos, termos); cadastro no CEMADEN; decisão de onde roda | 1 semana |
| 1 | Coletor + armazenamento bruto | Coleta rodando sem parar | 1 semana |
| 2 | Parser + base normalizada + geocodificação | Tabela limpa de ocorrências | 1 semana |
| 3 | Página MVP no site | Mapa + ranking publicados | 1 semana |
| 4 | Chuva x alagamento | Análise de limiar por ponto | 2 semanas |
| 5 | Modelo + artigo | Modelo com backtesting e artigo explicando | 2+ semanas |

Estimativas em meio período. MVP (etapas 0–3): 3 a 4 semanas.

## 7. Riscos e mitigação

| Risco | Mitigação |
|---|---|
| Site do CGE muda o layout e quebra o coletor | Guardar o bruto; testes do parser com snapshots reais; alarme de silêncio |
| Termos de uso não verificados | Ler termos/robots; coleta educada (baixa frequência, identificação no user-agent); citar a fonte; se preciso, pedir os dados via Lei de Acesso à Informação |
| Histórico do CGE não consultável | Base cresce por temporada; pedir o histórico ao CGE via LAI |
| Pessoas usarem o mapa como alerta em emergência | Aviso visível de que não é alerta oficial, link para o CGE, horário da última atualização; MVP focado no histórico, não no "ao vivo" |
| Geocodificação imprecisa | Revisar manualmente os pontos recorrentes (devem ser poucas centenas) |
| Sazonalidade | Começar já, aproveitando a temporada atual |

## 8. Custos

- **Tempo:** MVP em 3–4 semanas; com modelo e artigo, 6–8 semanas (meio período).
- **Infra:** R$0 (GitHub Actions + Cloudflare Pages) a ~R$25/mês (VPS).

## 9. Perguntas em aberto

1. ~~O histórico do CGE permite consultar datas passadas? Desde quando?~~ Sim, desde pelo menos 2005 (falta achar a data mais antiga).
2. ~~Que campos cada ponto de alagamento tem?~~ Início/fim, local, sentido, referência, zona/subprefeitura; status só por ícone; sem ID nem coordenada.
3. ~~O que dizem os termos de uso do site do CGE?~~ Sem página de termos; `robots.txt` libera `/v3/`.
4. ~~Onde roda o coletor: GitHub Actions ou VPS?~~ GitHub Actions (ver DECISOES.md e [docs/infra.md](docs/infra.md)).
5. Mostrar "ao vivo" no site ou só o histórico?
6. Nome e URL no site (ex.: `gor.dev.br/projetos/alagamentos`).
7. ~~Repositório público desde o início?~~ Sim.

## 10. Similares

- O próprio mapa do CGE — só o dia corrente.
- [Serviço da Maplink com dados do CGE](https://catracalivre.com.br/criatividade/site-gratuito-localiza-pontos-de-alagamento-em-sp-e-sugere-rotas-alternativas/) — de 2013, status atual incerto.
- Waze — relatos de usuários, sem base histórica aberta.
- Estudos acadêmicos sobre alagamentos em SP com dados do CGE (ex.: [UFPE](https://www3.ufpe.br/visimgeo/includes/Anais_final_10_09_16/artigos/Todos_Artigos/026_2.pdf), [Redalyc](https://www.redalyc.org/pdf/3939/393937733006.pdf)).

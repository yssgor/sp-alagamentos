# Fonte: CGE — página de alagamentos

Investigação feita em 03/11/2026 (etapa 0). Responde às perguntas 1–3 da seção 9 do PLANO.

## Endereços

| O quê | URL |
|---|---|
| Alagamentos do dia | `https://www.cgesp.org/v3/alagamentos.jsp` |
| Alagamentos de uma data | `https://www.cgesp.org/v3/alagamentos.jsp?dataBusca=DD%2FMM%2FAAAA` |
| Classificação dos pontos | `https://www.cgesp.org/v3/alagamentos-classificacao.jsp` |
| Feed RSS | `https://www.cgesp.org/v3/feed_rss.jsp` — **são notícias/boletins, não pontos de alagamento** (15 itens, `noticias.jsp?id=N`) |

## 1. Dá para consultar datas passadas? Desde quando?

**Sim**, pelo parâmetro `dataBusca` (data em `dd/mm/aaaa`, barras codificadas como `%2F`).

| Data testada | Pontos | Observação |
|---|---|---|
| 25/05/2005 | 100 (39 transitáveis, 61 intransitáveis) | o histórico chega pelo menos a 2005 |
| 08/12/2009 | 124 | |
| 10/02/2020 | 129 (46 transitáveis, 83 intransitáveis) | |
| 01/11/2026 | 0 | houve "estado de atenção", mas nenhum ponto registrado |

Dia sem ocorrência mostra o texto exato: `Não há registros de alagamentos para essa data.`

**Consequência importante:** como cada ponto já traz horário de início e fim, uma consulta por dia (no dia seguinte) basta para ter o histórico completo. A coleta a cada 5–10 min só é necessária se quisermos mostrar "ao vivo" ou detectar edições posteriores do CGE.

## 2. Campos de cada ponto

Agrupados por **zona** (Norte, Centro, Oeste, Leste, Sul, Sudeste) e **subprefeitura** (ex.: Casa Verde, Lapa, Sé).

Exemplo real (10/02/2020, Casa Verde):

```
De 03:25 a 09:47
MARGINAL TIETE ASCB EXPRESSA (CASA VERDE)
Sentido: A SENNA/CASTELO B - EXP
Referência: PTE DO LIMAO-ADHEMAR F. DA SILVA
```

| Campo | Presente? | Nota |
|---|---|---|
| Horário início / fim | sim | "De HH:MM a HH:MM"; fim pode faltar se ainda ativo |
| Local (endereço) | sim | texto livre, abreviado (AV., R, PTE, ASCB…) |
| Sentido | sim | ex.: "BAIRRO/CENTRO", "AMBOS" |
| Referência | sim | cruzamento ou ponto de referência |
| Zona / subprefeitura | sim | pelo agrupamento da tabela |
| Transitável / intransitável | sim | classe CSS do `<li>` de status |
| Ativo / inativo | sim | idem |
| ID, coordenada, link | **não** | precisamos criar um ID estável e geocodificar |

Estrutura confirmada com o HTML real (implementada em `collector/internal/parser`):
- zona: `<h1 class="tit-bairros">Zona Norte</h1>`
- subprefeitura: `<td class="bairro ...">Casa Verde <hr/></td>`
- ponto: `<div class="ponto-de-alagamento">`, contendo:
  - status como classe CSS: `<li class="inativo-intransitavel">` (`ativo|inativo` + `transitavel|intransitavel`)
  - `<li class="arial-descr-alag col-local">De 03:25 a 09:47<br/>LOCAL</li>`
  - `<li class="arial-descr-alag">Sentido: ...<br/>Referência: ...</li>`
- resumo no fim: `<h1>Transitáveis = 46 pts</h1>` / `<h1>Intransitáveis = 83 pts</h1>`, usado pelo parser para conferir a contagem.
- Página em UTF-8.
- Ponto sem horário de fim aparece como `De 14:31 a` (com o "a" sobrando) e com status `ativo-*`. Isso vale também para dias antigos: são pontos que o CGE **nunca encerrou** no registro (342 de 19.550 no histórico). Na análise, tratar "sem fim" como dado faltante, não como alagamento em curso.
- A página também muda conforme o tempo/horário (imagem de fundo `CC_dia`, `CH_noite`…), então pode haver uma versão nova do HTML sem nenhuma mudança nos pontos.

## 3. Termos de uso

- `robots.txt`: **libera** `/v3/` para todos os robôs (bloqueia caminhos que começam com `/c` e `/p`, e o SemrushBot por inteiro).
- Não foi encontrada página de termos de uso. Seguimos coleta educada: 1 requisição por vez, intervalo de vários segundos, user-agent identificado, citar "Fonte: CGE – Prefeitura de São Paulo".

## Pendências

- [x] Snapshots reais em `collector/testdata/`: dia cheio (10/02/2020), dia vazio e dia com ponto ativo (14/09/2026).
- [x] Medir o tamanho real: 12 KB (dia vazio), 90 KB (10/02/2020, 129 pontos).
- [ ] Confirmar a data mais antiga com dados: o primeiro ponto do backfill é de 03/01/2005; falta testar 2004 e antes.
- [x] O site responde aos runners do GitHub Actions (teste de 04/11/2026). A contagem de `div.ponto-de-alagamento` bateu com os 129 pontos de 10/02/2020, então esse seletor continua válido.

## Primeiro processamento do histórico (05/11/2026)

- 7.965 dias processados (01/01/2005 a 05/11/2026), 0 erros de parsing.
- 19.550 ocorrências em 1.893 dias com alagamento; 34% intransitáveis.
- Duração mediana de 87 min; 90% terminam em até ~6,5 h.
- Maior dia: 10/02/2020 (129 pontos).
- 14.395 combinações distintas de local + sentido + referência, mas só 4.028 textos de local: o mesmo ponto aparece com grafias diferentes (`AV.` x `AV`, `(CASA VERDE)` no fim…). **Normalizar antes de qualquer ranking.**
- Faltaram os dias 31/12 de 2005 a 2018 (bug de horário de verão no laço do backfill, corrigido); recoletar com o workflow `coleta` e a lista de datas.

## Referências

- Scraper anterior da SMDU/Prefeitura: https://github.com/geoinfo-smdu/cgesp_scrap

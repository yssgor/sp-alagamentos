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
| Transitável / intransitável | **só por ícone** | o parser precisa ler o `src`/classe da imagem de cada ponto |
| Ativo / inativo | sim | idem, por ícone/legenda |
| ID, coordenada, link | **não** | precisamos criar um ID estável e geocodificar |

Seletores HTML usados por um scraper anterior (a confirmar com snapshot atual):
- bloco do ponto: `div.ponto-de-alagamento`
- local e horário: `li.arial-descr-alag.col-local`
- sentido/referência: `li` com o texto `Referência:`
- subprefeitura: primeiro `td` da tabela que envolve o bloco

## 3. Termos de uso

- `robots.txt`: **libera** `/v3/` para todos os robôs (bloqueia caminhos que começam com `/c` e `/p`, e o SemrushBot por inteiro).
- Não foi encontrada página de termos de uso. Seguimos coleta educada: 1 requisição por vez, intervalo de vários segundos, user-agent identificado, citar "Fonte: CGE – Prefeitura de São Paulo".

## Pendências

- [ ] Salvar snapshots HTML reais em `collector/testdata/` (dia com muitos pontos, dia vazio, dia com ponto ativo) — o proxy desta sessão bloqueou o acesso direto; baixar pelo navegador ou pelo coletor.
- [x] Medir o tamanho real: 12 KB (dia vazio), 90 KB (10/02/2020, 129 pontos).
- [ ] Confirmar a data mais antiga com dados.
- [x] O site responde aos runners do GitHub Actions (teste de 04/11/2026). A contagem de `div.ponto-de-alagamento` bateu com os 129 pontos de 10/02/2020, então esse seletor continua válido.

## Referências

- Scraper anterior da SMDU/Prefeitura: https://github.com/geoinfo-smdu/cgesp_scrap

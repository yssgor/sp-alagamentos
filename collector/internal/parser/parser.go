// Package parser transforma o HTML bruto do CGE em ocorrências normalizadas.
//
// Estrutura da página (ver docs/fontes/cge.md):
//
//	<h1 class="tit-bairros">Zona Norte</h1>
//	<table class="tb-pontos-de-alagamentos">
//	  <td class="bairro ...">Casa Verde <hr/></td>
//	  <div class="ponto-de-alagamento">
//	    <li class="inativo-intransitavel" ...></li>
//	    <li class="arial-descr-alag col-local">De 03:25 a 09:47<br/>MARGINAL TIETE ...</li>
//	    <li class="arial-descr-alag">Sentido: ...<br/>Referência: ...</li>
//	  </div>
//	...
//	<div class="alag-resumo"> <h1>Transitáveis = 46 pts</h1> <h1>Intransitáveis = 83 pts</h1> ...
//
// O parser usa só a biblioteca padrão. Qualquer formato inesperado gera erro,
// para que mudanças no site do CGE sejam percebidas em vez de virar dado errado.
package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SemRegistros é o texto exibido pelo CGE em dias sem alagamento.
const SemRegistros = "Não há registros de alagamentos para essa data."

// Ocorrencia é um ponto de alagamento num dia.
type Ocorrencia struct {
	ID            string     // estável: hash de data + local + sentido + referência + início
	Data          time.Time  // dia consultado (00:00, horário de SP)
	Inicio        time.Time  // horário de início
	Fim           *time.Time // nil se o CGE não registrou o fim (ponto ativo ou nunca encerrado)
	Zona          string     // Zona Norte, Centro, Zona Oeste, Zona Leste, Zona Sul, Zona Sudeste
	Subprefeitura string     // ex.: Casa Verde
	Local         string     // endereço como veio, ex.: "AV. BRAZ LEME (CASA VERDE)"
	Sentido       string     // ex.: "BAIRRO/CENTRO", "AMBOS"
	Referencia    string     // cruzamento ou ponto de referência
	Intransitavel bool
	Ativo         bool
}

var (
	// marcadores, na ordem em que aparecem no documento
	reMarcador = regexp.MustCompile(`(?s)<h1 class="tit-bairros">(.*?)</h1>` +
		`|<td class="bairro[^"]*">(.*?)<hr` +
		`|<div class="ponto-de-alagamento">(.*?)</div>`)

	reStatus  = regexp.MustCompile(`<li class="(ativo|inativo)-(transitavel|intransitavel)"`)
	reLocal   = regexp.MustCompile(`(?s)col-local">(.*?)<br\s*/?>(.*?)</li>`)
	reSentido = regexp.MustCompile(`(?s)Sentido:(.*?)<br\s*/?>\s*Referência:(.*?)</li>`)
	// "De 03:25 a 09:47" (encerrado) ou "De 14:31 a" (sem fim registrado)
	reHorario = regexp.MustCompile(`^De (\d{2}):(\d{2})(?: a(?: (\d{2}):(\d{2}))?)?$`)
	reResumo  = regexp.MustCompile(`<h1>(?:Transitáveis|Intransitáveis) = (\d+) pts</h1>`)
)

// Parse lê o HTML de um dia e devolve as ocorrências.
// dia deve estar em 00:00 no fuso de São Paulo.
func Parse(dia time.Time, pagina []byte) ([]Ocorrencia, error) {
	doc := string(pagina)

	var (
		ocorrencias         []Ocorrencia
		zona, subprefeitura string
	)
	for _, m := range reMarcador.FindAllStringSubmatch(doc, -1) {
		switch {
		case m[1] != "":
			zona = limpar(m[1])
		case m[2] != "":
			subprefeitura = limpar(m[2])
		default:
			o, err := parsePonto(dia, m[3])
			if err != nil {
				return nil, fmt.Errorf("%s, ponto %d (%s/%s): %w",
					dia.Format("2006-01-02"), len(ocorrencias)+1, zona, subprefeitura, err)
			}
			o.Zona, o.Subprefeitura = zona, subprefeitura
			ocorrencias = append(ocorrencias, o)
		}
	}

	if len(ocorrencias) == 0 && !strings.Contains(doc, SemRegistros) {
		return nil, fmt.Errorf("%s: nenhum ponto encontrado e sem a mensagem %q — layout mudou?",
			dia.Format("2006-01-02"), SemRegistros)
	}

	// conferência: a soma do resumo da página tem que bater com o que foi lido
	if resumo := reResumo.FindAllStringSubmatch(doc, -1); len(resumo) > 0 {
		total := 0
		for _, r := range resumo {
			n, _ := strconv.Atoi(r[1])
			total += n
		}
		if total != len(ocorrencias) {
			return nil, fmt.Errorf("%s: resumo da página diz %d pontos, parser leu %d",
				dia.Format("2006-01-02"), total, len(ocorrencias))
		}
	}
	return ocorrencias, nil
}

func parsePonto(dia time.Time, bloco string) (Ocorrencia, error) {
	var o Ocorrencia
	o.Data = dia

	st := reStatus.FindStringSubmatch(bloco)
	if st == nil {
		return o, fmt.Errorf("status (ativo/inativo, transitável) não encontrado")
	}
	o.Ativo = st[1] == "ativo"
	o.Intransitavel = st[2] == "intransitavel"

	loc := reLocal.FindStringSubmatch(bloco)
	if loc == nil {
		return o, fmt.Errorf("horário/local não encontrado")
	}
	horario := limpar(loc[1])
	o.Local = limpar(loc[2])

	h := reHorario.FindStringSubmatch(horario)
	if h == nil {
		return o, fmt.Errorf("horário em formato desconhecido: %q", horario)
	}
	o.Inicio = horaNoDia(dia, h[1], h[2])
	if h[3] != "" {
		fim := horaNoDia(dia, h[3], h[4])
		if fim.Before(o.Inicio) { // passou da meia-noite
			fim = fim.AddDate(0, 0, 1)
		}
		o.Fim = &fim
	}

	if s := reSentido.FindStringSubmatch(bloco); s != nil {
		o.Sentido = limpar(s[1])
		o.Referencia = limpar(s[2])
	} else {
		return o, fmt.Errorf("sentido/referência não encontrado")
	}

	o.ID = gerarID(o)
	return o, nil
}

func horaNoDia(dia time.Time, hh, mm string) time.Time {
	h, _ := strconv.Atoi(hh)
	m, _ := strconv.Atoi(mm)
	return time.Date(dia.Year(), dia.Month(), dia.Day(), h, m, 0, 0, dia.Location())
}

// limpar decodifica entidades HTML e normaliza espaços.
func limpar(s string) string {
	return strings.Join(strings.Fields(html.UnescapeString(s)), " ")
}

// gerarID cria um identificador estável: a mesma ocorrência coletada em
// momentos diferentes (ou reprocessada) recebe o mesmo ID.
func gerarID(o Ocorrencia) string {
	chave := strings.Join([]string{
		o.Data.Format("2006-01-02"), o.Inicio.Format("15:04"), o.Local, o.Sentido, o.Referencia,
	}, "|")
	h := sha256.Sum256([]byte(chave))
	return hex.EncodeToString(h[:])[:12]
}

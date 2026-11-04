// Package parser transforma o HTML bruto do CGE em ocorrências normalizadas.
// Etapa 2 do PLANO — ainda não implementado. Ver docs/fontes/cge.md.
package parser

import (
	"errors"
	"time"
)

// Ocorrencia é um ponto de alagamento num dia.
type Ocorrencia struct {
	ID            string     // estável: hash de data + local + sentido + referência + início
	Data          time.Time  // dia consultado
	Inicio        time.Time  // "De HH:MM ..."
	Fim           *time.Time // nil se ainda ativo
	Zona          string     // Norte, Centro, Oeste, Leste, Sul, Sudeste
	Subprefeitura string     // ex.: Casa Verde
	Local         string     // endereço como veio, ex.: "AV. BRAZ LEME (CASA VERDE)"
	Sentido       string     // ex.: "BAIRRO/CENTRO", "AMBOS"
	Referencia    string     // cruzamento ou ponto de referência
	Intransitavel bool       // lido do ícone do ponto
	Ativo         bool       // lido do ícone do ponto
}

var ErrNaoImplementado = errors.New("parser: ainda não implementado")

// Parse lê o HTML de um dia. Seletores candidatos (confirmar com snapshot real):
//   - bloco do ponto: div.ponto-de-alagamento
//   - local e horário: li.arial-descr-alag.col-local
//   - sentido/referência: li que contém "Referência:"
//   - subprefeitura: primeiro td da tabela que envolve o bloco
func Parse(dia time.Time, html []byte) ([]Ocorrencia, error) {
	return nil, ErrNaoImplementado
}

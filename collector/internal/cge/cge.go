// Package cge baixa a página de alagamentos do CGE (Centro de Gerenciamento
// de Emergências Climáticas da Prefeitura de São Paulo).
package cge

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURL   = "https://www.cgesp.org/v3/alagamentos.jsp"
	UserAgent = "sp-alagamentos/0.1 (+https://gor.dev.br; coleta educada para pesquisa)"
	// SemRegistros é o texto exibido pelo CGE em dias sem alagamento.
	SemRegistros = "Não há registros de alagamentos para essa data."
)

// URLDoDia monta a URL da consulta histórica para uma data
// (parâmetro dataBusca no formato dd/mm/aaaa).
func URLDoDia(dia time.Time) string {
	q := url.Values{}
	q.Set("dataBusca", dia.Format("02/01/2006"))
	return BaseURL + "?" + q.Encode()
}

// Cliente faz requisições ao CGE com timeout e user-agent identificado.
type Cliente struct {
	HTTP *http.Client
}

func NovoCliente() *Cliente {
	return &Cliente{HTTP: &http.Client{Timeout: 30 * time.Second}}
}

// BaixarDia devolve o HTML bruto da página de alagamentos de uma data.
func (c *Cliente) BaixarDia(ctx context.Context, dia time.Time) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URLDoDia(dia), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cge: status %d para %s", resp.StatusCode, dia.Format("2006-01-02"))
	}
	return io.ReadAll(io.LimitReader(resp.Body, 10<<20))
}

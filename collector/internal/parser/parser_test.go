package parser

import (
	"compress/gzip"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

var sp, _ = time.LoadLocation("America/Sao_Paulo")

func lerTestdata(t *testing.T, nome string) []byte {
	t.Helper()
	f, err := os.Open("../../testdata/" + nome)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// 10/02/2020: 129 pontos, todos inativos (46 transitáveis, 83 intransitáveis).
func TestParseDiaCheio(t *testing.T) {
	dia := time.Date(2020, 2, 10, 0, 0, 0, 0, sp)
	oc, err := Parse(dia, lerTestdata(t, "2020-02-10.html.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if len(oc) != 129 {
		t.Fatalf("pontos = %d, want 129", len(oc))
	}

	intrans, ativos := 0, 0
	porZona := map[string]int{}
	ids := map[string]bool{}
	for _, o := range oc {
		if o.Intransitavel {
			intrans++
		}
		if o.Ativo {
			ativos++
		}
		porZona[o.Zona]++
		if ids[o.ID] {
			t.Errorf("ID repetido: %s (%s)", o.ID, o.Local)
		}
		ids[o.ID] = true
		if o.Local == "" || o.Sentido == "" || o.Subprefeitura == "" || o.Fim == nil {
			t.Errorf("campo vazio: %+v", o)
		}
	}
	if intrans != 83 || ativos != 0 {
		t.Errorf("intransitáveis = %d (want 83), ativos = %d (want 0)", intrans, ativos)
	}
	// totais por zona conferidos na página
	want := map[string]int{"Zona Norte": 31, "Centro": 17, "Zona Oeste": 49, "Zona Leste": 14, "Zona Sul": 13, "Zona Sudeste": 5}
	for z, n := range want {
		if porZona[z] != n {
			t.Errorf("%s = %d, want %d", z, porZona[z], n)
		}
	}

	// primeiro ponto, conferido à mão
	p := oc[0]
	if p.Zona != "Zona Norte" || p.Subprefeitura != "Casa Verde" ||
		p.Local != "MARGINAL TIETE ASCB EXPRESSA (CASA VERDE)" ||
		p.Sentido != "A SENNA/CASTELO B - EXP" ||
		p.Referencia != "PTE DO LIMAO-ADHEMAR F. DA SILVA" ||
		p.Inicio.Format("15:04") != "03:25" || p.Fim.Format("15:04") != "09:47" ||
		!p.Intransitavel || p.Ativo {
		t.Errorf("primeiro ponto inesperado: %+v", p)
	}
}

func TestParseDiaVazio(t *testing.T) {
	dia := time.Date(2026, 11, 4, 0, 0, 0, 0, sp)
	oc, err := Parse(dia, lerTestdata(t, "2026-11-04-vazio.html.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if len(oc) != 0 {
		t.Fatalf("pontos = %d, want 0", len(oc))
	}
}

func TestParseLayoutDesconhecido(t *testing.T) {
	_, err := Parse(time.Now(), []byte("<html><body>manutenção</body></html>"))
	if err == nil || !strings.Contains(err.Error(), "layout mudou") {
		t.Fatalf("esperava erro de layout, veio %v", err)
	}
}

func TestPontoAtivoEMeiaNoite(t *testing.T) {
	dia := time.Date(2026, 1, 15, 0, 0, 0, 0, sp)
	bloco := `<li class="ativo-transitavel" title="Ativo Transitável"></li>
		<li class="arial-descr-alag col-local">De 23:10 a <br/>R. TESTE &amp; CIA</li>
		<li class="arial-descr-alag">Sentido: AMBOS<br/>Referência: AV. X</li>`
	o, err := parsePonto(dia, bloco)
	if err != nil {
		t.Fatal(err)
	}
	if !o.Ativo || o.Intransitavel || o.Fim != nil || o.Local != "R. TESTE & CIA" {
		t.Errorf("ponto ativo inesperado: %+v", o)
	}

	bloco = strings.Replace(bloco, "De 23:10 a ", "De 23:10 a 00:40", 1)
	o, err = parsePonto(dia, bloco)
	if err != nil {
		t.Fatal(err)
	}
	if o.Fim == nil || o.Fim.Day() != 16 || o.Fim.Sub(o.Inicio) != 90*time.Minute {
		t.Errorf("fim após meia-noite: %v → %v", o.Inicio, o.Fim)
	}
}

// 14/09/2026: página real com ponto ativo e sem horário de fim ("De 14:31 a").
func TestParsePontoAtivoReal(t *testing.T) {
	dia := time.Date(2026, 9, 14, 0, 0, 0, 0, sp)
	oc, err := Parse(dia, lerTestdata(t, "2026-09-14-ativo.html.gz"))
	if err != nil {
		t.Fatal(err)
	}
	var achou bool
	for _, o := range oc {
		if o.Local == "AV CARD MOTTA" {
			achou = true
			if !o.Ativo || o.Intransitavel || o.Fim != nil || o.Inicio.Format("15:04") != "14:31" {
				t.Errorf("ponto ativo inesperado: %+v", o)
			}
		}
	}
	if !achou {
		t.Fatal("ponto AV CARD MOTTA não encontrado")
	}
}

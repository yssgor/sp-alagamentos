// Comando coletor: baixa páginas de alagamento do CGE e guarda o HTML bruto.
//
//	coletor dia [-data AAAA-MM-DD|hoje|ontem]   # padrão: ontem (horário de SP)
//	coletor backfill -de AAAA-MM-DD -ate AAAA-MM-DD [-intervalo 5s]
//	coletor processar [-raiz ../data/raw] [-saida ../data/processed/ocorrencias.csv]
//
// Sai com código 1 se alguma coleta falhar, para o GitHub Actions avisar.
package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/yssgor/sp-alagamentos/collector/internal/armazem"
	"github.com/yssgor/sp-alagamentos/collector/internal/cge"
	"github.com/yssgor/sp-alagamentos/collector/internal/parser"
)

var sp = carregarFuso()

func carregarFuso() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*60*60)
	}
	return loc
}

func main() {
	if len(os.Args) < 2 {
		uso()
	}
	switch os.Args[1] {
	case "dia":
		cmdDia(os.Args[2:])
	case "backfill":
		cmdBackfill(os.Args[2:])
	case "processar":
		cmdProcessar(os.Args[2:])
	default:
		uso()
	}
}

func uso() {
	fmt.Fprintln(os.Stderr, "uso: coletor dia [-data AAAA-MM-DD|hoje|ontem] | coletor backfill -de AAAA-MM-DD -ate AAAA-MM-DD [-intervalo 5s] | coletor processar [-raiz DIR] [-saida ARQ.csv]")
	os.Exit(2)
}

func cmdDia(args []string) {
	fs := flag.NewFlagSet("dia", flag.ExitOnError)
	data := fs.String("data", "ontem", "data a coletar: AAAA-MM-DD, hoje ou ontem")
	raiz := fs.String("raiz", "../data/raw", "pasta dos snapshots brutos")
	fs.Parse(args)

	if err := coletar(context.Background(), cge.NovoCliente(), armazem.Bruto{Raiz: *raiz}, parseData(*data)); err != nil {
		os.Exit(1)
	}
}

func cmdBackfill(args []string) {
	fs := flag.NewFlagSet("backfill", flag.ExitOnError)
	de := fs.String("de", "", "primeira data (AAAA-MM-DD)")
	ate := fs.String("ate", "", "última data (AAAA-MM-DD)")
	intervalo := fs.Duration("intervalo", 5*time.Second, "pausa entre requisições (coleta educada)")
	raiz := fs.String("raiz", "../data/raw", "pasta dos snapshots brutos")
	fs.Parse(args)
	if *de == "" || *ate == "" {
		uso()
	}

	cli, store := cge.NovoCliente(), armazem.Bruto{Raiz: *raiz}
	total, falhas := 0, 0
	for _, d := range diasEntre(parseData(*de), parseData(*ate)) {
		total++
		if err := coletar(context.Background(), cli, store, d); err != nil {
			falhas++
		}
		time.Sleep(*intervalo)
	}
	log.Printf("backfill: %d dias, %d falhas", total, falhas)
	if falhas > 0 {
		os.Exit(1)
	}
}

// cmdProcessar lê o snapshot mais recente de cada dia e grava todas as
// ocorrências num CSV. Dias com erro de parsing são listados e o comando
// termina com código 1, mas os demais dias são gravados.
func cmdProcessar(args []string) {
	fs := flag.NewFlagSet("processar", flag.ExitOnError)
	raiz := fs.String("raiz", "../data/raw", "pasta dos snapshots brutos")
	saida := fs.String("saida", "../data/processed/ocorrencias.csv", "arquivo CSV de saída")
	fs.Parse(args)

	snaps, err := armazem.Bruto{Raiz: *raiz}.UltimosPorDia(sp)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(*saida), 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(*saida)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"id", "data", "inicio", "fim", "zona", "subprefeitura", "local",
		"sentido", "referencia", "intransitavel", "ativo"})

	ocorrencias, falhas := 0, 0
	for _, s := range snaps {
		html, err := armazem.Ler(s.Caminho)
		if err == nil {
			var oc []parser.Ocorrencia
			oc, err = parser.Parse(s.Dia, html)
			for _, o := range oc {
				fim := ""
				if o.Fim != nil {
					fim = o.Fim.Format(time.RFC3339)
				}
				w.Write([]string{o.ID, o.Data.Format("2006-01-02"), o.Inicio.Format(time.RFC3339), fim,
					o.Zona, o.Subprefeitura, o.Local, o.Sentido, o.Referencia,
					strconv.FormatBool(o.Intransitavel), strconv.FormatBool(o.Ativo)})
			}
			ocorrencias += len(oc)
		}
		if err != nil {
			falhas++
			log.Printf("erro: %v (%s)", err, s.Caminho)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
	log.Printf("processar: %d dias, %d ocorrências, %d dias com erro → %s", len(snaps), ocorrencias, falhas, *saida)
	if falhas > 0 {
		os.Exit(1)
	}
}

func coletar(ctx context.Context, cli *cge.Cliente, store armazem.Bruto, dia time.Time) error {
	html, err := cli.BaixarDia(ctx, dia)
	if err != nil {
		log.Printf("%s: erro: %v", dia.Format("2006-01-02"), err)
		return err
	}
	caminho, gravou, err := store.Salvar(dia, time.Now(), html)
	if err != nil {
		log.Printf("%s: erro ao salvar: %v", dia.Format("2006-01-02"), err)
		return err
	}
	status := "sem mudança"
	if gravou {
		status = "salvo"
	}
	log.Printf("%s: %d bytes, %s (%s)", dia.Format("2006-01-02"), len(html), status, caminho)
	return nil
}

// parseData devolve a data ao meio-dia de SP. Meio-dia e não meia-noite porque,
// nos anos com horário de verão (até 2019), a meia-noite do dia em que o relógio
// adiantava não existia: o Go a empurrava para 01:00 e o laço do backfill
// pulava o último dia do período.
func parseData(s string) time.Time {
	agora := time.Now().In(sp)
	hoje := time.Date(agora.Year(), agora.Month(), agora.Day(), 12, 0, 0, 0, sp)
	switch s {
	case "hoje":
		return hoje
	case "ontem":
		return hoje.AddDate(0, 0, -1)
	}
	d, err := time.ParseInLocation("2006-01-02", s, sp)
	if err != nil {
		log.Fatalf("data inválida %q: %v", s, err)
	}
	return time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, sp)
}

// diasEntre lista as datas de de até ate, inclusive.
func diasEntre(de, ate time.Time) []time.Time {
	var dias []time.Time
	for d := de; !d.After(ate); d = d.AddDate(0, 0, 1) {
		dias = append(dias, d)
	}
	return dias
}

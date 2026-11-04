// Comando coletor: baixa páginas de alagamento do CGE e guarda o HTML bruto.
//
//	coletor dia [-data AAAA-MM-DD|hoje|ontem]   # padrão: ontem (horário de SP)
//	coletor backfill -de AAAA-MM-DD -ate AAAA-MM-DD [-intervalo 5s]
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yssgor/sp-alagamentos/collector/internal/armazem"
	"github.com/yssgor/sp-alagamentos/collector/internal/cge"
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
	default:
		uso()
	}
}

func uso() {
	fmt.Fprintln(os.Stderr, "uso: coletor dia [-data AAAA-MM-DD|hoje|ontem] | coletor backfill -de AAAA-MM-DD -ate AAAA-MM-DD [-intervalo 5s]")
	os.Exit(2)
}

func cmdDia(args []string) {
	fs := flag.NewFlagSet("dia", flag.ExitOnError)
	data := fs.String("data", "ontem", "data a coletar: AAAA-MM-DD, hoje ou ontem")
	raiz := fs.String("raiz", "../data/raw", "pasta dos snapshots brutos")
	fs.Parse(args)

	coletar(context.Background(), cge.NovoCliente(), armazem.Bruto{Raiz: *raiz}, parseData(*data))
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
	for d := parseData(*de); !d.After(parseData(*ate)); d = d.AddDate(0, 0, 1) {
		coletar(context.Background(), cli, store, d)
		time.Sleep(*intervalo)
	}
}

func coletar(ctx context.Context, cli *cge.Cliente, store armazem.Bruto, dia time.Time) {
	html, err := cli.BaixarDia(ctx, dia)
	if err != nil {
		log.Printf("%s: erro: %v", dia.Format("2006-01-02"), err)
		return
	}
	caminho, gravou, err := store.Salvar(dia, time.Now(), html)
	if err != nil {
		log.Printf("%s: erro ao salvar: %v", dia.Format("2006-01-02"), err)
		return
	}
	status := "sem mudança"
	if gravou {
		status = "salvo"
	}
	log.Printf("%s: %d bytes, %s (%s)", dia.Format("2006-01-02"), len(html), status, caminho)
}

func parseData(s string) time.Time {
	hoje := time.Now().In(sp)
	hoje = time.Date(hoje.Year(), hoje.Month(), hoje.Day(), 0, 0, 0, 0, sp)
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
	return d
}

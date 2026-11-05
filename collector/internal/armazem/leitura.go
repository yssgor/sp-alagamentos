package armazem

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Snapshot é a versão mais recente salva para um dia.
type Snapshot struct {
	Dia     time.Time
	Caminho string
}

// UltimosPorDia lista, para cada dia com snapshots, o arquivo mais recente
// (o estado final do dia), em ordem cronológica.
func (b Bruto) UltimosPorDia(loc *time.Location) ([]Snapshot, error) {
	pastas, err := filepath.Glob(filepath.Join(b.Raiz, "cge", "[0-9][0-9][0-9][0-9]", "[0-9][0-9]", "[0-9][0-9]"))
	if err != nil {
		return nil, err
	}
	sort.Strings(pastas)

	var out []Snapshot
	for _, p := range pastas {
		rel, _ := filepath.Rel(filepath.Join(b.Raiz, "cge"), p)
		dia, err := time.ParseInLocation("2006/01/02", filepath.ToSlash(rel), loc)
		if err != nil {
			continue
		}
		ultimo, err := ultimoArquivo(p)
		if err != nil {
			return nil, err
		}
		if ultimo != "" {
			out = append(out, Snapshot{Dia: dia, Caminho: ultimo})
		}
	}
	return out, nil
}

// Ler devolve o HTML descomprimido de um snapshot.
func Ler(caminho string) ([]byte, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

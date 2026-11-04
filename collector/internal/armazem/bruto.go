// Package armazem guarda os snapshots brutos exatamente como vieram,
// para que o parser possa ser reexecutado quando mudar.
package armazem

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Bruto grava snapshots em <Raiz>/cge/AAAA/MM/DD/<horario>_<hash>.html.gz.
type Bruto struct {
	Raiz string
}

// CaminhoDoDia devolve a pasta dos snapshots de uma data consultada.
func (b Bruto) CaminhoDoDia(dia time.Time) string {
	return filepath.Join(b.Raiz, "cge", dia.Format("2006"), dia.Format("01"), dia.Format("02"))
}

func hashCurto(conteudo []byte) string {
	h := sha256.Sum256(conteudo)
	return hex.EncodeToString(h[:])[:12]
}

// Salvar grava o snapshot se ele for diferente do último salvo para o dia
// (idempotente: rodar duas vezes com o mesmo conteúdo não duplica).
// Devolve o caminho gravado e se houve gravação.
func (b Bruto) Salvar(dia, coletadoEm time.Time, conteudo []byte) (string, bool, error) {
	pasta := b.CaminhoDoDia(dia)
	if err := os.MkdirAll(pasta, 0o755); err != nil {
		return "", false, err
	}
	hash := hashCurto(conteudo)
	if ultimo, _ := ultimoArquivo(pasta); ultimo != "" && filepath.Ext(ultimo) == ".gz" &&
		bytes.Contains([]byte(filepath.Base(ultimo)), []byte(hash)) {
		return ultimo, false, nil
	}
	nome := fmt.Sprintf("%s_%s.html.gz", coletadoEm.UTC().Format("20060102T150405Z"), hash)
	caminho := filepath.Join(pasta, nome)

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(conteudo); err != nil {
		return "", false, err
	}
	if err := zw.Close(); err != nil {
		return "", false, err
	}
	tmp := caminho + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return "", false, err
	}
	return caminho, true, os.Rename(tmp, caminho)
}

func ultimoArquivo(pasta string) (string, error) {
	nomes, err := filepath.Glob(filepath.Join(pasta, "*.html.gz"))
	if err != nil || len(nomes) == 0 {
		return "", err
	}
	sort.Strings(nomes)
	return nomes[len(nomes)-1], nil
}

package armazem

import (
	"testing"
	"time"
)

func TestSalvarIdempotente(t *testing.T) {
	b := Bruto{Raiz: t.TempDir()}
	dia := time.Date(2020, 2, 10, 0, 0, 0, 0, time.UTC)
	agora := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)

	_, gravou, err := b.Salvar(dia, agora, []byte("<html>a</html>"))
	if err != nil || !gravou {
		t.Fatalf("primeira gravação: gravou=%v err=%v", gravou, err)
	}
	_, gravou, err = b.Salvar(dia, agora.Add(time.Minute), []byte("<html>a</html>"))
	if err != nil || gravou {
		t.Fatalf("conteúdo igual não deveria gravar: gravou=%v err=%v", gravou, err)
	}
	_, gravou, err = b.Salvar(dia, agora.Add(2*time.Minute), []byte("<html>b</html>"))
	if err != nil || !gravou {
		t.Fatalf("conteúdo novo deveria gravar: gravou=%v err=%v", gravou, err)
	}
}

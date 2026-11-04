package cge

import (
	"testing"
	"time"
)

func TestURLDoDia(t *testing.T) {
	dia := time.Date(2020, 2, 10, 0, 0, 0, 0, time.UTC)
	want := "https://www.cgesp.org/v3/alagamentos.jsp?dataBusca=10%2F02%2F2020"
	if got := URLDoDia(dia); got != want {
		t.Fatalf("URLDoDia = %q, want %q", got, want)
	}
}

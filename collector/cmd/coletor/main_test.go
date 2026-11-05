package main

import "testing"

// Regressão: em 2018 o horário de verão começou em 04/11 à meia-noite.
// O laço antigo pulava 31/12 em todos os anos com horário de verão.
func TestDiasEntreComHorarioDeVerao(t *testing.T) {
	dias := diasEntre(parseData("2018-10-01"), parseData("2018-12-31"))
	if len(dias) != 92 {
		t.Fatalf("dias = %d, want 92", len(dias))
	}
	if got := dias[len(dias)-1].Format("2006-01-02"); got != "2018-12-31" {
		t.Fatalf("último dia = %s, want 2018-12-31", got)
	}
	for i, d := range dias {
		if d.Hour() != 12 {
			t.Fatalf("dia %d (%s) às %d h, want 12 h", i, d.Format("2006-01-02"), d.Hour())
		}
	}
}

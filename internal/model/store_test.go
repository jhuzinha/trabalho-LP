package model

import "testing"

func TestNewStoreValidatesFields(t *testing.T) {
	if _, err := NewStore("", "http://example.com"); err == nil {
		t.Fatal("esperava erro para nome vazio")
	}

	if _, err := NewStore("Loja", ""); err == nil {
		t.Fatal("esperava erro para URL vazia")
	}

	store, err := NewStore(" Loja ", " http://example.com ")
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if store.Name != "Loja" || store.URL != "http://example.com" {
		t.Fatalf("valores não foram normalizados: %#v", store)
	}
}

func TestResultOK(t *testing.T) {
	ok := NewSuccessResult("Loja", 10, 1, 200)
	if !ok.OK() {
		t.Fatal("resultado de sucesso deveria ser OK")
	}

	timeout := NewTimeoutResult("Loja", 1, "timeout")
	if timeout.OK() {
		t.Fatal("resultado com timeout não deveria ser OK")
	}

	err := NewErrorResult("Loja", 1, 500, "erro")
	if err.OK() {
		t.Fatal("resultado com erro não deveria ser OK")
	}
}

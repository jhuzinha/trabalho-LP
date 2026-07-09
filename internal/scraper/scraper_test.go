package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"trabalho-lp/internal/model"
)

func TestFetchSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"price":123.45}`))
	}))
	defer server.Close()

	scraper := newTestScraper(t, 500*time.Millisecond, 2)
	store, _ := model.NewStore("Teste", server.URL)

	result := scraper.Fetch(context.Background(), store)
	if !result.OK() {
		t.Fatalf("esperava sucesso, obtido erro: %#v", result)
	}
	if result.Price != 123.45 {
		t.Fatalf("preço esperado 123.45, obtido %.2f", result.Price)
	}
}

func TestFetchHTTPStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "erro interno", http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := newTestScraper(t, 500*time.Millisecond, 2)
	store, _ := model.NewStore("Teste", server.URL)

	result := scraper.Fetch(context.Background(), store)
	if result.OK() {
		t.Fatalf("não esperava sucesso para HTTP 500: %#v", result)
	}
	if result.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status esperado 500, obtido %d", result.StatusCode)
	}
}

func TestFetchTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{"price":10}`))
	}))
	defer server.Close()

	scraper := newTestScraper(t, 20*time.Millisecond, 2)
	store, _ := model.NewStore("Lenta", server.URL)

	result := scraper.Fetch(context.Background(), store)
	if !result.TimedOut {
		t.Fatalf("esperava timeout, obtido: %#v", result)
	}
}

func TestRunConcurrentReturnsAllResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"price":99.90}`))
	}))
	defer server.Close()

	scraper := newTestScraper(t, 500*time.Millisecond, 2)
	stores := []model.Store{
		{Name: "A", URL: server.URL},
		{Name: "B", URL: server.URL},
		{Name: "C", URL: server.URL},
	}

	results, _ := scraper.RunConcurrent(context.Background(), stores)
	if len(results) != len(stores) {
		t.Fatalf("esperava %d resultados, obtive %d", len(stores), len(results))
	}
	for _, result := range results {
		if !result.OK() {
			t.Fatalf("resultado deveria ser OK: %#v", result)
		}
	}
}

func TestNewConfigValidation(t *testing.T) {
	if _, err := NewConfig(0, 1); err == nil {
		t.Fatal("esperava erro para timeout inválido")
	}
	if _, err := NewConfig(time.Second, 0); err == nil {
		t.Fatal("esperava erro para workers inválido")
	}
}

func newTestScraper(t *testing.T, timeout time.Duration, workers int) *Scraper {
	t.Helper()
	config, err := NewConfig(timeout, workers)
	if err != nil {
		t.Fatal(err)
	}
	scraper, err := NewScraper(config)
	if err != nil {
		t.Fatal(err)
	}
	return scraper
}

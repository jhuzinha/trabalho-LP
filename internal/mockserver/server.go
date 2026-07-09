// Package mockserver fornece um servidor HTTP local que simula lojas virtuais.
// O objetivo é permitir uma demonstração reprodutível sem depender de internet.
package mockserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"trabalho-lp/internal/model"
)

// StoreConfig descreve uma loja simulada.
type StoreConfig struct {
	Path  string
	Name  string
	Price float64
	Delay time.Duration
}

// Server encapsula o servidor HTTP local.
type Server struct {
	port   int
	stores []StoreConfig
	server *http.Server
}

// NewServer é o construtor do servidor mock.
func NewServer(port int, stores []StoreConfig) (*Server, error) {
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("porta inválida: %d", port)
	}
	if len(stores) == 0 {
		return nil, errors.New("é necessário configurar pelo menos uma loja")
	}

	s := &Server{port: port, stores: stores}
	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           s.routes(),
		ReadHeaderTimeout: 2 * time.Second,
	}
	return s, nil
}

// DefaultStoreConfigs retorna lojas com latências fixas.
// As latências fixas deixam o comparativo entre Go e Python mais justo e reprodutível.
// Algumas lojas são propositalmente lentas para demonstrar timeout e cancelamento.
func DefaultStoreConfigs() []StoreConfig {
	return []StoreConfig{
		{Path: "/amazon", Name: "Amazon", Price: 1299.90, Delay: 900 * time.Millisecond},
		{Path: "/mercado-livre", Name: "Mercado Livre", Price: 1249.00, Delay: 1600 * time.Millisecond},
		{Path: "/americanas", Name: "Americanas", Price: 1399.00, Delay: 1100 * time.Millisecond},
		{Path: "/magazine-luiza", Name: "Magazine Luiza", Price: 1279.90, Delay: 500 * time.Millisecond},
		{Path: "/casas-bahia", Name: "Casas Bahia", Price: 1350.00, Delay: 2100 * time.Millisecond},
		{Path: "/shopee", Name: "Shopee", Price: 1199.00, Delay: 800 * time.Millisecond},
		{Path: "/kabum", Name: "KaBuM!", Price: 1280.00, Delay: 600 * time.Millisecond},
		{Path: "/submarino", Name: "Submarino", Price: 1320.00, Delay: 2300 * time.Millisecond},
		{Path: "/aliexpress", Name: "AliExpress", Price: 1100.00, Delay: 3000 * time.Millisecond},
		{Path: "/carrefour", Name: "Carrefour", Price: 1310.00, Delay: 2800 * time.Millisecond},
	}
}

// Start inicia o servidor. Ele bloqueia até receber cancelamento do contexto.
func (s *Server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.server.Shutdown(shutdownCtx)
	}()

	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// WaitUntilReady espera o endpoint /health responder.
func (s *Server) WaitUntilReady(ctx context.Context) error {
	return WaitForHealth(ctx, s.port)
}

// WaitForHealth espera o endpoint /health responder em uma porta específica.
func WaitForHealth(ctx context.Context, port int) error {
	client := &http.Client{Timeout: 150 * time.Millisecond}
	url := fmt.Sprintf("http://localhost:%d/health", port)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resp, err := client.Get(url)
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
		}
	}
}

// Stores retorna as lojas como endpoints HTTP prontos para o scraper consultar.
func (s *Server) Stores() ([]model.Store, error) {
	return StoresForPort(s.port, s.stores)
}

// StoresForPort monta os endpoints HTTP para uma porta específica.
// Isso permite que o scraper em Go e o coletor em Python consultem o mesmo mock server.
func StoresForPort(port int, configs []StoreConfig) ([]model.Store, error) {
	stores := make([]model.Store, 0, len(configs))
	for _, cfg := range configs {
		url := fmt.Sprintf("http://localhost:%d%s", port, cfg.Path)
		store, err := model.NewStore(cfg.Name, url)
		if err != nil {
			return nil, err
		}
		stores = append(stores, store)
	}
	return stores, nil
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/stores", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s.stores)
	})

	for _, store := range s.stores {
		cfg := store
		mux.HandleFunc(cfg.Path, func(w http.ResponseWriter, r *http.Request) {
			timer := time.NewTimer(cfg.Delay)
			defer timer.Stop()

			select {
			case <-r.Context().Done():
				return
			case <-timer.C:
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"store": cfg.Name,
				"price": cfg.Price,
			})
		})
	}

	return mux
}

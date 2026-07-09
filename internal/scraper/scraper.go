// Package scraper implementa a parte principal do projeto:
// consultas HTTP sequenciais e concorrentes usando goroutines, channels e worker pool.
package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"trabalho-lp/internal/model"
)

// Config concentra os parâmetros de execução do scraper.
type Config struct {
	Timeout time.Duration
	Workers int
}

// NewConfig é o construtor de Config.
func NewConfig(timeout time.Duration, workers int) (Config, error) {
	if timeout <= 0 {
		return Config{}, errors.New("timeout deve ser maior que zero")
	}
	if workers <= 0 {
		return Config{}, errors.New("quantidade de workers deve ser maior que zero")
	}
	return Config{Timeout: timeout, Workers: workers}, nil
}

// Scraper consulta lojas e coleta preços.
type Scraper struct {
	client *http.Client
	config Config
}

// NewScraper é o construtor idiomático de Go para Scraper.
func NewScraper(config Config) (*Scraper, error) {
	if config.Timeout <= 0 {
		return nil, errors.New("timeout deve ser maior que zero")
	}
	if config.Workers <= 0 {
		return nil, errors.New("quantidade de workers deve ser maior que zero")
	}

	return &Scraper{
		client: &http.Client{},
		config: config,
	}, nil
}

// Fetch consulta uma loja. O context.WithTimeout cancela a requisição de verdade
// quando ela ultrapassa o tempo limite.
func (s *Scraper) Fetch(ctx context.Context, store model.Store) model.Result {
	start := time.Now()

	requestCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, store.URL, nil)
	if err != nil {
		return model.NewErrorResult(store.Name, time.Since(start), 0, err.Error())
	}
	req.Header.Set("User-Agent", "trabalho-lp-go-scraper/1.0")

	resp, err := s.client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		if errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return model.NewTimeoutResult(store.Name, elapsed, fmt.Sprintf("sem resposta em %.1fs", s.config.Timeout.Seconds()))
		}
		return model.NewErrorResult(store.Name, elapsed, 0, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.NewErrorResult(store.Name, elapsed, resp.StatusCode, fmt.Sprintf("status HTTP inválido: %d", resp.StatusCode))
	}

	var payload struct {
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return model.NewErrorResult(store.Name, time.Since(start), resp.StatusCode, err.Error())
	}
	if payload.Price <= 0 {
		return model.NewErrorResult(store.Name, time.Since(start), resp.StatusCode, "preço inválido ou ausente")
	}

	return model.NewSuccessResult(store.Name, payload.Price, time.Since(start), resp.StatusCode)
}

// RunSequential consulta uma loja por vez.
// Ele serve como baseline para comparar com a versão concorrente.
func (s *Scraper) RunSequential(ctx context.Context, stores []model.Store) ([]model.Result, time.Duration) {
	start := time.Now()
	results := make([]model.Result, 0, len(stores))

	for _, store := range stores {
		results = append(results, s.Fetch(ctx, store))
	}

	return results, time.Since(start)
}

// RunConcurrent consulta várias lojas em paralelo usando worker pool.
//
// Fluxo:
//   - jobs recebe as lojas que precisam ser consultadas;
//   - N workers rodam como goroutines;
//   - cada worker consome jobs e envia Result para results;
//   - o canal results é fechado quando todos os workers terminam.
func (s *Scraper) RunConcurrent(ctx context.Context, stores []model.Store) ([]model.Result, time.Duration) {
	start := time.Now()
	jobs := make(chan model.Store)
	results := make(chan model.Result)

	workerCount := s.config.Workers
	if workerCount > len(stores) {
		workerCount = len(stores)
	}
	if workerCount == 0 {
		return nil, time.Since(start)
	}

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for id := 1; id <= workerCount; id++ {
		go s.worker(ctx, id, jobs, results, &wg)
	}

	go func() {
		defer close(jobs)
		for _, store := range stores {
			select {
			case <-ctx.Done():
				return
			case jobs <- store:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	collected := make([]model.Result, 0, len(stores))
	for result := range results {
		collected = append(collected, result)
	}

	return collected, time.Since(start)
}

func (s *Scraper) worker(ctx context.Context, id int, jobs <-chan model.Store, results chan<- model.Result, wg *sync.WaitGroup) {
	defer wg.Done()
	_ = id // o id fica disponível para logs/debug sem poluir a saída principal.

	for store := range jobs {
		result := s.Fetch(ctx, store)
		select {
		case <-ctx.Done():
			return
		case results <- result:
		}
	}
}

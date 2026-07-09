package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"trabalho-lp/internal/mockserver"
	"trabalho-lp/internal/model"
	"trabalho-lp/internal/report"
	"trabalho-lp/internal/scraper"
)

func main() {
	product := flag.String("produto", "iPhone 15 Pro 256GB", "produto exibido na demonstração")
	port := flag.Int("porta", 18080, "porta do servidor mock local")
	workers := flag.Int("workers", 5, "quantidade de workers/goroutines no pool")
	timeout := flag.Duration("timeout", 2500*time.Millisecond, "timeout por loja")
	outputDir := flag.String("saida", "reports", "pasta para salvar CSV e JSON")
	skipSequential := flag.Bool("sem-sequencial", false, "executa apenas a versão concorrente")
	useExistingServer := flag.Bool("usar-servidor-existente", false, "não inicia servidor mock interno; usa um servidor já rodando na porta informada")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stores []model.Store
	var err error
	if *useExistingServer {
		readyCtx, stopWaiting := context.WithTimeout(ctx, 2*time.Second)
		defer stopWaiting()
		if err := mockserver.WaitForHealth(readyCtx, *port); err != nil {
			log.Fatalf("não encontrei servidor mock em http://localhost:%d: %v", *port, err)
		}

		stores, err = mockserver.StoresForPort(*port, mockserver.DefaultStoreConfigs())
		if err != nil {
			log.Fatal(err)
		}
	} else {
		server, err := mockserver.NewServer(*port, mockserver.DefaultStoreConfigs())
		if err != nil {
			log.Fatal(err)
		}

		serverErr := make(chan error, 1)
		go func() {
			serverErr <- server.Start(ctx)
		}()

		readyCtx, stopWaiting := context.WithTimeout(ctx, 2*time.Second)
		defer stopWaiting()
		if err := server.WaitUntilReady(readyCtx); err != nil {
			select {
			case serverStartErr := <-serverErr:
				log.Fatalf("não foi possível iniciar o servidor mock na porta %d: %v", *port, serverStartErr)
			default:
				log.Fatalf("servidor mock não ficou pronto: %v", err)
			}
		}

		stores, err = server.Stores()
		if err != nil {
			log.Fatal(err)
		}
	}

	config, err := scraper.NewConfig(*timeout, *workers)
	if err != nil {
		log.Fatal(err)
	}
	priceScraper, err := scraper.NewScraper(config)
	if err != nil {
		log.Fatal(err)
	}

	realWorkers := *workers
	if realWorkers > len(stores) {
		realWorkers = len(stores)
	}
	report.PrintHeader(*product, len(stores), realWorkers, *timeout)

	fmt.Println("  Disparando consultas concorrentes com worker pool...")
	fmt.Println()
	concurrentResults, concurrentTime := priceScraper.RunConcurrent(ctx, stores)
	report.PrintRun("Execução concorrente", concurrentResults, concurrentTime, *timeout)

	var sequentialResults []model.Result
	var sequentialTime time.Duration
	if !*skipSequential {
		fmt.Println("  Executando baseline sequencial para comparação...")
		fmt.Println()
		sequentialResults, sequentialTime = priceScraper.RunSequential(ctx, stores)
		report.PrintRun("Execução sequencial", sequentialResults, sequentialTime, *timeout)
	}

	report.PrintSummary(concurrentResults, concurrentTime, sequentialResults, sequentialTime)

	csvPath := filepath.Join(*outputDir, "resultados_concorrentes.csv")
	jsonPath := filepath.Join(*outputDir, "resultados_concorrentes.json")
	if err := report.SaveCSV(csvPath, concurrentResults); err != nil {
		log.Fatalf("erro ao salvar CSV: %v", err)
	}
	if err := report.SaveJSON(jsonPath, concurrentResults); err != nil {
		log.Fatalf("erro ao salvar JSON: %v", err)
	}

	fmt.Printf("  Relatórios salvos em:\n")
	fmt.Printf("  - %s\n", csvPath)
	fmt.Printf("  - %s\n", jsonPath)
}

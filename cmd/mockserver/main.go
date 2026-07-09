package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"trabalho-lp/internal/mockserver"
)

func main() {
	port := flag.Int("porta", 18080, "porta do servidor mock local")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server, err := mockserver.NewServer(*port, mockserver.DefaultStoreConfigs())
	if err != nil {
		log.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(ctx)
	}()

	readyCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.WaitUntilReady(readyCtx); err != nil {
		select {
		case startErr := <-errCh:
			log.Fatalf("não foi possível iniciar o servidor mock na porta %d: %v", *port, startErr)
		default:
			log.Fatalf("servidor mock não ficou pronto: %v", err)
		}
	}

	fmt.Printf("Servidor mock rodando em http://localhost:%d\n", *port)
	fmt.Println("Endpoints disponíveis em /stores e /health.")
	fmt.Println("Pressione Ctrl+C para encerrar.")

	select {
	case <-ctx.Done():
		fmt.Println("\nEncerrando servidor mock...")
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	}
}

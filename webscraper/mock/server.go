// Package mock simula servidores de lojas com latências distintas por loja.
package mock

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// StoreConfig descreve uma loja simulada com faixa de latência própria.
type StoreConfig struct {
	Path     string
	Name     string
	Price    float64
	MinDelay int // ms
	MaxDelay int // ms
}

// Stores é a lista de lojas consultadas pelo scraper.
// AliExpress e Carrefour têm delay alto para demonstrar o timeout do select.
var Stores = []StoreConfig{
	{"/amazon", "Amazon", 1299.90, 300, 1400},
	{"/mercadolivre", "Mercado Livre", 1249.00, 600, 2000},
	{"/americanas", "Americanas", 1399.00, 400, 1600},
	{"/magazineluiza", "Magazine Luiza", 1279.90, 200, 800},
	{"/casasbahia", "Casas Bahia", 1350.00, 800, 2200},
	{"/shopee", "Shopee", 1199.00, 300, 1200},
	{"/aliexpress", "AliExpress", 1100.00, 2700, 3200}, // sempre acima do timeout → descartada
	{"/submarino", "Submarino", 1320.00, 500, 2400},
	{"/kabum", "KaBuM!", 1280.00, 200, 900},
	{"/carrefour", "Carrefour", 1310.00, 2000, 3000}, // geralmente acima do timeout
}

// StartServer sobe um servidor HTTP que imita as lojas com latência variável.
func StartServer(port int) {
	mux := http.NewServeMux()

	for _, store := range Stores {
		s := store
		mux.HandleFunc(s.Path, func(w http.ResponseWriter, r *http.Request) {
			spread := s.MaxDelay - s.MinDelay
			delay := time.Duration(s.MinDelay+rand.Intn(spread)) * time.Millisecond
			time.Sleep(delay)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"store": s.Name,
				"price": s.Price,
			})
		})
	}

	addr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(addr, mux)
}

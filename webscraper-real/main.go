package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// ── Cores ANSI ────────────────────────────────────────────────────────────────
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
)

const itemTimeout = 6 * time.Second

// ── Structs da API DummyJSON ──────────────────────────────────────────────────

type SearchResponse struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
}

type Product struct {
	ID                 int     `json:"id"`
	Title              string  `json:"title"`
	Brand              string  `json:"brand"`
	Price              float64 `json:"price"`
	DiscountPercentage float64 `json:"discountPercentage"`
	Rating             float64 `json:"rating"`
	Stock              int     `json:"stock"`
	Category           string  `json:"category"`
}

// ── Result: o que cada goroutine devolve pelo canal ──────────────────────────

type Result struct {
	Product  Product
	Elapsed  time.Duration
	Err      error
	TimedOut bool
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

// ── Funções de busca ─────────────────────────────────────────────────────────

// searchProducts busca produtos na API pública DummyJSON.
func searchProducts(query string, limit int) ([]Product, error) {
	endpoint := fmt.Sprintf(
		"https://dummyjson.com/products/search?q=%s&limit=%d",
		url.QueryEscape(query), limit,
	)
	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("falha na busca: %w", err)
	}
	defer resp.Body.Close()

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar: %w", err)
	}
	return result.Products, nil
}

// fetchProduct busca os detalhes atualizados de um produto pelo ID.
//
// Padrão SELECT: goroutine interna faz o request HTTP; select aguarda
// a resposta ou o timer de timeout — o que chegar primeiro vence.
func fetchProduct(p Product, ch chan<- Result) {
	start := time.Now()
	done := make(chan Result, 1) // buffer 1: goroutine interna não bloqueia

	// Goroutine interna — isolada, só faz o HTTP request
	go func() {
		endpoint := fmt.Sprintf("https://dummyjson.com/products/%d", p.ID)
		resp, err := httpClient.Get(endpoint)
		if err != nil {
			done <- Result{Product: p, Err: err, Elapsed: time.Since(start)}
			return
		}
		defer resp.Body.Close()

		var detail Product
		if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
			done <- Result{Product: p, Err: err, Elapsed: time.Since(start)}
			return
		}
		done <- Result{Product: detail, Elapsed: time.Since(start)}
	}()

	// ── SELECT: resposta real vs timeout ─────────────────────────────────────
	select {
	case r := <-done: // API respondeu dentro do prazo ✓
		ch <- r
	case <-time.After(itemTimeout): // timer disparou primeiro ⚠
		ch <- Result{
			Product:  p,
			TimedOut: true,
			Elapsed:  itemTimeout,
			Err:      fmt.Errorf("sem resposta em %.0fs", itemTimeout.Seconds()),
		}
	}
}

// ── Helpers visuais ───────────────────────────────────────────────────────────

func timeBar(elapsed, total time.Duration, width int) string {
	ratio := float64(elapsed) / float64(total)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func divider() {
	fmt.Println(dim + "  " + strings.Repeat("─", 62) + reset)
}

func stars(rating float64) string {
	full := int(rating)
	s := strings.Repeat("★", full) + strings.Repeat("☆", 5-full)
	return s
}

// ── Impressão de resultados ───────────────────────────────────────────────────

func printResult(r Result) {
	const barW = 16
	bar := timeBar(r.Elapsed, itemTimeout, barW)
	name := truncate(fmt.Sprintf("%s %s", r.Product.Brand, r.Product.Title), 38)

	switch {
	case r.TimedOut:
		fmt.Printf("  %s⚠  %-38s  TIMEOUT  %s[%s]%s  %.2fs\n",
			yellow+bold, name, reset+yellow, bar, reset, r.Elapsed.Seconds())
	case r.Err != nil:
		fmt.Printf("  %s✗  %-38s  ERRO     %s[%s]%s  %.2fs\n",
			red+bold, name, reset+red, bar, reset, r.Elapsed.Seconds())
	default:
		discounted := r.Product.Price * (1 - r.Product.DiscountPercentage/100)
		discount := ""
		if r.Product.DiscountPercentage > 0 {
			discount = fmt.Sprintf("  %s-%d%%%s", yellow, int(r.Product.DiscountPercentage), reset)
		}
		fmt.Printf("  %s✓%s  %-38s  $%7.2f  %s[%s]%s  %.2fs%s\n",
			green+bold, reset, name, discounted,
			dim, bar, reset, r.Elapsed.Seconds(), discount)
	}
}

func printSummary(results []Result, wallTime time.Duration) {
	var serialTime float64
	var timedOut int
	var valid []Result

	for _, r := range results {
		serialTime += r.Elapsed.Seconds()
		if r.TimedOut {
			timedOut++
		}
		if r.Err == nil && !r.TimedOut {
			valid = append(valid, r)
		}
	}

	// Ordena pelo preço final (com desconto)
	finalPrice := func(p Product) float64 {
		return p.Price * (1 - p.DiscountPercentage/100)
	}
	sort.Slice(valid, func(i, j int) bool {
		return finalPrice(valid[i].Product) < finalPrice(valid[j].Product)
	})

	// Ranking
	fmt.Println()
	fmt.Println(bold + "  Ranking de preços:" + reset)
	divider()
	medals := []string{"🥇", "🥈", "🥉"}
	for i, r := range valid {
		medal := "   "
		if i < len(medals) {
			medal = medals[i]
		}
		color := ""
		if i == 0 {
			color = green + bold
		}
		p := r.Product
		name := truncate(fmt.Sprintf("%s %s", p.Brand, p.Title), 38)
		price := finalPrice(p)
		fmt.Printf("  %s%s %d. %-38s  $%7.2f  %s%s estoque: %d%s\n",
			color, medal, i+1, name, price, dim, stars(p.Rating), p.Stock, reset)
	}

	// Performance
	speedup := serialTime / wallTime.Seconds()
	fmt.Println()
	fmt.Println(bold + "  Performance:" + reset)
	divider()
	fmt.Printf("  %s⏱  Concorrente:%s    %.2fs\n", cyan+bold, reset, wallTime.Seconds())
	fmt.Printf("  %s🐌 Serial (est.):%s  %.2fs\n", dim, reset, serialTime)
	fmt.Printf("  %s🚀 Speedup:%s        %.1fx mais rápido!\n", green+bold, reset, speedup)
	if timedOut > 0 {
		fmt.Printf("  %s⚠  Timeouts:%s       %d item(s) descartado(s)\n", yellow+bold, reset, timedOut)
	}

	// Melhor oferta
	if len(valid) > 0 {
		best := valid[0]
		worst := valid[len(valid)-1]
		price := finalPrice(best.Product)
		fmt.Println()
		divider()
		fmt.Printf("  %s★  Melhor oferta:%s  %s %s — $%.2f\n",
			green+bold, reset, best.Product.Brand, best.Product.Title, price)
		if len(valid) > 1 {
			diff := finalPrice(worst.Product) - price
			fmt.Printf("  %s   Economia:%s      $%.2f em relação ao mais caro\n", green, reset, diff)
		}
	}
	fmt.Println()
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	enableWindowsANSI()

	query := "iphone"
	limit := 10

	fmt.Println(bold + cyan + "╔══════════════════════════════════════════════════════════════╗" + reset)
	fmt.Println(bold+cyan+"║"+reset + bold+"   Web Scraper Concorrente  —  API Real (DummyJSON)          "+bold+cyan+"║"+reset)
	fmt.Println(bold+cyan+"║"+reset + dim+"   Goroutines  ·  Channels  ·  Select  ·  Timeout            "+bold+cyan+"║"+reset)
	fmt.Println(bold + cyan + "╚══════════════════════════════════════════════════════════════╝" + reset)
	fmt.Println()
	fmt.Printf("  %sBusca%s     %q\n", bold, reset, query)
	fmt.Printf("  %sResultados%s %d produtos\n", bold, reset, limit)
	fmt.Printf("  %sTimeout%s   %.0fs por request  %s(select + time.After)%s\n", bold, reset, itemTimeout.Seconds(), dim, reset)
	fmt.Printf("  %sAPI%s       dummyjson.com  %s(pública, sem autenticação)%s\n", bold, reset, dim, reset)
	fmt.Println()

	// 1. Busca inicial (sequencial — obtém a lista de produtos)
	fmt.Print(dim + "  [1/2] Buscando produtos..." + reset)
	t0 := time.Now()
	products, err := searchProducts(query, limit)
	if err != nil {
		fmt.Printf("\n  %s✗ %v%s\n\n", red, err, reset)
		return
	}
	fmt.Printf(" %s✓%s  %d produtos  %s(%.2fs)%s\n\n",
		green, reset, len(products), dim, time.Since(t0).Seconds(), reset)

	if len(products) == 0 {
		fmt.Println(red + "  Nenhum produto encontrado." + reset)
		return
	}

	// 2. Busca de detalhes em paralelo — uma goroutine por produto
	fmt.Printf(bold+"  [2/2] Buscando detalhes de %d produtos em paralelo:\n"+reset, len(products))
	divider()

	ch := make(chan Result, len(products))
	concStart := time.Now()

	for _, p := range products {
		go fetchProduct(p, ch) // disparo simultâneo
	}

	// 3. Coleta pelo canal — ordem = velocidade da API, não da lista
	var results []Result
	for range products {
		r := <-ch // bloqueia até qualquer goroutine enviar
		results = append(results, r)
		printResult(r)
	}

	printSummary(results, time.Since(concStart))
}

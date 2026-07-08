package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"webscraper/mock"
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

// ── Configurações ─────────────────────────────────────────────────────────────
const (
	serverPort   = 8080
	storeTimeout = 2500 * time.Millisecond
	barWidth     = 22
)

// Result carrega o resultado de um único scraping.
type Result struct {
	Store    string
	Price    float64
	Elapsed  time.Duration
	Err      error
	TimedOut bool
}

// scrapePrice demonstra o SELECT do Go: duas "corridas" simultâneas —
// a goroutine interna faz o request HTTP e o time.After marca o timeout.
// Quem chegar primeiro ao select vence; a outra é ignorada.
func scrapePrice(name, url string, ch chan<- Result) {
	start := time.Now()
	done := make(chan Result, 1) // buffer 1 evita goroutine leak

	// Goroutine interna: isolada, faz apenas o request HTTP
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			done <- Result{Store: name, Err: err, Elapsed: time.Since(start)}
			return
		}
		defer resp.Body.Close()

		var data struct {
			Price float64 `json:"price"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			done <- Result{Store: name, Err: err, Elapsed: time.Since(start)}
			return
		}
		done <- Result{Store: name, Price: data.Price, Elapsed: time.Since(start)}
	}()

	// ── SELECT: o coração deste projeto ──────────────────────────────────────
	// Go monitora os dois canais ao mesmo tempo e reage ao primeiro que tiver dado.
	select {
	case r := <-done: // HTTP respondeu dentro do prazo ✓
		ch <- r
	case <-time.After(storeTimeout): // timer disparou antes do HTTP ⚠
		ch <- Result{Store: name, TimedOut: true, Elapsed: storeTimeout,
			Err: fmt.Errorf("sem resposta em %.1fs", storeTimeout.Seconds())}
	}
}

// timeBar renderiza uma barra proporcional a elapsed/total.
func timeBar(elapsed, total time.Duration) string {
	ratio := float64(elapsed) / float64(total)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(barWidth))
	return strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
}

func printHeader() {
	line := strings.Repeat("═", 54)
	fmt.Println(bold + cyan + "╔" + line + "╗" + reset)
	fmt.Println(bold+cyan+"║"+reset + bold+"    Web Scraper Concorrente  —  Go + Channels       "+reset + bold+cyan+"║"+reset)
	fmt.Println(bold+cyan+"║"+reset + dim+"    Goroutines  ·  Channels  ·  Select  ·  Timeout  "+reset + bold+cyan+"║"+reset)
	fmt.Println(bold + cyan + "╚" + line + "╝" + reset)
}

func divider() string {
	return dim + "  " + strings.Repeat("─", 54) + reset
}

func main() {
	enableWindowsANSI()

	go mock.StartServer(serverPort)
	time.Sleep(100 * time.Millisecond)

	printHeader()
	fmt.Println()
	fmt.Printf("  %sProduto%s   iPhone 15 Pro (256GB)\n", bold, reset)
	fmt.Printf("  %sTimeout%s   %.1fs por loja  %s(lojas mais lentas são descartadas)%s\n",
		bold, reset, storeTimeout.Seconds(), dim, reset)
	fmt.Printf("  %sLojas%s     %d consultadas em paralelo\n", bold, reset, len(mock.Stores))
	fmt.Println()
	fmt.Println(dim + "  Disparando " + fmt.Sprintf("%d", len(mock.Stores)) + " goroutines simultaneamente..." + reset)
	fmt.Println()

	// Canal com buffer: goroutines enviam sem bloquear umas nas outras
	ch := make(chan Result, len(mock.Stores))
	start := time.Now()

	// Uma goroutine por loja — tudo dispara ao mesmo tempo
	for _, store := range mock.Stores {
		url := fmt.Sprintf("http://localhost:%d%s", serverPort, store.Path)
		go scrapePrice(store.Name, url, ch)
	}

	fmt.Println(bold + "  Resultados em tempo real  " + dim + "(ordem = velocidade da loja)" + reset)
	fmt.Println(divider())

	var results []Result
	for range mock.Stores {
		r := <-ch // bloqueia até qualquer goroutine enviar — ordem não determinística
		results = append(results, r)
		printResult(r)
	}

	wallTime := time.Since(start)
	printSummary(results, wallTime)
}

func printResult(r Result) {
	bar := timeBar(r.Elapsed, storeTimeout)

	switch {
	case r.TimedOut:
		fmt.Printf("  %s⚠  %-14s  TIMEOUT      %s[%s]%s  %.2fs\n",
			yellow+bold, r.Store, reset+yellow, bar, reset, r.Elapsed.Seconds())
	case r.Err != nil:
		fmt.Printf("  %s✗  %-14s  ERRO         %s[%s]%s  %.2fs\n",
			red+bold, r.Store, reset+red, bar, reset, r.Elapsed.Seconds())
	default:
		fmt.Printf("  %s✓  %-14s  R$ %8.2f  %s[%s]%s  %.2fs\n",
			green+bold, r.Store, r.Price, reset+green, bar, reset, r.Elapsed.Seconds())
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
		if r.Err == nil {
			valid = append(valid, r)
		}
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i].Price < valid[j].Price })

	// Ranking
	fmt.Println()
	fmt.Println(bold + "  Ranking de preços:" + reset)
	fmt.Println(divider())
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
		fmt.Printf("  %s%s %d. %-14s  R$ %.2f%s\n", color, medal, i+1, r.Store, r.Price, reset)
	}

	// Performance
	speedup := serialTime / wallTime.Seconds()
	fmt.Println()
	fmt.Println(bold + "  Performance:" + reset)
	fmt.Println(divider())
	fmt.Printf("  %s⏱  Concorrente:%s    %.2fs\n", cyan+bold, reset, wallTime.Seconds())
	fmt.Printf("  %s🐌 Serial (est.):%s  %.2fs\n", dim, reset, serialTime)
	fmt.Printf("  %s🚀 Speedup:%s        %.1fx mais rápido!\n", green+bold, reset, speedup)
	if timedOut > 0 {
		fmt.Printf("  %s⚠  Timeouts:%s       %d loja(s) descartada(s) pelo select\n", yellow+bold, reset, timedOut)
	}

	// Melhor oferta
	if len(valid) > 0 {
		best, worst := valid[0], valid[len(valid)-1]
		fmt.Println()
		fmt.Println(divider())
		fmt.Printf("  %s★  Melhor oferta:%s  %s — R$ %.2f\n", green+bold, reset, best.Store, best.Price)
		if len(valid) > 1 {
			fmt.Printf("  %s   Economia:%s       R$ %.2f em relação à pior oferta\n", green, reset, worst.Price-best.Price)
		}
	}
	fmt.Println()
}

// enableWindowsANSI habilita o processamento de cores ANSI no console do Windows.
func enableWindowsANSI() {
	if runtime.GOOS != "windows" {
		return
	}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getMode := kernel32.NewProc("GetConsoleMode")
	setMode := kernel32.NewProc("SetConsoleMode")

	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	getMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	setMode.Call(uintptr(handle), uintptr(mode|0x0004)) // ENABLE_VIRTUAL_TERMINAL_PROCESSING
}

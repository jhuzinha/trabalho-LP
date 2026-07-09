// Package report concentra a saída de terminal e a escrita de arquivos.
package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"trabalho-lp/internal/model"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
)

const barWidth = 22

// PrintHeader mostra o contexto do projeto para a apresentação.
func PrintHeader(product string, stores, workers int, timeout time.Duration) {
	line := strings.Repeat("═", 66)
	fmt.Println(bold + cyan + "╔" + line + "╗" + reset)
	fmt.Println(bold + cyan + "║" + reset + bold + "      Comparador de Preços Concorrente em Go                 " + reset + bold + cyan + "║" + reset)
	fmt.Println(bold + cyan + "║" + reset + dim + "      Goroutines · Channels · Worker Pool · Timeout          " + reset + bold + cyan + "║" + reset)
	fmt.Println(bold + cyan + "╚" + line + "╝" + reset)
	fmt.Println()
	fmt.Printf("  %sProduto:%s       %s\n", bold, reset, product)
	fmt.Printf("  %sDomínio:%s       comparação de preços em e-commerce\n", bold, reset)
	fmt.Printf("  %sLojas:%s         %d endpoints simulados\n", bold, reset, stores)
	fmt.Printf("  %sWorkers:%s       %d goroutines consumidoras\n", bold, reset, workers)
	fmt.Printf("  %sTimeout:%s       %.1fs por loja\n", bold, reset, timeout.Seconds())
	fmt.Println()
}

// PrintRun imprime os resultados de uma execução.
func PrintRun(title string, results []model.Result, elapsed time.Duration, timeout time.Duration) {
	fmt.Println(bold + "  " + title + reset)
	fmt.Println(divider())
	for _, result := range results {
		PrintResult(result, timeout)
	}
	fmt.Println(divider())
	fmt.Printf("  %sTempo total:%s %.2fs\n\n", cyan+bold, reset, elapsed.Seconds())
}

// PrintResult imprime uma única loja.
func PrintResult(result model.Result, timeout time.Duration) {
	bar := timeBar(result.Elapsed, timeout)

	switch {
	case result.TimedOut:
		fmt.Printf("  %s⚠  %-15s  TIMEOUT      %s[%s]%s  %.2fs\n",
			yellow+bold, result.Store, reset+yellow, bar, reset, result.Elapsed.Seconds())
	case !result.OK():
		fmt.Printf("  %s✗  %-15s  ERRO         %s[%s]%s  %.2fs  %s\n",
			red+bold, result.Store, reset+red, bar, reset, result.Elapsed.Seconds(), result.ErrorMessage)
	default:
		fmt.Printf("  %s✓  %-15s  R$ %8.2f  %s[%s]%s  %.2fs\n",
			green+bold, result.Store, result.Price, reset+green, bar, reset, result.Elapsed.Seconds())
	}
}

// PrintSummary exibe ranking, economia e comparação de tempo.
func PrintSummary(concurrent []model.Result, concurrentTime time.Duration, sequential []model.Result, sequentialTime time.Duration) {
	valid := validResults(concurrent)
	sort.Slice(valid, func(i, j int) bool { return valid[i].Price < valid[j].Price })

	fmt.Println(bold + "  Ranking de preços válidos:" + reset)
	fmt.Println(divider())
	if len(valid) == 0 {
		fmt.Println(red + "  Nenhuma loja respondeu com preço válido." + reset)
		return
	}

	medals := []string{"🥇", "🥈", "🥉"}
	for i, result := range valid {
		medal := "  "
		if i < len(medals) {
			medal = medals[i]
		}
		color := ""
		suffix := ""
		if i == 0 {
			color = green + bold
			suffix = reset
		}
		fmt.Printf("  %s%s %2d. %-15s R$ %.2f%s\n", color, medal, i+1, result.Store, result.Price, suffix)
	}

	best := valid[0]
	worst := valid[len(valid)-1]

	fmt.Println()
	fmt.Println(bold + "  Performance:" + reset)
	fmt.Println(divider())
	fmt.Printf("  %sConcorrente:%s %.2fs\n", cyan+bold, reset, concurrentTime.Seconds())
	if len(sequential) > 0 {
		speedup := sequentialTime.Seconds() / concurrentTime.Seconds()
		fmt.Printf("  %sSequencial:%s   %.2fs\n", dim, reset, sequentialTime.Seconds())
		fmt.Printf("  %sSpeedup:%s      %.1fx mais rápido\n", green+bold, reset, speedup)
	}
	fmt.Printf("  %sTimeouts:%s     %d loja(s) descartada(s)\n", yellow+bold, reset, countTimeouts(concurrent))

	fmt.Println()
	fmt.Println(divider())
	fmt.Printf("  %sMelhor oferta:%s %s — R$ %.2f\n", green+bold, reset, best.Store, best.Price)
	if len(valid) > 1 {
		fmt.Printf("  %sEconomia:%s      R$ %.2f em relação à pior oferta válida\n", green, reset, worst.Price-best.Price)
	}
	fmt.Println()
}

func divider() string {
	return dim + "  " + strings.Repeat("─", 66) + reset
}

func timeBar(elapsed, total time.Duration) string {
	if total <= 0 {
		return strings.Repeat("░", barWidth)
	}
	ratio := float64(elapsed) / float64(total)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(barWidth))
	return strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
}

func validResults(results []model.Result) []model.Result {
	valid := make([]model.Result, 0, len(results))
	for _, result := range results {
		if result.OK() {
			valid = append(valid, result)
		}
	}
	return valid
}

func countTimeouts(results []model.Result) int {
	total := 0
	for _, result := range results {
		if result.TimedOut {
			total++
		}
	}
	return total
}

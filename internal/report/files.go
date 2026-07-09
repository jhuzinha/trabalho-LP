package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"trabalho-lp/internal/model"
)

// SaveCSV grava os resultados em arquivo CSV.
func SaveCSV(path string, results []model.Result) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"loja", "preco", "tempo_ms", "status_http", "timeout", "erro"}); err != nil {
		return err
	}

	for _, result := range results {
		record := []string{
			result.Store,
			fmt.Sprintf("%.2f", result.Price),
			strconv.FormatInt(result.ElapsedMS, 10),
			strconv.Itoa(result.StatusCode),
			strconv.FormatBool(result.TimedOut),
			result.ErrorMessage,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return writer.Error()
}

// SaveJSON grava os resultados em arquivo JSON indentado.
func SaveJSON(path string, results []model.Result) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

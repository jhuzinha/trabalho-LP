package model

import "time"

// Result armazena o resultado de uma consulta HTTP a uma loja.
type Result struct {
	Store        string        `json:"store"`
	Price        float64       `json:"price"`
	Elapsed      time.Duration `json:"elapsed"`
	ElapsedMS    int64         `json:"elapsed_ms"`
	StatusCode   int           `json:"status_code"`
	TimedOut     bool          `json:"timed_out"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// NewSuccessResult cria um resultado válido.
func NewSuccessResult(store string, price float64, elapsed time.Duration, statusCode int) Result {
	return Result{
		Store:      store,
		Price:      price,
		Elapsed:    elapsed,
		ElapsedMS:  elapsed.Milliseconds(),
		StatusCode: statusCode,
	}
}

// NewErrorResult cria um resultado de erro sem encerrar o programa inteiro.
func NewErrorResult(store string, elapsed time.Duration, statusCode int, message string) Result {
	return Result{
		Store:        store,
		Elapsed:      elapsed,
		ElapsedMS:    elapsed.Milliseconds(),
		StatusCode:   statusCode,
		ErrorMessage: message,
	}
}

// NewTimeoutResult cria um resultado específico para timeout.
func NewTimeoutResult(store string, elapsed time.Duration, message string) Result {
	return Result{
		Store:        store,
		Elapsed:      elapsed,
		ElapsedMS:    elapsed.Milliseconds(),
		TimedOut:     true,
		ErrorMessage: message,
	}
}

// OK indica se a loja respondeu com sucesso e preço válido.
func (r Result) OK() bool {
	return !r.TimedOut && r.ErrorMessage == ""
}

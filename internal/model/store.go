package model

import (
	"errors"
	"strings"
)

// Store representa uma loja que será consultada pelo scraper.
type Store struct {
	Name string
	URL  string
}

// NewStore é o construtor idiomático de Go para Store.
// Ele centraliza a validação e impede que o restante do sistema receba dados inválidos.
func NewStore(name, url string) (Store, error) {
	name = strings.TrimSpace(name)
	url = strings.TrimSpace(url)

	if name == "" {
		return Store{}, errors.New("nome da loja não pode ser vazio")
	}
	if url == "" {
		return Store{}, errors.New("URL da loja não pode ser vazia")
	}

	return Store{Name: name, URL: url}, nil
}

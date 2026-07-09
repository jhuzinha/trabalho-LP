# Web Scraper Concorrente em Go

Trabalho da disciplina de **Linguagens de Programação**: um comparador de preços que
consulta várias lojas **ao mesmo tempo**, demonstrando o modelo de concorrência de Go
(`goroutines`, `channels` e `select`) e comparando-o com C e Python.

## Módulos

| Pasta | O que faz |
|-------|-----------|
| [`webscraper/`](webscraper/) | Demo com **10 lojas simuladas** por um servidor HTTP local (determinística, sem internet). |
| [`webscraper-real/`](webscraper-real/) | Consulta a API pública **DummyJSON** com dados reais. |
| [`python/`](python/) | Mesmo problema em Python (sequencial × threads × pool) para comparação. |

## Documentação

- **Wiki do projeto** — [`docs/wiki.html`](docs/wiki.html)
- **Apresentação (slides)** — [`docs/apresentacao.html`](docs/apresentacao.html)

## Como executar

Requer Go 1.21+ (e Python 3 para o módulo de comparação).

```bash
# módulo 1 — lojas simuladas (não precisa de internet)
cd webscraper && go run .

# módulo 2 — API real (requer internet)
cd webscraper-real && go run .

# módulo 3 — comparação em Python
cd python && python3 comparacao.py

# compilar um binário (ex.: Windows a partir do Linux)
GOOS=windows go build -o webscraper.exe .
```

## Equipe

| Integrante | Parte do código Go | Componente |
|------------|--------------------|------------|
| Gustavo | Servidor mock (`mock/server.go`) | Wiki / documentação |
| Jhuliana | Orquestração concorrente (`main.go`) | Slides / apresentação |
| Marcelo | `select` + timeout (`scrapePrice`) | Comparação com C |
| Enzo | Scraper da API real (`webscraper-real`) | Edição de vídeo |
| Pedro Galeno | Saída no terminal + portabilidade | Comparação Python × Go |

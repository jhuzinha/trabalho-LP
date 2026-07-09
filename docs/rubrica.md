# Rubrica do professor aplicada ao projeto

## Máximo 10 minutos

A apresentação deve focar no essencial: problema, Go, concorrência, comparação com Python, código, demo e conclusão.

## Linguagem: histórico e versão

Falar que Go foi criada no Google, anunciada em 2009 e que Go 1 saiu em 2012. O projeto usa `go 1.21` no `go.mod`.

## Projeto: premissas, usuário e domínio

- Domínio: comparação de preços em e-commerce.
- Usuário: pessoa que quer encontrar a melhor oferta rapidamente.
- Premissas: lojas respondem por HTTP com JSON; algumas lojas são lentas; timeout evita travamento; ranking final mostra a melhor oferta.

## Construtores

Go não tem construtor igual Java/C++. O padrão idiomático é usar funções `NewTipo`.

Exemplos no projeto:

- `model.NewStore`
- `scraper.NewConfig`
- `scraper.NewScraper`
- `mockserver.NewServer`

## Legibilidade

O projeto foi separado em pacotes:

- `cmd/scraper`: entrada da aplicação Go principal;
- `cmd/mockserver`: servidor mock independente para comparação com Python;
- `internal/model`: structs principais;
- `internal/scraper`: lógica de scraping em Go;
- `internal/mockserver`: servidor de lojas simuladas;
- `internal/report`: impressão e arquivos;
- `python`: versão Python para comparação.

## Capacidade de escrita

O programa escreve relatórios em:

- `reports/resultados_concorrentes.csv`
- `reports/resultados_concorrentes.json`
- `reports/python_resultados_concorrentes.csv`
- `reports/python_resultados_concorrentes.json`

## Confiabilidade

Pontos implementados:

- validação nos construtores;
- tratamento explícito de erros;
- validação de `StatusCode` HTTP;
- `context.WithTimeout` para cancelar requisição lenta em Go;
- timeout em `urllib.request.urlopen` no Python;
- `go test ./...`;
- `go test -race ./...`.

## Custo e outros

O projeto usa worker pool em Go. Isso evita criar goroutines ilimitadas e controla o custo de CPU, memória e rede.

Na comparação, Python usa `ThreadPoolExecutor`, que é aceitável para esse problema por ser I/O-bound. Para CPU-bound, Python normalmente sofre mais por causa do GIL, enquanto Go tende a escalar melhor com paralelismo real em múltiplos núcleos.

## Exemplos

Arquivos principais para mostrar:

```text
internal/scraper/scraper.go
python/comparador_python.py
```

No Go, mostrar:

- `RunConcurrent`;
- criação dos workers com `go s.worker(...)`;
- canais `jobs` e `results`;
- `sync.WaitGroup`;
- `context.WithTimeout` dentro de `Fetch`.

No Python, mostrar:

- `ThreadPoolExecutor`;
- `as_completed`;
- timeout no `urlopen`.

## Projeto

O projeto contém:

- servidor mock local;
- coletor principal em Go;
- comparação sequencial vs concorrente;
- coletor equivalente em Python;
- geração de CSV/JSON;
- testes automatizados.

## Site / demonstração

A demo principal é no terminal:

```bash
make compare  # Linux/macOS/Git Bash/WSL
# Windows: .\scripts\comparar_go_python.bat
```

Alternativa só com Go:

```bash
go run ./cmd/scraper
```

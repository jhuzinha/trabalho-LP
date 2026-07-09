# Roteiro de apresentação em até 10 minutos

## 0:00 a 0:40 — Introdução

Este projeto é um comparador de preços concorrente feito principalmente em Go. A ideia é simular várias lojas virtuais e consultar todas ao mesmo tempo para encontrar a melhor oferta.

Também foi criada uma versão em Python para comparar os modelos de concorrência e discutir quando Go vale mais a pena.

## 0:40 a 1:30 — Linguagem Go

Go é uma linguagem criada no Google, anunciada em 2009 e com Go 1 lançado em 2012. Ela é compilada, tipada estaticamente e tem suporte nativo a concorrência com goroutines e channels.

Fala pronta:

> Go foi escolhido porque o projeto envolve várias chamadas HTTP independentes. Esse tipo de problema combina muito com goroutines, channels e worker pool.

## 1:30 a 2:20 — Domínio e usuário

O domínio é e-commerce e comparação de preços. O usuário é alguém que deseja encontrar a melhor oferta rapidamente. A premissa é que cada loja responde por HTTP com JSON, mas algumas lojas podem ser lentas.

## 2:20 a 3:20 — Arquitetura

Mostrar a estrutura:

```text
Servidor mock -> lojas simuladas
Scraper Go -> consulta lojas com worker pool
Coletor Python -> consulta mesmas lojas com ThreadPoolExecutor
Report -> imprime ranking e salva CSV/JSON
```

Fala pronta:

> O servidor mock deixa a demonstração reprodutível, porque não dependemos da internet nem de APIs externas. As lojas têm atrasos fixos, então Go e Python consultam o mesmo cenário.

## 3:20 a 5:10 — Concorrência em Go

Abrir `internal/scraper/scraper.go` e mostrar:

- `RunConcurrent`;
- canal `jobs`;
- canal `results`;
- `go s.worker(...)`;
- `sync.WaitGroup`.

Explicação curta:

> Em vez de consultar uma loja por vez, o programa cria workers em goroutines. Cada worker pega uma loja no canal de jobs, faz a requisição HTTP e envia o resultado para o canal de results.

Explicação importante:

> O worker pool controla o custo. Sem esse limite, se tivéssemos 10 mil lojas, poderíamos criar 10 mil goroutines de uma vez. Com workers, limitamos a concorrência de forma previsível.

## 5:10 a 6:00 — Timeout e confiabilidade

Mostrar `Fetch` e explicar:

> Cada requisição usa `context.WithTimeout`. Se uma loja demora demais, a requisição é cancelada e o programa continua normalmente. Além disso, o código valida status HTTP, JSON e preço.

Também mencionar:

```bash
go test ./...
go test -race ./...
```

## 6:00 a 6:40 — Construtores

Mostrar:

- `NewStore`
- `NewConfig`
- `NewScraper`
- `NewServer`

Fala pronta:

> Go não tem construtor formal como Java. O padrão idiomático é usar funções `NewTipo`, que inicializam e validam as structs.

## 6:40 a 7:40 — Comparação com Python

Abrir `python/comparador_python.py` e mostrar:

- `ThreadPoolExecutor`;
- `fetch_price`;
- `run_concurrent`;
- escrita de CSV/JSON.

Fala pronta:

> A versão Python resolve o mesmo problema usando threads. Como esse problema é I/O-bound, ou seja, passa muito tempo esperando HTTP, Python também consegue ter bom ganho de concorrência.

## 7:40 a 8:50 — Demonstração

Rodar:

```bash
make compare  # Linux/macOS/Git Bash/WSL
# Windows: .\scripts\comparar_go_python.bat
```

Ou manualmente:

```bash
go run ./cmd/scraper
python3 python/comparador_python.py --mock-interno
python3 python/comparador_python.py
```

Mostrar:

- execução concorrente em Go;
- execução sequencial em Go;
- execução concorrente em Python;
- execução sequencial em Python;
- speedup;
- ranking;
- arquivos CSV e JSON.

## 8:50 a 9:30 — Quando usar Go e quando usar Python

Fala pronta:

> A conclusão não é que Go sempre vence. Para scripts rápidos, dados e machine learning, Python pode ser mais produtivo. Mas para serviços concorrentes, rede, infraestrutura e aplicações que precisam lidar com muitas conexões ao mesmo tempo, Go costuma ser uma escolha mais natural.

Resumo:

```text
Go: serviços, concorrência, infraestrutura, APIs, CLIs robustas.
Python: protótipos, automação simples, análise de dados, machine learning.
```

## 9:30 a 10:00 — Conclusão

Fala pronta:

> O projeto mostra que Go facilita a construção de sistemas concorrentes. Com goroutines, channels e worker pool, conseguimos consultar várias lojas ao mesmo tempo, reduzir o tempo total e ainda manter controle de erro, timeout e custo computacional. A comparação com Python mostra que a escolha da linguagem depende do tipo de problema, mas Go se destaca quando concorrência é parte central da solução.

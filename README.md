# Comparador de Preços Concorrente em Go

Projeto de Linguagens de Programação desenvolvido principalmente em **Go** para demonstrar concorrência/paralelismo com **goroutines**, **channels**, **worker pool** e **timeout com context**.

A versão em **Python** foi adicionada como comparação didática. Ela resolve o mesmo problema usando `ThreadPoolExecutor`, permitindo explicar quando Go vale mais a pena e quando Python também atende bem.

## 1. Ideia do projeto

O sistema simula um comparador de preços de e-commerce.

Ele sobe um servidor HTTP local com várias lojas fictícias. Cada loja tem um preço e uma latência fixa. Depois, o scraper consulta essas lojas e monta um ranking com as melhores ofertas.

A parte principal do projeto é comparar quatro execuções:

1. **Go sequencial:** consulta uma loja por vez.
2. **Go concorrente:** consulta várias lojas ao mesmo tempo usando goroutines, channels e worker pool.
3. **Python sequencial:** consulta uma loja por vez.
4. **Python concorrente:** consulta várias lojas ao mesmo tempo usando `ThreadPoolExecutor`.

## 2. Domínio, usuário e premissas

**Domínio:** comparação de preços em e-commerce.

**Usuário:** pessoa que deseja encontrar rapidamente a melhor oferta entre várias lojas.

**Premissas:**

- cada loja expõe um endpoint HTTP que retorna JSON;
- cada resposta contém um preço;
- lojas podem responder rápido, devagar ou estourar timeout;
- respostas lentas não devem travar o programa inteiro;
- o programa deve gerar ranking e salvar os resultados em arquivo;
- Go é a implementação principal;
- Python é usado apenas como comparação entre linguagens/modelos de concorrência.

## 3. Recursos de Go usados

- `goroutine`: execução concorrente dos workers;
- `channel`: comunicação entre fila de jobs, workers e resultados;
- `sync.WaitGroup`: espera segura pelo fim dos workers;
- `context.WithTimeout`: cancelamento real de requisições lentas;
- `struct`: representação de lojas, resultados e configurações;
- construtores `New...`: `NewStore`, `NewConfig`, `NewScraper`, `NewServer`;
- tratamento explícito de erros;
- testes automatizados;
- `go test -race` para verificar race conditions.

## 4. Comparação com Python

A versão Python fica em:

```text
python/comparador_python.py
```

Ela usa apenas biblioteca padrão:

- `urllib.request` para HTTP;
- `ThreadPoolExecutor` para concorrência de I/O;
- `dataclass` para modelar lojas e resultados;
- `csv` e `json` para escrita dos relatórios.

### Ideia da comparação

Para este problema, a maior parte do tempo é gasta esperando resposta HTTP. Esse é um cenário de **I/O-bound**. Em problemas assim, Python com threads pode funcionar bem, porque as threads passam muito tempo esperando rede.

Go continua sendo uma escolha muito forte porque:

- goroutines são leves;
- channels deixam a comunicação entre tarefas mais explícita;
- o runtime de Go foi feito pensando em concorrência;
- o binário final é simples de distribuir;
- o mesmo modelo escala bem para muitos workers, servidores e serviços de rede.

Python pode valer a pena quando:

- o script é pequeno;
- a equipe já domina Python;
- o foco é prototipação rápida;
- existe dependência forte de bibliotecas de dados, automação ou machine learning.

Go tende a valer mais quando:

- o sistema vai rodar em produção por muito tempo;
- precisa lidar com muitas conexões concorrentes;
- precisa de binário simples, desempenho previsível e baixo overhead;
- a concorrência é parte central do problema.

Go **não é sempre melhor**. Para análise de dados, notebooks, machine learning e scripts rápidos, Python muitas vezes é mais produtivo. Para serviços concorrentes, CLIs robustas, servidores e ferramentas de infraestrutura, Go geralmente fica mais interessante.

## 5. Histórico rápido da linguagem Go

Go, também chamada de Golang, foi criada no Google por Robert Griesemer, Rob Pike e Ken Thompson. A linguagem foi anunciada publicamente em 2009 e a versão Go 1 foi lançada em 2012.

A proposta da linguagem é ser simples, compilada, tipada estaticamente, eficiente e com suporte nativo a concorrência. Por isso ela é muito usada em servidores, CLIs, infraestrutura, microsserviços e sistemas de rede.

Este projeto usa `go 1.21` no `go.mod`.

## 6. Estrutura do projeto

```text
trabalho-LP-melhorado/
├── cmd/
│   ├── mockserver/
│   │   └── main.go
│   └── scraper/
│       └── main.go
├── internal/
│   ├── mockserver/
│   │   └── server.go
│   ├── model/
│   │   ├── result.go
│   │   ├── store.go
│   │   └── store_test.go
│   ├── report/
│   │   ├── console.go
│   │   └── files.go
│   └── scraper/
│       ├── scraper.go
│       └── scraper_test.go
├── python/
│   └── comparador_python.py
├── scripts/
│   ├── comparar_go_python.sh
│   ├── comparar_go_python.ps1
│   ├── comparar_go_python.bat
│   ├── testar_windows.ps1
│   └── testar_windows.bat
├── docs/
│   ├── comparacao-go-python.md
│   ├── roteiro-apresentacao.md
│   └── rubrica.md
├── reports/
├── go.mod
├── Makefile
└── README.md
```

## 7. Como executar só a versão Go

Na raiz do projeto:

```bash
go run ./cmd/scraper
```

No Windows PowerShell, o mesmo comando funciona:

```powershell
go run ./cmd/scraper
```

Ou, se tiver `make` instalado no Linux/macOS/Git Bash/WSL:

```bash
make run
```

## 8. Como executar o comparativo Go vs Python

### Windows PowerShell

No Windows, `make` normalmente não vem instalado. Use um destes comandos na raiz do projeto:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\comparar_go_python.ps1
```

ou:

```powershell
.\scripts\comparar_go_python.bat
```

### Linux, macOS, Git Bash ou WSL

```bash
make compare
```

ou:

```bash
./scripts/comparar_go_python.sh
```

Esses scripts:

1. rodam o scraper em Go com servidor mock interno;
2. rodam o coletor em Python com servidor mock interno;
3. usam as mesmas lojas, preços e latências nas duas versões;
4. salvam os relatórios em `reports/`.

Essa forma evita bloqueios do Windows/App Control, porque não usa `Start-Process` para abrir outro executável em segundo plano.

### Opção manual em um terminal

```bash
go run ./cmd/scraper
python3 python/comparador_python.py --mock-interno
```

No Windows, caso `python3` não funcione, use:

```powershell
py -3 python/comparador_python.py --mock-interno
```

ou:

```powershell
python python/comparador_python.py --mock-interno
```

### Opção manual com servidor separado, se quiser demonstrar arquitetura cliente/servidor

Terminal 1:

```bash
go run ./cmd/mockserver
```

Terminal 2:

```bash
go run ./cmd/scraper -usar-servidor-existente
python3 python/comparador_python.py
```

## 9. Parâmetros opcionais

Go:

```bash
go run ./cmd/scraper -workers=5 -timeout=2500ms -porta=18080
```

Python no Linux/macOS/Git Bash/WSL:

```bash
python3 python/comparador_python.py --workers 5 --timeout 2.5 --porta 18080
```

Python no Windows:

```powershell
python python/comparador_python.py --workers 5 --timeout 2.5 --porta 18080
```

Exemplos:

```bash
# Mais workers em Go
go run ./cmd/scraper -workers=10

# Mais workers em Python
python3 python/comparador_python.py --mock-interno --workers 10

# Timeout menor, para forçar mais descartes
go run ./cmd/scraper -timeout=1500ms
python3 python/comparador_python.py --mock-interno --timeout 1.5

# Rodar Go usando servidor mock já aberto
go run ./cmd/scraper -usar-servidor-existente
```

## 10. Como testar

```bash
go test ./...
```

Com detector de race condition:

```bash
go test -race ./...
```

No Windows, também dá para usar:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\testar_windows.ps1
```

ou:

```powershell
.\scripts\testar_windows.bat
```

No Linux/macOS/Git Bash/WSL, se tiver `make`:

```bash
make test
make race
```

## 11. O que a demonstração mostra

A saída mostra:

- consultas concorrentes chegando em ordem diferente;
- lojas descartadas por timeout;
- ranking de melhores preços;
- tempo total concorrente;
- tempo total sequencial;
- speedup obtido com concorrência;
- comparação entre Go e Python;
- CSV e JSON gerados em `reports/`.

## 12. Relação com a rubrica

| Critério | Onde aparece no projeto |
|---|---|
| Linguagem: histórico e versão | README e apresentação |
| Premissas, usuário e domínio | README e roteiro |
| Construtores | `NewStore`, `NewConfig`, `NewScraper`, `NewServer` |
| Legibilidade | Pacotes separados, nomes claros, `gofmt` |
| Capacidade de escrita | Geração de CSV e JSON em Go e Python |
| Confiabilidade | Erros explícitos, HTTP status, timeout real, testes e race detector |
| Custo e outros | Worker pool limita goroutines, CPU, memória e rede |
| Exemplos | Código com goroutines, channels, WaitGroup e ThreadPoolExecutor |
| Projeto | Comparador completo com mock server, relatório e comparação entre linguagens |
| Demonstração | `./scripts/comparar_go_python.bat`, `make compare` ou `go run ./cmd/scraper` |

## 13. Frase curta para explicar o projeto

> O projeto é um comparador de preços concorrente em Go. Ele consulta várias lojas simuladas ao mesmo tempo usando goroutines e channels, controla a quantidade de goroutines com worker pool, cancela lojas lentas com context timeout e compara o resultado com uma versão Python baseada em ThreadPoolExecutor.

## 14. Conclusão para a apresentação

A conclusão não deve ser “Go é sempre melhor”. Uma conclusão mais madura é:

> Go é muito forte quando concorrência, rede e execução em produção são partes centrais do problema. Python também consegue resolver bem esse caso por ser I/O-bound, mas a concorrência em Go é mais natural, leve e integrada à linguagem. Por isso Go costuma ser uma escolha melhor para serviços concorrentes e ferramentas de infraestrutura, enquanto Python continua excelente para prototipação, automação e ciência de dados.

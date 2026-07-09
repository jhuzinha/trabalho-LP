# Comparação entre Go e Python no projeto

## 1. O que está sendo comparado

O projeto compara a coleta de preços em várias lojas simuladas.

Cada loja é um endpoint HTTP local. Algumas respondem rápido, outras demoram e algumas estouram timeout.

As duas linguagens resolvem o mesmo problema:

```text
Lojas HTTP -> coletor concorrente -> ranking de preços -> CSV/JSON
```

## 2. Como Go resolve

A versão Go usa:

- goroutines;
- channels;
- worker pool;
- `sync.WaitGroup`;
- `context.WithTimeout`.

Arquivo principal:

```text
internal/scraper/scraper.go
```

A ideia é criar workers em goroutines. Cada worker recebe lojas pelo canal `jobs`, consulta a loja e envia o resultado pelo canal `results`.

## 3. Como Python resolve

A versão Python usa:

- `ThreadPoolExecutor`;
- `as_completed`;
- timeout no `urllib.request.urlopen`;
- `dataclass`;
- CSV/JSON.

Arquivo principal:

```text
python/comparador_python.py
```

A ideia é criar um pool de threads. Cada thread executa uma requisição HTTP e devolve o resultado.

## 4. Resultado esperado

Como o problema é I/O-bound, ou seja, passa muito tempo esperando rede/HTTP, Python também consegue ganhar bastante desempenho com threads.

Nesse tipo de problema, a diferença de tempo entre Go e Python pode não ser enorme, porque o gargalo principal não é cálculo pesado de CPU, e sim espera de resposta.

Mesmo assim, Go continua tendo vantagens importantes:

- goroutines são mais leves que threads tradicionais;
- channels fazem parte da linguagem;
- o runtime de Go foi desenhado para concorrência;
- `context` facilita cancelamento e timeout;
- o binário final é simples de distribuir;
- tende a escalar melhor em serviços concorrentes reais.

## 5. Quando Go vale mais a pena

Go tende a valer mais quando:

- concorrência é parte central do sistema;
- existem muitas conexões simultâneas;
- o projeto é um servidor, API, proxy, crawler, CLI robusta ou ferramenta de infraestrutura;
- você quer compilar e distribuir um binário único;
- precisa de desempenho previsível e baixo overhead.

## 6. Quando Python pode ser melhor

Python pode ser melhor quando:

- o projeto é pequeno ou exploratório;
- a equipe já conhece Python;
- o foco é análise de dados, automação ou machine learning;
- existem bibliotecas Python que resolvem grande parte do problema;
- produtividade inicial importa mais que performance e distribuição.

## 7. Go é sempre melhor?

Não.

A conclusão correta é mais equilibrada:

> Go não é sempre melhor que Python. A escolha depende do problema. Para scripts rápidos, dados e machine learning, Python costuma ser mais produtivo. Para serviços concorrentes, infraestrutura e aplicações de rede, Go normalmente é uma escolha mais natural.

## 8. Frase pronta para a apresentação

> Neste projeto, tanto Go quanto Python conseguem melhorar o tempo da coleta porque o problema é I/O-bound. A diferença é que Go oferece concorrência como parte central da linguagem, com goroutines, channels e context, enquanto Python resolve usando uma biblioteca de threads. Por isso Go é mais interessante quando concorrência e rede são o núcleo da aplicação, mas Python ainda pode ser mais produtivo em scripts, protótipos e análise de dados.

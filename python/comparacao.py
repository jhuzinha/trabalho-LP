# -*- coding: utf-8 -*-
import io
import sys
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8")

"""
Web Scraper Concorrente — Comparação Python vs Go
Disciplina: Linguagens de Programação

Três implementações do mesmo problema:
  1. Sequencial      — uma requisição por vez (baseline)
  2. Threads + Queue — equivalente Python ao Go (goroutines + channels)
  3. ThreadPoolExecutor — versão mais moderna/idiomática do Python

API: dummyjson.com (pública, sem autenticação)
"""

import json
import queue
import threading
import time
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from urllib.parse import quote

# ── Cores ANSI ────────────────────────────────────────────────────────────────
RESET  = "\033[0m"
BOLD   = "\033[1m"
DIM    = "\033[2m"
RED    = "\033[31m"
GREEN  = "\033[32m"
YELLOW = "\033[33m"
CYAN   = "\033[36m"

TIMEOUT  = 6.0
API_BASE = "https://dummyjson.com"
QUERY    = "iphone"
LIMIT    = 8

# ── Helpers ───────────────────────────────────────────────────────────────────

def time_bar(elapsed: float, total: float, width: int = 16) -> str:
    ratio = min(elapsed / total, 1.0)
    filled = int(ratio * width)
    return "█" * filled + "░" * (width - filled)

def truncate(s: str, max_len: int) -> str:
    return s if len(s) <= max_len else s[:max_len - 3] + "..."

def divider():
    print(DIM + "  " + "─" * 62 + RESET)

def final_price(p: dict) -> float:
    return p["price"] * (1 - p.get("discountPercentage", 0) / 100)

# ── Funções de acesso à API ───────────────────────────────────────────────────

HEADERS = {"User-Agent": "Mozilla/5.0 (compatible; scraper-demo/1.0)"}

def _get(url: str, timeout: float) -> dict:
    req = urllib.request.Request(url, headers=HEADERS)
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read())

def search_products(query: str, limit: int) -> list[dict]:
    url = f"{API_BASE}/products/search?q={quote(query)}&limit={limit}"
    return _get(url, timeout=10)["products"]

def fetch_product(product: dict) -> dict:
    """Busca detalhes de um produto. Retorna dict com ok, product, elapsed."""
    start = time.perf_counter()
    try:
        url = f"{API_BASE}/products/{product['id']}"
        detail = _get(url, timeout=TIMEOUT)
        return {"ok": True, "product": detail, "elapsed": time.perf_counter() - start}
    except Exception as e:
        return {"ok": False, "product": product, "elapsed": time.perf_counter() - start, "error": str(e)}

# ── Implementação 1: Sequencial ───────────────────────────────────────────────

def run_sequencial(products: list[dict]) -> tuple[list[dict], float]:
    """
    Python sequencial — uma requisição por vez.
    Tempo total ≈ soma de todos os tempos individuais.
    """
    results = []
    start = time.perf_counter()
    for p in products:
        r = fetch_product(p)
        results.append(r)
        name = truncate(f"{r['product'].get('brand','')} {r['product']['title']}", 38)
        bar  = time_bar(r["elapsed"], TIMEOUT)
        if r["ok"]:
            print(f"  {GREEN}✓{RESET}  {name:<38}  ${final_price(r['product']):7.2f}"
                  f"  {DIM}[{bar}]{RESET}  {r['elapsed']:.2f}s")
        else:
            print(f"  {RED}✗  {name:<38}  ERRO{RESET}  {r['elapsed']:.2f}s")
    return results, time.perf_counter() - start

# ── Implementação 2: Threads + Queue ─────────────────────────────────────────

def _thread_worker(product: dict, q: queue.Queue) -> None:
    """Worker que roda em uma thread separada e coloca o resultado na Queue."""
    q.put(fetch_product(product))

def run_threads(products: list[dict]) -> tuple[list[dict], float]:
    """
    Python com threading.Thread + queue.Queue.

    Equivalência com Go:
      threading.Thread  ≈  goroutine
      queue.Queue       ≈  channel
      q.get()           ≈  <-ch   (bloqueia até ter dado)

    Limitação: o GIL (Global Interpreter Lock) do Python impede
    paralelismo real em CPU. Para I/O (HTTP), o GIL é liberado —
    então threads funcionam bem aqui, mas são mais pesadas que goroutines.
    """
    q: queue.Queue = queue.Queue()
    start = time.perf_counter()

    # Dispara uma thread por produto — equivalente ao "go scrapePrice(...)"
    for p in products:
        t = threading.Thread(target=_thread_worker, args=(p, q), daemon=True)
        t.start()

    # Coleta resultados conforme chegam — ordem não determinística
    results = []
    for _ in products:
        r = q.get()  # bloqueia até qualquer thread enviar — igual ao <-ch do Go
        results.append(r)
        name = truncate(f"{r['product'].get('brand','')} {r['product']['title']}", 38)
        bar  = time_bar(r["elapsed"], TIMEOUT)
        if r["ok"]:
            print(f"  {GREEN}✓{RESET}  {name:<38}  ${final_price(r['product']):7.2f}"
                  f"  {DIM}[{bar}]{RESET}  {r['elapsed']:.2f}s")
        else:
            print(f"  {RED}✗  {name:<38}  ERRO{RESET}  {r['elapsed']:.2f}s")

    return results, time.perf_counter() - start

# ── Implementação 3: ThreadPoolExecutor ──────────────────────────────────────

def run_executor(products: list[dict]) -> tuple[list[dict], float]:
    """
    Python moderno com concurrent.futures.ThreadPoolExecutor.
    Mais idiomático que Thread + Queue, mas menos explícito que o Go.
    """
    results = []
    start = time.perf_counter()

    with ThreadPoolExecutor(max_workers=len(products)) as pool:
        futures = {pool.submit(fetch_product, p): p for p in products}
        for future in as_completed(futures):  # yield conforme ficam prontos
            r = future.result()
            results.append(r)
            name = truncate(f"{r['product'].get('brand','')} {r['product']['title']}", 38)
            bar  = time_bar(r["elapsed"], TIMEOUT)
            if r["ok"]:
                print(f"  {GREEN}✓{RESET}  {name:<38}  ${final_price(r['product']):7.2f}"
                      f"  {DIM}[{bar}]{RESET}  {r['elapsed']:.2f}s")
            else:
                print(f"  {RED}✗  {name:<38}  ERRO{RESET}  {r['elapsed']:.2f}s")

    return results, time.perf_counter() - start

# ── Comparação final ──────────────────────────────────────────────────────────

def print_comparison(t_seq: float, t_threads: float, t_exec: float, t_go_ref: float = 0.29):
    print(f"\n{BOLD}  Comparação de Performance:{RESET}")
    divider()

    rows = [
        ("Python Sequencial",        t_seq,     "❌ sem concorrência"),
        ("Python Thread + Queue",    t_threads, "⚠️  GIL libera só em I/O"),
        ("Python ThreadPoolExecutor",t_exec,    "⚠️  GIL libera só em I/O"),
        ("Go Goroutines + Channels", t_go_ref,  "✅ paralelismo real, 2KB/goroutine"),
    ]

    print(f"  {'Implementação':<30}  {'Tempo':>7}  {'Speedup vs seq':>14}  Observação")
    divider()
    for name, t, obs in rows:
        speedup = t_seq / t if t > 0 else 0
        bar_color = GREEN if "Go" in name else (YELLOW if t < t_seq else RED)
        print(f"  {bar_color}{name:<30}{RESET}  {t:>6.2f}s  {speedup:>13.1f}x  {DIM}{obs}{RESET}")

    print(f"""
{BOLD}  Diferenças fundamentais Go vs Python:{RESET}
{divider.__doc__ or ""}
  {CYAN}Go{RESET}
    go fetchProduct(p, ch)     {DIM}// goroutine: 2KB de memória, roda em múltiplos cores{RESET}
    r := <-ch                  {DIM}// channel: tipo nativo da linguagem{RESET}
    select {{ case r := <-ch:   {DIM}// select: monitora múltiplos canais{RESET}
             case <-time.After(...): }}

  {YELLOW}Python{RESET}
    t = threading.Thread(...)  {DIM}# thread do SO: ~1MB de memória{RESET}
    t.start()
    r = q.get()                {DIM}# Queue: biblioteca padrão{RESET}
    {DIM}# timeout: q.get(timeout=N) — menos elegante que select{RESET}
    {DIM}# GIL: só uma thread executa Python por vez (I/O libera){RESET}
""")

# ── Main ──────────────────────────────────────────────────────────────────────

def main():
    print(BOLD + CYAN + "╔══════════════════════════════════════════════════════════════╗" + RESET)
    print(BOLD + CYAN + "║" + RESET + BOLD + "   Web Scraper — Comparação Python vs Go                     " + BOLD + CYAN + "║" + RESET)
    print(BOLD + CYAN + "║" + RESET + DIM  + "   Sequencial  ·  Threads+Queue  ·  ThreadPoolExecutor       " + BOLD + CYAN + "║" + RESET)
    print(BOLD + CYAN + "╚══════════════════════════════════════════════════════════════╝" + RESET)
    print()

    # Busca inicial
    print(DIM + f"  Buscando {LIMIT} produtos ({QUERY!r}) em dummyjson.com..." + RESET, end="", flush=True)
    t0 = time.perf_counter()
    products = search_products(QUERY, LIMIT)
    print(f" {GREEN}✓{RESET}  {len(products)} produtos  {DIM}({time.perf_counter()-t0:.2f}s){RESET}\n")

    if not products:
        print(RED + "  Nenhum produto encontrado." + RESET)
        return

    # ── 1. Sequencial ─────────────────────────────────────────────────────────
    print(BOLD + "  [1/3] Sequencial (sem concorrência):" + RESET)
    divider()
    _, t_seq = run_sequencial(products)
    print(f"\n  {RED}Tempo total: {t_seq:.2f}s{RESET}  {DIM}(soma de todas as requisições){RESET}\n")

    # ── 2. Threads + Queue ────────────────────────────────────────────────────
    print(BOLD + "  [2/3] Threads + Queue  (equivalente Python ao goroutine+channel):" + RESET)
    divider()
    _, t_threads = run_threads(products)
    print(f"\n  {YELLOW}Tempo total: {t_threads:.2f}s{RESET}  {DIM}({t_seq/t_threads:.1f}x mais rápido que sequencial){RESET}\n")

    # ── 3. ThreadPoolExecutor ─────────────────────────────────────────────────
    print(BOLD + "  [3/3] ThreadPoolExecutor  (idiomático Python moderno):" + RESET)
    divider()
    _, t_exec = run_executor(products)
    print(f"\n  {YELLOW}Tempo total: {t_exec:.2f}s{RESET}  {DIM}({t_seq/t_exec:.1f}x mais rápido que sequencial){RESET}\n")

    # ── Comparação final ──────────────────────────────────────────────────────
    print_comparison(t_seq, t_threads, t_exec)

if __name__ == "__main__":
    main()

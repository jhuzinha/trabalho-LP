#!/usr/bin/env python3
"""
Coletor de preços em Python para comparação didática com a versão em Go.

A implementação usa apenas biblioteca padrão:
- urllib.request para HTTP;
- concurrent.futures.ThreadPoolExecutor para concorrência de I/O;
- http.server para subir um servidor mock interno quando necessário;
- csv/json para escrita dos relatórios.

Importante: este arquivo não substitui o projeto principal em Go. Ele serve para
mostrar, na apresentação, como o mesmo problema pode ser resolvido em Python e
quais diferenças aparecem em sintaxe, modelo de concorrência, custo e execução.
"""

from __future__ import annotations

import argparse
import csv
import json
import socket
import sys
import threading
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import asdict, dataclass
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any
from pathlib import Path
from typing import Iterable, Optional


@dataclass(frozen=True)
class Store:
    name: str
    url: str


@dataclass(frozen=True)
class StoreConfig:
    path: str
    name: str
    price: float
    delay_seconds: float


@dataclass
class Result:
    store: str
    price: float
    elapsed_ms: int
    status_code: int = 0
    timed_out: bool = False
    error_message: str = ""

    @property
    def ok(self) -> bool:
        return not self.timed_out and self.error_message == ""


def default_store_configs() -> list[StoreConfig]:
    # Mesmas lojas, preços e atrasos usados pelo servidor mock em Go.
    return [
        StoreConfig("/amazon", "Amazon", 1299.90, 0.9),
        StoreConfig("/mercado-livre", "Mercado Livre", 1249.00, 1.6),
        StoreConfig("/americanas", "Americanas", 1399.00, 1.1),
        StoreConfig("/magazine-luiza", "Magazine Luiza", 1279.90, 0.5),
        StoreConfig("/casas-bahia", "Casas Bahia", 1350.00, 2.1),
        StoreConfig("/shopee", "Shopee", 1199.00, 0.8),
        StoreConfig("/kabum", "KaBuM!", 1280.00, 0.6),
        StoreConfig("/submarino", "Submarino", 1320.00, 2.3),
        StoreConfig("/aliexpress", "AliExpress", 1100.00, 3.0),
        StoreConfig("/carrefour", "Carrefour", 1310.00, 2.8),
    ]


def default_stores(port: int) -> list[Store]:
    base_url = f"http://localhost:{port}"
    return [Store(config.name, f"{base_url}{config.path}") for config in default_store_configs()]


# Em Windows, quando o cliente aplica timeout e fecha a conexão antes de o
# servidor mock responder, o http.server pode registrar uma exceção como
# ConnectionAbortedError/ConnectionResetError/BrokenPipeError. Para a demonstração,
# isso é esperado: representa justamente uma loja lenta sendo descartada pelo cliente.
_EXPECTED_DISCONNECT_ERRORS = (BrokenPipeError, ConnectionAbortedError, ConnectionResetError)


class QuietThreadingHTTPServer(ThreadingHTTPServer):
    daemon_threads = True

    def handle_error(self, request: Any, client_address: object) -> None:
        # Evita stack traces no terminal quando o cliente fecha a conexão por timeout.
        exc = sys.exc_info()[1]
        if isinstance(exc, _EXPECTED_DISCONNECT_ERRORS):
            return
        super().handle_error(request, client_address)


def make_mock_handler(configs: list[StoreConfig]) -> type[BaseHTTPRequestHandler]:
    routes = {config.path: config for config in configs}

    class MockHandler(BaseHTTPRequestHandler):
        def log_message(self, format: str, *args: object) -> None:  # noqa: A002
            # Evita poluir a demonstração no terminal.
            return

        def _send_json(self, status_code: int, payload: object) -> None:
            body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
            try:
                self.send_response(status_code)
                self.send_header("Content-Type", "application/json; charset=utf-8")
                self.send_header("Content-Length", str(len(body)))
                self.end_headers()
                self.wfile.write(body)
            except _EXPECTED_DISCONNECT_ERRORS:
                # Quando o cliente aplica timeout, ele fecha a conexão antes de a loja lenta
                # responder. Isso é esperado na demonstração e não deve poluir o terminal.
                return

        def do_GET(self) -> None:  # noqa: N802 - nome exigido pelo BaseHTTPRequestHandler.
            if self.path == "/health":
                self._send_json(200, {"status": "ok"})
                return

            if self.path == "/stores":
                self._send_json(200, [asdict(config) for config in configs])
                return

            config = routes.get(self.path)
            if config is None:
                self._send_json(404, {"error": "loja não encontrada"})
                return

            time.sleep(config.delay_seconds)
            self._send_json(200, {"store": config.name, "price": config.price})

    return MockHandler


def start_internal_mock_server(port: int) -> QuietThreadingHTTPServer:
    handler = make_mock_handler(default_store_configs())
    server = QuietThreadingHTTPServer(("127.0.0.1", port), handler)
    thread = threading.Thread(target=server.serve_forever, name="python-mock-server", daemon=True)
    thread.start()
    wait_for_server(port, timeout_seconds=2.0)
    return server


def wait_for_server(port: int, timeout_seconds: float = 2.0) -> None:
    health_url = f"http://localhost:{port}/health"
    deadline = time.perf_counter() + timeout_seconds
    last_error: Optional[Exception] = None

    while time.perf_counter() < deadline:
        try:
            with urllib.request.urlopen(health_url, timeout=0.2) as response:
                if response.status == 200:
                    return
        except Exception as exc:  # noqa: BLE001 - aqui queremos mostrar erro claro ao usuário.
            last_error = exc
        time.sleep(0.05)

    raise RuntimeError(
        f"não encontrei o servidor mock em {health_url}. "
        "Use --mock-interno ou rode: go run ./cmd/mockserver"
    ) from last_error


def fetch_price(store: Store, timeout_seconds: float) -> Result:
    start = time.perf_counter()
    request = urllib.request.Request(
        store.url,
        headers={"User-Agent": "trabalho-lp-python-scraper/1.0"},
        method="GET",
    )

    try:
        with urllib.request.urlopen(request, timeout=timeout_seconds) as response:
            elapsed_ms = int((time.perf_counter() - start) * 1000)
            if response.status != 200:
                return Result(
                    store=store.name,
                    price=0.0,
                    elapsed_ms=elapsed_ms,
                    status_code=response.status,
                    error_message=f"status HTTP inválido: {response.status}",
                )

            payload = json.loads(response.read().decode("utf-8"))
            price = float(payload.get("price", 0.0))
            if price <= 0:
                return Result(
                    store=store.name,
                    price=0.0,
                    elapsed_ms=elapsed_ms,
                    status_code=response.status,
                    error_message="preço inválido ou ausente",
                )

            return Result(
                store=store.name,
                price=price,
                elapsed_ms=elapsed_ms,
                status_code=response.status,
            )
    except (TimeoutError, socket.timeout) as exc:
        elapsed_ms = int((time.perf_counter() - start) * 1000)
        return Result(
            store=store.name,
            price=0.0,
            elapsed_ms=elapsed_ms,
            timed_out=True,
            error_message=f"sem resposta em {timeout_seconds:.1f}s: {exc}",
        )
    except urllib.error.URLError as exc:
        elapsed_ms = int((time.perf_counter() - start) * 1000)
        reason = getattr(exc, "reason", exc)
        timed_out = isinstance(reason, TimeoutError) or "timed out" in str(reason).lower()
        return Result(
            store=store.name,
            price=0.0,
            elapsed_ms=elapsed_ms,
            timed_out=timed_out,
            error_message=f"erro HTTP: {reason}",
        )
    except Exception as exc:  # noqa: BLE001 - erro é convertido para resultado, sem derrubar toda a coleta.
        elapsed_ms = int((time.perf_counter() - start) * 1000)
        return Result(
            store=store.name,
            price=0.0,
            elapsed_ms=elapsed_ms,
            error_message=str(exc),
        )


def run_sequential(stores: Iterable[Store], timeout_seconds: float) -> tuple[list[Result], float]:
    start = time.perf_counter()
    results = [fetch_price(store, timeout_seconds) for store in stores]
    return results, time.perf_counter() - start


def run_concurrent(
    stores: list[Store], timeout_seconds: float, workers: int
) -> tuple[list[Result], float]:
    start = time.perf_counter()
    results: list[Result] = []

    # ThreadPoolExecutor é adequado para I/O em Python, porque as threads ficam
    # esperando rede/HTTP na maior parte do tempo. Para CPU pesada, o GIL muda a análise.
    with ThreadPoolExecutor(max_workers=workers) as executor:
        futures = [executor.submit(fetch_price, store, timeout_seconds) for store in stores]
        for future in as_completed(futures):
            results.append(future.result())

    return results, time.perf_counter() - start


def sorted_results(results: list[Result]) -> list[Result]:
    return sorted(results, key=lambda r: (not r.ok, r.price if r.ok else float("inf")))


def print_run(title: str, results: list[Result], elapsed_seconds: float) -> None:
    print(f"\n=== {title} ===")
    print(f"Tempo total: {elapsed_seconds:.2f}s\n")

    for result in sorted_results(results):
        if result.ok:
            print(
                f"{result.store:<18} R$ {result.price:>8.2f} "
                f"{result.elapsed_ms / 1000:>5.2f}s"
            )
        elif result.timed_out:
            print(f"{result.store:<18} TIMEOUT   {result.elapsed_ms / 1000:>5.2f}s")
        else:
            print(f"{result.store:<18} ERRO      {result.error_message}")


def print_summary(
    concurrent_results: list[Result],
    concurrent_time: float,
    sequential_results: list[Result],
    sequential_time: float,
) -> None:
    print("\n=== Resumo Python ===")
    ok_results = [r for r in concurrent_results if r.ok]
    if ok_results:
        best = min(ok_results, key=lambda r: r.price)
        print(f"Melhor oferta: {best.store} por R$ {best.price:.2f}")
    print(f"Tempo concorrente: {concurrent_time:.2f}s")

    if sequential_results:
        print(f"Tempo sequencial:  {sequential_time:.2f}s")
        if concurrent_time > 0:
            print(f"Speedup:           {sequential_time / concurrent_time:.2f}x")


def save_reports(output_dir: Path, results: list[Result]) -> None:
    output_dir.mkdir(parents=True, exist_ok=True)
    csv_path = output_dir / "python_resultados_concorrentes.csv"
    json_path = output_dir / "python_resultados_concorrentes.json"

    rows = [asdict(result) for result in sorted_results(results)]

    with csv_path.open("w", newline="", encoding="utf-8") as csv_file:
        writer = csv.DictWriter(
            csv_file,
            fieldnames=["store", "price", "elapsed_ms", "status_code", "timed_out", "error_message"],
        )
        writer.writeheader()
        writer.writerows(rows)

    with json_path.open("w", encoding="utf-8") as json_file:
        json.dump(rows, json_file, ensure_ascii=False, indent=2)

    print("\nRelatórios Python salvos em:")
    print(f"- {csv_path}")
    print(f"- {json_path}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Coletor Python para comparação com Go")
    parser.add_argument("--porta", type=int, default=18080, help="porta do servidor mock")
    parser.add_argument("--workers", type=int, default=5, help="quantidade de threads no pool")
    parser.add_argument("--timeout", type=float, default=2.5, help="timeout por loja em segundos")
    parser.add_argument("--saida", default="reports", help="pasta para salvar CSV e JSON")
    parser.add_argument(
        "--sem-sequencial",
        action="store_true",
        help="executa apenas a versão concorrente em Python",
    )
    parser.add_argument(
        "--mock-interno",
        action="store_true",
        help="sobe um servidor mock interno em Python, sem depender do servidor Go separado",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    if args.workers <= 0:
        raise SystemExit("--workers deve ser maior que zero")
    if args.timeout <= 0:
        raise SystemExit("--timeout deve ser maior que zero")

    server: Optional[QuietThreadingHTTPServer] = None
    try:
        if args.mock_interno:
            print(f"Subindo servidor mock interno em Python na porta {args.porta}...")
            server = start_internal_mock_server(args.porta)
        else:
            wait_for_server(args.porta)

        stores = default_stores(args.porta)
        real_workers = min(args.workers, len(stores))

        print("=== Coletor Python ===")
        print(f"Lojas: {len(stores)}")
        print(f"Threads/workers: {real_workers}")
        print(f"Timeout por loja: {args.timeout:.1f}s")

        concurrent_results, concurrent_time = run_concurrent(stores, args.timeout, real_workers)
        print_run("Execução concorrente com ThreadPoolExecutor", concurrent_results, concurrent_time)

        sequential_results: list[Result] = []
        sequential_time = 0.0
        if not args.sem_sequencial:
            sequential_results, sequential_time = run_sequential(stores, args.timeout)
            print_run("Execução sequencial", sequential_results, sequential_time)

        print_summary(concurrent_results, concurrent_time, sequential_results, sequential_time)
        save_reports(Path(args.saida), concurrent_results)
    finally:
        if server is not None:
            server.shutdown()
            server.server_close()


if __name__ == "__main__":
    main()

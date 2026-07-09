#!/usr/bin/env bash
set -euo pipefail

PORTA="${PORTA:-18080}"
WORKERS="${WORKERS:-5}"
TIMEOUT_GO="${TIMEOUT_GO:-2500ms}"
TIMEOUT_PYTHON="${TIMEOUT_PYTHON:-2.5}"

cd "$(dirname "$0")/.."
mkdir -p reports

echo "==================== GO ===================="
echo "Rodando Go com servidor mock interno..."
go run ./cmd/scraper -porta="$PORTA" -workers="$WORKERS" -timeout="$TIMEOUT_GO"

echo
echo "================== PYTHON =================="
echo "Rodando Python com servidor mock interno..."
if command -v python3 >/dev/null 2>&1; then
  python3 python/comparador_python.py --mock-interno --porta "$PORTA" --workers "$WORKERS" --timeout "$TIMEOUT_PYTHON"
elif command -v python >/dev/null 2>&1; then
  python python/comparador_python.py --mock-interno --porta "$PORTA" --workers "$WORKERS" --timeout "$TIMEOUT_PYTHON"
else
  echo "Python 3 não encontrado." >&2
  exit 1
fi

echo
echo "Comparação finalizada. Veja a pasta reports/ para CSV e JSON."

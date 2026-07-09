$ErrorActionPreference = "Stop"
$ProjectRoot = Resolve-Path (Join-Path $PSScriptRoot "..")

function Test-CommandExists {
    param([string]$Command)
    return $null -ne (Get-Command $Command -ErrorAction SilentlyContinue)
}

function Test-NativeCommand {
    param(
        [string]$Command,
        [string[]]$Arguments
    )

    if (-not (Test-CommandExists $Command)) {
        return $false
    }

    try {
        & $Command @Arguments *> $null
        return $LASTEXITCODE -eq 0
    } catch {
        return $false
    }
}

function Invoke-PythonCompile {
    if (Test-NativeCommand "py" @("-3", "--version")) {
        & py -3 -m py_compile python/comparador_python.py
        return
    }

    if (Test-NativeCommand "python" @("--version")) {
        & python -m py_compile python/comparador_python.py
        return
    }

    if (Test-NativeCommand "python3" @("--version")) {
        & python3 -m py_compile python/comparador_python.py
        return
    }

    throw "Python não foi encontrado. Instale o Python 3 e marque 'Add python.exe to PATH', ou use o Python Launcher py."
}

Push-Location $ProjectRoot
try {
    Write-Host "Rodando go test ./..." -ForegroundColor Cyan
    & go test ./...

    Write-Host ""
    Write-Host "Rodando go test -race ./..." -ForegroundColor Cyan
    & go test -race ./...

    Write-Host ""
    Write-Host "Validando sintaxe do Python" -ForegroundColor Cyan
    Invoke-PythonCompile

    Write-Host ""
    Write-Host "Tudo certo." -ForegroundColor Green
} finally {
    Pop-Location
}

param(
    [int]$Porta = 18080,
    [int]$Workers = 5,
    [string]$TimeoutGo = "2500ms",
    [double]$TimeoutPython = 2.5
)

$ErrorActionPreference = "Stop"

$ProjectRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$ReportsDir = Join-Path $ProjectRoot "reports"
New-Item -ItemType Directory -Force -Path $ReportsDir | Out-Null

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

    $oldPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        & $Command @Arguments *> $null
        return $LASTEXITCODE -eq 0
    } catch {
        return $false
    } finally {
        $ErrorActionPreference = $oldPreference
    }
}

function Invoke-PythonScript {
    param(
        [string]$ScriptPath,
        [int]$Porta,
        [int]$Workers,
        [double]$TimeoutPython
    )

    $scriptArgs = @(
        "-u",
        $ScriptPath,
        "--mock-interno",
        "--porta", "$Porta",
        "--workers", "$Workers",
        "--timeout", "$TimeoutPython"
    )

    # No Windows, o comando "python" às vezes é só um atalho da Microsoft Store.
    # Por isso tentamos primeiro o launcher oficial "py -3".
    if (Test-NativeCommand "py" @("-3", "--version")) {
        & py -3 @scriptArgs
        return
    }

    if (Test-NativeCommand "python" @("--version")) {
        & python @scriptArgs
        return
    }

    if (Test-NativeCommand "python3" @("--version")) {
        & python3 @scriptArgs
        return
    }

    throw "Python não foi encontrado. Instale o Python 3 em https://www.python.org/downloads/ e marque a opção 'Add python.exe to PATH', ou instale o Python Launcher para usar o comando py."
}

if (-not (Test-CommandExists "go")) {
    throw "Go não foi encontrado no PATH. Instale o Go ou reinicie o terminal depois da instalação."
}

Push-Location $ProjectRoot
try {
    Write-Host "==================== GO ====================" -ForegroundColor Green
    $goArgs = @(
        "run",
        "./cmd/scraper",
        "-porta=$($Porta)",
        "-workers=$($Workers)",
        "-timeout=$TimeoutGo"
    )
    & go @goArgs

    Write-Host ""
    Write-Host "================== PYTHON ==================" -ForegroundColor Green
    Write-Host "Rodando Python com servidor mock interno, sem Start-Process..." -ForegroundColor Cyan
    Invoke-PythonScript `
        -ScriptPath "python/comparador_python.py" `
        -Porta $Porta `
        -Workers $Workers `
        -TimeoutPython $TimeoutPython

    Write-Host ""
    Write-Host "Comparação finalizada. Veja a pasta reports/ para CSV e JSON." -ForegroundColor Cyan
} finally {
    Pop-Location
}

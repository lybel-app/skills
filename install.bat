@echo off
REM ============================================================
REM  Lybel Skills - Windows Installer
REM  Baixa e instala a skill lybel-docs no Claude Desktop/Code
REM ============================================================

setlocal enabledelayedexpansion

echo.
echo ============================================================
echo   Instalador Lybel Skills (lybel-docs)
echo ============================================================
echo.

REM Checa se PowerShell esta disponivel
where powershell >nul 2>&1
if errorlevel 1 (
    echo [ERRO] PowerShell nao encontrado. Este instalador precisa do PowerShell.
    echo        O PowerShell vem instalado por padrao no Windows 10/11.
    echo.
    pause
    exit /b 1
)

REM Executa o bloco PowerShell principal
powershell -NoProfile -ExecutionPolicy Bypass -Command ^
    "$ErrorActionPreference = 'Stop';" ^
    "try {" ^
        "$InstallDir = Join-Path $env:USERPROFILE '.claude\skills\lybel-docs';" ^
        "$TempZip = Join-Path $env:TEMP 'lybel-docs.zip';" ^
        "$TempExtract = Join-Path $env:TEMP 'lybel-docs-extract';" ^
        "$Url = 'https://github.com/lybel-app/skills/releases/latest/download/lybel-docs-windows-amd64.zip';" ^
        "Write-Host '[1/4] Preparando diretorio de instalacao...' -ForegroundColor Cyan;" ^
        "if (Test-Path $InstallDir) { Remove-Item $InstallDir -Recurse -Force };" ^
        "New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null;" ^
        "Write-Host '      -> ' $InstallDir -ForegroundColor DarkGray;" ^
        "Write-Host '[2/4] Baixando ultima versao...' -ForegroundColor Cyan;" ^
        "Write-Host '      -> ' $Url -ForegroundColor DarkGray;" ^
        "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12;" ^
        "Invoke-WebRequest -Uri $Url -OutFile $TempZip -UseBasicParsing;" ^
        "Write-Host '[3/4] Extraindo arquivos...' -ForegroundColor Cyan;" ^
        "if (Test-Path $TempExtract) { Remove-Item $TempExtract -Recurse -Force };" ^
        "Expand-Archive -Path $TempZip -DestinationPath $TempExtract -Force;" ^
        "Copy-Item -Path (Join-Path $TempExtract '*') -Destination $InstallDir -Recurse -Force;" ^
        "Write-Host '[4/4] Limpando arquivos temporarios...' -ForegroundColor Cyan;" ^
        "Remove-Item $TempZip -Force -ErrorAction SilentlyContinue;" ^
        "Remove-Item $TempExtract -Recurse -Force -ErrorAction SilentlyContinue;" ^
        "Write-Host '';" ^
        "Write-Host '============================================================' -ForegroundColor Green;" ^
        "Write-Host '  Instalacao concluida com sucesso!' -ForegroundColor Green;" ^
        "Write-Host '============================================================' -ForegroundColor Green;" ^
        "Write-Host '';" ^
        "Write-Host 'Proximos passos:' -ForegroundColor Yellow;" ^
        "Write-Host '  1. Reinicie o Claude Desktop (feche e abra de novo).';" ^
        "Write-Host '  2. Abra a aba Code no Claude Desktop.';" ^
        "Write-Host '  3. Configure a integracao Atlassian (OAuth) em Settings > Integrations.';" ^
        "Write-Host '  4. Pergunte em linguagem natural, ex:';" ^
        "Write-Host '     \"onde cadastro um novo advogado?\"' -ForegroundColor DarkGray;" ^
        "Write-Host '     \"me da a pagina de parceiros\"' -ForegroundColor DarkGray;" ^
        "Write-Host '';" ^
    "} catch {" ^
        "Write-Host '';" ^
        "Write-Host '============================================================' -ForegroundColor Red;" ^
        "Write-Host '  ERRO na instalacao' -ForegroundColor Red;" ^
        "Write-Host '============================================================' -ForegroundColor Red;" ^
        "Write-Host '';" ^
        "Write-Host $_.Exception.Message -ForegroundColor Red;" ^
        "Write-Host '';" ^
        "Write-Host 'Dicas:' -ForegroundColor Yellow;" ^
        "Write-Host '  - Verifique sua conexao com a internet.';" ^
        "Write-Host '  - Se a primeira release ainda nao foi publicada, aguarde ou contate o Diego.';" ^
        "Write-Host '  - Tente rodar este instalador como Administrador.';" ^
        "Write-Host '';" ^
        "exit 1" ^
    "}"

if errorlevel 1 (
    echo.
    echo A instalacao falhou. Leia a mensagem acima e tente novamente.
    echo.
    pause
    exit /b 1
)

echo.
pause
exit /b 0

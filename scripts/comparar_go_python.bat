@echo off
setlocal
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0comparar_go_python.ps1" %*
endlocal

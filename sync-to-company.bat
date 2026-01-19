@echo off
REM Sync mini-game-server to company q-connect-server
REM 빠른 실행용 배치 파일

cd /d "%~dp0"
powershell -ExecutionPolicy Bypass -File "sync-to-company.ps1"
pause

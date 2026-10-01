@echo off
title Backup PostgreSQL - Presensi

set "PG_BIN=D:\DatabasePostgree\bin"
set "DB_HOST=localhost"
set "DB_PORT=5432"
set "DB_USER=postgres"
set "DB_PASSWORD=12345"
set "DB_NAME=presensi"
set "BACKUP_FILE=%~dp0presensi.bak"

set "PGPASSWORD=%DB_PASSWORD%"

echo.
echo ==========================================
echo       BACKUP DATABASE POSTGRESQL
echo ==========================================
echo.
echo Database : %DB_NAME%
echo Output   : %BACKUP_FILE%
echo.

if not exist "%PG_BIN%\pg_dump.exe" (
    echo ERROR: pg_dump.exe tidak ditemukan.
    pause
    exit /b 1
)

"%PG_BIN%\pg_dump.exe" ^
    -h "%DB_HOST%" ^
    -p "%DB_PORT%" ^
    -U "%DB_USER%" ^
    -Fc ^
    -f "%BACKUP_FILE%" ^
    "%DB_NAME%"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ==========================================
    echo BACKUP BERHASIL
    echo ==========================================
    echo.
) else (
    echo.
    echo ==========================================
    echo BACKUP GAGAL
    echo ==========================================
    echo.
)

set "PGPASSWORD="

pause
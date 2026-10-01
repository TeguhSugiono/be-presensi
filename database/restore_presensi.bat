@echo off
title Restore PostgreSQL - Presensi

set "PG_BIN=D:\DatabasePostgree\bin"
set "DB_HOST=localhost"
set "DB_PORT=5432"
set "DB_USER=postgres"
set "DB_PASSWORD=12345"
set "DB_NAME=presensi"

set "BACKUP_FILE=%~dp0presensi.bak"

set "PGPASSWORD=%DB_PASSWORD%"

for /f %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyy-MM-dd_HH-mm-ss"') do set "BACKUP_DATE=%%i"

set "CURRENT_BACKUP=%~dp0presensi_backup_%BACKUP_DATE%.bak"

echo.
echo ==========================================
echo       RESTORE DATABASE POSTGRESQL
echo ==========================================
echo.
echo Database       : %DB_NAME%
echo Restore source : %BACKUP_FILE%
echo Safety backup  : %CURRENT_BACKUP%
echo.

if not exist "%BACKUP_FILE%" (
    echo ERROR: presensi.bak tidak ditemukan.
    echo.
    pause
    exit /b 1
)

if not exist "%PG_BIN%\pg_dump.exe" (
    echo ERROR: pg_dump.exe tidak ditemukan.
    echo.
    pause
    exit /b 1
)

if not exist "%PG_BIN%\dropdb.exe" (
    echo ERROR: dropdb.exe tidak ditemukan.
    echo.
    pause
    exit /b 1
)

if not exist "%PG_BIN%\createdb.exe" (
    echo ERROR: createdb.exe tidak ditemukan.
    echo.
    pause
    exit /b 1
)

if not exist "%PG_BIN%\pg_restore.exe" (
    echo ERROR: pg_restore.exe tidak ditemukan.
    echo.
    pause
    exit /b 1
)

echo.
echo [1/4] Backup database saat ini...
echo.

"%PG_BIN%\pg_dump.exe" ^
    -h "%DB_HOST%" ^
    -p "%DB_PORT%" ^
    -U "%DB_USER%" ^
    -Fc ^
    -f "%CURRENT_BACKUP%" ^
    "%DB_NAME%"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo BACKUP DATABASE SAAT INI GAGAL.
    echo Restore dibatalkan demi keamanan.
    echo.
    pause
    exit /b 1
)

echo.
echo Backup safety berhasil:
echo %CURRENT_BACKUP%
echo.

echo [2/4] Menghapus database lama...
echo.

"%PG_BIN%\dropdb.exe" ^
    -h "%DB_HOST%" ^
    -p "%DB_PORT%" ^
    -U "%DB_USER%" ^
    --if-exists ^
    "%DB_NAME%"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo GAGAL menghapus database.
    echo.
    pause
    exit /b 1
)

echo.
echo [3/4] Membuat database baru...
echo.

"%PG_BIN%\createdb.exe" ^
    -h "%DB_HOST%" ^
    -p "%DB_PORT%" ^
    -U "%DB_USER%" ^
    "%DB_NAME%"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo GAGAL membuat database baru.
    echo.
    pause
    exit /b 1
)

echo.
echo [4/4] Restore database...
echo.

"%PG_BIN%\pg_restore.exe" ^
    -h "%DB_HOST%" ^
    -p "%DB_PORT%" ^
    -U "%DB_USER%" ^
    -d "%DB_NAME%" ^
    "%BACKUP_FILE%"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ==========================================
    echo RESTORE BERHASIL
    echo ==========================================
    echo.
    echo Safety backup:
    echo %CURRENT_BACKUP%
) else (
    echo.
    echo ==========================================
    echo RESTORE GAGAL
    echo ==========================================
    echo.
    echo Database sudah dibuat ulang tetapi restore gagal.
    echo Safety backup masih tersedia di:
    echo %CURRENT_BACKUP%
)

set "PGPASSWORD="

echo.
pause
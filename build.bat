@echo off
REM RWMod Monitor Build Script for Windows

setlocal EnableDelayedExpansion

set APP_NAME=rwmod-monitor
if "%VERSION%"=="" set VERSION=1.0.0

echo Building RWMod Monitor...
echo ==========================
echo.

REM Build for current platform (Windows)
echo Building for Windows (current architecture)...
go build -ldflags "-s -w" -o %APP_NAME%.exe
if %errorlevel% neq 0 (
    echo Error building for Windows
    exit /b %errorlevel%
)
echo ✓ Built: %APP_NAME%.exe
echo.

REM Build for Linux (amd64)
echo Building for Linux (amd64)...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o %APP_NAME%-linux-amd64
if %errorlevel% neq 0 (
    echo Error building for Linux amd64
    exit /b %errorlevel%
)
echo ✓ Built: %APP_NAME%-linux-amd64
echo.

REM Build for Linux (arm64)
echo Building for Linux (arm64)...
set GOOS=linux
set GOARCH=arm64
go build -ldflags "-s -w" -o %APP_NAME%-linux-arm64
if %errorlevel% neq 0 (
    echo Error building for Linux arm64
    exit /b %errorlevel%
)
echo ✓ Built: %APP_NAME%-linux-arm64
echo.

REM Build for macOS (amd64)
echo Building for macOS (amd64)...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w" -o %APP_NAME%-darwin-amd64
if %errorlevel% neq 0 (
    echo Error building for macOS amd64
    exit /b %errorlevel%
)
echo ✓ Built: %APP_NAME%-darwin-amd64
echo.

REM Build for macOS (arm64)
echo Building for macOS (arm64)...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "-s -w" -o %APP_NAME%-darwin-arm64
if %errorlevel% neq 0 (
    echo Error building for macOS arm64
    exit /b %errorlevel%
)
echo ✓ Built: %APP_NAME%-darwin-arm64
echo.

REM Build for Windows (amd64)
echo Building for Windows (amd64)...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o %APP_NAME%-windows-amd64.exe
if %errorlevel% neq 0 (
    echo Error building for Windows amd64
    exit /b %errorlevel%
)
echo ✓ Built: %APP_NAME%-windows-amd64.exe
echo.

REM Build for Windows (arm64)
echo Building for Windows (arm64)...
set GOOS=windows
set GOARCH=arm64
go build -ldflags "-s -w" -o %APP_NAME%-windows-arm64.exe
if %errorlevel% neq 0 (
    echo Error building for Windows arm64
    exit /b %errorlevel%
)
echo ✓ Built: %APP_NAME%-windows-arm64.exe
echo.

echo ==========================
echo Build completed successfully!
echo ==========================
echo.
echo Output files:
dir /B %APP_NAME%* 2>nul | findstr /V ".go"

endlocal

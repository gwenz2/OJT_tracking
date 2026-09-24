@echo off
REM Convenience wrapper for Windows users without `make`.
REM Usage: scripts\dev.bat backend | frontend | migrate-up | migrate-down | seed | test | lint

if "%1"=="" goto help
if "%1"=="backend" goto backend
if "%1"=="frontend" goto frontend
if "%1"=="migrate-up" goto migrateup
if "%1"=="migrate-down" goto migratedown
if "%1"=="seed" goto seed
if "%1"=="test" goto test
if "%1"=="lint" goto lint
goto help

:backend
cd backend && go run ./cmd/api
goto :eof

:frontend
cd frontend && npm run dev
goto :eof

:migrateup
cd backend && go run ./cmd/api --migrate-up
goto :eof

:migratedown
cd backend && go run ./cmd/api --migrate-down
goto :eof

:seed
cd backend && go run ./cmd/api --seed
goto :eof

:test
cd backend && go test ./...
cd ../frontend && npm test -- --run
goto :eof

:lint
cd backend && go vet ./...
cd ../frontend && npm run lint
goto :eof

:help
echo Usage: scripts\dev.bat [backend^|frontend^|migrate-up^|migrate-down^|seed^|test^|lint]

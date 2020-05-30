@echo off

REM Clean up old build artifacts.
del registry.exe registry

REM Build the server for Windows.
set GOOS=windows
set GOARCH=amd64
go build -o registry.exe

REM Build the server for Linux.
set GOOS=linux
set GOARCH=amd64
go build -o registry

REM Back to Windows for next build.
set GOOS=windows
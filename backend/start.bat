@echo off
cd /d "%~dp0"
set PATH=%PATH%;C:\Program Files\Go\bin
go run main.go
pause

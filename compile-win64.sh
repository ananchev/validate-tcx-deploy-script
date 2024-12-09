#!/bin/sh
GOOS=windows GOARCH=amd64 go build -o bin/scripts-check.exe main.go
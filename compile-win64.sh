#!/bin/sh
GOOS=windows GOARCH=amd64 go build  -ldflags "-w -s" -o bin/scripts-check.exe main.go
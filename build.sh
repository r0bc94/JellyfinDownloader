#!/bin/bash
#
# Simple script which build the Jellyfin Downloader Executable
# for different Operating Systems. 

# Compile for Windows
echo "Building Windows Binary..."
GOOS=windows GOARCH=amd64 go build -o ./dist/jellyfindownloader.exe main.go

# Compile for Linux
echo "Building Linux Binary..."
GOOS=linux GOARCH=amd64 go build -o ./dist/jellyfindownloader

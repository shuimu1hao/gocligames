#!/usr/bin/env bash
# gocligames 三平台构建脚本：Termux/Linux 本机 + Linux amd64 + Windows amd64
# 用法：bash build.sh
set -e
cd "$(dirname "$0")"
mkdir -p bin
echo "==> [1/3] 构建本机 (Termux/Linux: $(go env GOOS)/$(go env GOARCH)) ..."
CGO_ENABLED=0 go build -trimpath -o bin/bounce ./cmd/bounce
CGO_ENABLED=0 go build -trimpath -o bin/xiuxian ./cmd/xiuxian
CGO_ENABLED=0 go build -trimpath -o bin/jianghu ./cmd/jianghu
CGO_ENABLED=0 go build -trimpath -o bin/dart ./cmd/dart
CGO_ENABLED=0 go build -trimpath -o bin/closed ./cmd/closed
echo "==> [2/3] 构建 Linux amd64 ..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o bin/bounce-linux-amd64 ./cmd/bounce
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o bin/jianghu-linux-amd64 ./cmd/jianghu
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o bin/dart-linux-amd64 ./cmd/dart
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o bin/closed-linux-amd64 ./cmd/closed
echo "==> [3/3] 构建 Windows amd64 ..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -o bin/bounce-windows-amd64.exe ./cmd/bounce
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -o bin/jianghu-windows-amd64.exe ./cmd/jianghu
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -o bin/dart-windows-amd64.exe ./cmd/dart
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -o bin/closed-windows-amd64.exe ./cmd/closed
echo ""
echo "构建完成："
ls -lh bin/

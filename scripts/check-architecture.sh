#!/usr/bin/env bash
# Guarda da regra de dependência: internal/core não pode conhecer adapters nem infraestrutura.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
core="$root/internal/core"

forbidden='github.com/clodoaldomarques/ledger-events/internal/adapters|github.com/clodoaldomarques/ledger-events/config|github.com/aws/|github.com/clodoaldomarques/core-sdk/pkg/aws|github.com/labstack/echo|github.com/clodoaldomarques/core-sdk/pkg/otel/echo|github.com/go-sql-driver/mysql|github.com/sony/gobreaker'

if [ ! -d "$core" ]; then
	echo "arch-check: internal/core ainda não existe; nada a verificar"
	exit 0
fi

if matches="$(grep -R -n -E "$forbidden" "$core" --include='*.go' || true)"; [ -n "$matches" ]; then
	echo "arch-check: imports proibidos em internal/core:" >&2
	echo "$matches" >&2
	exit 1
fi

echo "arch-check: OK"

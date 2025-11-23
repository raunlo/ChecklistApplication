#!/bin/bash
set -e

# 1. Generate OpenAPI code first (types/interfaces needed by Wire)
go generate -tags="oapigen" ./...

# 2. Then generate Wire code (which may depend on generated types)
go generate -tags="wireinject" ./...

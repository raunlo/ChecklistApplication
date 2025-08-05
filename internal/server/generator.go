package server

// This file is used to generate the code from the open-api specification
// v1 version API
//go:generate   go tool  github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o ./v1/checklist/server.gen.go -config ./v1/checklist/cfg.yaml ./../../openapi/api_v1.yaml
//go:generate  go tool  github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o ./v1/checklistItem/server.gen.go -config ./v1/checklistItem/cfg.yaml ./../../openapi/api_v1.yaml

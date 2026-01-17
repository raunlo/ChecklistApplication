//go:build oapigen
// +build oapigen

package server

// This file is used to generate the code from the open-api specification
// v1 version API
//go:generate   go tool  github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o ./v1/checklist/server.gen.go -config ./v1/checklist/cfg.yaml ./../../openapi/api_v1.yaml
//go:generate  go tool  github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o ./v1/checklistItem/server.gen.go -config ./v1/checklistItem/cfg.yaml ./../../openapi/api_v1.yaml
//go:generate  go tool  github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o ./v1/sse/server.gen.go -config ./v1/sse/cfg.yaml ./../../openapi/api_v1.yaml
//go:generate  go tool  github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o ./v1/user/server.gen.go -config ./v1/user/cfg.yaml ./../../openapi/api_v1.yaml
//go:generate  go tool  github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -o ./v1/session/server.gen.go -config ./v1/session/cfg.yaml ./../../openapi/api_v1.yaml

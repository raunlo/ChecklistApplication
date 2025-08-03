FROM golang:1.24-alpine as builder

ENV GO111MODULE=on

WORKDIR /app

#Download deps
COPY go.mod go.sum ./
RUN go mod download


COPY internal ./internal/
COPY cmd ./cmd/
COPY application.yaml application.yaml
COPY openapi ./openapi/
RUN go generate ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go  build  -o checklistapp ./cmd/app.go

FROM scratch

ENV GIN_MODE=release

COPY --from=builder /app/checklistapp /app
COPY --from=builder /app/application.yaml application.yaml
COPY --from=builder /app/openapi/ ./openapi/

EXPOSE 8080

ENTRYPOINT [ "/app" ]
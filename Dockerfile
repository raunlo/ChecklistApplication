FROM golang:1.25-alpine as builder

ENV GO111MODULE=on

WORKDIR /app

#Download deps
COPY go.mod go.sum ./
RUN go mod download


COPY internal ./internal/
COPY cmd ./cmd/
COPY application.yaml application.yaml
COPY openapi ./openapi/
COPY generate.sh ./

RUN chmod +x generate.sh
RUN sh ./generate.sh

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go  build  -o checklistapp ./cmd/app.go

FROM gcr.io/distroless/static-debian11

ENV GIN_MODE=release

COPY --from=builder /app/checklistapp /app
COPY --from=builder /app/application.yaml application.yaml
COPY --from=builder /app/openapi/ ./openapi/

EXPOSE 8080

ENTRYPOINT [ "/app" ]
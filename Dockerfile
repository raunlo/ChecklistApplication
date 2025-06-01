FROM golang:1.21-alpine as builder

ENV GO111MODULE=on

WORKDIR /app

#Download deps
COPY go.mod go.sum ./
RUN go mod download


COPY internal ./internal/
COPY cmd ./cmd/
COPY application.yaml application.yaml
RUN go generate ./...

RUN CGO_ENABLED=0 GOOS=linux go  build  -o checklistapp ./cmd/app.go

run ls -a

FROM scratch

ENV GIN_MODE=release

COPY --from=builder /app/checklistapp /app
COPY --from=builder /app/application.yaml application.yaml

EXPOSE 8080

ENTRYPOINT [ "./app" ]
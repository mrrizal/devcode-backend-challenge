# ---------------------------------------------------------------------
#  Builder
# ---------------------------------------------------------------------
FROM golang:1.17.2-buster as builder

COPY . /app
WORKDIR /app/

RUN GOARCH=amd64 \
    GOOS=linux \
    CGO_ENABLED=0 \
    go build -v -o main main.go

# ---------------------------------------------------------------------
#  Runner
# ---------------------------------------------------------------------
FROM alpine:3.11
COPY --from=builder /app/main /app/main

WORKDIR /app

CMD [ "/app/main" ]
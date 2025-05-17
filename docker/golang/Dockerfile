#======= Stage 1 ===========
FROM golang:alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

#======= Stage 2 ==========

FROM alpine

RUN adduser -S -D -H appuser
USER appuser

WORKDIR /api

COPY --from=builder /build/main /api/main

CMD [ "./main" ]

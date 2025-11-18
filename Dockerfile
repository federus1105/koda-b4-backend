FROM golang:1.25-alpine AS builder

WORKDIR /build


COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o senjakopikiri ./cmd/main.go

FROM alpine:latest 

WORKDIR /backend

COPY --from=builder /build/senjakopikiri /backend/senjakopikiri

EXPOSE 8006

ENTRYPOINT ["/backend/senjakopikiri"]
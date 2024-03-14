FROM golang:1.22.1 AS base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o node node/node.go


# Start a new stage to create a minimal image
FROM alpine:latest

WORKDIR /

COPY --from=base /app/node .

ENTRYPOINT ["./node"]
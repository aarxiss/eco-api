FROM golang:alpine AS builder

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

RUN go mod tidy

RUN swag init --parseDependency --parseInternal

RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/myapp .

EXPOSE 8080

CMD ["./myapp"]
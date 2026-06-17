FROM golang:1.26-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["sh", "-c", "go mod tidy && go run ./cmd/api"]

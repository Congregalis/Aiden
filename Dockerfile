FROM golang:1.25-alpine

WORKDIR /app

RUN apk add --no-cache bash ca-certificates tzdata postgresql-client

COPY go.mod ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["go", "run", "./cmd/aiden"]

FROM golang:1.14-alpine

WORKDIR /app/server
COPY go.* ./
RUN go mod download
COPY . .
RUN go build

CMD ["./main"]
EXPOSE 8080
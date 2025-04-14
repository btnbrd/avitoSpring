FROM golang:1.23

WORKDIR /app

COPY . .

RUN go mod tidy && go build -o /build ./cmd/app/main.go && go clean -cache -modcache

EXPOSE 8080

CMD ["/build"]

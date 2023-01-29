FROM golang:1.17-alpine
RUN mkdir -p /app
WORKDIR /app
ADD . /app
RUN go build -o /api cmd/api/api.go
CMD ["/api"]
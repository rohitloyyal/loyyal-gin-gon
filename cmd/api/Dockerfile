FROM golang:1.19-alpine
RUN mkdir -p /app
WORKDIR /app
ADD . /app
RUN go build -o /api cmd/api/api.go
CMD ["/api"]
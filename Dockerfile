FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/github.com/hyphengolang/noughts-and-crosses/
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o /go/bin/monolith ./cmd/monolith/
FROM scratch
COPY --from=builder /go/bin/monolith .
COPY .env .
EXPOSE 8080
CMD ["/monolith"]
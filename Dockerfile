# syntax=docker/dockerfile:1

FROM golang:latest as builder
WORKDIR /home/adoublef
COPY go.mod go.sum ./
RUN go mod download 
COPY . .
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o monolith ./cmd/monolith

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /home/adoublef/monolith .
EXPOSE 8080
CMD ["./monolith"]
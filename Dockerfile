FROM golang:1.17-buster as builder
WORKDIR /app/
COPY . .
ENV CGO_ENABLED=0
RUN go build -o bin/ringss cmd/ringss/main.go

FROM alpine:3.14
RUN apk add --no-cache monit
WORKDIR /app/
COPY --from=builder /app/bin .

CMD ["./ringss"]

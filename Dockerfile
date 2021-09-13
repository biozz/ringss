FROM golang:1.17-buster as builder
WORKDIR /app/
COPY . .
RUN go build -o bin/ringss cmd/ringss/main.go

FROM scratch
WORKDIR /app/
COPY --from=builder /app/bin .
ENTRYPOINT ["./ringss"]

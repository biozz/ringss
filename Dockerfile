FROM golang:1.17-buster as builder
WORKDIR /app/
COPY . .
ENV CGO_ENABLED=0
RUN update-ca-certificates && \
        [ -f .git-credentials ] && git config --global credential.helper 'store --file '$PWD'/.git-credentials'; \
    go build -o bin/ringss cmd/ringss/main.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
WORKDIR /app/
COPY --from=builder /app/bin .
ENTRYPOINT ["./ringss"]

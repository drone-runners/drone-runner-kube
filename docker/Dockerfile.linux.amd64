FROM golang:1.14-alpine as builder

ENV CGO_ENABLED 0
ENV GOOS linux 
ENV GOARCH amd64

RUN apk add -U --no-cache ca-certificates

WORKDIR /src/app
COPY . .
RUN go mod download && \
    go build -o release/linux/amd64/drone-runner-kube

FROM alpine:3.6
EXPOSE 3000

ENV GODEBUG netdns=go
ENV DRONE_PLATFORM_OS linux
ENV DRONE_PLATFORM_ARCH amd64

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/app/release/linux/amd64/drone-runner-kube /bin/drone-runner-kube

ENTRYPOINT ["/bin/drone-runner-kube"]

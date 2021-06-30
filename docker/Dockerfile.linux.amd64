FROM alpine:3.13 as alpine
RUN apk add -U --no-cache ca-certificates

FROM alpine:3.13
EXPOSE 3000

ENV GODEBUG netdns=go
ENV DRONE_PLATFORM_OS linux
ENV DRONE_PLATFORM_ARCH amd64

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ADD release/linux/amd64/drone-runner-kube /bin/
ENTRYPOINT ["/bin/drone-runner-kube"]
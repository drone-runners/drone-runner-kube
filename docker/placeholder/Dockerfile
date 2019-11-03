FROM golang:alpine as alpine
ADD docker/placeholder/main.go /tmp/main.go
RUN apk add -U --no-cache ca-certificates
WORKDIR /tmp
RUN go build -o /tmp/placeholder

FROM scratch
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine /tmp/placeholder /bin/sh
ENTRYPOINT ["/bin/sh"]

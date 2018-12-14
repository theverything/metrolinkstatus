FROM golang:latest AS builder

RUN apt-get update && apt-get install -y ca-certificates

ENV GOOS linux
ENV GOARCH amd64

RUN go get github.com/theverything/metrolinkstatus

########################################

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /src/cli/cli /cli

CMD ["/cli"]
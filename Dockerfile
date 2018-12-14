FROM golang:alpine AS builder

ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED 0

RUN apk --update add --no-cache ca-certificates git

RUN go get -u github.com/theverything/metrolinkstatus/cli

########################################

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /go/bin/cli /cli

ENTRYPOINT [ "/cli" ]
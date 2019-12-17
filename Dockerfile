FROM golang:1.13-alpine as builder

RUN apk update && apk add git

WORKDIR /go/src/github.com/EDSN-NL/irma-oidc-server

RUN go get -u github.com/gobuffalo/packr/packr

COPY . .

RUN packr build -o /tmp/irma-oidc-server .

FROM alpine:3.10

COPY --from=builder /tmp/irma-oidc-server /

ENTRYPOINT /irma-oidc-server

FROM golang as builder
ENV CGO_ENABLED=0
WORKDIR /basic-auth-proxy
COPY main.go tls.crt tls.key /basic-auth-proxy/
RUN go build . && rm main.go

FROM alpine:latest as certs
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/* /etc/ssl/certs/
ENTRYPOINT ["/basic-auth-proxy"]
COPY --from=builder /basic-auth-proxy/* /

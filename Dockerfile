FROM golang:1.13-alpine as builder
WORKDIR /go/src/github.com/adamdecaf/odfw
RUN apk add -U make
RUN adduser -D -g '' --shell /bin/false adam
COPY . .
RUN make build
USER adam

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/adamdecaf/odfw/bin/odfw /bin/odfw
COPY --from=builder /etc/passwd /etc/passwd
USER adam
EXPOSE 8888
ENTRYPOINT ["/bin/odfw"]

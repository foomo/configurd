FROM golang:alpine AS builder

COPY main.go .

RUN go build -o /helloworld .

FROM alpine:latest

COPY --from=builder /helloworld /helloworld

ENTRYPOINT ["/helloworld"]
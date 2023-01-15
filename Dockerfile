FROM golang:latest

RUN go install github.com/mikerybka/server@latest

ENTRYPOINT [ "server" ]

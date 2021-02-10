FROM alpine:latest

ARG VERSION

COPY dist/lambda-container-exec_v${VERSION}_linux_amd64 /usr/local/bin/lambda-container-exec

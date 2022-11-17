FROM golang:1.19
WORKDIR /go/src/github.com/rymdhund/whazza/
COPY ./ .
RUN rm -rf build && \
    mkdir build && \
    GOOS=linux go build -ldflags "-linkmode external -extldflags -static" -a -o build ./...

# ---------------#

FROM alpine:latest

# See http://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN apk --no-cache add ca-certificates sqlite && \
    mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

WORKDIR /root/
COPY --from=0 /go/src/github.com/rymdhund/whazza/build .
CMD ["./whazza"]

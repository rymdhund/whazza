FROM golang:1.14.4
WORKDIR /go/src/github.com/rymdhund/whazza/
COPY ./ .
RUN rm -rf build && \
    mkdir build && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build ./...

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/rymdhund/whazza/build .
CMD ["./whazza"]

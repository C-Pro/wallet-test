FROM golang:1.12.5-alpine3.9 as builder
ADD . /build
WORKDIR /build
RUN GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -o wallet .

FROM scratch
EXPOSE 8080
COPY --from=builder /build/wallet /
CMD ["/wallet"]

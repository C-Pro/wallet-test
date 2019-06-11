FROM golang:1.12.5-alpine3.9 as builder
ADD . /build
WORKDIR /build
RUN CGO_ENABLED=0 go build -o wallet . 

FROM scratch
EXPOSE 8080
COPY --from=builder /build/wallet /
CMD ["/wallet"]

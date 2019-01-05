# Builder
FROM    golang:latest as BUILDER
RUN     mkdir -p /go/src/github.com/tchaudhry91/cloudinventory
ENV     GO111MODULE=on
WORKDIR /go/src/github.com/tchaudhry91/cloudinventory
RUN     go get
COPY    . .
ENV     CGO_ENABLED=0
RUN     go build

# Final Image
FROM       alpine
RUN        apk update && apk add --no-cache ca-certificates
COPY       --from=BUILDER /go/src/github.com/tchaudhry91/cloudinventory/cloudinventory /bin/
ENTRYPOINT [ "/bin/cloudinventory" ]

# Builder
FROM    golang:latest as BUILDER
RUN     mkdir -p /go/src/github.com/tchaudhry91/cloudinventory
WORKDIR /go/src/github.com/tchaudhry91/cloudinventory
COPY    . .
RUN     go get -d -v ./...
RUN     go get -u golang.org/x/lint/golint
ENV     CGO_ENABLED=0
RUN     make all
RUN     go build

# Final Image
FROM       alpine
RUN        apk update && apk add --no-cache ca-certificates
COPY       --from=BUILDER /go/src/github.com/tchaudhry91/cloudinventory/cloudinventory /bin/
ENTRYPOINT [ "/bin/cloudinventory" ]

# Builder
FROM    golang:latest as BUILDER
RUN     mkdir -p /go/src/github.com/adobe/cloudinventory
ENV     GO111MODULE=on
WORKDIR /go/src/github.com/adobe/cloudinventory

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY    . .

ENV     CGO_ENABLED=0
RUN     go build

# Final Image
FROM       alpine
RUN        apk update && apk add --no-cache ca-certificates
COPY       --from=BUILDER /go/src/github.com/adobe/cloudinventory/cloudinventory /bin/
ENTRYPOINT [ "/bin/cloudinventory" ]

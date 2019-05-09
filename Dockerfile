# Builder
FROM    golang:latest as BUILDER
RUN     mkdir -p /go/src/github.com/adobe/cloudinventory
ENV     GO111MODULE=on
WORKDIR /go/src/github.com/adobe/cloudinventory
COPY    . .
RUN  go get -u github.com/Azure/azure-sdk-for-go
RUN go get github.com/Azure/go-autorest/autorest
RUN go get github.com/Azure/go-autorest/autorest/azure
RUN go get -u github.com/dimchansky/utfbom
RUN go get -u github.com/mitchellh/go-homedir
RUN go get -u golang.org/x/crypto/pkcs12
RUN go get -u github.com/spf13/pflag
RUN go get -u github.com/spf13/cobra

ENV     CGO_ENABLED=0
RUN     go build

# Final Image
FROM       alpine
RUN        apk update && apk add --no-cache ca-certificates
COPY       --from=BUILDER /go/src/github.com/adobe/cloudinventory/cloudinventory /bin/
ENTRYPOINT [ "/bin/cloudinventory" ]

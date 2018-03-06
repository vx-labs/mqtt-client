FROM vxlabs/glide as builder
WORKDIR $GOPATH/src/github.com/vx-labs/mqtt-client
COPY glide* ./
RUN glide install -v
COPY . .
RUN go test ./... && go build -o /bin/mqtt-client ./main.go

FROM alpine
COPY --from=builder /bin/mqtt-client /bin/mqtt-client
WORKDIR /var/lib/mqtt-client
RUN touch config.yaml
ENTRYPOINT ["/bin/mqtt-client"]
RUN apk -U add ca-certificates && \
    rm -rf /var/cache/apk/*

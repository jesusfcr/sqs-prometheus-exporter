FROM golang:1.13.3

WORKDIR /go/src/github.com/nadeemjamali/sqs-prometheus-exporter/
 
COPY .  .

RUN go get -d ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sqs-prometheus-exporter .

FROM alpine

COPY --from=0 /go/src/github.com/nadeemjamali/sqs-prometheus-exporter /

RUN apk --update add ca-certificates && \
	rm -rf /var/cache/apk/*

EXPOSE 9434

CMD ["/sqs-prometheus-exporter"]

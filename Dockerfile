FROM golang:1.15-alpine3.13 as builder

WORKDIR /src

ADD go.* .

RUN go mod download

COPY .  .

RUN GOOS=linux GOARCH=amd64 go build -o sqs-prometheus-exporter .

FROM alpine

RUN apk --update add ca-certificates && \
	rm -rf /var/cache/apk/*

COPY --from=builder /src /

EXPOSE 9434

CMD ["/sqs-prometheus-exporter"]

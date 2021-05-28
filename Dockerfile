FROM golang:1.15-alpine3.13 as builder

WORKDIR /src

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY .  .

RUN GOOS=linux GOARCH=amd64 go build -o sqs-prometheus-exporter .

FROM alpine

RUN apk --update add ca-certificates && \
	rm -rf /var/cache/apk/*

ARG BUILD_RFC3339="1970-01-01T00:00:00Z"
ARG COMMIT="local"

ENV BUILD_RFC3339 "$BUILD_RFC3339"
ENV COMMIT "$COMMIT"

COPY --from=builder /src /

EXPOSE 9434

CMD ["/sqs-prometheus-exporter"]

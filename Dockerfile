FROM golang:1.11 as builder

LABEL maintainer="CoolDuke <me@coolduke.com>"

WORKDIR /go/src/github.com/coolduke/prometheus-fritzbox-exporter

COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/prometheus-fritzbox-exporter .

######## Start a new stage from scratch #######
FROM alpine:latest  

RUN apk --no-cache add ca-certificates && \
    adduser -D prometheus

WORKDIR /home/prometheus/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /go/bin/prometheus-fritzbox-exporter /home/prometheus/

EXPOSE 9742

CMD ["./prometheus-fritzbox-exporter"]

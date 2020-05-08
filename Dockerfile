FROM golang:1.12.5-stretch as builder

RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /kafka2influxdb main.go

FROM alpine

RUN apk update \
    && apk add --no-cache ca-certificates \
    && update-ca-certificates

ENV KFK2INF_PORT=""
ENV KFK2INF_KAFKA_ADDR=""
ENV KFK2INF_KAFKA_TOPIC=""
ENV KFK2INF_KAFKA_SCHEMA_REGISTRY=""
ENV KFK2INF_INFLUXDB_ADDR=""
ENV KFK2INF_INFLUXDB_NAME=""
ENV KFK2INF_INFLUXDB_USER=""
ENV KFK2INF_INFLUXDB_PASSWORD=""
ENV KFK2INF_WITH_SASL=""
ENV KFK2INF_KERBEROS_CONFIG_PATH=""
ENV KFK2INF_KERBEROS_SERVICE_NAME=""
ENV KFK2INF_KERBEROS_USERNAME=""
ENV KFK2INF_KERBEROS_PASSWORD=""
ENV KFK2INF_KERBEROS_REALM=""

COPY --from=builder /kafka2influxdb /

CMD [ "/kafka2influxdb", "serve" ]

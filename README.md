# Kafka2InfluxDB


# How it works

The Kafka2InfluxDB is responsible for consuming all topics (based on a wildcard term) in a Kafka queue and persisting in an InfluxDB.

## Consumer
The consumer is prepared to receive messages of type avro or json according to the scheme below

### Message schema
```sh
$ {
  "name": "movbb",
  "type": "record",
  "namespace": "br.com.bb.interactws",
  "fields": [
    {
      "name": "key",
      "type": "string"
    },
    {
      "name": "value",
      "type": {
        "name": "value",
        "type": "record",
        "fields": [
          {
            "name": "dateTime",
            "type": "string"
          },
          {
            "name": "lat",
            "type": "string"
          },
          {
            "name": "lon",
            "type": "string"
          },
          {
            "name": "mci",
            "type": "string"
          },
          {
            "name": "type",
            "type": "string"
          }
        ]
      }
    }
  ]
}
```
### Message sample
```sh
$ {
    "key":"owner/movbb/thing/297145674599/node/location",
    "value":{
        "dateTime":"2020-04-08T00:04:08Z",
        "lat":"-5.5222581",
        "lon":"-47.4573297",
        "mci":"297145674599",
        "type":"gps"
    }
}
```


## REST Features

  - Queries persisted data over a period of time
  - Insert new data
  - Standardize data according to [InteractWS]() specifications


### Get by period

```sh
$ GET /owner/:owner/thing/:thing/node/:node?startDateTime=:startDate&endDateTime=:endDate
```


#### Parameters

| Parameters    | Required | Type          | Description |
|---------------|----------|---------------|-------------|
| owner         | true     | alphanumeric  | Owner name  |
| thing         | true     | alphanumeric  | Thing name  |
| node          | true     | alphanumeric  | Node name   |
| startDateTime | true     | date RFC 3339 | Start date  |
| endDateTime   | true     | date RFC 3339 | End date    |


#### Request
```sh
$ curl --request GET \
  --url 'http://localhost:8000/owner/movbb/thing/297145674599/node/location?endDateTime=2020-12-31T21%3A10%3A54Z&startDateTime=2020-01-01T00%3A00%3A00Z'
```

#### Result

```sh
$ [
  {
    "owner": "movbb",
    "thing": "297145674599",
    "node": "location",
    "attributes": {
      "lat": "-5.5222581",
      "lon": "-47.4573297",
      "mci": "297145674599",
      "type": "gps"
    },
    "dateTime": "2020-04-08T00:04:08Z"
  }
]
```

### Create a point

```sh
$ POST /owner/:owner/thing/:thing/node/:node
```

#### Parameters

| Parameters    | Required | Type          | Description |
|---------------|----------|---------------|-------------|
| owner         | true     | alphanumeric  | Owner name  |
| thing         | true     | alphanumeric  | Thing name  |
| node          | true     | alphanumeric  | Node name   |


#### Request
```sh
$ curl --request POST \
  --url http://localhost:8000/owner/movbb/thing/297145674599/node/location \
  --header 'content-type: application/json' \
  --data '{
        "dateTime":"2020-04-08T00:04:08Z",
        "lat":"-5.5222581",
        "lon":"-47.4573297",
        "mci":"297145674599",
        "type":"gps"
  }'
```

#### Result

```sh
$ State point created
```

### Installation

Kafka2InfluxDB requires [Golang](https://golang.org/dl/) v1.12 and a [Kafka](https://kafka.apache.org/) service to run.

Install the dependencies.

```sh
$ cd kafka2influxdb
$ go build
```

### Configurations

Some parameters are required to running the api, these parameters can be passed via the command line or environment variables as described below


| ENV                           | Command | Required | Default  | Description                                        |
|-------------------------------|---------|----------|----------|----------------------------------------------------|
| KFK2INF_PORT                  | -p      | false    | 7070     | Api service port                                   |
| KFK2INF_KAFKA_ADDR            | -k      | true     | null     | Kafka host address                                 |
| KFK2INF_KAFKA_TOPIC           | -t      | true     | /owner/* | Kafka topic (wildcard)                             |
| KFK2INF_KAFKA_SCHEMA_REGISTRY | -e      | true     | null     | Kafka schema registry                              |
| KFK2INF_INFLUXDB_ADDR         | -i      | true     | null     | InfluxDB host address                              |
| KFK2INF_INFLUXDB_ADDR         | -i      | true     | null     | InfluxDB host address                              |
| KFK2INF_INFLUXDB_NAME         | -n      | true     | null     | InfluxDB database name                             |
| KFK2INF_INFLUXDB_USER         | -u      | true     | null     | InfluxDB username                                  |
| KFK2INF_INFLUXDB_PASSWORD     | -s      | true     | null     | InfluxDB password                                  |
| KFK2INF_LOG_LEVEL             | -l      | false    | info     | Log level (debug, info, warn, error, fatal, panic) |
| KFK2INF_WITH_SASL             | -w      | false    | false    | Enable/Disable SASL Kafka Security.                |
| KFK2INF_KERBEROS_CONFIG_PATH  | -c      | true     | null     | Kerberos config path                               |
| KFK2INF_KERBEROS_SERVICE_NAME | -d      | true     | null     | Kerberos service name                              |
| KFK2INF_KERBEROS_USERNAME     | -f      | true     | null     | Kerberos username                                  |
| KFK2INF_KERBEROS_PASSWORD     | -g      | true     | null     | Kerberos password                                  |
| KFK2INF_KERBEROS_REALM        | -r      | false    | KERBEROS | Kerberos realm                                     |


## How to run

With all dependencies installed, start the server.

Run in localhost

```sh
$ cd kafka2influxdb
$ go run . serve \
-p=8000 \
-k=localhost:9092 \
-e=localhost:8081 \
-t=owner \
-i=http://localhost:8086 \
-n=influxdbName \
-u=influxdbUser \
-s=influxdbPassword \
-c=/krb5.conf \
-f=kerberosUser \
-g=kerberosPassword \
-l=debug
```

For production environments

```sh
$ go run . serve
$ ENV KFK2INF_PORT="8000"
$ ENV KFK2INF_KAFKA_ADDR="localhost:9092"
$ ENV KFK2INF_KAFKA_SCHEMA_REGISTRY="localhost:8081"
$ ENV KFK2INF_KAFKA_TOPIC="topicName"
$ ENV KFK2INF_INFLUXDB_ADDR="http://localhost:8086"
$ ENV KFK2INF_INFLUXDB_NAME="influxdbName"
$ ENV KFK2INF_INFLUXDB_USER="influxdbUser"
$ ENV KFK2INF_INFLUXDB_PASSWORD="influxdbPassword"
$ ENV KFK2INF_KERBEROS_CONFIG_PATH="/krb5.conf"
$ ENV KFK2INF_KERBEROS_USERNAME="kerberosUser"
$ ENV KFK2INF_KERBEROS_PASSWORD="kerberosPassword"
```


### Docker
Kafka2InfluxDB is very easy to install and deploy in a Docker container.

By default, the Docker will expose port 8000, so change this within the Dockerfile if necessary. When ready, simply use the Dockerfile to build the image.

```sh
cd kafka2influxdb
docker build -t labbsr0x/kafka2influxdb:latest .
```
This will create the Kafka2InfluxDB image and pull in the necessary dependencies.

Once done, run the Docker image and map the port to whatever you wish on your host. In this example, we simply map port 8080 of the host to port 8000 of the Docker (or whatever port was exposed in the Dockerfile):

```sh
docker run -d -p 8000:8080 --restart="always" <youruser>/kafka2influxdb:<version>
```

Verify the deployment by navigating to your server address in your preferred browser.

```sh
127.0.0.1:8080
```

### Todos

 - Write MORE Tests
 - Shard the data to appropriate influxdb instances
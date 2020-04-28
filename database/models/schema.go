package models

var SchemaModel = `{
    "name": "movbb",
    "type": "record",
    "namespace": "interactws",
    "fields": [
      {
        "name": "key",
        "type": "string"
      },
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
        "name": "type",
        "type": "string"
      },
      {
        "name": "mci",
        "type": "string"
      }
    ]
  }`

type Schema struct {
	Key      string `avro:"key"`
	DateTime string `avro:"dateTime"`
	Lat      string `avro:"lat"`
	Lon      string `avro:"lon"`
	Type     string `avro:"type"`
	Mci      string `avro:"mci"`
}

type SchemaJson struct {
	Key      string `json:"key"`
	DateTime string `json:"dateTime"`
	Lat      string `json:"lat"`
	Lon      string `json:"lon"`
	Type     string `json:"type"`
	Mci      string `json:"mci"`
}

type SchemaData struct {
	DateTime string `avro:"dateTime"`
	Lat      string `avro:"lat"`
	Lon      string `avro:"lon"`
	Mci      string `avro:"mci"`
	Type     string `avro:"type"`
}

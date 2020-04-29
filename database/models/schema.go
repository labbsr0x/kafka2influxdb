package models

var SchemaModel = `{
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
}`

type Schema struct {
	Key   string     `avro:"key"`
	Value SchemaData `avro:"value"`
}

type SchemaData struct {
	DateTime string `avro:"dateTime"`
	Lat      string `avro:"lat"`
	Lon      string `avro:"lon"`
	Mci      string `avro:"mci"`
	Type     string `avro:"type"`
}

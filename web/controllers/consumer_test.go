package controllers

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/docker/docker/pkg/testutil/assert"
	"github.com/hamba/avro"
)

func TestByteArrayToInt(t *testing.T) {
	bs, _ := hex.DecodeString("0000000002")
	var schemaID int32
	buf := bytes.NewBuffer(bs[1:5])
	binary.Read(buf, binary.BigEndian, &schemaID)
	assert.Equal(t, schemaID, int32(2))
}

func TestByteArrayMessageToAvro(t *testing.T) {
	bs, _ := hex.DecodeString("28323032302d30342d30385430303a32333a30305a162d32322e37313938363833162d34372e363531333938311831383632323036383039323206677073")
	schema, err := avro.Parse("{\"type\":\"record\",\"name\":\"movbb\",\"namespace\":\"br.com.bb.interactws\",\"fields\":[{\"name\":\"dateTime\",\"type\":\"string\"},{\"name\":\"lat\",\"type\":\"string\"},{\"name\":\"lon\",\"type\":\"string\"},{\"name\":\"mci\",\"type\":\"string\"},{\"name\":\"type\",\"type\":\"string\"}]}")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	var record map[string]interface{}
	err = avro.Unmarshal(schema, bs, &record)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Println(record)
}

func TestByteArrayKeyToAvro(t *testing.T) {
	bs, _ := hex.DecodeString("4e6f776e65722f74657374652f7468696e672f616263313233342f6e6f64652f6c6f636174696f6e")
	schema, err := avro.Parse("{ \"type\": \"string\", \"name\": \"keyschema\", \"fields\": [] }")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	var record string
	err = avro.Unmarshal(schema, bs, &record)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Println(record)
}

// Test how to Decoding Message
// https://docs.confluent.io/current/schema-registry/serdes-develop/index.html#wire-format
func TestGetByteArrayMagicByteAndSchemaID(t *testing.T) {
	bs, _ := hex.DecodeString("000000000228323032302d30342d30385430303a32333a30305a162d32322e37313938363833162d34372e363531333938311831383632323036383039323206677073")
	magicByte := bs[0]
	schemaID := binary.BigEndian.Uint32(bs[1:5])
	msg := bs[5:]

	fmt.Printf("Magic Byte: %b\n", magicByte)
	fmt.Printf("SchemaID: %d\n", schemaID)
	fmt.Printf("Message: %s\n", msg)
}

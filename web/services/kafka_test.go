package services

import (
	"fmt"
	"testing"
	"encoding/json"
)

func TestGetPayloadData(t *testing.T){
	payload := `{"schema":"{\"type\":\"record\",\"name\":\"movbb\",\"namespace\":\"br.com.bb.interactws\",\"fields\":[{\"name\":\"dateTime\",\"type\":\"string\"},{\"name\":\"lat\",\"type\":\"string\"},{\"name\":\"lon\",\"type\":\"string\"},{\"name\":\"mci\",\"type\":\"string\"},{\"name\":\"type\",\"type\":\"string\"}]}"}`
	var response map[string]interface{}
	_ = json.Unmarshal([]byte(payload), &response)
	_, found := response["schema"]
	if found {
		_ = json.Unmarshal([]byte(response["schema"].(string)), &response)
	}
	fmt.Println(response)
}
package models

import (
	"time"
)

// StatePoint addresses a ower/thing/node state representation
type StatePoint struct {
	Owner      string            `json:"owner"`
	Thing      string            `json:"thing"`
	Node       string            `json:"node"`
	Attributes map[string]string `json:"attributes"`
	DateTime   time.Time         `json:"dateTime"`
}

package models

import (
	"time"
)

type Data struct {
	DateTime      time.Time
	StartDateTime time.Time
	EndDateTime   time.Time
	Tags          map[string]string
	Fields        map[string]string
}

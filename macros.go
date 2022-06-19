package main

import (
	"time"
)

var macros_template = map[string]func() string {
	"@timestamp": timestamp,
}

var macros = map[string]func() string {
	"@timestamp": timestamp,
}

func timestamp() string {
	return time.Now().Format("2006-01-02")
}

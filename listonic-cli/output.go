package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Output struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func printSuccess(data any) {
	out := Output{Success: true, Data: data}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}

func printError(err error) {
	out := Output{Success: false, Error: err.Error()}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
	os.Exit(1)
}

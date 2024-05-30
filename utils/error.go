package utils

import (
	"encoding/json"
	"os"
)

func ReadErrDict(filename string) map[string]string {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	jsonData := make(map[string]string)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil
	}
	return jsonData
}

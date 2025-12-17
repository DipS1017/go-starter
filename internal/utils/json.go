package utils

import (
	"encoding/json"
)

func MapToJSON(messageData map[string]interface{}) (string, error) {
	bytes, err := json.Marshal(messageData)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

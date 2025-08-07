package internal

import (
	"encoding/json"
	"os"
)

var tokens = map[string]string{}

func LoadToken(key string) string {
	if len(tokens) == 0 {
		f, err := os.Open("scaleconnect.json")
		if err != nil {
			return ""
		}
		defer f.Close()

		_ = json.NewDecoder(f).Decode(&tokens)
	}

	return tokens[key]
}

func SaveToken(key string, value string) {
	tokens[key] = value

	f, err := os.Create("scaleconnect.json")
	if err != nil {
		return
	}
	defer f.Close()

	_ = json.NewEncoder(f).Encode(&tokens)
}

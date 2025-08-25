package internal

import (
	"encoding/json"
	"os"
	"strings"
)

var tokens = map[string]string{}

func LoadToken(key string) string {
	key = replaceKey(key)

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
	key = replaceKey(key)

	tokens[key] = value

	f, err := os.Create("scaleconnect.json")
	if err != nil {
		return
	}
	defer f.Close()

	_ = json.NewEncoder(f).Encode(&tokens)
}

func replaceKey(key string) string {
	key, value, _ := strings.Cut(key, ":")
	switch key {
	case AccMiFitness, AccXiaomiHome:
		return AccXiaomi + ":" + value
	}
	return key
}

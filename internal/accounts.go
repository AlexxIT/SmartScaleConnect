package internal

import (
	"errors"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
	"github.com/AlexxIT/SmartScaleConnect/pkg/garmin"
	"github.com/AlexxIT/SmartScaleConnect/pkg/picooc"
	"github.com/AlexxIT/SmartScaleConnect/pkg/tanita"
	"github.com/AlexxIT/SmartScaleConnect/pkg/xiaomi"
	"github.com/AlexxIT/SmartScaleConnect/pkg/zepp"
)

var accounts = map[string]core.Account{}

func GetAccount(fields []string) (core.Account, error) {
	key := fields[0] + ":" + fields[1]
	if account, ok := accounts[key]; ok {
		return account, nil
	}

	account, err := getAccount(fields, key)
	if err != nil {
		return nil, err
	}

	accounts[key] = account

	return account, nil
}

func getAccount(fields []string, key string) (core.Account, error) {
	var acc core.Account

	switch fields[0] {
	case "garmin":
		acc = garmin.NewClient()
	case "picooc":
		acc = picooc.NewClient()
	case "tanita":
		acc = tanita.NewClient()
	case "xiaomi":
		acc = xiaomi.NewClient(xiaomi.AppMiFitness)
	case "zepp/xiaomi":
		acc = zepp.NewClient()
	default:
		return nil, errors.New("unsupported type: " + fields[0])
	}

	if acc, ok := acc.(core.AccountWithToken); ok {
		if token := LoadToken(key); token != "" {
			if err := acc.LoginWithToken(token); err == nil {
				return acc, nil
			}
		}
	}

	if err := acc.Login(fields[1], fields[2]); err != nil {
		return nil, err
	}

	if acc, ok := acc.(core.AccountWithToken); ok {
		SaveToken(key, acc.Token())
	}

	return acc, nil
}

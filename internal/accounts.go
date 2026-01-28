package internal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
	"github.com/AlexxIT/SmartScaleConnect/pkg/garmin"
	"github.com/AlexxIT/SmartScaleConnect/pkg/picooc"
	"github.com/AlexxIT/SmartScaleConnect/pkg/tanita"
	"github.com/AlexxIT/SmartScaleConnect/pkg/xiaomi"
	"github.com/AlexxIT/SmartScaleConnect/pkg/zepp"
)

const (
	AccGarmin     = "garmin"
	AccMiFitness  = "mifitness"
	AccPicooc     = "picooc"
	AccTanita     = "tanita"
	AccXiaomi     = "xiaomi"
	AccXiaomiHome = "xiaomihome"
	AccZeppXiaomi = "zepp/xiaomi"
)

var accounts map[string]core.Account
var cacheTS time.Time

func GetAccount(fields []string) (core.Account, error) {
	if now := time.Now(); now.After(cacheTS) {
		accounts = map[string]core.Account{}
		cacheTS = now.Add(23 * time.Hour)
	}

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
	case AccGarmin:
		acc = garmin.NewClient()
	case AccPicooc:
		acc = picooc.NewClient()
	case AccTanita:
		acc = tanita.NewClient()
	case AccXiaomi, AccMiFitness:
		acc = xiaomi.NewClient(xiaomi.AppMiFitness)
	case AccXiaomiHome:
		acc = xiaomi.NewClient(xiaomi.AppXiaomiHome)
	case AccZeppXiaomi:
		acc = zepp.NewClient()
	default:
		return nil, errors.New("unsupported type: " + fields[0])
	}

	if accWithToken, ok := acc.(core.AccountWithToken); ok {
		if token := LoadToken(key); token != "" {
			if err := accWithToken.LoginWithToken(token); err == nil {
				return acc, nil
			}
		}
	}

	if err := acc.Login(fields[1], fields[2]); err != nil {
		return handleLoginError(err, acc, fields, key)
	}

	if accWithToken, ok := acc.(core.AccountWithToken); ok {
		saveTokenForAccount(accWithToken, fields[0], key)
	}

	return acc, nil
}

func handleLoginError(err error, acc core.Account, fields []string, key string) (core.Account, error) {
	loginErr, ok := err.(*xiaomi.LoginError)
	if !ok {
		return nil, err
	}

	xiaomiClient, ok := acc.(*xiaomi.Client)
	if !ok {
		return nil, err
	}

	accountType := fields[0]

	if len(loginErr.Captcha) > 0 {
		if isInteractive() {
			return handleCaptchaInteractive(xiaomiClient, loginErr, fields[1], fields[2], accountType, key)
		}
		return nil, fmt.Errorf("captcha required: run with -it flags for interactive mode")
	}

	if loginErr.VerifyPhone != "" || loginErr.VerifyEmail != "" {
		if isInteractive() {
			return handle2FAInteractive(xiaomiClient, loginErr, accountType, key)
		}
		verifyTarget := loginErr.VerifyPhone
		if verifyTarget == "" {
			verifyTarget = loginErr.VerifyEmail
		}
		return nil, fmt.Errorf("2FA verification required: code sent to %s, run with -it flags for interactive mode", verifyTarget)
	}

	return nil, err
}

func saveTokenForAccount(acc core.AccountWithToken, accountType, key string) {
	token := acc.Token()
	SaveToken(key, token)

	if !isXiaomiAccount(accountType) {
		return
	}

	xiaomiAcc, ok := acc.(interface{ UserToken() (string, string) })
	if !ok {
		return
	}

	accountID, _ := xiaomiAcc.UserToken()
	accountKey := AccXiaomi + ":" + accountID
	SaveToken(accountKey, token)
}

func isXiaomiAccount(accountType string) bool {
	return accountType == AccXiaomi || accountType == AccXiaomiHome || accountType == AccMiFitness
}

func isInteractive() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	mode := fileInfo.Mode()
	return (mode&os.ModeCharDevice) != 0 || (mode&os.ModeNamedPipe) != 0
}

func handleCaptchaInteractive(client *xiaomi.Client, loginErr *xiaomi.LoginError, username, password, accountType, key string) (core.Account, error) {
	fmt.Println("\nCaptcha required")

	if err := os.WriteFile("captcha.png", loginErr.Captcha, 0644); err == nil {
		fmt.Println("Captcha image saved to captcha.png")
	}

	fmt.Print("Enter captcha code: ")
	reader := bufio.NewReader(os.Stdin)
	captcha, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read captcha code: %w", err)
	}
	captcha = strings.TrimSpace(captcha)

	if err := client.LoginWithCaptcha(captcha); err != nil {
		if loginErr2, ok := err.(*xiaomi.LoginError); ok {
			if loginErr2.VerifyPhone != "" || loginErr2.VerifyEmail != "" {
				return handle2FAInteractive(client, loginErr2, accountType, key)
			}
		}
		return nil, fmt.Errorf("captcha verification failed: %w", err)
	}

	saveTokenForAccount(client, accountType, key)
	return client, nil
}

func handle2FAInteractive(client *xiaomi.Client, loginErr *xiaomi.LoginError, accountType, key string) (core.Account, error) {
	fmt.Println("\n2FA verification required")
	if loginErr.VerifyPhone != "" {
		fmt.Printf("Verification code sent to phone: %s\n", loginErr.VerifyPhone)
	} else if loginErr.VerifyEmail != "" {
		fmt.Printf("Verification code sent to email: %s\n", loginErr.VerifyEmail)
	}

	fmt.Print("Enter verification code: ")
	reader := bufio.NewReader(os.Stdin)
	code, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read verification code: %w", err)
	}
	code = strings.TrimSpace(code)

	if err := client.LoginWithVerify(code); err != nil {
		return nil, fmt.Errorf("2FA verification failed: %w", err)
	}

	saveTokenForAccount(client, accountType, key)
	return client, nil
}

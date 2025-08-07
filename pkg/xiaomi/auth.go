package xiaomi

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
)

const (
	AppXiaomiHome = "xiaomiio"
	AppMiFitness  = "miothealth"
)

func (c *Client) Login(username, password string) error {
	res1, err := c.serviceLogin()
	if err != nil {
		return err
	}

	res2, err := c.serviceLogin2(res1, username, password)
	if err != nil {
		return err
	}

	return c.serviceLogin3(res2.Location)
}

type loginResponse1 struct {
	//ServiceParam   string      `json:"serviceParam"`
	Qs string `json:"qs"`
	//Code           int         `json:"code"`
	//Description    string      `json:"description"`
	//SecurityStatus int         `json:"securityStatus"`
	Sign string `json:"_sign"`
	Sid  string `json:"sid"`
	//Result         string      `json:"result"`
	//CaptchaUrl     interface{} `json:"captchaUrl"`
	Callback string `json:"callback"`
	//Location       string      `json:"location"`
	//Pwd            int         `json:"pwd"`
	//Child          int         `json:"child"`
	//Desc           string      `json:"desc"`
}

func (c *Client) serviceLogin() (*loginResponse1, error) {
	res, err := c.client.Get("https://account.xiaomi.com/pass/serviceLogin?_json=true&sid=" + c.sid)
	if err != nil {
		return nil, err
	}

	body, err := readLoginResponse(res)
	if err != nil {
		return nil, err
	}

	var res1 loginResponse1
	if err = json.Unmarshal(body, &res1); err != nil {
		return nil, err
	}

	return &res1, nil
}

type loginResponse2 struct {
	//Qs             string      `json:"qs"`
	Ssecurity []byte `json:"ssecurity"`
	//Code           int         `json:"code"`
	PassToken string `json:"passToken"`
	//Description    string      `json:"description"`
	//SecurityStatus int         `json:"securityStatus"`
	//Nonce          int64       `json:"nonce"`
	UserId int64 `json:"userId"`
	//CUserId        string      `json:"cUserId"`
	//Result         string      `json:"result"`
	//Psecurity      string      `json:"psecurity"`
	//CaptchaUrl     interface{} `json:"captchaUrl"`
	Location string `json:"location"`
	//Pwd            int         `json:"pwd"`
	//Child          int         `json:"child"`
	//Desc           string      `json:"desc"`
}

func (c *Client) serviceLogin2(res1 *loginResponse1, username, password string) (*loginResponse2, error) {
	hash := fmt.Sprintf("%X", md5.Sum([]byte(password)))

	form := url.Values{
		"_json":    {"true"},
		"hash":     {hash},
		"sid":      {res1.Sid},
		"callback": {res1.Callback},
		"_sign":    {res1.Sign},
		"qs":       {res1.Qs},
		"user":     {username},
	}

	req, err := http.NewRequest(
		"POST", "https://account.xiaomi.com/pass/serviceLoginAuth2", strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", "deviceId="+core.RandString(16, 62))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := readLoginResponse(res)
	if err != nil {
		return nil, err
	}

	var res2 loginResponse2
	if err = json.Unmarshal(body, &res2); err != nil {
		return nil, err
	}

	c.passToken = res2.PassToken
	c.ssecurity = res2.Ssecurity
	c.userID = res2.UserId

	return &res2, nil
}

func (c *Client) serviceLogin3(location string) error {
	res, err := c.client.Get(location)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	for _, s := range res.Header["Set-Cookie"] {
		s, _, _ = strings.Cut(s, ";")
		if len(c.cookies) > 0 {
			c.cookies += "; "
		}
		c.cookies += s
	}

	return nil
}

func (c *Client) OAuth2(params, username, password string) (string, error) {
	res1, err := c.oauth2Authorize(params)
	if err != nil {
		return "", err
	}

	res2, err := c.serviceLogin2(res1, username, password)
	if err != nil {
		return "", err
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: time.Minute,
		Jar:     jar, // important to support cookies
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) == 2 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	res, err := client.Get(res2.Location)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	location := res.Header.Get("Location")
	_, code, _ := strings.Cut(location, "=")

	return code, nil
}

func (c *Client) oauth2Authorize(params string) (*loginResponse1, error) {
	res, err := c.client.Get("https://account.xiaomi.com/oauth2/authorize?" + params)
	if err != nil {
		return nil, err
	}

	body, err := readLoginResponse(res)
	if err != nil {
		return nil, err
	}

	var json1 struct {
		Data struct {
			OauthLoginUrl string `json:"oauthLoginUrl"`
		} `json:"data"`
	}

	if err = json.Unmarshal(body, &json1); err != nil {
		return nil, err
	}

	res, err = c.client.Get(json1.Data.OauthLoginUrl)
	if err != nil {
		return nil, err
	}

	body, err = readLoginResponse(res)
	if err != nil {
		return nil, err
	}

	var res1 loginResponse1
	if err = json.Unmarshal(body, &res1); err != nil {
		return nil, err
	}

	return &res1, nil
}

func (c *Client) Request(apiURL, params string) (data []byte, err error) {
	for range 2 {
		if data, err = c.request(apiURL, params); err != nil {
			time.Sleep(time.Second)
			continue
		}
		return
	}
	return c.request(apiURL, params)
}

func (c *Client) request(apiURL, params string) ([]byte, error) {
	form := url.Values{"data": {params}}

	nonce := GenNonce()
	signedNonce := GenSignedNonce(c.ssecurity, nonce)

	// 1. gen hash for data param
	form.Set("rc4_hash__", GenSignature64("POST", apiURL, form, signedNonce))

	// 2. encrypt data and hash params
	for _, v := range form {
		ciphertext, err := Crypt(signedNonce, []byte(v[0]))
		if err != nil {
			return nil, err
		}
		v[0] = base64.StdEncoding.EncodeToString(ciphertext)
	}

	// 3. add signature for encrypted data and hash params
	form.Set("signature", GenSignature64("POST", apiURL, form, signedNonce))

	// 4. add nonce
	form.Set("_nonce", base64.StdEncoding.EncodeToString(nonce))

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", c.cookies)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return nil, errors.New(res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return nil, err
	}

	plaintext, err := Crypt(signedNonce, ciphertext)
	if err != nil {
		return nil, err
	}

	var res1 struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Result  json.RawMessage `json:"result"`
	}
	if err = json.Unmarshal(plaintext, &res1); err != nil {
		return nil, err
	}

	if res1.Code != 0 {
		return nil, errors.New("xiaomi: " + res1.Message)
	}

	return res1.Result, nil
}

func (c *Client) LoginWithToken(token string) error {
	userID, passToken, _ := strings.Cut(token, ":")

	req, err := http.NewRequest("GET", "https://account.xiaomi.com/pass/serviceLogin?_json=true&sid="+c.sid, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Cookie", fmt.Sprintf("userId=%s; passToken=%s", userID, passToken))

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	body, err := readLoginResponse(res)
	if err != nil {
		return err
	}

	var res2 loginResponse2
	if err = json.Unmarshal(body, &res2); err != nil {
		return err
	}

	c.passToken = res2.PassToken
	c.ssecurity = res2.Ssecurity
	c.userID = res2.UserId

	return c.serviceLogin3(res2.Location)
}

func (c *Client) Token() string {
	return fmt.Sprintf("%d:%s", c.userID, c.passToken)
}

const loginPrefix = "&&&START&&&"

func readLoginResponse(res *http.Response) ([]byte, error) {
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(body, []byte(loginPrefix)) {
		return nil, errors.New("xiaomi: wrong loginPrefix")
	}

	return body[len(loginPrefix):], nil
}

func GenNonce() []byte {
	ts := time.Now().UnixMilli() / 60000

	nonce := make([]byte, 12)
	_, _ = rand.Read(nonce[:8])
	binary.BigEndian.PutUint32(nonce[8:], uint32(ts))
	return nonce
}

func GenSignedNonce(ssecurity, nonce []byte) []byte {
	hasher := sha256.New()
	hasher.Write(ssecurity)
	hasher.Write(nonce)
	return hasher.Sum(nil)
}

func Crypt(key, plaintext []byte) ([]byte, error) {
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cipher.XORKeyStream(make([]byte, 1024), make([]byte, 1024))

	ciphertext := make([]byte, len(plaintext))
	cipher.XORKeyStream(ciphertext, plaintext)

	return ciphertext, nil
}

var prefixes = [4]string{
	"https://hlth.io.mi.com/healthapp", "https://hlth.io.mi.com",
	"https://api.io.mi.com/healthapp", "https://api.io.mi.com/app",
}

func GenSignature64(method, url string, values url.Values, signedNonce []byte) string {
	for _, s := range prefixes {
		if strings.HasPrefix(url, s) {
			url = url[len(s):]
			break
		}
	}

	s := method + "&" + url
	for k, v := range values {
		s += "&" + k + "=" + v[0]
	}
	s += "&" + base64.StdEncoding.EncodeToString(signedNonce)

	hasher := sha1.New()
	hasher.Write([]byte(s))
	signature := hasher.Sum(nil)

	return base64.StdEncoding.EncodeToString(signature)
}

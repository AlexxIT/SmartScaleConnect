package zepp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/AlexxIT/SmartScaleConnect/pkg/xiaomi"
	"github.com/google/uuid"
)

const paramsZeppLife = "_json=true&" +
	"client_id=428135909242707968&" +
	"pt=1&" +
	"redirect_uri=https://api-mifit-cn.huami.com/huami.health.loginview.do&" +
	"response_type=code"

func (c *Client) Login(username, password string) error {
	if c.nativeAccount {
		return c.loginAccount(username, password)
	}
	client := xiaomi.NewClient("")
	code, err := client.OAuth2(paramsZeppLife, username, password)
	if err != nil {
		return err
	}

	return c.exchangeToken(code, "request_token", "xiaomi-hm-mifit")
}

// NewAccountClient authenticates a native Zepp email/password account.
// NewClient retains the existing Xiaomi OAuth login.
func NewAccountClient() *Client {
	c := NewClient()
	c.nativeAccount = true
	return c
}

func (c *Client) loginAccount(username, password string) error {
	if !strings.Contains(username, "@") {
		return errors.New("zepp: native login currently requires an email address")
	}
	const redirectURI = "https://s3-us-west-2.amazonaws.com/hm-registration/successsignin.html"
	form := url.Values{
		"emailOrPhone": {username}, "password": {password},
		"state": {"REDIRECTION"}, "client_id": {"HuaMi"},
		"redirect_uri": {redirectURI}, "region": {"us-west-2"},
		"token": {"access", "refresh"}, "country_code": {"GB"},
	}
	req, err := http.NewRequest("POST", "https://api-user.zepp.com/v2/registrations/tokens", bytes.NewReader(encryptAccountForm(form.Encode())))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("app_name", "com.xiaomi.hm.health")
	req.Header.Set("Accept-Language", "en-GB")
	req.Header.Set("X-HM-EKV", "1")
	req.Header.Set("appname", "com.xiaomi.hm.health")
	req.Header.Set("cv", "50800_6.12.0")
	req.Header.Set("v", "2.0")
	req.Header.Set("appplatform", "android_phone")
	req.Header.Set("User-Agent", "MiFit6.12.0 (Android; Android 6.0; Density/2.0)")
	req.Header.Set("X-Request-Id", uuid.NewString())
	// The redirect contains credentials. Inspect it without following it.
	client := *c.client
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	res, err := client.Do(req)
	if err != nil {
		return errors.New("zepp: account login request failed")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusFound && res.StatusCode != http.StatusSeeOther {
		return fmt.Errorf("zepp: account login failed (HTTP %d); check credentials or required verification in the app", res.StatusCode)
	}
	location, err := url.Parse(res.Header.Get("Location"))
	if err != nil {
		return errors.New("zepp: invalid account login redirect")
	}
	expected, _ := url.Parse(redirectURI)
	if location.Scheme != expected.Scheme || location.Host != expected.Host || location.Path != expected.Path {
		return errors.New("zepp: account login requires verification or returned an unexpected redirect")
	}
	if location.Query().Get("error") != "" {
		return errors.New("zepp: account login rejected; check credentials or verification in the app")
	}
	token := location.Query().Get("access")
	if token == "" {
		return errors.New("zepp: account login returned no access token")
	}
	return c.exchangeToken(token, "access_token", "huami")
}

func (c *Client) exchangeToken(code, grantType, thirdName string) error {
	country, version := "CN", "6.14.0"
	if c.nativeAccount {
		country, version = "GB", "6.12.0"
	}
	form := url.Values{
		"app_name": {"com.xiaomi.hm.health"}, "app_version": {version},
		"code": {code}, "country_code": {country}, "device_id": {uuid.NewString()},
		"device_model": {"phone"}, "dn": {"api-mifit.zepp.com"},
		"grant_type": {grantType}, "third_name": {thirdName},
	}
	if c.nativeAccount {
		form.Set("allow_registration", "false")
	}
	res, err := c.client.Post("https://account.zepp.com/v2/client/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return errors.New("zepp: app token request failed")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("zepp: app login failed (HTTP %d)", res.StatusCode)
	}
	var result struct {
		Result    string `json:"result"`
		TokenInfo struct {
			AppToken string `json:"app_token"`
			UserID   string `json:"user_id"`
		} `json:"token_info"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return errors.New("zepp: invalid app login response")
	}
	if result.Result != "ok" || result.TokenInfo.AppToken == "" || result.TokenInfo.UserID == "" {
		return errors.New("zepp: app login failed or returned incomplete credentials")
	}
	c.appToken, c.userID = result.TokenInfo.AppToken, result.TokenInfo.UserID
	return nil
}

func (c *Client) LoginWithToken(token string) error {
	c.userID, c.appToken, _ = strings.Cut(token, ":")
	return c.GetFamilyMembers()
}

func (c *Client) Token() string {
	return c.userID + ":" + c.appToken
}

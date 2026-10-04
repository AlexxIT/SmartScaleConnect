package zepp

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"testing"
)

type authTransport func(*http.Request) (*http.Response, error)

func (f authTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func authResponse(status int, location, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Location": {location}}, Body: io.NopCloser(strings.NewReader(body))}
}

func TestNativeLogin(t *testing.T) {
	c := NewAccountClient()
	calls := 0
	c.client.Transport = authTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			if r.Header.Get("X-HM-EKV") != "1" {
				t.Fatal("missing encryption version")
			}
			data, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			block, _ := aes.NewCipher([]byte(accountKey))
			cipher.NewCBCDecrypter(block, []byte(accountIV)).CryptBlocks(data, data)
			padding := int(data[len(data)-1])
			data = data[:len(data)-padding]
			r.Body = io.NopCloser(strings.NewReader(string(data)))
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		switch calls {
		case 1:
			if r.URL.Host != "api-user.zepp.com" || r.URL.Path != "/v2/registrations/tokens" || r.Form.Get("emailOrPhone") != "a+b@example.com" || r.Form.Get("password") != "p&+= word" {
				t.Fatal("incorrect credential request")
			}
			if len(r.Form["token"]) != 2 || r.Header.Get("app_name") != "com.xiaomi.hm.health" {
				t.Fatal("missing passport fields")
			}
			return authResponse(303, r.Form.Get("redirect_uri")+"?access=token%2B%26%3D", ""), nil
		case 2:
			if r.URL.Host != "account.zepp.com" || r.URL.Path != "/v2/client/login" || r.Form.Get("code") != "token+&=" || r.Form.Get("grant_type") != "access_token" || r.Form.Get("third_name") != "huami" || r.Form.Get("allow_registration") != "false" {
				t.Fatal("incorrect token exchange")
			}
			return authResponse(200, "", `{"result":"ok","token_info":{"app_token":"app","user_id":"123"}}`), nil
		default:
			t.Fatal("unexpected request: redirect must not be followed")
			return nil, nil
		}
	})
	if err := c.Login("a+b@example.com", "p&+= word"); err != nil {
		t.Fatal(err)
	}
	if c.Token() != "123:app" || calls != 2 {
		t.Fatal("missing authenticated token")
	}
}

func TestNativeLoginRejectsFailures(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		location string
	}{
		{"bad password", 401, ""},
		{"verification", 303, "https://account.zepp.com/verify?access=secret"},
		{"missing token", 303, "https://s3-us-west-2.amazonaws.com/hm-registration/successsignin.html?error=invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewAccountClient()
			calls := 0
			c.client.Transport = authTransport(func(*http.Request) (*http.Response, error) {
				calls++
				return authResponse(tc.status, tc.location, ""), nil
			})
			if err := c.Login("a@example.com", "secret"); err == nil || strings.Contains(err.Error(), "secret") {
				t.Fatal("expected a safe login error")
			}
			if calls != 1 || c.appToken != "" {
				t.Fatal("failed login must not exchange or store tokens")
			}
		})
	}
}

func TestAppLoginRequiresCompleteTokens(t *testing.T) {
	for _, body := range []string{`{"result":"error"}`, `{"result":"ok","token_info":{"user_id":"123"}}`, `not JSON`} {
		c := NewAccountClient()
		c.client.Transport = authTransport(func(*http.Request) (*http.Response, error) { return authResponse(200, "", body), nil })
		if err := c.exchangeToken("secret", "access_token", "huami"); err == nil {
			t.Fatal("expected rejection of invalid tokens")
		}
		if c.appToken != "" || c.userID != "" {
			t.Fatal("failed response stored credentials")
		}
	}
}

// Captured from the APK's native Crypto.method03 using fake data, not generated
// by the Go implementation. Covers partial, full and multiple AES blocks.
func TestAccountEncryptionNativeVectors(t *testing.T) {
	for _, tc := range []struct{ plain, encrypted string }{
		{"a", "cd7325045d6e606d9e99d3e0aff24713"},
		{"1234567890123456", "b5aae451b9ccb50866d6603c9da1013fc1755f86588b56bb646f8ad748165492"},
		{"emailOrPhone=test%40example.com&password=fake", "6926acbfe3c7d62ed6a50c5d5cf753efe277f3807ae82cacda727657cb055ecb7070690ae3404000f917166a2d9f3794"},
	} {
		if got := hex.EncodeToString(encryptAccountForm(tc.plain)); got != tc.encrypted {
			t.Fatalf("native encryption mismatch: %s", got)
		}
	}
}

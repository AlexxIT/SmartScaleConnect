package picooc

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const api = "https://api2.picooc-int.com/v1/api/"

func (c *Client) Login(username, password string) error {
	form := c.values("user_login_new")

	var req1 struct {
		AppVer    string `json:"appver"`
		Timestamp string `json:"timestamp"`
		Lang      string `json:"lang"`
		Method    string `json:"method"`
		Timezone  string `json:"timezone"`
		Sign      string `json:"sign"`
		PushToken string `json:"push_token"`
		DeviceID  string `json:"device_id"`
		Rec       struct {
			AppChannel  string `json:"app_channel"`
			AppVer      string `json:"app_version"`
			Email       string `json:"email"`
			Password    string `json:"password"`
			Phone       string `json:"phone"`
			PhoneSystem string `json:"phone_system"`
			PhoneType   string `json:"phone_type"`
		} `json:"req"`
	}
	req1.AppVer = form.Get("appver")
	req1.Timestamp = form.Get("timestamp")
	req1.Lang = form.Get("lang")
	req1.Method = form.Get("method")
	req1.Sign = form.Get("sign")
	req1.PushToken = form.Get("push_token")
	req1.DeviceID = form.Get("device_id")
	req1.Rec.AppVer = form.Get("appver")
	req1.Rec.Email = username
	req1.Rec.Password = password

	data, err := json.Marshal(req1)
	if err != nil {
		return err
	}

	form.Set("reqData", string(data))

	res, err := c.client.Post(
		api+"account/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var res1 struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Resp struct {
			UserID string `json:"user_id"`
			RoleID string `json:"role_id"`
			Roles  []struct {
				RoleID   string `json:"role_id"`
				RoleName string `json:"role_name"`
			} `json:"roles"`
			//EndBodyIndex json.RawMessage `json:"end_body_index"`
		} `json:"resp"`
	}

	if err = json.NewDecoder(res.Body).Decode(&res1); err != nil {
		return err
	}

	if res1.Code != 0 {
		return errors.New("picooc: login error: " + res1.Msg)
	}

	c.userID = res1.Resp.UserID

	c.roleIDs = map[string]string{"": res1.Resp.RoleID}
	for _, role := range res1.Resp.Roles {
		c.roleIDs[role.RoleName] = role.RoleID
	}

	return nil
}

const appVer = "i4.1.11.0"

func (c *Client) values(method string) url.Values {
	if c.deviceID == "" {
		c.deviceID = strings.ToUpper(uuid.NewString())
	}

	timestamp := strconv.Itoa(int(time.Now().Unix()))
	sign := upperMD5(c.deviceID + upperMD5(timestamp+upperMD5(method)+upperMD5(appVer)))

	return url.Values{
		"appver":     {appVer},
		"timestamp":  {timestamp},
		"lang":       {"en"},
		"method":     {method},
		"timezone":   {""}, // don't know how to get right value
		"sign":       {sign},
		"push_token": {"android::" + c.deviceID},
		"device_id":  {c.deviceID},
	}
}

func upperMD5(s string) string {
	return fmt.Sprintf("%X", md5.Sum([]byte(s)))
}

package tanita

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		client: &http.Client{Timeout: time.Minute, Jar: jar},
	}
}

func (c *Client) Login(username, password string) error {
	res, err := c.client.Get("https://mytanita.eu/en/user/login")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	token := core.Between(string(body), `name="token" value="`, `"`)

	form := fmt.Sprintf(
		"mail=%s&password=%s&token=%s&login=Login",
		url.QueryEscape(username), url.QueryEscape(password), token,
	)
	res, err = c.client.Post(
		"https://mytanita.eu/en/user/processlogin", "application/x-www-form-urlencoded", strings.NewReader(form),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(string(body), "<title>My trends - myTanita</title>") {
		msg := core.Between(string(body), `<li>`, `</li>`)
		return errors.New("tanita: " + msg)
	}

	return nil
}

func (c *Client) GetAllWeights() ([]*core.Weight, error) {
	// VERY long operation
	res, err := c.client.Get("https://mytanita.eu/en/user/export-csv")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var weights []*core.Weight

	f := csv.NewReader(res.Body)
	if _, err = f.Read(); err != nil {
		return nil, err
	}

	for {
		// Date,"Weight (kg)",BMI,"Body Fat (%)","Visc Fat","Muscle Mass (kg)","Muscle Quality","Bone Mass (kg)","BMR (kcal)","Metab Age","Body Water (%)","Physique Rating","Muscle mass - right arm","Muscle mass - left arm","Muscle mass - right leg","Muscle mass - left leg","Muscle mass - trunk","Muscle quality - right arm","Muscle quality - left arm","Muscle quality - right leg","Muscle quality - left leg","Muscle quality - trunk","Body fat (%) - right arm","Body fat (%) - left arm","Body fat (%) - right leg","Body fat (%) - left leg","Body fat (%) - trunk","Heart rate"
		line, err := f.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		ts, err := time.Parse(time.DateTime, line[0]) // parse as UTC
		if err != nil {
			return nil, err
		}

		w := &core.Weight{
			Date:            ts,
			Weight:          parseFloat(line[1]),
			BMI:             parseFloat(line[2]),
			BodyFat:         parseFloat(line[3]),
			VisceralFat:     int(parseFloat(line[4])),
			MuscleMass:      parseFloat(line[5]),
			BoneMass:        parseFloat(line[7]),
			BasalMetabolism: int(parseFloat(line[8])),
			MetabolicAge:    int(parseFloat(line[9])),
			BodyWater:       parseFloat(line[10]),
			PhysiqueRating:  int(parseFloat(line[11])),
		}
		weights = append(weights, w)
	}

	return weights, nil
}

func parseFloat(s string) float32 {
	if s == "-" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 32)
	return float32(f)
}

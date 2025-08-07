package garmin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
	"github.com/gomodule/oauth1/oauth"
)

type Client struct {
	client *http.Client

	oauthToken  string
	oauthSecret string

	oauthClient *oauth.Client
	accessToken string
	expiresTime time.Time

	weightID map[int64]string
}

func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		client:   &http.Client{Timeout: time.Minute, Jar: jar},
		weightID: make(map[int64]string),
	}
}

func (c *Client) Get(api string) (*http.Response, error) {
	req, err := http.NewRequest("GET", "https://connectapi.garmin.com/"+api, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) Delete(api string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", "https://connectapi.garmin.com/"+api, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) PostFile(api, filename string, data []byte) (*http.Response, error) {
	buf := bytes.NewBuffer(nil)
	w := multipart.NewWriter(buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	if _, err = part.Write(data); err != nil {
		return nil, err
	}

	if err = w.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://connectapi.garmin.com/"+api, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", w.FormDataContentType())

	return c.do(req)
}

func (c *Client) Upload(filename string, data []byte) error {
	res, err := c.PostFile("upload-service/upload", filename, data)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return errors.New(res.Status)
	}

	return nil
}

func (c *Client) GetAllWeights() ([]*core.Weight, error) {
	return c.GetWeight("1970-01-01", time.Now().Format(time.DateOnly))
}

// GetWeight - start and end format: 2025-07-28
func (c *Client) GetWeight(start, end string) ([]*core.Weight, error) {
	path := fmt.Sprintf("weight-service/weight/range/%s/%s?includeAll=true", start, end)
	res, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data struct {
		DailyWeightSummaries []struct {
			AllWeightMetrics []struct {
				SamplePk       int64   `json:"samplePk"`       // 1753533352952
				Date           int64   `json:"date"`           // 1753544137000
				CalendarDate   string  `json:"calendarDate"`   // 2025-07-26
				Weight         float32 `json:"weight"`         // 63900.0
				BMI            float32 `json:"bmi"`            // 21.600000381469727
				BodyFat        float32 `json:"bodyFat"`        // 8.1
				BodyWater      float32 `json:"bodyWater"`      // 68.3
				BoneMass       float32 `json:"boneMass"`       // 3099
				MuscleMass     float32 `json:"muscleMass"`     // 55599
				PhysiqueRating int     `json:"physiqueRating"` // 4
				VisceralFat    int     `json:"visceralFat"`    // 9
				MetabolicAge   float64 `json:"metabolicAge"`   // 1104492410000
				SourceType     string  `json:"sourceType"`     // INDEX_SCALE
				TimestampGMT   int64   `json:"timestampGMT"`   // 1753533337000
				WeightDelta    float32 `json:"weightDelta"`
			} `json:"allWeightMetrics"`
		} `json:"dailyWeightSummaries"`
	}

	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	var weights []*core.Weight

	for _, day := range data.DailyWeightSummaries {
		for _, metric := range day.AllWeightMetrics {
			ts := metric.TimestampGMT
			if ts == 0 {
				ts = metric.Date
			}

			// important for delete function
			c.weightID[ts] = fmt.Sprintf("%s/byversion/%d", metric.CalendarDate, metric.SamplePk)

			w := &core.Weight{
				Date:   time.UnixMilli(ts),
				Weight: metric.Weight / 1000,

				BMI:       metric.BMI,
				BodyFat:   metric.BodyFat,
				BodyWater: metric.BodyWater,
				BoneMass:  metric.BoneMass / 1000,

				MetabolicAge:   int(metric.MetabolicAge / 31536e6),
				PhysiqueRating: metric.PhysiqueRating,
				VisceralFat:    metric.VisceralFat,

				SkeletalMuscleMass: metric.MuscleMass / 1000,

				Source: metric.SourceType,
			}
			weights = append(weights, w)
		}
	}
	return weights, nil
}

func (c *Client) DeleteWeight(weight *core.Weight) error {
	weightID, ok := c.weightID[weight.Date.UnixMilli()]
	if !ok {
		return errors.New("garmin: weight not exist")
	}

	res, err := c.Delete("weight-service/weight/" + weightID)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

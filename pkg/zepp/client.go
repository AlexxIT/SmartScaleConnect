package zepp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
)

type Client struct {
	client *http.Client

	appToken string // for auth
	userID   string // for some requests

	family map[string]int64
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{Timeout: time.Minute},
	}
}

func (c *Client) GetAllWeights() ([]*core.Weight, error) {
	return c.GetFilterWeights("")
}

func (c *Client) GetFilterWeights(name string) ([]*core.Weight, error) {
	familyID, err := c.GetFamilyID(name)
	if err != nil {
		return nil, err
	}

	var weights []*core.Weight

	for ts := time.Now().Unix(); ts > 0; {
		// 200 is maximum
		url := fmt.Sprintf(
			"https://api-mifit.zepp.com/users/%s/members/%d/weightRecords?limit=200&toTime=%d",
			c.userID, familyID, ts,
		)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("apptoken", c.appToken)

		res, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		var res1 struct {
			Items []struct {
				UserId        string `json:"userId"`
				MemberId      string `json:"memberId"`
				DeviceSource  int    `json:"deviceSource"`
				AppName       string `json:"appName"`
				GeneratedTime int64  `json:"generatedTime"`
				WeightType    int    `json:"weightType"`
				DeviceId      string `json:"deviceId"`
				Summary       struct {
					Weight           float32 `json:"weight"`
					Height           float32 `json:"height"`
					BMI              float32 `json:"bmi"`
					FatRate          float32 `json:"fatRate,omitempty"`
					BodyWaterRate    float32 `json:"bodyWaterRate,omitempty"`
					BoneMass         float32 `json:"boneMass,omitempty"`
					Metabolism       float32 `json:"metabolism,omitempty"` // basal metabolism 1358 kcal
					MuscleRate       float32 `json:"muscleRate,omitempty"`
					MuscleAge        int     `json:"muscleAge,omitempty"`
					ProteinRatio     float32 `json:"proteinRatio,omitempty"`
					StandBodyWeight  float32 `json:"standBodyWeight,omitempty"` // ideal body weight
					VisceralFat      float32 `json:"visceralFat,omitempty"`
					Impedance        int     `json:"impedance,omitempty"`
					EncryptImpedance string  `json:"encryptImpedance,omitempty"`
					BodyScore        int     `json:"bodyScore,omitempty"`
					BodyStyle        int     `json:"bodyStyle,omitempty"`
					DeviceType       int     `json:"deviceType"`
					Source           int     `json:"source"`
				} `json:"summary"`
				CreateTime int64 `json:"createTime"`
			} `json:"items"`
			Next int64 `json:"next"`
		}

		if err = json.NewDecoder(res.Body).Decode(&res1); err != nil {
			return nil, err
		}

		for _, item := range res1.Items {
			// don't know what it means, but WeightType=3 has broken weight values
			if item.WeightType != 0 {
				continue
			}

			w := &core.Weight{
				Date:      time.Unix(item.GeneratedTime, 0),
				Weight:    item.Summary.Weight,
				BMI:       item.Summary.BMI,
				BodyFat:   item.Summary.FatRate,
				BodyWater: item.Summary.BodyWaterRate,
				BoneMass:  item.Summary.BoneMass,

				MuscleMass:     item.Summary.MuscleRate, // don't know wny name is rate?!
				MetabolicAge:   item.Summary.MuscleAge,
				PhysiqueRating: item.Summary.BodyStyle,
				VisceralFat:    int(item.Summary.VisceralFat),

				BasalMetabolism: int(item.Summary.Metabolism),

				User:   name,
				Source: item.DeviceId,
			}
			weights = append(weights, w)
		}

		ts = res1.Next
	}

	return weights, nil
}

func (c *Client) GetFamilyID(name string) (int64, error) {
	if name == "" {
		return -1, nil
	}

	if c.family == nil {
		if err := c.GetFamilyMembers(); err != nil {
			return 0, err
		}
	}

	if fid, ok := c.family[name]; ok {
		return fid, nil
	}

	return 0, errors.New("zepp: can't find family member: " + name)
}

func (c *Client) GetFamilyMembers() error {
	req, err := http.NewRequest(
		"POST", "https://api-mifit.zepp.com/huami.health.scale.familymember.get.json",
		strings.NewReader("fuid=all&userid="+c.userID),
	)
	if err != nil {
		return err
	}

	req.Header.Add("apptoken", c.appToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("zepp: " + res.Status)
	}

	var res1 struct {
		//Code    int    `json:"code"`
		//Message string `json:"message"`
		Data struct {
			//Total int `json:"total"`
			List []struct {
				//Uid      string `json:"uid"`
				Fuid     int64  `json:"fuid"`
				Nickname string `json:"nickname"`
				//City          string  `json:"city"`
				//Brithday      string  `json:"brithday"`
				//Gender        int     `json:"gender"`
				//Height        int     `json:"height"`
				//Weight        float32 `json:"weight"`
				//Targetweight  float32 `json:"targetweight"`
				//LastModify    int     `json:"last_modify"`
				//ScaleAvatarId int     `json:"scale_avatar_id,omitempty"`
				//MeasureMode   int     `json:"measure_mode,omitempty"`
			} `json:"list"`
		} `json:"data"`
	}

	if err = json.NewDecoder(res.Body).Decode(&res1); err != nil {
		return err
	}

	c.family = make(map[string]int64)
	for _, item := range res1.Data.List {
		c.family[item.Nickname] = item.Fuid
	}

	return nil
}

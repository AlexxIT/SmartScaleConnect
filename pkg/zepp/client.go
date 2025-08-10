package zepp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
			Items []Record `json:"items"`
			Next  int64    `json:"next"`
		}

		if err = json.NewDecoder(res.Body).Decode(&res1); err != nil {
			return nil, err
		}

		for _, record := range res1.Items {
			// don't know what it means, but WeightType=3 has broken weight values
			if record.WeightType != 0 {
				continue
			}

			w := &core.Weight{
				Date:      time.Unix(record.GeneratedTime, 0),
				Weight:    record.Summary.Weight,
				BMI:       record.Summary.BMI,
				BodyFat:   record.Summary.FatRate,
				BodyWater: record.Summary.BodyWaterRate,
				BoneMass:  record.Summary.BoneMass,

				MuscleMass:     record.Summary.MuscleRate, // don't know wny name is rate?!
				MetabolicAge:   record.Summary.MuscleAge,
				PhysiqueRating: record.Summary.BodyStyle,
				ProteinMass:    record.Summary.Weight * record.Summary.ProteinRatio / 100,
				VisceralFat:    int(record.Summary.VisceralFat),

				BasalMetabolism: int(record.Summary.Metabolism),
				BodyScore:       record.Summary.BodyScore,
				Height:          record.Summary.Height,
				Impedance:       record.Summary.Impedance,

				User:   name,
				Source: record.DeviceId,
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
				Height int `json:"height"`
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

func (c *Client) AddWeights(weights []*core.Weight) error {
	if len(weights) == 0 {
		return nil
	}

	var records []*Record
	for _, weight := range weights {
		familyID, err := c.GetFamilyID(weight.User)
		if err != nil {
			return err
		}

		r := &Record{
			DataSource:    dataSource,
			DeviceId:      weight.Source,
			DeviceSource:  deviceSource,
			GeneratedTime: weight.Date.Unix(),
			MemberId:      strconv.FormatInt(familyID, 10),
			UserId:        c.userID,
			WeightType:    0,
			Summary: RecordSummary{
				Weight:        weight.Weight,
				Height:        weight.Height,
				BMI:           weight.BMI,
				FatRate:       weight.BodyFat,
				BodyWaterRate: weight.BodyWater,
				BoneMass:      weight.BoneMass,
				Metabolism:    float32(weight.BasalMetabolism),
				MuscleRate:    weight.MuscleMass,
				MuscleAge:     weight.MetabolicAge,
				ProteinRatio:  weight.ProteinMass / weight.Weight * 100,
				VisceralFat:   float32(weight.VisceralFat),
				BodyScore:     weight.BodyScore,
				BodyStyle:     weight.PhysiqueRating,
				DeviceType:    deviceType,
				Source:        source,

				//StandBodyWeight:  64.4,
				//Impedance:        482,
				//EncryptImpedance: "482",
			},
		}
		records = append(records, r)
	}

	body, err := json.Marshal(records)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST", "https://api-mifit.zepp.com/users/"+c.userID+"/members/-1/weightRecords", bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Add("apptoken", c.appToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("zepp: add weights error: " + res.Status)
	}

	return nil
}

func (c *Client) DeleteWeight(weight *core.Weight) error {
	familyID, err := c.GetFamilyID(weight.User)
	if err != nil {
		return err
	}

	data := fmt.Sprintf(`[{"ts":%d,"fuid":"%d"}]`, weight.Date.Unix(), familyID)

	form := url.Values{"dt": {"1"}, "jsondata": {data}, "userid": {c.userID}}
	req, err := http.NewRequest(
		"POST", "https://api-mifit.zepp.com/huami.health.scale.delete.json", strings.NewReader(form.Encode()),
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
		return errors.New("zepp: delete weight error" + res.Status)
	}

	return nil
}

func (c *Client) Equal(w1, w2 *core.Weight) bool {
	return equalFloat(w1.Weight, w2.Weight) &&
		equalFloat(w1.BMI, w2.BMI) &&
		equalFloat(w1.BodyFat, w2.BodyFat) &&
		equalFloat(w1.BodyWater, w2.BodyWater) &&
		equalFloat(w1.BoneMass, w2.BoneMass) &&
		equalFloat(w1.MuscleMass, w2.MuscleMass) &&
		w1.MetabolicAge == w2.MetabolicAge &&
		w1.PhysiqueRating == w2.PhysiqueRating &&
		w1.VisceralFat == w2.VisceralFat &&
		w1.BasalMetabolism == w2.BasalMetabolism &&
		w1.BodyScore == w2.BodyScore &&
		equalFloat(w1.Height, w2.Height) &&
		w1.Impedance == w2.Impedance
}

const (
	dataSource   = 1   // Weight
	deviceSource = 102 // ???
	deviceType   = 1   // Weight
	source       = 1   // ???
)

type Record struct {
	DataSource    int    `json:"dataSource"`    // 1 = WEIGHT
	DeviceId      string `json:"deviceId"`      // uppercase mac without colons
	DeviceSource  int    `json:"deviceSource"`  // 102
	GeneratedTime int64  `json:"generatedTime"` // unix
	MemberId      string `json:"memberId"`      // -1 for main user
	UserId        string `json:"userId"`
	WeightType    int    `json:"weightType"` // 0 - add, 1 - change or delete?

	Summary RecordSummary `json:"summary"`
}

type RecordSummary struct {
	Weight           float32 `json:"weight"`
	Height           float32 `json:"height,omitempty"`
	BMI              float32 `json:"bmi,omitempty"`
	FatRate          float32 `json:"fatRate,omitempty"`
	BodyWaterRate    float32 `json:"bodyWaterRate,omitempty"`
	BoneMass         float32 `json:"boneMass,omitempty"`
	Metabolism       float32 `json:"metabolism,omitempty"`
	MuscleRate       float32 `json:"muscleRate,omitempty"`
	MuscleAge        int     `json:"muscleAge,omitempty"`
	ProteinRatio     float32 `json:"proteinRatio,omitempty"`
	StandBodyWeight  float32 `json:"standBodyWeight,omitempty"`
	VisceralFat      float32 `json:"visceralFat,omitempty"`
	Impedance        int     `json:"impedance,omitempty"`
	EncryptImpedance string  `json:"encryptImpedance,omitempty"`
	BodyScore        int     `json:"bodyScore,omitempty"`
	BodyStyle        int     `json:"bodyStyle,omitempty"`
	DeviceType       int     `json:"deviceType"` // 1 - WEIGHT
	Source           int     `json:"source"`     // 1 - without Age, BodyBalanceScore, OneFootMeasureTime, SyncHealth

	//Age                int `json:"age,omitempty"`
	//BodyBalanceScore   int     `json:"bodyBalanceScore"`
	//OneFootMeasureTime float32 `json:"oneFootMeasureTime"`
	//SyncHealth         int     `json:"syncHealth"` // 1 - ???
}

func equalFloat(f1, f2 float32) bool {
	if f1 == f2 {
		return true
	}
	if f1 > f2 {
		return f1-f2 < 0.1
	}
	return f2-f1 < 0.1
}

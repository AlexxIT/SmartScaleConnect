package picooc

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
)

type Client struct {
	client *http.Client

	deviceID string
	userID   string

	roleIDs map[string]string
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
	roleID, ok := c.roleIDs[name]
	if !ok {
		return nil, errors.New("picooc: unknown user: " + name)
	}

	var weights []*core.Weight

	params := c.values("bodyIndexList")
	//params.Set("orderType", "-1")
	params.Set("pageSize", "1000")
	params.Set("time", params.Get("timestamp"))
	params.Set("userId", c.userID)
	params.Set("roleId", roleID)

	for {
		res, err := c.client.Get(api + "bodyIndex/bodyIndexList?" + params.Encode())
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		var res1 struct {
			//Code   int    `json:"code"`
			//Msg    string `json:"msg"`
			//Method string `json:"method"`
			//Result struct {
			//	Code    int    `json:"code"`
			//	Message string `json:"message"`
			//} `json:"result"`
			Resp struct {
				Records []struct {
					BodyTime         int64   `json:"bodyTime"`
					DataType         int     `json:"dataType"`
					BodyIndexId      int     `json:"body_index_id"`
					RoleId           int     `json:"role_id"`
					BodyFat          float32 `json:"body_fat"`
					Weight           float32 `json:"weight"`
					BMI              float32 `json:"bmi"`
					VisceralFatLevel int     `json:"visceral_fat_level"`
					MuscleRace       float32 `json:"muscle_race"`
					BodyAge          int     `json:"body_age"`
					BoneMass         float32 `json:"bone_mass"`
					BasicMetabolism  int     `json:"basic_metabolism"`
					WaterRace        float32 `json:"water_race"`
					SkeletalMuscle   float32 `json:"skeletal_muscle"`
					LocalTime        int     `json:"local_time"`
					SubcutaneousFat  float32 `json:"subcutaneous_fat"`
					ServerTime       int     `json:"server_time"`
					ServerId         int     `json:"server_id"`
					IsDel            int     `json:"is_del"`
					Abnormal         struct {
						Weight       int `json:"weight"`
						Time         int `json:"time"`
						AbnormalFlag int `json:"abnormal_flag"`
						BodyFat      int `json:"body_fat"`
					} `json:"abnormal"`
					AbnormalFlag          int      `json:"abnormal_flag"`
					ElectricResistance    int      `json:"electric_resistance"`
					IsManuallyAdd         int      `json:"is_manually_add"`
					IsFirstDay            int      `json:"is_first_day"`
					LandmarkIcons         []string `json:"landmarkIcons"`
					MAC                   string   `json:"mac"`
					AnchorWeight          int      `json:"anchor_weight"`
					AnchorBata            int      `json:"anchor_bata"`
					CorrectionValueR      int      `json:"correction_value_r"`
					BodyFatReferenceValue float32  `json:"body_fat_reference_value"`
					LabelMarker           int      `json:"label_marker"`
					DataSources           int      `json:"data_sources"`
				} `json:"records"`
				LastTime int  `json:"lastTime"`
				Continue bool `json:"continue"`
			} `json:"resp"`
		}

		if err = json.NewDecoder(res.Body).Decode(&res1); err != nil {
			return nil, err
		}

		for _, v1 := range res1.Resp.Records {
			if v1.AbnormalFlag != 0 || v1.IsDel != 0 {
				continue
			}

			w := &core.Weight{
				Date:   time.Unix(v1.BodyTime, 0),
				Weight: v1.Weight,

				BMI:       v1.BMI,
				BodyFat:   v1.BodyFat,
				BodyWater: v1.WaterRace,
				BoneMass:  v1.BoneMass,

				MetabolicAge: v1.BodyAge, // 0
				VisceralFat:  v1.VisceralFatLevel,

				BasalMetabolism:    v1.BasicMetabolism,
				SkeletalMuscleMass: v1.SkeletalMuscle, // 0

				User:   name,
				Source: v1.MAC,
			}
			weights = append(weights, w)
		}

		if !res1.Resp.Continue {
			break
		}

		params.Set("time", strconv.Itoa(res1.Resp.LastTime))
	}

	return weights, nil
}

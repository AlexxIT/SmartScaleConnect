package xiaomi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
)

type Client struct {
	client *http.Client

	sid       string // for login
	cookies   string // for auth
	userID    int64  // for some requests
	ssecurity []byte // for encryption
	passToken string
}

func NewClient(app string) *Client {
	return &Client{
		client: &http.Client{Timeout: time.Minute},
		sid:    app,
	}
}

func (c *Client) GetAllWeights() ([]*core.Weight, error) {
	return c.getAllWeights("")
}

func (c *Client) getAllWeights(region string) ([]*core.Weight, error) {
	var weights []*core.Weight

	ts := time.Now().Add(24 * time.Hour).Unix()
	params := fmt.Sprintf(`{"start_time":1,"end_time":%d,"key":"weight"}`, ts)

	for {
		// this request depends on user region
		data, err := c.Request(baseURL(region), "/app/v1/data/get_fitness_data_by_time", params)
		if err != nil {
			return nil, err
		}

		var res1 struct {
			DataList []struct {
				Sid        string `json:"sid"`
				Key        string `json:"key"`
				Time       int    `json:"time"`
				Value      string `json:"value"`
				ZoneOffset int    `json:"zone_offset"`
				UpdateTime int    `json:"update_time"`
				ZoneName   string `json:"zone_name,omitempty"`
			} `json:"data_list"`
			HasMore bool   `json:"has_more"`
			NextKey string `json:"next_key"`
		}

		if err = json.Unmarshal(data, &res1); err != nil {
			return nil, err
		}

		for _, v1 := range res1.DataList {
			if v1.Key != "weight" {
				continue
			}

			var res2 struct {
				BasalMetabolism  int     `json:"basal_metabolism"`   // S400, Eight
				BMI              float32 `json:"bmi"`                // S400, Eight
				BodyAge          int     `json:"body_age"`           // S400, Eight
				BodyFatRate      float32 `json:"body_fat_rate"`      // S400, Eight
				BodyMoistureMass float32 `json:"body_moisture_mass"` // S400, Eight
				BodyScore        int     `json:"body_score"`         // S400, Eight
				BoneMass         float32 `json:"bone_mass"`          // S400, Eight
				BoneRate         float32 `json:"bone_rate"`          // S400, Eight
				BPM              int     `json:"bpm"`                // S400, Eight
				FatControl       float32 `json:"fat_control"`        // S400, Eight
				MoistureRate     float32 `json:"moisture_rate"`      // S400, Eight
				MuscleControl    float32 `json:"muscle_control"`     // S400, Eight
				MuscleMass       float32 `json:"muscle_mass"`        // S400, Eight
				MuscleRate       float32 `json:"muscle_rate"`        // S400, Eight
				ProteinMass      float32 `json:"protein_mass"`       // S400, Eight
				ProteinRate      float32 `json:"protein_rate"`       // S400, Eight
				Somatotype       int     `json:"somatotype"`         // S400, Eight
				StandardWeight   int     `json:"standard_weight"`    // S400, Eight
				StandardWeightV2 float32 `json:"standard_weight_v2"` // S400, Eight
				Time             int     `json:"time"`               // S400, Eight
				VisceralFat      float32 `json:"visceral_fat"`       // S400, Eight
				Weight           float32 `json:"weight"`             // S400, Eight
				WeightControl    float32 `json:"weight_control"`     // S400, Eight
				Whr              float32 `json:"whr"`                // S400, Eight

				//FatFreeBody        float32 `json:"fat_free_body"`        // S400
				//ScoreStandardType  int     `json:"score_standard_type"`  // S400
				SkeletalMuscleMass float32 `json:"skeletal_muscle_mass"` // S400

				//BodyShape                 int     `json:"body_shape"`                   // Eight
				//FatMass                   float32 `json:"fat_mass"`                     // Eight
				//LeftLowerLimbFatMass      float32 `json:"left_lower_limb_fat_mass"`     // Eight
				//LeftLowerLimbFatRank      int     `json:"left_lower_limb_fat_rank"`     // Eight
				//LeftLowerLimbMuscleMass   int     `json:"left_lower_limb_muscle_mass"`  // Eight
				//LeftLowerLimbMuscleRank   int     `json:"left_lower_limb_muscle_rank"`  // Eight
				//LeftUpperLimbFatMass      float32 `json:"left_upper_limb_fat_mass"`     // Eight
				//LeftUpperLimbFatRank      int     `json:"left_upper_limb_fat_rank"`     // Eight
				//LeftUpperLimbMuscleMass   float32 `json:"left_upper_limb_muscle_mass"`  // Eight
				//LeftUpperLimbMuscleRank   int     `json:"left_upper_limb_muscle_rank"`  // Eight
				//LimbsFatBalance           int     `json:"limbs_fat_balance"`            // Eight
				//LimbsMuscleBalance        int     `json:"limbs_muscle_balance"`         // Eight
				//LimbsSkeletalMuscleIndex  float32 `json:"limbs_skeletal_muscle_index"`  // Eight
				//LowerLimbFatBalance       int     `json:"lower_limb_fat_balance"`       // Eight
				//LowerLimbMuscleBalance    int     `json:"lower_limb_muscle_balance"`    // Eight
				//RecommendedCaloriesIntake int     `json:"recommended_calories_intake"`  // Eight
				//RightLowerLimbFatMass     float32 `json:"right_lower_limb_fat_mass"`    // Eight
				//RightLowerLimbFatRank     int     `json:"right_lower_limb_fat_rank"`    // Eight
				//RightLowerLimbMuscleMass  float32 `json:"right_lower_limb_muscle_mass"` // Eight
				//RightLowerLimbMuscleRank  int     `json:"right_lower_limb_muscle_rank"` // Eight
				//RightUpperLimbFatMass     float32 `json:"right_upper_limb_fat_mass"`    // Eight
				//RightUpperLimbFatRank     int     `json:"right_upper_limb_fat_rank"`    // Eight
				//RightUpperLimbMuscleMass  float32 `json:"right_upper_limb_muscle_mass"` // Eight
				//RightUpperLimbMuscleRank  int     `json:"right_upper_limb_muscle_rank"` // Eight
				//TrunkFatMass              float32 `json:"trunk_fat_mass"`               // Eight
				//TrunkFatRank              int     `json:"trunk_fat_rank"`               // Eight
				//TrunkMuscleMass           float32 `json:"trunk_muscle_mass"`            // Eight
				//TrunkMuscleRank           int     `json:"trunk_muscle_rank"`            // Eight
				//UpperLimbFatBalance       int     `json:"upper_limb_fat_balance"`       // Eight
				//UpperLimbMuscleBalance    int     `json:"upper_limb_muscle_balance"`    // Eight
			}

			if err = json.Unmarshal([]byte(v1.Value), &res2); err != nil {
				return nil, err
			}

			w := &core.Weight{
				Date:      time.Unix(int64(res2.Time), 0), // 1732550224
				Weight:    res2.Weight,                    // 69.8
				BMI:       res2.BMI,                       // 23.6
				BodyFat:   res2.BodyFatRate,               // 19.6
				BodyWater: res2.MoistureRate,              // 51
				BoneMass:  res2.BoneMass,                  // 2.8

				MetabolicAge: res2.BodyAge,    // 36
				MuscleMass:   res2.MuscleMass, // 53.3
				ProteinMass:  res2.ProteinMass,
				VisceralFat:  int(res2.VisceralFat),

				BasalMetabolism:    res2.BasalMetabolism,
				BodyScore:          res2.BodyScore,
				HeartRate:          res2.BPM,
				SkeletalMuscleMass: res2.SkeletalMuscleMass,

				Source: v1.Sid, // blt.3.xxx
			}

			weights = append(weights, w)
		}

		if !res1.HasMore {
			break
		}

		params = fmt.Sprintf(`{"start_time":1,"end_time":%d,"key":"weight","next_key":%q}`, ts, res1.NextKey)
	}

	return weights, nil
}

func (c *Client) GetFamilyMembers() (map[int64]string, error) {
	params := `{"eco_api":"eco/scale/account/list"}`
	data, err := c.Request(baseURL(""), "/app/v1/eco/api_proxy", params)
	if err != nil {
		return nil, err
	}

	var res1 struct {
		Code          int    `json:"code"`
		Message       string `json:"message"`
		DetailMessage string `json:"detailMessage"`
		Result        []struct {
			Uid              string `json:"uid"`
			AccountId        string `json:"accountId"`
			Name             string `json:"name"`
			Icon             string `json:"icon"`
			Type             int    `json:"type"`
			Sex              string `json:"sex"`
			Height           string `json:"height"`
			WeightTarget     string `json:"weightTarget"`
			Birth            string `json:"birth"`
			CreationTime     int64  `json:"creationTime"`
			AccountCode      int    `json:"accountCode"`
			DeviceId         string `json:"deviceId"`
			WeightUpdateTime int64  `json:"weightUpdateTime"`
		} `json:"result"`
	}

	if err = unmarshalProxyResp(data, &res1); err != nil {
		return nil, err
	}

	accounts := make(map[int64]string)

	for _, v := range res1.Result {
		i, _ := strconv.ParseInt(v.AccountId, 10, 64)
		accounts[i] = v.Name
	}

	return accounts, nil
}

// GetFilterWeights filter can be region or scale model
func (c *Client) GetFilterWeights(filter string) ([]*core.Weight, error) {
	// check if the filter is a region
	if s := baseURL(filter); s != "" {
		return c.getAllWeights(filter)
	}

	var weights []*core.Weight

	ts := time.Now().UnixMilli()

	for {
		// model is important, did may be zero
		params := fmt.Sprintf(
			`{"param":{"endTime":1,"beginTime":%d},"model":"%s","did":0,"uid":%d}`,
			ts, filter, c.userID,
		)
		params = fmt.Sprintf(`{"eco_api":"eco/scale/getData","params":%q}`, params)
		// this request works only for main (CN) region
		data, err := c.Request(baseURL(""), "/app/v1/eco/api_proxy", params)
		if err != nil {
			return nil, err
		}

		var res1 struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Result  []struct {
				Model       string `json:"model"`
				Uid         int64  `json:"uid"`
				AccountId   int64  `json:"accountId"`
				Did         string `json:"did"`
				CreateTime  int64  `json:"createTime"`
				Data        string `json:"data"`
				DataVersion int    `json:"dataVersion"`
				Sn          string `json:"sn"`
				FromSource  int    `json:"fromSource"`
			} `json:"result"`
		}

		if err = unmarshalProxyResp(data, &res1); err != nil {
			return nil, err
		}

		for _, v1 := range res1.Result {
			switch v1.FromSource {
			case 1:
				var v2 struct {
					Weight    float32 `json:"weight"` // 87.8 kg
					BMI       float32 `json:"bmi"`    // 25.7 points
					BodyFat   float32 `json:"bfp"`    // 22.9 %
					BodyWater float32 `json:"bwp"`    // 58.8 $
					BoneMass  float32 `json:"bmc"`    // 3.7 kg

					MetabolicAge int     `json:"ma"`  // 55 years
					MuscleMass   float32 `json:"slm"` // 63.9 kg
					BodyType     int     `json:"bt"`  // 4
					ProteinMass  float32 `json:"pm"`  // 11.6 kg
					VisceralFat  int     `json:"vfl"` // 9 points

					BMR                int     `json:"bmr"`        // 1832 kcal
					BodyScore          int     `json:"sbc"`        // 80 points
					HeartRate          int     `json:"heartRate"`  // 73 bpm
					SkeletalMuscleMass float32 `json:"smm"`        // 37.6 kg
					ReportFrom         string  `json:"reportFrom"` // Regular

					//UserID             int     `json:"miid"`       // 1234567890
					//Duid               int     `json:"duid"`       // 6 ?
					//UserType           int     `json:"userType"`   // 1 ?
					//Status             int     `json:"status"`     // 0 ?
					//Time               int64   `json:"time"`       // 1755927448
					//ProteinPercent     float32 `json:"pp"`         // 13.2 %
					//IdealWeight        float32 `json:"swt"`        // 73.5 kg
					//MuscleCorrection   float32 `json:"mc"`         // -5.2
					//WeightCorrection   float32 `json:"wc"`         // -14.3
					//FatCorrection      float32 `json:"fc"`         // -9.1
					//WHR                float32 `json:"whr"`        // 1.3
					//MusclePercent      float32 `json:"slp"`        // 72.9 %
					//BoneMassPercentage float32 `json:"bmcp"`       // 4.2 %
					//FatMass            float32 `json:"bfm"`        // 20.1 kg
					//LeanBodyMass       float32 `json:"ffm"`        // 67.6 kg
					//BodyWaterMass      float32 `json:"bwm"`        // 51.5 kg
					//BodyRes            float32 `json:"bodyRes"`    // 384.1
					//BodyRes2           float32 `json:"bodyRes2"`   // 357.5
					//Idx                int     `json:"idx"`        // -1
					User struct {
						//Uid              string `json:"uid"`
						//Sex              string `json:"sex"`
						//Birth            int64  `json:"birth"`
						//AccountId        string `json:"accountId"`
						//Icon             string `json:"icon"`
						//Height           string `json:"height"`
						Name string `json:"name"`
						//Type             int    `json:"type"`
						//AccountCode      int    `json:"accountCode"`
						//CreationTime     int64  `json:"creationTime"`
						//WeightTarget     string `json:"weightTarget"`
						//WeightUpdateTime int64  `json:"weightUpdateTime"`
						//UpdateTime       int64  `json:"updateTime"`
					} `json:"user"`
				}

				if err = json.Unmarshal([]byte(v1.Data), &v2); err != nil {
					return nil, err
				}

				// v2.Time has bugs. For "UserEditor" it has seconds, for Claimed it has milliseconds
				w := &core.Weight{
					Date:      time.UnixMilli(v1.CreateTime),
					Weight:    v2.Weight,
					BMI:       v2.BMI,
					BodyFat:   v2.BodyFat,
					BodyWater: v2.BodyWater,
					BoneMass:  v2.BoneMass,

					MetabolicAge:   v2.MetabolicAge,
					MuscleMass:     v2.MuscleMass,
					PhysiqueRating: v2.BodyType,
					ProteinMass:    v2.ProteinMass,
					VisceralFat:    v2.VisceralFat,

					BasalMetabolism:    v2.BMR,
					BodyScore:          v2.BodyScore,
					HeartRate:          v2.HeartRate,
					Height:             0,
					SkeletalMuscleMass: v2.SkeletalMuscleMass,

					User:   v2.User.Name,
					Source: v2.ReportFrom,
				}
				weights = append(weights, w)
			case 3:
				var v2 struct {
					BMI         string `json:"bmi"`
					BodyRes     string `json:"bodyRes"`
					BodyRes2    string `json:"bodyRes2"`
					BodyResData string `json:"bodyResData"`
					HeartRate   int    `json:"heartRate"`
					Mid         string `json:"mid"`
					Time        string `json:"time"`
					User        struct {
						//AccountId        int64   `json:"accountId"`
						//Birth            int64   `json:"birth"`
						//CreateTime       int64   `json:"createTime"`
						//Height           int     `json:"height"`
						//Icon             string  `json:"icon"`
						//Id               int     `json:"id"`
						Name string `json:"name"`
						//Sex              int     `json:"sex"`
						//Type             int     `json:"type"`
						//UserId           int64   `json:"userId"`
						//WeightTarget     float32 `json:"weightTarget"`
						//WeightUpdateTime int     `json:"weightUpdateTime"`
					} `json:"user"`
					Weight string `json:"weight"`
				}

				if err = json.Unmarshal([]byte(v1.Data), &v2); err != nil {
					return nil, err
				}

				w := &core.Weight{
					Date:      time.UnixMilli(parseInt64(v2.Time)),
					Weight:    parseFloat(v2.Weight),
					BMI:       parseFloat(v2.BMI),
					HeartRate: v2.HeartRate,
					User:      v2.User.Name,
					Source:    v1.Did,
				}

				if v2.BodyResData != "" {
					var v3 struct {
						BodyFatRate        string `json:"bfp"`  // 12.1
						MuscleMass         string `json:"slm"`  // 32.2
						MoistureRate       string `json:"bwp"`  // 52.1
						BoneMass           string `json:"bmc"`  // 1.6
						VisceralFat        string `json:"vfl"`  // 5
						ProteinRate        string `json:"pp"`   // 31
						SkeletalMuscleMass string `json:"smm"`  // 15.39
						BMI                string `json:"bmi"`  // 19.1
						StandardWeightV2   string `json:"swt"`  // 46.2
						MuscleControl      string `json:"mc"`   // 3.5
						WeightControl      string `json:"wc"`   // 5.3
						FatControl         string `json:"fc"`   // 2
						WHR                string `json:"whr"`  // 1
						Wl                 string `json:"wl"`   // 68.2
						Hl                 string `json:"hl"`   // 70
						BasalMetabolic     string `json:"bmr"`  // 1143
						Bt                 string `json:"bt"`   // 1
						BodyAge            string `json:"ma"`   // 14
						BodyScore          string `json:"sbc"`  // 86
						MuscleRate         string `json:"slp"`  // 84
						BoneRate           string `json:"bmcp"` // 3.9
						FatMass            string `json:"bfm"`  // 4.9
						FatFreeBody        string `json:"ffm"`  // 35.8
						BodyMoistureMass   string `json:"bwm"`  // 21.2
						ProteinMass        string `json:"pm"`   // 12.6
						Smi                string `json:"smi"`  // 7.2
					}

					if err = json.Unmarshal([]byte(v2.BodyResData), &v3); err != nil {
						return nil, err
					}

					w.BodyFat = parseFloat(v3.BodyFatRate)
					w.BodyWater = parseFloat(v3.MoistureRate)
					w.BoneMass = parseFloat(v3.BoneMass)

					w.MetabolicAge = parseInt(v3.BodyAge)
					w.MuscleMass = parseFloat(v3.MuscleMass)
					w.ProteinMass = parseFloat(v3.ProteinMass)
					w.VisceralFat = parseInt(v3.VisceralFat)

					w.BasalMetabolism = parseInt(v3.BasalMetabolic)
					w.BodyScore = parseInt(v3.BodyScore)
					w.SkeletalMuscleMass = parseFloat(v3.SkeletalMuscleMass)
				}

				weights = append(weights, w)
			}
		}

		if len(res1.Result) < 20 {
			break
		}

		ts = res1.Result[19].CreateTime
	}

	return weights, nil
}

func baseURL(region string) string {
	switch region {
	case "", "cn":
		return "https://hlth.io.mi.com"
	case "de", "i2", "ru", "sg", "us":
		return "https://" + region + ".hlth.io.mi.com"
	}
	return ""
}

func unmarshalProxyResp(data []byte, v any) error {
	var res1 struct {
		Resp string `json:"resp"`
	}

	if err := json.Unmarshal(data, &res1); err != nil {
		return err
	}

	return json.Unmarshal([]byte(res1.Resp), v)
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func parseInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func parseFloat(s string) float32 {
	f, _ := strconv.ParseFloat(s, 64)
	return float32(f)
}

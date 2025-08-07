package fitbit

import (
	"archive/zip"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
)

func Read(path string) ([]*core.Weight, error) {
	zipFile, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	var weights []*core.Weight

	for _, file := range zipFile.File {
		if !strings.Contains(file.Name, "/Personal & Account/weight") {
			continue
		}

		var rc io.ReadCloser
		if rc, err = file.Open(); err != nil {
			return nil, err
		}
		defer rc.Close()

		var res1 []struct {
			LogId  int64   `json:"logId"`  // 1659948335000
			Weight float32 `json:"weight"` // 165.8
			Bmi    float32 `json:"bmi"`    // 25.13
			Fat    float32 `json:"fat"`    // 18.76300048828125
			Date   string  `json:"date"`   // "08/08/22"
			Time   string  `json:"time"`   // "08:45:35"
			Source string  `json:"source"` // "Aria"
		}
		if err = json.NewDecoder(rc).Decode(&res1); err != nil {
			return nil, err
		}

		for _, v1 := range res1 {
			w := &core.Weight{
				Date:    time.UnixMilli(v1.LogId), // UTC as source
				Weight:  v1.Weight * LBS2KG,
				BMI:     v1.Bmi,
				BodyFat: v1.Fat,
				Source:  v1.Source,
			}
			weights = append(weights, w)
		}
	}

	return weights, nil
}

const LBS2KG = 0.45359237

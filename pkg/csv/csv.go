package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
)

const Header = "Date,Weight," +
	"BMI,BodyFat,BodyWater,BoneMass," +
	"MetabolicAge,MuscleMass,PhysiqueRating,VisceralFat," +
	"BasalMetabolism,HeartRate,SkeletalMuscleMass," +
	"User,Source\n"

func Read(r io.Reader) ([]*core.Weight, error) {
	cr := csv.NewReader(r)
	header, err := cr.Read()
	if err != nil {
		return nil, err
	}

	var weights []*core.Weight

	for {
		record, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		var w core.Weight

		for i, s := range header {
			switch s {
			case "Date":
				w.Date = parseDate(record[i])
			case "Weight":
				w.Weight = parseFloat(record[i])
			case "BMI":
				w.BMI = parseFloat(record[i])
			case "BodyFat":
				w.BodyFat = parseFloat(record[i])
			case "BodyWater":
				w.BodyWater = parseFloat(record[i])
			case "BoneMass":
				w.BoneMass = parseFloat(record[i])
			case "MetabolicAge":
				w.MetabolicAge = parseInt(record[i])
			case "MuscleMass":
				w.MuscleMass = parseFloat(record[i])
			case "PhysiqueRating":
				w.PhysiqueRating = parseInt(record[i])
			case "ProteinMass":
				w.ProteinMass = parseFloat(record[i])
			case "VisceralFat":
				w.VisceralFat = parseInt(record[i])
			case "BasalMetabolism":
				w.BasalMetabolism = parseInt(record[i])
			case "HeartRate":
				w.HeartRate = parseInt(record[i])
			case "SkeletalMuscleMass":
				w.SkeletalMuscleMass = parseFloat(record[i])
			case "User":
				w.User = record[i]
			case "Source":
				w.Source = record[i]
			}
		}

		weights = append(weights, &w)
	}

	return weights, nil
}

func parseDate(s string) time.Time {
	t, _ := time.ParseInLocation(time.DateTime, s, time.Local) // local time!!!
	return t
}

func parseFloat(s string) float32 {
	f, _ := strconv.ParseFloat(s, 32)
	return float32(f)
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func Write(w io.Writer, weights []*core.Weight) error {
	if _, err := w.Write([]byte(Header)); err != nil {
		return err
	}
	for _, weight := range weights {
		if _, err := w.Write(Marshal(weight)); err != nil {
			return err
		}
	}
	return nil
}

func Marshal(weight *core.Weight) []byte {
	b := make([]byte, 0, 128)

	b = appendDate(b, weight.Date)
	b = appendFloat(b, weight.Weight)

	b = appendFloat(b, weight.BMI)
	b = appendFloat(b, weight.BodyFat)
	b = appendFloat(b, weight.BodyWater)
	b = appendFloat(b, weight.BoneMass)

	b = appendInt(b, weight.MetabolicAge)
	b = appendFloat(b, weight.MuscleMass)
	b = appendInt(b, weight.PhysiqueRating)
	b = appendFloat(b, weight.ProteinMass)
	b = appendInt(b, weight.VisceralFat)

	b = appendInt(b, weight.BasalMetabolism)
	b = appendInt(b, weight.HeartRate)
	b = appendFloat(b, weight.SkeletalMuscleMass)

	b = appendString(b, weight.User)
	b = appendString(b, weight.Source)

	return append(b, '\n')
}

func appendDate(b []byte, v time.Time) []byte {
	return v.AppendFormat(b, time.DateTime) // local time!!!
}

func appendFloat(b []byte, v float32) []byte {
	if v == 0 {
		return append(b, ',')
	}
	return fmt.Appendf(b, ",%.2f", v)
}

func appendInt(b []byte, v int) []byte {
	if v == 0 {
		return append(b, ',')
	}
	return fmt.Appendf(b, ",%d", v)
}

func appendString(b []byte, v string) []byte {
	b = append(b, ',')
	return append(b, v...)
}

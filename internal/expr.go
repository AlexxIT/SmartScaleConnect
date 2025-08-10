package internal

import (
	"fmt"
	"reflect"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func Expr(config map[string]string, weights []*core.Weight) error {
	programs := map[string]*vm.Program{}

	for key, input := range config {
		var opt expr.Option

		switch key {
		case "Date":
			opt = expr.AsAny()
		case "Weight", "BMI", "BodyFat", "BodyWater", "BoneMass", "MuscleMass", "ProteinMass", "Height", "SkeletalMuscleMass":
			opt = expr.AsFloat64()
		case "MetabolicAge", "PhysiqueRating", "VisceralFat", "BasalMetabolism", "BodyScore", "HeartRate":
			opt = expr.AsInt()
		case "User", "Source":
			opt = expr.AsKind(reflect.String)
		}

		program, err := expr.Compile(input, opt)
		if err != nil {
			return err
		}

		programs[key] = program
	}

	for _, weight := range weights {
		for key, program := range programs {
			v, err := expr.Run(program, weight)
			if err != nil {
				return err
			}

			switch key {
			case "Date":
				date, ok := v.(time.Time)
				if !ok {
					return fmt.Errorf("invalid date: %v", v)
				}
				weight.Date = date
			case "Weight":
				weight.Weight = float32(v.(float64))
			case "BMI":
				weight.BMI = float32(v.(float64))
			case "BodyFat":
				weight.BodyFat = float32(v.(float64))
			case "BodyWater":
				weight.BodyWater = float32(v.(float64))
			case "BoneMass":
				weight.BoneMass = float32(v.(float64))
			case "MetabolicAge":
				weight.MetabolicAge = v.(int)
			case "MuscleMass":
				weight.MuscleMass = float32(v.(float64))
			case "PhysiqueRating":
				weight.PhysiqueRating = v.(int)
			case "ProteinMass":
				weight.ProteinMass = float32(v.(float64))
			case "VisceralFat":
				weight.VisceralFat = v.(int)
			case "BasalMetabolism":
				weight.BasalMetabolism = v.(int)
			case "BodyScore":
				weight.BodyScore = v.(int)
			case "HeartRate":
				weight.HeartRate = v.(int)
			case "Height":
				weight.Height = float32(v.(float64))
			case "SkeletalMuscleMass":
				weight.SkeletalMuscleMass = float32(v.(float64))
			case "User":
				weight.User = v.(string)
			case "Source":
				weight.Source = v.(string)
			}
		}
	}

	return nil
}

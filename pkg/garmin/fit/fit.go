package fit

import (
	"io"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
	"github.com/muktihari/fit/encoder"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
)

func WriteWeight(w io.Writer, weights ...*core.Weight) error {
	file := filedef.NewWeight()
	file.FileId.Type = typedef.FileWeight
	file.FileId.Manufacturer = typedef.ManufacturerGarmin
	file.FileId.Product = 2429      // scale
	file.FileId.SerialNumber = 1234 // any

	for _, weight := range weights {
		scale := mesgdef.NewWeightScale(nil)
		scale.Timestamp = weight.Date
		scale.Weight = typedef.Weight(weight.Weight * 100)

		if weight.BMI != 0 {
			scale.Bmi = uint16(weight.BMI * 10)
		}
		if weight.BodyFat != 0 {
			scale.PercentFat = uint16(weight.BodyFat * 100)
		}
		if weight.BodyWater != 0 {
			scale.PercentHydration = uint16(weight.BodyWater * 100)
		}
		if weight.BoneMass != 0 {
			scale.BoneMass = uint16(weight.BoneMass * 100)
		}

		if weight.MetabolicAge != 0 {
			scale.MetabolicAge = uint8(weight.MetabolicAge)
		}
		if weight.SkeletalMuscleMass != 0 {
			scale.MuscleMass = uint16(weight.SkeletalMuscleMass * 100)
		}
		if weight.PhysiqueRating != 0 {
			scale.PhysiqueRating = uint8(weight.PhysiqueRating)
		}
		if weight.VisceralFat != 0 {
			scale.VisceralFatRating = uint8(weight.VisceralFat)
		}

		if weight.BasalMetabolism != 0 {
			scale.BasalMet = uint16(weight.BasalMetabolism * 4)
		}

		//scale.ActiveMet = 0
		//scale.VisceralFatMass = 0

		file.WeightScales = append(file.WeightScales, scale)
	}

	// Convert back to FIT protocol messages
	fit := file.ToFIT(nil)
	return encoder.New(w).Encode(&fit)
}

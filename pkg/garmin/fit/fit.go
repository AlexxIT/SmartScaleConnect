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

		scale.Bmi = uint16(weight.BMI * 10)
		scale.PercentFat = uint16(weight.BodyFat * 100)
		scale.PercentHydration = uint16(weight.BodyWater * 100)
		scale.BoneMass = uint16(weight.BoneMass * 100)

		scale.MetabolicAge = uint8(weight.MetabolicAge)
		scale.MuscleMass = uint16(weight.SkeletalMuscleMass * 100)
		scale.PhysiqueRating = uint8(weight.PhysiqueRating)
		scale.VisceralFatRating = uint8(weight.VisceralFat)

		scale.BasalMet = uint16(weight.BasalMetabolism * 4)

		//scale.ActiveMet = 0
		//scale.VisceralFatMass = 0

		file.WeightScales = append(file.WeightScales, scale)
	}

	// Convert back to FIT protocol messages
	fit := file.ToFIT(nil)
	return encoder.New(w).Encode(&fit)
}

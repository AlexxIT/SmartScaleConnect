package garmin

import (
	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
)

func Equal(w1, w2 *core.Weight) bool {
	return equalFloat(w1.Weight, w2.Weight) &&
		equalFloat(w1.BMI, w2.BMI) &&
		equalFloat(w1.BodyFat, w2.BodyFat) &&
		equalFloat(w1.BodyWater, w2.BodyWater) &&
		equalFloat(w1.BoneMass, w2.BoneMass) &&
		//w1.MetabolicAge == w2.MetabolicAge &&
		w1.PhysiqueRating != w2.PhysiqueRating &&
		w1.VisceralFat != w2.VisceralFat &&
		equalFloat(w1.SkeletalMuscleMass, w2.SkeletalMuscleMass)
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
